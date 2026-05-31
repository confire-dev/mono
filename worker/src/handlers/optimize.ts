import type { Env, OptimizeRequest, OptimizeResponse } from '../types.js'
import { handle } from '../engine.js'
import { authenticate } from '../lib/auth.js'
import { cacheKey, cacheGet, cachePut } from '../lib/cache.js'
import { trackEvent } from '../lib/analytics.js'
import { MAX_RAW_PAYLOAD_BYTES, getUsageThisPeriod, recordOptimization } from '../lib/supabase.js'
import { getPlan } from '../lib/plans.js'
import { canUseRemoteOptimizer, entitlementMessage } from '../lib/entitlement.js'

export async function handleOptimize(request: Request, env: Env): Promise<Response> {
  // ── 1. Parse + validate body ──────────────────────────────────────────────
  let body: OptimizeRequest
  try {
    body = await request.json() as OptimizeRequest
  } catch {
    return Response.json({ error: 'invalid JSON' }, { status: 400 })
  }
  if (!body.event) {
    return Response.json({ error: 'missing event' }, { status: 400 })
  }

  // ── 2. Absolute payload cap (before auth, before plan lookup) ─────────────
  const rawOutputSize    = JSON.stringify(body.event.tool?.output ?? '').length
  const rawTokenEstimate = Math.round(rawOutputSize / 4)
  if (rawOutputSize > MAX_RAW_PAYLOAD_BYTES) {
    return Response.json({
      result: { kind: 'passthrough' },
      _warning: `Payload too large (${Math.round(rawOutputSize / 1024)}KB). Max absolute limit is ${Math.round(MAX_RAW_PAYLOAD_BYTES / 1024 / 1024)}MB.`,
    } satisfies OptimizeResponse)
  }

  // ── 3. Authenticate ───────────────────────────────────────────────────────
  const auth = await authenticate(request, env)
  if (!auth.ok) {
    return Response.json({ error: auth.error }, { status: auth.status })
  }
  const { user } = auth

  // ── 4. Rate limit check (fast — Cloudflare edge, no DB) ─────────────────
  const rateLimited = await checkRateLimit(env, user.plan, user.id)
  if (rateLimited) {
    return Response.json(
      { error: 'rate_limit_exceeded', message: 'Too many requests. See confire.dev/docs/rate-limits' },
      { status: 429, headers: { 'Retry-After': '60' } },
    )
  }

  // ── 5. Resolve plan + current usage ──────────────────────────────────────
  const plan = await getPlan(user.plan, env)

  const cfg = env.SUPABASE_URL && env.SUPABASE_SERVICE_KEY
    ? { url: env.SUPABASE_URL, serviceKey: env.SUPABASE_SERVICE_KEY }
    : null

  const usageThisPeriod = cfg
    ? await getUsageThisPeriod(cfg, user.id, plan.id)
    : { id: '', cloudOptimizationsUsed: 0, cloudTokensUsed: 0, localOptimizationsCount: 0, savedTokens: 0 }

  // ── 6. Entitlement check (plan capabilities + limits) ─────────────────────
  const optimizer = body.event.tool?.name ?? 'unknown'
  const check = canUseRemoteOptimizer({
    plan,
    optimizer,
    payloadBytes:       rawOutputSize,
    rawTokensEstimate:  rawTokenEstimate,
    usageThisPeriod: {
      cloudOptimizationsUsed: usageThisPeriod.cloudOptimizationsUsed,
      cloudTokensUsed:        usageThisPeriod.cloudTokensUsed,
    },
  })

  if (!check.allowed) {
    return Response.json({
      result: {
        kind: 'add-context',
        context: entitlementMessage(check, plan.name),
      },
    } satisfies OptimizeResponse)
  }

  // ── 6. Cache lookup (content-hash) ────────────────────────────────────────
  const key = await cacheKey(body.event.tool?.output ?? body.event)
  if (env.CACHE) {
    const cached = await cacheGet(env.CACHE, key)
    if (cached) {
      return Response.json({ result: cached } satisfies OptimizeResponse)
    }
  }

  // ── 7. Optimize ───────────────────────────────────────────────────────────
  const result = handle(body.event)

  // ── 8. Cache the result ───────────────────────────────────────────────────
  if (env.CACHE) {
    cachePut(env.CACHE, key, result).catch(() => {})
  }

  // ── 9. Record usage + analytics (async — don't block response) ───────────
  if (result.kind !== 'passthrough' && cfg) {
    recordOptimization(cfg, {
      userId:      user.id,
      planId:      plan.id,
      sessionId:   body.event.session?.id ?? '',
      toolType:    body.event.tool?.mcpServer ?? body.event.tool?.name ?? 'unknown',
      integration: body.event.host ?? 'claude_code',
      optimizer:   result.stats?.optimizer ?? 'unknown',
      rawBytes:    result.stats?.beforeBytes ?? rawOutputSize,
      optimizedBytes: result.stats?.afterBytes ?? rawOutputSize,
      wasCached:   false,
      analyticsConsented: false,
    }).catch(() => {})

    trackEvent(env.AE, env.AMPLITUDE_KEY, {
      userId:      user.id,
      email:       user.email,
      eventType:   'tool_optimized',
      toolName:    body.event.tool?.name,
      optimizer:   result.stats?.optimizer,
      beforeBytes: result.stats?.beforeBytes,
      afterBytes:  result.stats?.afterBytes,
      sessionId:   body.event.session?.id,
      host:        body.event.host,
    })
  }

  // ── 10. Approaching-limit nudge ───────────────────────────────────────────
  const responseResult = { ...result } as typeof result & { _warning?: string }
  if (check.approachingLimit && check.approachingLimitMessage) {
    responseResult._warning = `⚠️ ${check.approachingLimitMessage} · confire.dev/upgrade`
  }

  return Response.json({ result: responseResult } satisfies OptimizeResponse)
}

// Select the rate limiter binding for the user's plan tier and check the limit.
// Returns true (blocked) if the user has exceeded their per-minute request rate.
// If no binding is configured (local dev), allows through.
function checkRateLimit(env: Env, planId: string, userId: string): Promise<boolean> {
  let limiter: RateLimit | undefined
  if (planId === 'free') {
    limiter = env.RL_FREE
  } else if (planId === 'dev' || planId === 'dev_annual') {
    limiter = env.RL_DEV
  } else {
    // pro, pro_annual, enterprise
    limiter = env.RL_PRO
  }
  if (!limiter) return Promise.resolve(false)
  return limiter.limit({ key: userId }).then(r => !r.success)
}

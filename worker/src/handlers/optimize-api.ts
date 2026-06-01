// POST /v1/optimize — standalone optimizer API.
//
// Accepts any text/JSON and returns an optimized version, with token reduction stats.
// Callers use this before sending context to Groq, OpenAI, Gemini, or any LLM —
// paying us once to pay the LLM less on every call.
//
// Uses the same auth, rate limiting, entitlement, caching, and usage tracking
// as the Claude Code hook path. The only difference: a synthetic InterceptEvent
// is constructed from the caller's content + type hint, then passed to handle().

import type {
  Env, InterceptEvent, OptimizerApiRequest, OptimizerApiResponse, OptimizerType,
} from '../types.js'
import { handle } from '../engine.js'
import { authenticate } from '../lib/auth.js'
import { cacheKey, cacheGet, cachePut } from '../lib/cache.js'
import { trackEvent } from '../lib/analytics.js'
import { MAX_RAW_PAYLOAD_BYTES, getUsageThisPeriod, getCreditBalance, recordOptimization } from '../lib/supabase.js'
import { getPlan } from '../lib/plans.js'
import { canUseRemoteOptimizer, entitlementMessage } from '../lib/entitlement.js'

// Maps caller-supplied type hints to tool names the optimizer dispatch understands.
const TYPE_TO_TOOL: Record<OptimizerType, string> = {
  auto:       'generic',
  json:       'generic',
  html:       'WebFetch',
  search:     'websearch',
  bash:       'Bash',
  github:     'github_pr',
  slack:      'slack',
  figma:      'figma',
  jira:       'jira',
  notion:     'notion',
  confluence: 'confluence',
  clickup:    'clickup',
  amplitude:  'amplitude',
  zapier:     'zapier',
  playwright: 'playwright',
}

const MAX_TOTAL_REQUEST_BYTES = 100 * 1024 * 1024

export async function handleOptimizerApi(request: Request, env: Env): Promise<Response> {
  // ── 0. Total body size guard ──────────────────────────────────────────────
  const contentLength = parseInt(request.headers.get('Content-Length') ?? '', 10)
  if (!isNaN(contentLength) && contentLength > MAX_TOTAL_REQUEST_BYTES) {
    return Response.json(
      { error: 'payload_too_large', message: `Request too large (${Math.round(contentLength / 1024 / 1024)}MB).` },
      { status: 413 },
    )
  }

  // ── 1. Parse + validate ───────────────────────────────────────────────────
  let body: OptimizerApiRequest
  try {
    body = await request.json() as OptimizerApiRequest
  } catch {
    return Response.json({ error: 'invalid JSON' }, { status: 400 })
  }
  if (typeof body.content !== 'string' || !body.content) {
    return Response.json({ error: 'missing or empty content' }, { status: 400 })
  }
  if (body.content.length > MAX_RAW_PAYLOAD_BYTES) {
    return Response.json(
      { error: 'payload_too_large', message: `content exceeds ${Math.round(MAX_RAW_PAYLOAD_BYTES / 1024 / 1024)}MB limit.` },
      { status: 413 },
    )
  }

  // ── 2. Authenticate ───────────────────────────────────────────────────────
  const auth = await authenticate(request, env)
  if (!auth.ok) {
    return Response.json({ error: auth.error }, { status: auth.status })
  }
  const { user } = auth

  // ── 3. Rate limit ─────────────────────────────────────────────────────────
  const rateLimited = await checkRateLimit(env, user.plan, user.id)
  if (rateLimited) {
    return Response.json(
      { error: 'rate_limit_exceeded', message: 'Too many requests. See confire.dev/docs/rate-limits' },
      { status: 429, headers: { 'Retry-After': '60' } },
    )
  }

  // ── 4. Resolve plan + usage ───────────────────────────────────────────────
  const plan = await getPlan(user.plan, env)
  const cfg = env.SUPABASE_URL && env.SUPABASE_SERVICE_KEY
    ? { url: env.SUPABASE_URL, serviceKey: env.SUPABASE_SERVICE_KEY }
    : null
  const usageThisPeriod = cfg
    ? await getUsageThisPeriod(cfg, user.id, plan.id)
    : { id: '', cloudOptimizationsUsed: 0, cloudTokensUsed: 0, localOptimizationsCount: 0, savedTokens: 0 }

  const creditBalance = cfg
    ? await getCreditBalance(cfg, user.id)
    : null
  const purchasedCredits = creditBalance?.purchased_credits ?? 0

  // ── 5. Entitlement check ──────────────────────────────────────────────────
  const toolName = TYPE_TO_TOOL[body.type ?? 'auto'] ?? 'generic'
  const rawBytes = new TextEncoder().encode(body.content).length
  const check = canUseRemoteOptimizer({
    plan,
    optimizer: toolName,
    payloadBytes: rawBytes,
    rawTokensEstimate: Math.round(rawBytes / 4),
    usageThisPeriod: {
      cloudOptimizationsUsed: usageThisPeriod.cloudOptimizationsUsed,
      cloudTokensUsed:        usageThisPeriod.cloudTokensUsed,
    },
    purchasedCreditsRemaining: purchasedCredits,
  })
  if (!check.allowed) {
    return Response.json(
      { error: 'entitlement_denied', message: entitlementMessage(check, plan.name) },
      { status: 402 },
    )
  }

  // ── 6. Cache lookup ───────────────────────────────────────────────────────
  const key = await cacheKey(body.content)
  if (env.CACHE) {
    const cached = await cacheGet(env.CACHE, key)
    if (cached && cached.kind === 'replace-output' && typeof cached.toolOutput === 'string') {
      const cachedText = cached.toolOutput
      return Response.json({
        result:        cachedText,
        optimized:     true,
        input_chars:   body.content.length,
        output_chars:  cachedText.length,
        reduction_pct: Math.max(0, Math.round((1 - cachedText.length / body.content.length) * 100)),
        optimizer:     cached.stats?.optimizer ?? 'generic',
        cached:        true,
      } satisfies OptimizerApiResponse)
    }
  }

  // ── 7. Optimize ───────────────────────────────────────────────────────────
  const syntheticEvent: InterceptEvent = {
    host:     'optimizer-api',
    strategy: 'api',
    phase:    'tool.post',
    session:  { id: 'api' },
    tool:     { name: toolName, output: body.content, input: {}, isMcp: false },
  }
  const result = handle(syntheticEvent)

  // ── 8. Cache result ───────────────────────────────────────────────────────
  if (env.CACHE && result.kind !== 'passthrough') {
    cachePut(env.CACHE, key, result).catch(() => {})
  }

  // ── 9. Record usage (async) ───────────────────────────────────────────────
  if (result.kind !== 'passthrough' && cfg) {
    recordOptimization(cfg, {
      userId:         user.id,
      planId:         plan.id,
      sessionId:      'api',
      toolType:       toolName,
      integration:    'optimizer-api',
      optimizer:      result.stats?.optimizer ?? 'generic',
      rawBytes:       result.stats?.beforeBytes ?? rawBytes,
      optimizedBytes: result.stats?.afterBytes  ?? rawBytes,
      wasCached:      false,
      analyticsConsented: false,
      ...(check.usePurchasedCredit ? { usePurchasedCredit: true as const } : {}),
    }).catch(() => {})

    trackEvent(env.AE, env.AMPLITUDE_KEY, {
      userId:      user.id,
      email:       user.email,
      eventType:   'api_optimized',
      toolName,
      optimizer:   result.stats?.optimizer,
      beforeBytes: result.stats?.beforeBytes,
      afterBytes:  result.stats?.afterBytes,
      host:        'optimizer-api',
    })
  }

  // ── 10. Build response ────────────────────────────────────────────────────
  const optimizedText = result.kind === 'replace-output' && typeof result.toolOutput === 'string'
    ? result.toolOutput
    : body.content

  const inputChars  = body.content.length
  const outputChars = optimizedText.length

  const response: OptimizerApiResponse = {
    result:        optimizedText,
    optimized:     result.kind === 'replace-output',
    input_chars:   inputChars,
    output_chars:  outputChars,
    reduction_pct: Math.max(0, Math.round((1 - outputChars / inputChars) * 100)),
    optimizer:     result.stats?.optimizer ?? 'generic',
    cached:        false,
  }

  const res = Response.json(response)
  if (check.approachingLimit && check.approachingLimitMessage) {
    res.headers.set('X-Confire-Warning', check.approachingLimitMessage)
  }
  return res
}

function checkRateLimit(env: Env, planId: string, userId: string): Promise<boolean> {
  let limiter: RateLimit | undefined
  if (planId === 'free') limiter = env.RL_FREE
  else if (planId === 'dev' || planId === 'dev_annual') limiter = env.RL_DEV
  else limiter = env.RL_PRO
  if (!limiter) return Promise.resolve(false)
  return limiter.limit({ key: userId }).then(r => !r.success)
}

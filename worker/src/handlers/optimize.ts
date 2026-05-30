import type { Env, OptimizeRequest, OptimizeResponse } from '../types.js'
import { handle } from '../engine.js'
import { authenticate } from '../lib/auth.js'
import { cacheKey, cacheGet, cachePut } from '../lib/cache.js'
import { trackEvent } from '../lib/analytics.js'
// Usage recording is now done via /v1/events telemetry endpoint.
// The /optimize handler remains lean — auth + cache + optimize only.

export async function handleOptimize(request: Request, env: Env): Promise<Response> {
  // ── 1. Parse body ─────────────────────────────────────────────
  let body: OptimizeRequest
  try {
    body = await request.json() as OptimizeRequest
  } catch {
    return Response.json({ error: 'invalid JSON' }, { status: 400 })
  }
  if (!body.event) {
    return Response.json({ error: 'missing event' }, { status: 400 })
  }

  // ── 2. Authenticate + check limits ────────────────────────────
  const auth = await authenticate(request, env)
  if (!auth.ok) {
    return Response.json({ error: auth.error }, { status: auth.status })
  }
  const { user, usage } = auth

  if (!usage.ok) {
    // At limit: return passthrough (never block) + notification
    const result: OptimizeResponse = {
      result: {
        kind: 'add-context',
        context: `⚠️ Confire free tier reached (${usage.used}/${usage.limit} requests). Using local optimizer. Upgrade at confire.dev/upgrade`,
      },
    }
    return Response.json(result)
  }

  // ── 3. Cache lookup ───────────────────────────────────────────
  const key = await cacheKey(body.event.tool?.output ?? body.event)
  if (env.CACHE) {
    const cached = await cacheGet(env.CACHE, key)
    if (cached) {
      return Response.json({ result: cached } satisfies OptimizeResponse)
    }
  }

  // ── 4. Optimize ───────────────────────────────────────────────
  const result = handle(body.event)

  // ── 5. Cache the result ───────────────────────────────────────
  if (env.CACHE) {
    cachePut(env.CACHE, key, result).catch(() => {})
  }

  // ── 6. Track usage + analytics (async, don't block response) ──
  const bytesSaved = result.stats ? result.stats.beforeBytes - result.stats.afterBytes : 0
  if (result.kind !== 'passthrough') {
    const cfg = env.SUPABASE_URL && env.SUPABASE_SERVICE_KEY
      ? { url: env.SUPABASE_URL, serviceKey: env.SUPABASE_SERVICE_KEY }
      : null
    // Usage accounting is written via the /v1/events telemetry endpoint
  // which the daemon calls asynchronously after the hook returns.

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

  // ── 7. Warn approaching limit ─────────────────────────────────
  const responseResult = { ...result } as typeof result & { _warning?: string }
  if (usage.ok && usage.used >= usage.limit * 0.8) {
    responseResult._warning = `⚠️ ${usage.used}/${usage.limit} requests used · confire.dev/upgrade`
  }

  return Response.json({ result: responseResult } satisfies OptimizeResponse)
}

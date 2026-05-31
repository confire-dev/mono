// POST /v1/events — the Worker's trusted ingestion boundary.
//
// The CLI sends raw metrics here. The Worker:
//   1. Validates the API key (authenticates the caller)
//   2. Writes exact values to Supabase (billing truth / user dashboard)
//   3. Optionally forwards bucketed, sanitized values to Amplitude (analytics)
//
// The CLI is NOT trusted to self-report billing state.
// "analytics_consented: false" suppresses Amplitude but NOT Supabase writes.
// Supabase writes are required for the product to work (billing, dashboard, limits).

import type { Env } from '../types.js'
import { authenticate } from '../lib/auth.js'
import { recordOptimization, upsertCliSession, closeCliSession, writeAudit } from '../lib/supabase.js'
import { trackEvent } from '../lib/analytics.js'
import { byteBucket, reductionBucket, durationBucket, sanitizeToolType } from '../lib/buckets.js'
import { cacheKey, cacheGet, cachePut } from '../lib/cache.js'

// ── Wire format from the CLI daemon ─────────────────────────────────────────

interface TelemetryEvent {
  event_id:            string    // UUID — idempotency key
  event_type:          EventType
  cli_version?:        string
  integration?:        string    // 'claude_code', 'cursor', …
  session_id?:         string    // Claude Code session_id
  tool_type?:          string    // 'figma', 'bash', 'github_pr', …
  optimizer?:          string    // which optimizer handled it
  raw_bytes?:          number
  optimized_bytes?:    number
  duration_ms?:        number
  was_cached?:         boolean
  analytics_consented: boolean   // false = skip Amplitude only; Supabase still writes
}

type EventType =
  | 'session_start'
  | 'session_end'
  | 'tool_call_optimized'
  | 'tool_call_passthrough'
  | 'hook_installed'
  | 'daemon_started'

interface SessionEndPayload {
  total_tool_calls:  number
  optimized_calls:   number
  raw_bytes:         number
  optimized_bytes:   number
}

export async function handleTelemetry(request: Request, env: Env): Promise<Response> {
  // ── 1. Auth ──────────────────────────────────────────────────────────────
  const auth = await authenticate(request, env)
  if (!auth.ok) {
    return Response.json({ error: auth.error }, { status: auth.status })
  }
  const { user } = auth

  // ── 2. Parse body ────────────────────────────────────────────────────────
  let event: TelemetryEvent
  try {
    event = await request.json() as TelemetryEvent
  } catch {
    return Response.json({ error: 'invalid JSON' }, { status: 400 })
  }

  // ── 3. Idempotency check via KV cache ────────────────────────────────────
  // Prevent the daemon from double-counting on retry.
  if (event.event_id && env.CACHE) {
    const idempKey = `evt:${event.event_id}`
    const seen = await env.CACHE.get(idempKey)
    if (seen) return Response.json({ ok: true, deduplicated: true })
    env.CACHE.put(idempKey, '1', { expirationTtl: 604_800 }).catch(() => {}) // 7 days — covers retry window
  }

  const cfg = env.SUPABASE_URL && env.SUPABASE_SERVICE_KEY
    ? { url: env.SUPABASE_URL, serviceKey: env.SUPABASE_SERVICE_KEY }
    : null

  // ── 4. Route by event type ───────────────────────────────────────────────
  switch (event.event_type) {

    case 'session_start':
      if (cfg && event.session_id) {
        const deviceId   = request.headers.get('X-Confire-Device')
        const cliVersion = event.cli_version
        const integration = event.integration
        await upsertCliSession(cfg, event.session_id, user.id, {
          ...(deviceId   ? { deviceId }   : {}),
          ...(cliVersion ? { cliVersion } : {}),
          ...(integration ? { integration } : {}),
        })
      }
      maybeTrack(env, event, user.email, user.plan, 'cli_session_started')
      break

    case 'session_end': {
      const endData = event as TelemetryEvent & SessionEndPayload
      if (cfg && event.session_id) {
        await closeCliSession(cfg, event.session_id, {
          totalCalls:     (endData as unknown as Record<string,number>)['total_tool_calls'] ?? 0,
          optimizedCalls: (endData as unknown as Record<string,number>)['optimized_calls'] ?? 0,
          rawBytes:       (endData as unknown as Record<string,number>)['raw_bytes'] ?? 0,
          optimizedBytes: (endData as unknown as Record<string,number>)['optimized_bytes'] ?? 0,
        })
      }
      maybeTrack(env, event, user.email, user.plan, 'cli_session_ended')
      break
    }

    case 'tool_call_optimized':
      if (cfg && event.raw_bytes != null && event.optimized_bytes != null) {
        await recordOptimization(cfg, {
          userId:             user.id,
          planId:             user.plan ?? 'free',
          sessionId:          event.session_id ?? '',
          toolType:           event.tool_type ?? 'unknown',
          integration:        event.integration ?? 'claude_code',
          optimizer:          event.optimizer ?? 'unknown',
          rawBytes:           event.raw_bytes ?? 0,
          optimizedBytes:     event.optimized_bytes ?? 0,
          ...(event.duration_ms != null ? { durationMs: event.duration_ms } : {}),
          wasCached:          event.was_cached ?? false,
          analyticsConsented: event.analytics_consented,
        })
      }
      // Amplitude only if user consented to analytics
      if (event.analytics_consented) {
        maybeTrack(env, event, user.email, user.plan, 'tool_call_optimized')
      }
      break

    case 'hook_installed':
      if (cfg) await writeAudit(cfg, user.id, 'hook_installed', { cli_version: event.cli_version })
      maybeTrack(env, event, user.email, user.plan, 'hook_installed')
      break

    case 'daemon_started':
      maybeTrack(env, event, user.email, user.plan, 'daemon_started')
      break

    default:
      // Unknown event type — accept and discard (forward compat)
  }

  return Response.json({ ok: true })
}

// ── Amplitude forwarding with bucketed, sanitized values ─────────────────

function maybeTrack(
  env: Env,
  event: TelemetryEvent,
  email: string,
  plan: string,
  amplitudeEventType: string,
): void {
  if (!event.analytics_consented) return

  const rawBytes = event.raw_bytes ?? 0
  const optBytes = event.optimized_bytes ?? 0
  const ratio    = rawBytes > 0 ? (rawBytes - optBytes) / rawBytes : 0

  trackEvent(env.AE, env.AMPLITUDE_KEY, {
    userId:   email,    // Amplitude identifies by email (no PII in properties)
    email,
    eventType: amplitudeEventType,
    toolName:  event.tool_type ? sanitizeToolType(event.tool_type) : undefined,
    optimizer: event.optimizer,
    // Bucketed — never raw values in Amplitude
    beforeBytes: rawBytes > 0 ? parseInt(byteBucket(rawBytes)) || rawBytes : undefined,
    afterBytes:  optBytes > 0 ? parseInt(byteBucket(optBytes)) || optBytes : undefined,
    sessionId:   undefined,  // never send session_id to Amplitude
    host:        event.integration ?? 'claude_code',
  })
}

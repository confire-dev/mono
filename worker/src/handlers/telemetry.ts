// POST /v1/events — the Worker's trusted ingestion boundary.
//
// The CLI sends security and session events here. The Worker:
//   1. Validates the API key (authenticates the caller)
//   2. Writes to Supabase (dashboard, security audit trail)
//   3. Optionally forwards sanitized values to Amplitude (analytics)

import type { Env } from '../types.js'
import { authenticate } from '../lib/auth.js'
import { recordSecurityEvent, recordProvenanceEvent, upsertCliSession, closeCliSession, writeAudit } from '../lib/supabase.js'
import { trackEvent } from '../lib/analytics.js'
import { sanitizeToolType } from '../lib/buckets.js'
import { cacheKey, cacheGet, cachePut } from '../lib/cache.js'
import { getPlan } from '../lib/plans.js'

// ── Wire format from the CLI daemon ─────────────────────────────────────────

interface TelemetryEvent {
  event_id:            string    // UUID — idempotency key
  event_type:          EventType
  cli_version?:        string
  integration?:        string    // 'claude_code', 'cursor', …
  session_id?:         string
  tool_type?:          string
  // Security event fields
  risk_level?:         string
  action_taken?:       string
  pattern_matched?:    string
  sanitized?:          boolean
  secrets_redacted?:   number
  // Provenance event fields
  trust_level?:        string
  flags?:              string[]
  mcp_server?:         string
  origin_domain?:      string
  analytics_consented: boolean
  // Session end fields
  total_tool_calls?:   number
  blocked_calls?:      number
  reviewed_calls?:     number
  warned_calls?:       number
  sanitized_calls?:    number
  secrets_total?:      number
}

type EventType =
  | 'session_start'
  | 'session_end'
  | 'security_event'
  | 'provenance_event'
  | 'hook_installed'
  | 'daemon_started'
  | 'budget_exhaustion'

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
  if (event.event_id && env.CACHE) {
    const idempKey = `evt:${event.event_id}`
    const seen = await env.CACHE.get(idempKey)
    if (seen) return Response.json({ ok: true, deduplicated: true })
    env.CACHE.put(idempKey, '1', { expirationTtl: 604_800 }).catch(() => {})
  }

  const cfg = env.SUPABASE_URL && env.SUPABASE_SERVICE_KEY
    ? { url: env.SUPABASE_URL, serviceKey: env.SUPABASE_SERVICE_KEY }
    : null

  // ── 4. Route by event type ───────────────────────────────────────────────
  switch (event.event_type) {

    case 'session_start':
      if (cfg && event.session_id) {
        const deviceId    = request.headers.get('X-Confire-Device')
        const cliVersion  = event.cli_version
        const integration = event.integration
        await upsertCliSession(cfg, event.session_id, user.id, {
          ...(deviceId    ? { deviceId }    : {}),
          ...(cliVersion  ? { cliVersion }  : {}),
          ...(integration ? { integration } : {}),
        })
      }
      maybeTrack(env, event, user.email, user.plan, 'cli_session_started')
      break

    case 'session_end':
      if (cfg && event.session_id) {
        await closeCliSession(cfg, event.session_id, {
          totalCalls: event.total_tool_calls ?? 0,
        })
      }
      maybeTrack(env, event, user.email, user.plan, 'cli_session_ended')
      break

    case 'security_event': {
      const plan = await getPlan(user.plan_id, env)
      if (cfg && plan.features.securityEventHistory && event.tool_type && event.risk_level && event.action_taken) {
        await recordSecurityEvent(cfg, {
          userId:         user.id,
          sessionId:      event.session_id ?? '',
          toolName:       event.tool_type,
          eventType:      event.risk_level === 'HIGH' ? 'PROMPT_INJECTION' : 'DESTRUCTIVE_CMD',
          riskLevel:      event.risk_level,
          actionTaken:    event.action_taken,
          ...(event.pattern_matched ? { patternMatched: event.pattern_matched } : {}),
          bypassed:       false,
        })
      }
      if (event.analytics_consented) {
        maybeTrack(env, event, user.email, user.plan, 'security_event')
      }
      break
    }

    case 'provenance_event': {
      const plan = await getPlan(user.plan_id, env)
      if (cfg && plan.features.provenanceTracking && event.tool_type && event.trust_level) {
        await recordProvenanceEvent(cfg, {
          userId:         user.id,
          sessionId:      event.session_id ?? '',
          toolName:       event.tool_type,
          trustLevel:     event.trust_level,
          sanitized:      event.sanitized ?? false,
          redactionCount: event.secrets_redacted ?? 0,
          flags:          event.flags ?? [],
          ...(event.mcp_server    ? { mcpServer:    event.mcp_server }    : {}),
          ...(event.origin_domain ? { originDomain: event.origin_domain } : {}),
        })
      }
      break
    }

    case 'hook_installed':
      if (cfg) await writeAudit(cfg, user.id, 'hook_installed', { cli_version: event.cli_version })
      maybeTrack(env, event, user.email, user.plan, 'hook_installed')
      break

    case 'daemon_started':
      maybeTrack(env, event, user.email, user.plan, 'daemon_started')
      break

    case 'budget_exhaustion':
      // Ask-budget exhausted for a session — all subsequent warns now surface as review.
      // No DB write in v1; analytics forwarding only.
      if (event.analytics_consented) {
        maybeTrack(env, event, user.email, user.plan, 'budget_exhaustion')
      }
      break

    default:
      // Unknown event type — accept and discard (forward compat)
  }

  return Response.json({ ok: true })
}

// ── Amplitude forwarding ──────────────────────────────────────────────────

function maybeTrack(
  env: Env,
  event: TelemetryEvent,
  email: string,
  plan: string,
  amplitudeEventType: string,
): void {
  if (!event.analytics_consented) return

  trackEvent(env.AE, env.AMPLITUDE_KEY, {
    userId:    email,
    email,
    eventType: amplitudeEventType,
    toolName:  event.tool_type ? sanitizeToolType(event.tool_type) : undefined,
    sessionId: undefined,
    host:      event.integration ?? 'claude_code',
  })
}

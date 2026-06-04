import type { Env } from './types.js'
import { handleOptimize }        from './handlers/optimize.js'
import { handleOptimizerApi }    from './handlers/optimize-api.js'
import { handleSessionStart }    from './handlers/session.js'
import { handleGenerateKey, handleMe, handleRevokeKey, handleRevokeSelf } from './handlers/auth.js'
import { handleCreateCheckout }       from './handlers/checkout.js'
import { handleCreateTopup }          from './handlers/topup.js'
import { handleCancelSubscription }   from './handlers/cancel.js'
import { handleTelemetry }            from './handlers/telemetry.js'
import { handleStripeWebhook }   from './handlers/stripe.js'
import { syncPlansToKV }         from './lib/plans.js'
import { handleSupabaseWebhook } from './handlers/db-webhook.js'
import { handleGetPolicy, handlePatchPolicyGroups } from './handlers/policy.js'

const ALLOWED_ORIGINS = new Set([
  'https://confire.dev',
  'https://dev.confire.dev',
])

function corsHeaders(origin: string | null): Record<string, string> {
  const allowed = origin && ALLOWED_ORIGINS.has(origin) ? origin : 'https://confire.dev'
  return {
    'Access-Control-Allow-Origin':  allowed,
    'Access-Control-Allow-Methods': 'POST, GET, PATCH, OPTIONS',
    'Access-Control-Allow-Headers': 'Content-Type, Authorization, X-Confire-Sig, X-Confire-Timestamp, X-Confire-Device',
    'Vary': 'Origin',
  }
}

function withCors(res: Response, origin: string | null): Response {
  const h = new Headers(res.headers)
  for (const [k, v] of Object.entries(corsHeaders(origin))) h.set(k, v)
  return new Response(res.body, { status: res.status, statusText: res.statusText, headers: h })
}

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    const url    = new URL(request.url)
    const method = request.method

    const origin = request.headers.get('Origin')

    if (method === 'OPTIONS') {
      return new Response(null, { headers: corsHeaders(origin) })
    }

    const res = await route(request, url, method, env)
    return withCors(res, origin)
  },
}

async function route(request: Request, url: URL, method: string, env: Env): Promise<Response> {

    // ── Optimizer (core product) ─────────────────────────────────────────
    if (method === 'POST' && url.pathname === '/optimize') {
      return handleOptimize(request, env)
    }
    // ── Standalone Optimizer API (pre-LLM context reduction) ─────────────
    // Gated by OPTIMIZER_API_ENABLED env var — returns 404 until launched.
    // To enable: set OPTIMIZER_API_ENABLED="true" in wrangler.toml or via
    //   wrangler secret put OPTIMIZER_API_ENABLED
    if (method === 'POST' && url.pathname === '/v1/optimize') {
      if (env.OPTIMIZER_API_ENABLED !== 'true') {
        return new Response('Not Found', { status: 404 })
      }
      return handleOptimizerApi(request, env)
    }
    if (method === 'POST' && url.pathname === '/session/start') {
      return handleSessionStart(request, env)
    }

    // ── Telemetry ingestion (CLI → Supabase truth + Amplitude analytics) ─
    if (method === 'POST' && url.pathname === '/v1/events') {
      return handleTelemetry(request, env)
    }

    // ── Auth & account ────────────────────────────────────────────────────
    if (method === 'POST' && url.pathname === '/api/keys/generate') {
      return handleGenerateKey(request, env)
    }
    if (method === 'GET' && url.pathname === '/api/me') {
      return handleMe(request, env)
    }
    if (method === 'POST' && url.pathname === '/api/keys/revoke') {
      return handleRevokeKey(request, env)
    }
    if (method === 'POST' && url.pathname === '/api/keys/revoke-self') {
      return handleRevokeSelf(request, env)
    }

    // ── Policy / firewall group overrides ────────────────────────────────
    if (method === 'GET' && url.pathname === '/v1/policy') {
      return handleGetPolicy(request, env)
    }
    if (method === 'PATCH' && url.pathname === '/v1/policy/groups') {
      return handlePatchPolicyGroups(request, env)
    }

    // ── Billing / checkout ────────────────────────────────────────────────
    if (method === 'POST' && url.pathname === '/api/checkout/create') {
      return handleCreateCheckout(request, env)
    }
    if (method === 'POST' && url.pathname === '/api/topup/create') {
      return handleCreateTopup(request, env)
    }
    if (method === 'POST' && url.pathname === '/api/subscription/cancel') {
      return handleCancelSubscription(request, env)
    }

    // ── Stripe webhooks ───────────────────────────────────────────────────
    if (method === 'POST' && url.pathname === '/webhooks/stripe') {
      return handleStripeWebhook(request, env)
    }

    // ── Supabase DB webhook (plans table changes → re-sync KV) ───────────
    // Fired automatically by Supabase whenever a row in `plans` is changed.
    // No manual sync needed — the DB pushes changes to the Worker.
    if (method === 'POST' && url.pathname === '/webhooks/supabase') {
      return handleSupabaseWebhook(request, env)
    }

    // ── Admin: manual plan sync — disabled (use Supabase webhook instead) ──
    // if (method === 'POST' && url.pathname === '/admin/plans/sync') { ... }

    // ── Health ────────────────────────────────────────────────────────────
    if (method === 'GET' && url.pathname === '/health') {
      return Response.json({ ok: true, version: '0.1.0', env: env.ENVIRONMENT })
    }

    return new Response('Not Found', { status: 404 })
}

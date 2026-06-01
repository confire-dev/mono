import type { Env } from './types.js'
import { handleOptimize }        from './handlers/optimize.js'
import { handleOptimizerApi }    from './handlers/optimize-api.js'
import { handleSessionStart }    from './handlers/session.js'
import { handleGenerateKey, handleMe } from './handlers/auth.js'
import { handleCreateCheckout }       from './handlers/checkout.js'
import { handleCreateTopup }          from './handlers/topup.js'
import { handleTelemetry }            from './handlers/telemetry.js'
import { handleStripeWebhook }   from './handlers/stripe.js'
import { syncPlansToKV }         from './lib/plans.js'
import { handleSupabaseWebhook } from './handlers/db-webhook.js'

const CORS_HEADERS = {
  'Access-Control-Allow-Origin':  '*',
  'Access-Control-Allow-Methods': 'POST, GET, OPTIONS',
  'Access-Control-Allow-Headers': 'Content-Type, Authorization, X-Confire-Sig, X-Confire-Timestamp, X-Confire-Device',
}

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    const url    = new URL(request.url)
    const method = request.method

    if (method === 'OPTIONS') {
      return new Response(null, { headers: CORS_HEADERS })
    }

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

    // ── Billing / checkout ────────────────────────────────────────────────
    if (method === 'POST' && url.pathname === '/api/checkout/create') {
      return handleCreateCheckout(request, env)
    }
    if (method === 'POST' && url.pathname === '/api/topup/create') {
      return handleCreateTopup(request, env)
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

    // ── Admin: manual plan sync (fallback / debug) ────────────────────────
    // Normally not needed — the Supabase webhook handles it automatically.
    // Useful for: initial setup, debugging, webhook failure recovery.
    if (method === 'POST' && url.pathname === '/admin/plans/sync') {
      try {
        await syncPlansToKV(env)
        return Response.json({ ok: true, message: 'Plans synced from Supabase → KV' })
      } catch (e) {
        return Response.json({ error: String(e) }, { status: 500 })
      }
    }

    // ── Health ────────────────────────────────────────────────────────────
    if (method === 'GET' && url.pathname === '/health') {
      return Response.json({ ok: true, version: '0.1.0', env: env.ENVIRONMENT })
    }

    return new Response('Not Found', { status: 404 })
  },
}

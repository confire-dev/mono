import type { Env } from './types.js'
import { handleOptimize }        from './handlers/optimize.js'
import { handleSessionStart }    from './handlers/session.js'
import { handleGenerateKey, handleMe } from './handlers/auth.js'
import { handleTelemetry }       from './handlers/telemetry.js'
import { handleStripeWebhook }   from './handlers/stripe.js'

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

    // ── Stripe webhooks ───────────────────────────────────────────────────
    if (method === 'POST' && url.pathname === '/webhooks/stripe') {
      return handleStripeWebhook(request, env)
    }

    // ── Health ────────────────────────────────────────────────────────────
    if (method === 'GET' && url.pathname === '/health') {
      return Response.json({ ok: true, version: '0.1.0', env: env.ENVIRONMENT })
    }

    return new Response('Not Found', { status: 404 })
  },
}

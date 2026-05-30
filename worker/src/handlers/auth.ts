// Auth endpoints:
//   POST /api/keys/generate  — exchange a Supabase JWT for a long-lived API key
//   GET  /api/me             — return current user info for the CLI `whoami` command

import type { Env } from '../types.js'
import { upsertProfile, generateApiKey, checkUsage } from '../lib/supabase.js'
import { trackEvent } from '../lib/analytics.js'

// verifySupabaseJWT validates a Supabase access_token using Supabase's REST API.
// Returns {sub, email} on success, null on failure.
async function verifySupabaseJWT(token: string, supabaseUrl: string, anonKey: string): Promise<{sub:string; email:string} | null> {
  const res = await fetch(`${supabaseUrl}/auth/v1/user`, {
    headers: { 'apikey': anonKey, 'Authorization': `Bearer ${token}` },
  })
  if (!res.ok) return null
  const user = await res.json() as { id: string; email: string }
  if (!user.id || !user.email) return null
  return { sub: user.id, email: user.email }
}

export async function handleGenerateKey(request: Request, env: Env): Promise<Response> {
  if (!env.SUPABASE_URL || !env.SUPABASE_ANON_KEY || !env.SUPABASE_SERVICE_KEY) {
    return Response.json({ error: 'auth not configured' }, { status: 503 })
  }

  const auth = request.headers.get('Authorization')
  if (!auth?.startsWith('Bearer ')) {
    return Response.json({ error: 'missing Supabase JWT' }, { status: 401 })
  }
  const jwt = auth.slice(7).trim()

  const verified = await verifySupabaseJWT(jwt, env.SUPABASE_URL, env.SUPABASE_ANON_KEY)
  if (!verified) {
    return Response.json({ error: 'invalid or expired Supabase token' }, { status: 401 })
  }

  const cfg = { url: env.SUPABASE_URL, serviceKey: env.SUPABASE_SERVICE_KEY }
  const user = await upsertProfile(cfg, verified.sub, verified.email)
  const apiKey = await generateApiKey(cfg, user.id)
  const usage = await checkUsage(cfg, user.id, user.plan)

  trackEvent(env.AE, env.AMPLITUDE_KEY, {
    userId: user.id, email: user.email, eventType: 'api_key_generated',
  })

  return Response.json({
    apiKey,
    user: { email: user.email, plan: user.plan },
    usage: { used: usage.used, limit: usage.limit },
    message: `✓ Logged in as ${user.email} · ${user.plan} plan · ${usage.used}/${usage.limit} requests used`,
  })
}

export async function handleMe(request: Request, env: Env): Promise<Response> {
  if (!env.SUPABASE_URL || !env.SUPABASE_SERVICE_KEY) {
    return Response.json({ email: 'dev@local', plan: 'free', used: 0, limit: 500 })
  }

  const auth = request.headers.get('Authorization')
  if (!auth?.startsWith('Bearer ')) return Response.json({ error: 'unauthorized' }, { status: 401 })

  const { validateApiKey, checkUsage } = await import('../lib/supabase.js')
  const cfg = { url: env.SUPABASE_URL, serviceKey: env.SUPABASE_SERVICE_KEY }
  const user = await validateApiKey(cfg, auth.slice(7).trim())
  if (!user) return Response.json({ error: 'invalid key' }, { status: 401 })

  const usage = await checkUsage(cfg, user.id, user.plan)
  return Response.json({ email: user.email, plan: user.plan, used: usage.used, limit: usage.limit })
}

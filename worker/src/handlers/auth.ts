import type { Env } from '../types.js'
import { upsertProfile, getProfileById, generateApiKey, validateApiKey, revokeApiKey, revokeCurrentKey, getUsageThisPeriod, getCreditBalance } from '../lib/supabase.js'
import { getPlan } from '../lib/plans.js'
import { trackEvent } from '../lib/analytics.js'

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

  const cfg        = { url: env.SUPABASE_URL, serviceKey: env.SUPABASE_SERVICE_KEY }
  const user       = await upsertProfile(cfg, verified.sub, verified.email)
  const deviceId   = request.headers.get('X-Confire-Device') ?? undefined
  const deviceName = request.headers.get('X-Confire-Device-Name') ?? undefined
  const apiKey     = await generateApiKey(cfg, user.id, deviceId, deviceName)
  const plan   = await getPlan(user.plan, env)
  const usage  = await getUsageThisPeriod(cfg, user.id, plan.id)

  trackEvent(env.AE, env.AMPLITUDE_KEY, {
    userId: user.id, email: user.email, eventType: 'api_key_generated',
  })

  return Response.json({
    apiKey,
    user:  { email: user.email, plan: plan.name },
    usage: { used: usage.cloudOptimizationsUsed, limit: plan.limits.cloudOptimizationsMonthly },
    message: [
      `✓ Logged in as ${user.email}`,
      `${plan.name} plan`,
      `${usage.cloudOptimizationsUsed}/${plan.limits.cloudOptimizationsMonthly} cloud optimizations used`,
    ].join(' · '),
  })
}

// POST /api/keys/revoke-self — revokes the key in the Authorization header (CLI logout).
export async function handleRevokeSelf(request: Request, env: Env): Promise<Response> {
  if (!env.SUPABASE_URL || !env.SUPABASE_SERVICE_KEY) {
    return Response.json({ ok: true }) // dev mode — no-op
  }
  const auth = request.headers.get('Authorization')
  if (!auth?.startsWith('Bearer ')) return Response.json({ error: 'unauthorized' }, { status: 401 })
  const cfg = { url: env.SUPABASE_URL, serviceKey: env.SUPABASE_SERVICE_KEY }
  await revokeCurrentKey(cfg, auth.slice(7).trim())
  return Response.json({ ok: true })
}

export async function handleRevokeKey(request: Request, env: Env): Promise<Response> {
  if (!env.SUPABASE_URL || !env.SUPABASE_ANON_KEY || !env.SUPABASE_SERVICE_KEY) {
    return Response.json({ error: 'auth not configured' }, { status: 503 })
  }

  const auth = request.headers.get('Authorization')
  if (!auth?.startsWith('Bearer ')) return Response.json({ error: 'unauthorized' }, { status: 401 })

  const cfg      = { url: env.SUPABASE_URL, serviceKey: env.SUPABASE_SERVICE_KEY }
  const token    = auth.slice(7).trim()
  const verified = await verifySupabaseJWT(token, env.SUPABASE_URL, env.SUPABASE_ANON_KEY)
  if (!verified) return Response.json({ error: 'invalid session' }, { status: 401 })

  const { keyId } = await request.json() as { keyId?: string }
  if (!keyId) return Response.json({ error: 'keyId required' }, { status: 400 })

  await revokeApiKey(cfg, verified.sub, keyId)
  return Response.json({ ok: true })
}

export async function handleMe(request: Request, env: Env): Promise<Response> {
  if (!env.SUPABASE_URL || !env.SUPABASE_SERVICE_KEY) {
    return Response.json({ email: 'dev@local', plan: 'free', subscription_status: 'none', used: 0, limit: 500 })
  }

  const auth = request.headers.get('Authorization')
  if (!auth?.startsWith('Bearer ')) return Response.json({ error: 'unauthorized' }, { status: 401 })

  const cfg   = { url: env.SUPABASE_URL, serviceKey: env.SUPABASE_SERVICE_KEY }
  const token = auth.slice(7).trim()

  // Accept either a confire API key or a Supabase JWT (sent by the dashboard).
  let user = await validateApiKey(cfg, token)
  if (!user && env.SUPABASE_ANON_KEY) {
    const verified = await verifySupabaseJWT(token, env.SUPABASE_URL, env.SUPABASE_ANON_KEY)
    if (verified) {
      user = await getProfileById(cfg, verified.sub)
    }
  }
  if (!user) return Response.json({ error: 'invalid key' }, { status: 401 })

  const plan    = await getPlan(user.plan, env)
  const usage   = await getUsageThisPeriod(cfg, user.id, plan.id)
  const credits = await getCreditBalance(cfg, user.id)
  const purchasedCredits = credits?.purchased_credits ?? 0
  const planLimit = plan.limits.cloudOptimizationsMonthly

  return Response.json({
    email:               user.email,
    plan:                plan.id,
    planId:              plan.id,
    subscription_status: user.subscription_status,
    used:                usage.cloudOptimizationsUsed,
    limit:               planLimit,
    purchasedCredits,
    effectiveLimit:      planLimit + purchasedCredits,
    limits:              plan.limits,
    features:            plan.features,
  })
}

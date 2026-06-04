// POST /api/subscription/cancel
//
// Sets cancel_at_period_end = true on the user's active Stripe subscription.
// The subscription stays active until the period ends — no immediate revocation.
// Paid access is only removed when Stripe fires customer.subscription.deleted.

import type { Env } from '../types.js'
import { writeAudit } from '../lib/supabase.js'

async function verifySupabaseJWT(
  token: string,
  supabaseUrl: string,
  anonKey: string,
): Promise<{ sub: string; email: string } | null> {
  const res = await fetch(`${supabaseUrl}/auth/v1/user`, {
    headers: { apikey: anonKey, Authorization: `Bearer ${token}` },
  })
  if (!res.ok) return null
  const user = await res.json() as { id: string; email: string }
  if (!user.id || !user.email) return null
  return { sub: user.id, email: user.email }
}

export async function handleCancelSubscription(request: Request, env: Env): Promise<Response> {
  if (!env.STRIPE_SECRET_KEY) {
    return Response.json({ error: 'stripe_not_configured' }, { status: 503 })
  }
  if (!env.SUPABASE_URL || !env.SUPABASE_ANON_KEY || !env.SUPABASE_SERVICE_KEY) {
    return Response.json({ error: 'auth_not_configured' }, { status: 503 })
  }

  const authHeader = request.headers.get('Authorization')
  if (!authHeader?.startsWith('Bearer ')) {
    return Response.json({ error: 'missing_token' }, { status: 401 })
  }
  const jwt = authHeader.slice(7).trim()

  const verified = await verifySupabaseJWT(jwt, env.SUPABASE_URL, env.SUPABASE_ANON_KEY)
  if (!verified) {
    return Response.json({ error: 'invalid_token' }, { status: 401 })
  }

  // Fetch the user's profile to get the Stripe subscription ID
  const profileRes = await fetch(
    `${env.SUPABASE_URL}/rest/v1/profiles?id=eq.${encodeURIComponent(verified.sub)}&select=stripe_subscription_id,subscription_status,plan&limit=1`,
    {
      headers: {
        apikey:        env.SUPABASE_SERVICE_KEY,
        Authorization: `Bearer ${env.SUPABASE_SERVICE_KEY}`,
      },
    },
  )

  if (!profileRes.ok) {
    return Response.json({ error: 'profile_fetch_failed' }, { status: 500 })
  }

  const profiles = await profileRes.json() as Array<{
    stripe_subscription_id?: string
    subscription_status: string
    plan: string
  }>

  const profile = profiles[0]

  if (!profile) {
    return Response.json({ error: 'profile_not_found' }, { status: 404 })
  }

  if (!profile.stripe_subscription_id) {
    return Response.json({ error: 'no_active_subscription' }, { status: 422 })
  }

  const cancelable = profile.subscription_status === 'active' || profile.subscription_status === 'trialing'
  if (!cancelable) {
    return Response.json(
      { error: 'not_cancelable', message: `Subscription status is '${profile.subscription_status}'.` },
      { status: 422 },
    )
  }

  // Tell Stripe to cancel at period end
  const stripeRes = await fetch(
    `https://api.stripe.com/v1/subscriptions/${encodeURIComponent(profile.stripe_subscription_id)}`,
    {
      method:  'POST',
      headers: {
        Authorization:  `Bearer ${env.STRIPE_SECRET_KEY}`,
        'Content-Type': 'application/x-www-form-urlencoded',
      },
      body: 'cancel_at_period_end=true',
    },
  )

  if (!stripeRes.ok) {
    const err = await stripeRes.json() as { error?: { message?: string } }
    return Response.json(
      { error: 'stripe_error', message: err.error?.message ?? 'Failed to cancel subscription.' },
      { status: 500 },
    )
  }

  const sub = await stripeRes.json() as {
    cancel_at_period_end: boolean
    current_period_end: number
    status: string
  }

  const periodEnd = new Date(sub.current_period_end * 1000).toISOString()

  // Store cancel_at_period_end in our DB immediately (webhook will also sync this)
  await fetch(
    `${env.SUPABASE_URL}/rest/v1/profiles?id=eq.${encodeURIComponent(verified.sub)}`,
    {
      method:  'PATCH',
      headers: {
        apikey:         env.SUPABASE_SERVICE_KEY,
        Authorization:  `Bearer ${env.SUPABASE_SERVICE_KEY}`,
        'Content-Type': 'application/json',
        Prefer:         'return=minimal',
      },
      body: JSON.stringify({ cancel_at_period_end: true }),
    },
  )

  const cfg = { url: env.SUPABASE_URL, serviceKey: env.SUPABASE_SERVICE_KEY }
  await writeAudit(cfg, verified.sub, 'subscription_cancel_requested', {
    subscription_id: profile.stripe_subscription_id,
    period_end:      periodEnd,
    plan:            profile.plan,
  })

  return Response.json({ ok: true, periodEnd })
}

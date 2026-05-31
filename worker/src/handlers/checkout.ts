// POST /api/checkout/create
//
// Validates a plan slug server-side, then creates a Stripe Checkout session.
// The caller (Astro auth callback) passes the user's Supabase JWT.
// This handler re-validates the JWT, resolves the plan from KV, and only
// creates a checkout session for plans that are active and purchasable.
//
// Paid access is NEVER granted here — only the Stripe webhook does that.

import type { Env } from '../types.js'
import { getPlans } from '../lib/plans.js'
import { upsertProfile, writeAudit } from '../lib/supabase.js'

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

export async function handleCreateCheckout(request: Request, env: Env): Promise<Response> {
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

  let body: { planSlug?: string; successUrl?: string; cancelUrl?: string }
  try {
    body = await request.json()
  } catch {
    return Response.json({ error: 'invalid_json' }, { status: 400 })
  }

  const { planSlug, successUrl, cancelUrl } = body
  if (!planSlug || !successUrl || !cancelUrl) {
    return Response.json({ error: 'missing_fields' }, { status: 400 })
  }

  // Validate plan slug server-side against live plan data.
  // Never trust the slug from the client for anything beyond lookup.
  const plans = await getPlans(env)
  const plan  = plans[planSlug]

  if (!plan || plan.status !== 'active' || plan.billingMode === 'free' || !plan.stripe.priceId) {
    return Response.json(
      { error: 'invalid_plan', message: 'Plan is not available for purchase.' },
      { status: 422 },
    )
  }

  const cfg     = { url: env.SUPABASE_URL, serviceKey: env.SUPABASE_SERVICE_KEY }
  const profile = await upsertProfile(cfg, verified.sub, verified.email)

  // Prevent duplicate subscriptions for already-active paid users.
  if (profile.subscription_status === 'active' && profile.plan !== 'free') {
    const isSamePlan = profile.plan === plan.id
    return Response.json({
      redirect: '/dashboard',
      message:  isSamePlan ? 'already_subscribed' : 'plan_change_from_billing',
    })
  }

  // Build Stripe Checkout session via REST API (no SDK — stays V8-safe).
  const params = new URLSearchParams({
    mode:                              'subscription',
    'line_items[0][price]':            plan.stripe.priceId,
    'line_items[0][quantity]':         '1',
    success_url:                       successUrl,
    cancel_url:                        cancelUrl,
    client_reference_id:               verified.sub,
    'metadata[user_id]':               verified.sub,
    'metadata[plan_slug]':             plan.id,
    'metadata[price_id]':              plan.stripe.priceId,
    // Also embed on the subscription so the webhook can read metadata there too.
    'subscription_data[metadata][user_id]':   verified.sub,
    'subscription_data[metadata][plan_slug]': plan.id,
  })

  // Link to existing Stripe customer when available to avoid duplicate customers.
  if (profile.stripe_customer_id) {
    params.set('customer', profile.stripe_customer_id)
  } else {
    params.set('customer_email', verified.email)
  }

  const stripeRes = await fetch('https://api.stripe.com/v1/checkout/sessions', {
    method:  'POST',
    headers: {
      Authorization:  `Bearer ${env.STRIPE_SECRET_KEY}`,
      'Content-Type': 'application/x-www-form-urlencoded',
    },
    body: params.toString(),
  })

  if (!stripeRes.ok) {
    const err = await stripeRes.json() as { error?: { message?: string } }
    return Response.json(
      { error: 'stripe_error', message: err.error?.message ?? 'Checkout creation failed.' },
      { status: 500 },
    )
  }

  const session = await stripeRes.json() as { url: string; id: string }

  await writeAudit(cfg, verified.sub, 'checkout_session_created', {
    plan_slug:  plan.id,
    session_id: session.id,
  })

  return Response.json({ checkoutUrl: session.url })
}

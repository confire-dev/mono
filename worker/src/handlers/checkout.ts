// POST /api/checkout/create
//
// Creates a Stripe Checkout session for a subscription.
// Accepts planSlug + interval (monthly | annual).
// Paid access is NEVER granted here — only the Stripe webhook does that.

import type { Env } from '../types.js'
import { getPlans, getPriceForInterval } from '../lib/plans.js'
import type { BillingInterval } from '../lib/plans.js'
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

  let body: { planSlug?: string; interval?: string; successUrl?: string; cancelUrl?: string }
  try {
    body = await request.json()
  } catch {
    return Response.json({ error: 'invalid_json' }, { status: 400 })
  }

  const { planSlug, successUrl, cancelUrl } = body
  const interval: BillingInterval = body.interval === 'annual' ? 'annual' : 'monthly'

  if (!planSlug || !successUrl || !cancelUrl) {
    return Response.json({ error: 'missing_fields' }, { status: 400 })
  }

  const plans = await getPlans(env)
  const plan  = plans[planSlug]

  if (
    !plan ||
    plan.status !== 'active' ||
    plan.billingMode === 'free' ||
    plan.billingMode === 'enterprise' ||
    plan.stripe.checkoutMode !== 'subscription'
  ) {
    return Response.json(
      { error: 'invalid_plan', message: 'Plan is not available for purchase.' },
      { status: 422 },
    )
  }

  const priceId = getPriceForInterval(plan, interval)
  if (!priceId) {
    return Response.json(
      { error: 'invalid_interval', message: `No ${interval} price configured for this plan.` },
      { status: 422 },
    )
  }

  const cfg     = { url: env.SUPABASE_URL, serviceKey: env.SUPABASE_SERVICE_KEY }
  const profile = await upsertProfile(cfg, verified.sub, verified.email)

  // Prevent duplicate subscriptions for already-active paid users on the same plan+interval
  if (
    profile.plan === plan.id &&
    profile.billing_interval === interval &&
    (profile.subscription_status === 'active' || profile.subscription_status === 'trialing')
  ) {
    return Response.json({ redirect: '/dashboard', message: 'already_subscribed' })
  }

  const params = new URLSearchParams({
    mode:                                            'subscription',
    'line_items[0][price]':                          priceId,
    'line_items[0][quantity]':                       '1',
    success_url:                                     successUrl,
    cancel_url:                                      cancelUrl,
    client_reference_id:                             verified.sub,
    // Checkout session metadata
    'metadata[kind]':                                'subscription',
    'metadata[plan_id]':                             plan.id,
    'metadata[billing_interval]':                    interval,
    'metadata[credit_type]':                         'remote_optimization',
    // Subscription metadata (carried to webhook events)
    'subscription_data[metadata][kind]':             'subscription',
    'subscription_data[metadata][plan_id]':          plan.id,
    'subscription_data[metadata][billing_interval]': interval,
    'subscription_data[metadata][credit_type]':      'remote_optimization',
  })

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
    plan_id:    plan.id,
    interval,
    session_id: session.id,
    price_id:   priceId,
  })

  return Response.json({ checkoutUrl: session.url })
}

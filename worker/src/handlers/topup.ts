// POST /api/topup/create
//
// Creates a Stripe Checkout session for a one-time top-up pack purchase.
// The CLI calls this with the user's API key; we resolve the user, create
// the session, and return the checkout URL. Payment confirmation happens
// exclusively via the Stripe webhook — this handler never grants credits.
//
// Pack size: 5,000 requests per pack, $5 per pack.
// Quantity: caller-supplied (1–20), stored in session metadata for the webhook.

import type { Env } from '../types.js'
import { authenticate } from '../lib/auth.js'
import { writeAudit } from '../lib/supabase.js'

const REQUESTS_PER_PACK = 5_000
const MAX_PACKS         = 20

export async function handleCreateTopup(request: Request, env: Env): Promise<Response> {
  if (!env.STRIPE_SECRET_KEY || !env.STRIPE_TOPUP_PRICE_ID) {
    return Response.json({ error: 'topup_not_configured' }, { status: 503 })
  }

  const auth = await authenticate(request, env)
  if (!auth.ok) {
    return Response.json({ error: auth.error }, { status: auth.status })
  }
  const { user } = auth

  let quantity = 1
  try {
    const body = await request.json() as { quantity?: number }
    if (body.quantity && Number.isInteger(body.quantity)) {
      quantity = Math.max(1, Math.min(MAX_PACKS, body.quantity))
    }
  } catch { /* default to 1 pack */ }

  const totalCredits = quantity * REQUESTS_PER_PACK

  // Resolve origin for success/cancel URLs
  const origin    = new URL(request.url).origin
  const successUrl = `${origin.replace('api.confire.dev', 'confire.dev')}/billing/topup-success`
  const cancelUrl  = `${origin.replace('api.confire.dev', 'confire.dev')}/billing/cancelled`

  // Create Stripe Checkout session (payment mode = one-time charge)
  const params = new URLSearchParams({
    mode:                          'payment',
    'line_items[0][price]':        env.STRIPE_TOPUP_PRICE_ID,
    'line_items[0][quantity]':     String(quantity),
    client_reference_id:           user.id,
    success_url:                   successUrl,
    cancel_url:                    cancelUrl,
    // Metadata carried to checkout.session.completed webhook
    'metadata[type]':              'topup',
    'metadata[total_credits]':     String(totalCredits),
    'metadata[requests_per_pack]': String(REQUESTS_PER_PACK),
    'metadata[quantity]':          String(quantity),
  })

  // Attach existing Stripe customer if available (avoids duplicate customers)
  if (user.stripe_customer_id) {
    params.set('customer', user.stripe_customer_id)
  } else if (user.email) {
    params.set('customer_email', user.email)
  }

  const res = await fetch('https://api.stripe.com/v1/checkout/sessions', {
    method:  'POST',
    headers: {
      Authorization:  `Bearer ${env.STRIPE_SECRET_KEY}`,
      'Content-Type': 'application/x-www-form-urlencoded',
    },
    body: params.toString(),
  })

  if (!res.ok) {
    const err = await res.json() as { error?: { message?: string } }
    return Response.json(
      { error: 'stripe_error', message: err.error?.message ?? 'Stripe error' },
      { status: 502 },
    )
  }

  const session = await res.json() as { id: string; url: string }

  const cfg = env.SUPABASE_URL && env.SUPABASE_SERVICE_KEY
    ? { url: env.SUPABASE_URL, serviceKey: env.SUPABASE_SERVICE_KEY }
    : null

  if (cfg) {
    await writeAudit(cfg, user.id, 'topup_checkout_created', {
      session_id:     session.id,
      quantity,
      total_credits:  totalCredits,
    }).catch(() => {})
  }

  return Response.json({ url: session.url, totalCredits, quantity })
}

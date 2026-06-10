// POST /api/topup/create
//
// Creates a Stripe Checkout session for a one-time top-up credit pack.
// Price and pack config are loaded from the billing_items table — no env var.
// Payment confirmation and credit granting happen exclusively via the Stripe webhook.

import type { Env } from '../types.js'
import { authenticate } from '../lib/auth.js'
import { fetchBillingItem, writeAudit } from '../lib/supabase.js'

const TOPUP_ITEM_ID = 'topup_remote_optimization_pack'

export async function handleCreateTopup(request: Request, env: Env): Promise<Response> {
  if (!env.STRIPE_SECRET_KEY) {
    return Response.json({ error: 'topup_not_configured' }, { status: 503 })
  }
  if (!env.SUPABASE_URL || !env.SUPABASE_SERVICE_KEY) {
    return Response.json({ error: 'supabase_not_configured' }, { status: 503 })
  }

  const auth = await authenticate(request, env)
  if (!auth.ok) {
    return Response.json({ error: auth.error }, { status: auth.status })
  }
  const { user } = auth
  const cfg = { url: env.SUPABASE_URL, serviceKey: env.SUPABASE_SERVICE_KEY }

  // Load top-up config from DB (price ID lives in billing_items, not env)
  const item = await fetchBillingItem(cfg, TOPUP_ITEM_ID)
  if (!item) {
    return Response.json({ error: 'topup_not_configured' }, { status: 503 })
  }

  const { creditsPerUnit, maxQuantity, stripe } = item.config
  const maxQty = maxQuantity ?? 20

  let quantity = 1
  try {
    const body = await request.json() as { quantity?: number }
    if (body.quantity && Number.isInteger(body.quantity)) {
      quantity = Math.max(1, Math.min(maxQty, body.quantity))
    }
  } catch { /* default to 1 pack */ }

  const origin     = new URL(request.url).origin
  const successUrl = `${origin.replace('api.confire.dev', 'confire.dev')}/billing?topup=success`
  const cancelUrl  = `${origin.replace('api.confire.dev', 'confire.dev')}/billing?topup=cancelled`
  const successUrlDev = successUrl.replace('api-dev.confire.dev', 'dev.confire.dev')
  const cancelUrlDev  = cancelUrl.replace('api-dev.confire.dev', 'dev.confire.dev')

  const params = new URLSearchParams({
    mode:                                               'payment',
    'line_items[0][price]':                             stripe.priceId,
    'line_items[0][quantity]':                          String(quantity),
    'line_items[0][adjustable_quantity][enabled]':      'true',
    'line_items[0][adjustable_quantity][minimum]':      '1',
    'line_items[0][adjustable_quantity][maximum]':      String(maxQty),
    client_reference_id:                                user.id,
    success_url:                                        successUrl.includes('api-dev') ? successUrlDev : successUrl,
    cancel_url:                                         cancelUrl.includes('api-dev')  ? cancelUrlDev  : cancelUrl,
    'metadata[kind]':                                   'topup',
    'metadata[topup_id]':                               TOPUP_ITEM_ID,
    'metadata[credits_per_unit]':                       String(creditsPerUnit),
    'metadata[max_quantity]':                           String(maxQty),
    // quantity snapshotted here; webhook re-reads from line_items for accuracy
    'metadata[quantity]':                               String(quantity),
  })

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

  await writeAudit(cfg, user.id, 'topup_checkout_created', {
    session_id:    session.id,
    quantity,
    credits_per_unit: creditsPerUnit,
    price_id:      stripe.priceId,
  }).catch(() => {})

  return Response.json({ url: session.url, creditsPerUnit, quantity, maxQuantity: maxQty })
}

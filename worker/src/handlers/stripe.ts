// Stripe webhook handler — POST /webhooks/stripe
//
// Stripe is the source of truth for payments, invoices, and subscription state.
// This handler receives Stripe events, verifies the signature, and updates
// Supabase with the product-facing billing state needed for access control.
//
// v1: disputes, refunds, and failed-payment edge cases handled manually
// via the Stripe + Supabase dashboards. No custom admin panel yet.

import type { Env } from '../types.js'
import {
  handleSubscriptionUpdated,
  handleSubscriptionCanceled,
  writeAudit,
} from '../lib/supabase.js'
import { getPlans } from '../lib/plans.js'

// Relevant Stripe event types handled in v1.
type StripeEventType =
  | 'customer.subscription.created'
  | 'customer.subscription.updated'
  | 'customer.subscription.deleted'
  | 'invoice.paid'
  | 'invoice.payment_failed'

export async function handleStripeWebhook(request: Request, env: Env): Promise<Response> {
  if (!env.STRIPE_WEBHOOK_SECRET) {
    return Response.json({ error: 'webhook not configured' }, { status: 503 })
  }

  const body = await request.text()
  const sig  = request.headers.get('stripe-signature') ?? ''

  // Verify the Stripe webhook signature.
  const valid = await verifyStripeSignature(body, sig, env.STRIPE_WEBHOOK_SECRET)
  if (!valid) {
    return Response.json({ error: 'invalid signature' }, { status: 400 })
  }

  let event: { id: string; type: string; data: { object: Record<string, unknown> } }
  try {
    event = JSON.parse(body)
  } catch {
    return Response.json({ error: 'invalid JSON' }, { status: 400 })
  }

  if (!env.SUPABASE_URL || !env.SUPABASE_SERVICE_KEY) {
    return Response.json({ error: 'supabase not configured' }, { status: 503 })
  }
  const cfg = { url: env.SUPABASE_URL, serviceKey: env.SUPABASE_SERVICE_KEY }

  // Dispatch by event type. Unrecognised events are accepted and ignored
  // (Stripe requires 2xx or it retries indefinitely).
  switch (event.type as StripeEventType) {
    case 'customer.subscription.created':
    case 'customer.subscription.updated': {
      const sub = event.data.object
      await handleSubscriptionUpdated(cfg, {
        stripeEventId:      event.id,
        stripeCustomerId:   sub['customer'] as string,
        subscriptionId:     sub['id'] as string,
        status:             sub['status'] as string,
        planId:             await resolvePlanId(extractPriceId(sub), env),
        currentPeriodStart: sub['current_period_start'] as number,
        currentPeriodEnd:   sub['current_period_end'] as number,
      })
      await writeAudit(cfg, null, 'stripe_subscription_updated', {
        event_id:    event.id,
        customer_id: sub['customer'],
        status:      sub['status'],
      })
      break
    }

    case 'customer.subscription.deleted': {
      const sub = event.data.object
      await handleSubscriptionCanceled(cfg, sub['customer'] as string)
      await writeAudit(cfg, null, 'stripe_subscription_canceled', {
        event_id:    event.id,
        customer_id: sub['customer'],
      })
      break
    }

    case 'invoice.paid': {
      // Invoice paid = subscription renewed. The subscription.updated event
      // handles granting credits, so we just audit here.
      const inv = event.data.object
      await writeAudit(cfg, null, 'stripe_invoice_paid', {
        event_id:        event.id,
        customer_id:     inv['customer'],
        amount_paid:     inv['amount_paid'],
        subscription_id: inv['subscription'],
      })
      break
    }

    case 'invoice.payment_failed': {
      // Mark subscription as past_due. The subscription.updated event may also
      // fire, but we handle it here explicitly for clarity.
      const inv = event.data.object
      await writeAudit(cfg, null, 'stripe_payment_failed', {
        event_id:    event.id,
        customer_id: inv['customer'],
      })
      // Payment failure → subscription.updated will arrive with status=past_due
      // and update Supabase. No additional action needed in v1.
      break
    }

    default:
      // Unknown event — accept and ignore.
  }

  return Response.json({ received: true })
}

// ── Signature verification ─────────────────────────────────────────────────
// Stripe uses HMAC-SHA256 over `${timestamp}.${raw_body}`.
// Header: Stripe-Signature: t=1234,v1=abcdef...

async function verifyStripeSignature(
  rawBody: string,
  header: string,
  secret: string,
): Promise<boolean> {
  try {
    const parts = Object.fromEntries(
      header.split(',').map(p => p.split('=') as [string, string])
    )
    const ts  = parts['t']
    const v1  = parts['v1']
    if (!ts || !v1) return false

    // Reject stale webhooks (>5 minute clock skew)
    const delta = Date.now() / 1000 - parseInt(ts, 10)
    if (Math.abs(delta) > 300) return false

    const payload = `${ts}.${rawBody}`
    const key = await crypto.subtle.importKey(
      'raw',
      new TextEncoder().encode(secret),
      { name: 'HMAC', hash: 'SHA-256' },
      false,
      ['sign'],
    )
    const sigBuf  = await crypto.subtle.sign('HMAC', key, new TextEncoder().encode(payload))
    const sigHex  = Array.from(new Uint8Array(sigBuf))
      .map(b => b.toString(16).padStart(2, '0')).join('')

    // Constant-time comparison
    return timingSafeEqual(sigHex, v1)
  } catch {
    return false
  }
}

function timingSafeEqual(a: string, b: string): boolean {
  if (a.length !== b.length) return false
  let diff = 0
  for (let i = 0; i < a.length; i++) diff |= a.charCodeAt(i) ^ b.charCodeAt(i)
  return diff === 0
}

function extractPriceId(sub: Record<string, unknown>): string {
  const items = (sub['items'] as Record<string, unknown>)?.['data'] as Array<Record<string,unknown>> | undefined
  const price = items?.[0]?.['price'] as Record<string,string> | undefined
  return price?.['id'] ?? ''
}

// Resolve a Stripe price ID to an internal plan ID using the plans table
// (via the KV-cached getPlans). No env var hardcoding — price IDs live in the DB.
async function resolvePlanId(priceId: string, env: Env): Promise<string> {
  if (!priceId) return 'free'
  const plans = await getPlans(env)
  for (const plan of Object.values(plans)) {
    if (plan.stripe.priceId === priceId) return plan.id
  }
  return 'free'
}

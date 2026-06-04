// Stripe webhook handler — POST /webhooks/stripe
//
// Stripe is the source of truth for payments, invoices, and subscription state.
// Credits are granted on invoice.paid — not on subscription.created/updated —
// to avoid double-granting and to align with the actual payment event.
//
// Idempotency: every event is recorded in stripe_webhook_events before processing.
// If the record already exists the event is silently accepted and skipped.

import type { Env } from '../types.js'
import {
  handleSubscriptionUpdated,
  handleSubscriptionCanceled,
  handleInvoicePaid,
  grantPurchasedCredits,
  markWebhookEventProcessed,
  writeAudit,
} from '../lib/supabase.js'
import { getPlans } from '../lib/plans.js'
import type { BillingInterval } from '../lib/plans.js'

export async function handleStripeWebhook(request: Request, env: Env): Promise<Response> {
  if (!env.STRIPE_WEBHOOK_SECRET) {
    return Response.json({ error: 'webhook not configured' }, { status: 503 })
  }

  const body = await request.text()
  const sig  = request.headers.get('stripe-signature') ?? ''

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

  // Dispatch-level idempotency — skip events we've already processed.
  const isNew = await markWebhookEventProcessed(cfg, event.id, event.type)
  if (!isNew) {
    return Response.json({ received: true, skipped: true })
  }

  switch (event.type) {
    case 'checkout.session.completed': {
      const sess           = event.data.object
      const userId         = sess['client_reference_id'] as string | undefined
      const stripeCustomer = sess['customer'] as string | undefined
      const metadata       = (sess['metadata'] ?? {}) as Record<string, string>
      const kind           = metadata['kind'] ?? metadata['type'] // 'topup' or 'subscription'

      if (kind === 'topup' && userId) {
        // Top-up: read final quantity from line items (user may have changed it in Checkout)
        const lineItemsRes = await fetch(
          `https://api.stripe.com/v1/checkout/sessions/${sess['id']}/line_items?limit=1`,
          { headers: { Authorization: `Bearer ${env.STRIPE_SECRET_KEY}` } }
        )
        let quantity = parseInt(metadata['quantity'] ?? '1', 10)
        if (lineItemsRes.ok) {
          const li = await lineItemsRes.json() as { data: Array<{ quantity: number }> }
          quantity = li.data[0]?.quantity ?? quantity
        }

        const creditsPerUnit = parseInt(metadata['credits_per_unit'] ?? metadata['requests_per_pack'] ?? '5000', 10)
        const totalCredits = Math.max(1, Math.min(20, quantity)) * creditsPerUnit

        await grantPurchasedCredits(cfg, {
          userId,
          credits:       totalCredits,
          stripeEventId: event.id,
        })
        await writeAudit(cfg, userId, 'topup_credits_granted', {
          event_id:      event.id,
          total_credits: totalCredits,
          quantity,
        })
      }

      // Link Stripe customer to profile for both topup and subscription checkouts
      if (userId && stripeCustomer) {
        await linkStripeCustomer(env, cfg, userId, stripeCustomer, event.id)
      }
      break
    }

    case 'customer.subscription.created':
    case 'customer.subscription.updated': {
      const sub            = event.data.object
      const priceId        = extractPriceId(sub)
      const { planId, billingInterval } = await resolvePlanFromPrice(priceId, env)

      await handleSubscriptionUpdated(cfg, {
        stripeEventId:      event.id,
        stripeCustomerId:   sub['customer'] as string,
        subscriptionId:     sub['id'] as string,
        status:             sub['status'] as string,
        planId,
        billingInterval,
        currentPeriodStart: sub['current_period_start'] as number,
        currentPeriodEnd:   sub['current_period_end'] as number,
        cancelAtPeriodEnd:  !!(sub['cancel_at_period_end']),
      })
      await writeAudit(cfg, null, 'stripe_subscription_updated', {
        event_id:    event.id,
        customer_id: sub['customer'],
        status:      sub['status'],
        plan_id:     planId,
        interval:    billingInterval,
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
      const inv            = event.data.object
      const priceId        = extractInvoicePriceId(inv)
      const { planId, billingInterval } = await resolvePlanFromPrice(priceId, env)

      // Only grant credits for subscription invoices (not one-off top-up payments)
      if (planId !== 'free' && inv['billing_reason'] !== 'manual') {
        await handleInvoicePaid(cfg, {
          stripeCustomerId: inv['customer'] as string,
          stripeInvoiceId:  inv['id'] as string,
          planId,
          billingInterval,
        })
      }
      await writeAudit(cfg, null, 'stripe_invoice_paid', {
        event_id:        event.id,
        customer_id:     inv['customer'],
        amount_paid:     inv['amount_paid'],
        subscription_id: inv['subscription'],
        plan_id:         planId,
        interval:        billingInterval,
      })
      break
    }

    case 'invoice.payment_failed': {
      const inv = event.data.object
      await writeAudit(cfg, null, 'stripe_payment_failed', {
        event_id:    event.id,
        customer_id: inv['customer'],
      })
      break
    }

    default:
      // Unknown event — accept and ignore so Stripe doesn't retry.
  }

  return Response.json({ received: true })
}

async function linkStripeCustomer(
  env: Env,
  cfg: { url: string; serviceKey: string },
  userId: string,
  stripeCustomer: string,
  eventId: string,
): Promise<void> {
  await fetch(
    `${env.SUPABASE_URL}/rest/v1/profiles?id=eq.${encodeURIComponent(userId)}`,
    {
      method:  'PATCH',
      headers: {
        apikey:         env.SUPABASE_SERVICE_KEY!,
        Authorization:  `Bearer ${env.SUPABASE_SERVICE_KEY}`,
        'Content-Type': 'application/json',
        Prefer:         'return=minimal',
      },
      body: JSON.stringify({
        stripe_customer_id: stripeCustomer,
        updated_at:         new Date().toISOString(),
      }),
    },
  ).catch(() => {})
  await writeAudit(cfg, userId, 'stripe_customer_linked', {
    event_id:           eventId,
    stripe_customer_id: stripeCustomer,
  })
}

// ── Price → plan resolution ────────────────────────────────────────────────

async function resolvePlanFromPrice(
  priceId: string,
  env: Env,
): Promise<{ planId: string; billingInterval: BillingInterval }> {
  if (!priceId) return { planId: 'free', billingInterval: 'monthly' }
  const plans = await getPlans(env)
  for (const plan of Object.values(plans)) {
    if (plan.stripe.prices?.monthly === priceId) return { planId: plan.id, billingInterval: 'monthly' }
    if (plan.stripe.prices?.annual  === priceId) return { planId: plan.id, billingInterval: 'annual'  }
    if (plan.stripe.priceId         === priceId) {
      const interval: BillingInterval = plan.interval === 'annual' ? 'annual' : 'monthly'
      return { planId: plan.id, billingInterval: interval }
    }
  }
  return { planId: 'free', billingInterval: 'monthly' }
}

function extractPriceId(sub: Record<string, unknown>): string {
  const items = (sub['items'] as Record<string, unknown>)?.['data'] as Array<Record<string, unknown>> | undefined
  const price = items?.[0]?.['price'] as Record<string, string> | undefined
  return price?.['id'] ?? ''
}

function extractInvoicePriceId(inv: Record<string, unknown>): string {
  const lines = (inv['lines'] as Record<string, unknown>)?.['data'] as Array<Record<string, unknown>> | undefined
  const price = lines?.[0]?.['price'] as Record<string, string> | undefined
  return price?.['id'] ?? ''
}

// ── Signature verification ─────────────────────────────────────────────────

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
    const sigBuf = await crypto.subtle.sign('HMAC', key, new TextEncoder().encode(payload))
    const sigHex = Array.from(new Uint8Array(sigBuf))
      .map(b => b.toString(16).padStart(2, '0')).join('')

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

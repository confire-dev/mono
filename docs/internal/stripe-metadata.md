# Stripe Metadata Reference

All metadata fields set on Stripe objects by Confire.
Keep this in sync when adding new products or checkout flows.

---

## Products

### Confire Dev

Set these metadata fields on the product in the Stripe Dashboard:

| Key | Value |
|---|---|
| `product_kind` | `subscription` |
| `plan_id` | `dev` |
| `plan_name` | `Dev` |

---

## Prices

### Dev monthly

| Key | Value |
|---|---|
| `product_kind` | `subscription` |
| `plan_id` | `dev` |
| `plan_name` | `Dev` |
| `billing_interval` | `monthly` |

### Dev annual

| Key | Value |
|---|---|
| `product_kind` | `subscription` |
| `plan_id` | `dev` |
| `plan_name` | `Dev` |
| `billing_interval` | `annual` |

---

## Checkout Session metadata (set automatically by Worker)

### Subscription checkout

| Key | Value |
|---|---|
| `kind` | `subscription` |
| `plan_id` | `dev` |
| `billing_interval` | `monthly` or `annual` |

### subscription_data metadata (carried to all webhook events)

Same fields as above — duplicated onto the subscription so every webhook event carries them.

| Key | Value |
|---|---|
| `kind` | `subscription` |
| `plan_id` | `dev` |
| `billing_interval` | `monthly` or `annual` |

---

## Webhook events handled

| Stripe event | What the Worker does |
|---|---|
| `checkout.session.completed` | Links Stripe customer ID to profile |
| `customer.subscription.created` | Updates plan + billing interval in profile |
| `customer.subscription.updated` | Updates plan + billing interval + period end in profile |
| `customer.subscription.deleted` | Reverts plan to `free`, sets status `canceled` |
| `invoice.paid` | Writes audit log |
| `invoice.payment_failed` | Writes audit log |

Plan resolution in the webhook uses **price ID lookup** against `plans.config` in the DB — not metadata.
Metadata is for traceability only.

---

## Stripe object IDs

### Dev environment (test mode)

| Object | ID |
|---|---|
| Dev product | `prod_UdYngLggCJUx5B` |
| Dev monthly price | `price_1TeHfTCjsgbilsT2kKli4wU5` |
| Dev annual price | `price_1TeHggCjsgbilsT2a56012vt` |

### Production (live mode)

Create products in Stripe Dashboard, then update `plans` table in prod Supabase:

```sql
UPDATE plans
SET config = jsonb_set(jsonb_set(jsonb_set(config,
  '{stripe,productId}', '"prod_LIVE_PRODUCT_ID"'),
  '{stripe,prices,monthly}', '"price_LIVE_MONTHLY_ID"'),
  '{stripe,prices,annual}',  '"price_LIVE_ANNUAL_ID"')
WHERE id = 'dev';
```

Then trigger a KV sync via `POST /admin/plans/sync` (or the Supabase webhook fires automatically).
No Worker redeploy needed — price IDs live in KV, not in env vars.

---

## Stripe webhook endpoint setup

| Field | Dev | Production |
|---|---|---|
| Endpoint URL | `https://api.dev.confire.dev/webhooks/stripe` | `https://api.confire.dev/webhooks/stripe` |
| Events to listen for | `checkout.session.completed`, `customer.subscription.*`, `invoice.paid`, `invoice.payment_failed` | same |
| Signing secret | → `wrangler secret put STRIPE_WEBHOOK_SECRET --env dev` | → `wrangler secret put STRIPE_WEBHOOK_SECRET --env production` |

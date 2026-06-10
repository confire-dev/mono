# Confire Pricing — Setup Runbook

This document is the canonical runbook for recreating Stripe products/prices and
Supabase plan rows from scratch. Run in order. Do not skip the code changes section —
they must ship before the Supabase SQL or the Worker will throw on missing `credits`.

---

## Plans summary

| Plan | Monthly | Annual | Status |
|------|---------|--------|--------|
| Free | $0 | — | active |
| Dev | $12/mo | $99/yr | active |
| Team | $249/mo flat (≤10 seats) | — | hidden (early access) |

---

## Step 0 — Code changes (ship these first)

Before touching Stripe or Supabase, the following code changes are required.
The Worker reads `plan.credits` at runtime — if that field is missing the Worker throws.

### `worker/src/lib/plans.ts`

Remove the `credits` block from the `Plan` interface and `getCreditsForInterval`.
Add `maxCustomRules` to `limits`. Add new feature flags.

```typescript
// Remove entirely:
credits: {
  includedMonthly: number
  annual?: number
  rollover: boolean
  allowManualGrants: boolean
  allowPurchases: boolean
}

// Updated limits:
limits: {
  maxPayloadBytes: number
  retainedHistoryDays: number
  cliSessions: number
  maxCustomRules: number        // 0 = disabled, -1 = unlimited
}

// Updated features — keep existing, add:
features: {
  firewallEnabled: boolean
  customRules: boolean          // NEW — custom guardrail rules
  policySync: boolean           // NEW — confire policy pull
  usageDashboard: boolean
  advancedUsageDashboard: boolean
  cliSessionManagement: boolean
  exportData: boolean
  ssoSaml: boolean
  firewallGroupToggles: boolean
  securityEventHistory: boolean
  provenanceTracking: boolean
  sharedPolicies: boolean       // NEW — team shared guardrails
  teamDashboard: boolean        // NEW — cross-agent team view
  auditLogs: boolean            // NEW — tamper-evident history
  agentInventory: boolean       // NEW — device/agent registry
  fleetControls: boolean        // NEW — force-upgrade, lockdown
}

// Remove getCreditsForInterval — no longer needed.
```

### `worker/src/handlers/checkout.ts`

Remove `credit_type` from both metadata blocks. It was optimizer-era metadata
and is no longer meaningful.

```typescript
// Remove these two lines from the params block:
'metadata[credit_type]':                         'remote_optimization',
'subscription_data[metadata][credit_type]':      'remote_optimization',
```

### `worker/src/handlers/stripe.ts`

The `invoice.paid` handler calls `handleInvoicePaid` which grants credits.
Remove that call (or replace with a no-op). The credit grant is no longer needed.

```typescript
// Remove or no-op this block inside case 'invoice.paid':
if (planId !== 'free' && inv['billing_reason'] !== 'manual') {
  await handleInvoicePaid(cfg, { ... })
}
```

### `worker/src/lib/supabase.ts`

Update the `Plan` type:
```typescript
// Old:
export type Plan = 'free' | 'dev' | 'dev_annual' | 'pro' | 'pro_annual' | 'enterprise'

// New:
export type Plan = 'free' | 'dev' | 'team' | 'enterprise'
```

---

## Step 1 — Stripe (sandbox)

Run these with the Stripe CLI. Note every `prod_xxx` and `price_xxx` ID — you need
them for Step 3.

```bash
# ── Dev product ──────────────────────────────────────────────────────────────

stripe products create \
  --name="Confire Dev" \
  --description="Custom guardrails, firewall history, and dashboard for developers who want control." \
  --metadata[plan_id]=dev

# → copy the product ID: prod_DEV

stripe prices create \
  --product=prod_DEV \
  --currency=usd \
  --unit-amount=1200 \
  --recurring[interval]=month \
  --nickname="Dev Monthly" \
  --metadata[plan_id]=dev \
  --metadata[billing_interval]=monthly

# → copy the price ID: price_DEV_MONTHLY

stripe prices create \
  --product=prod_DEV \
  --currency=usd \
  --unit-amount=9900 \
  --recurring[interval]=year \
  --nickname="Dev Annual" \
  --metadata[plan_id]=dev \
  --metadata[billing_interval]=annual

# → copy the price ID: price_DEV_ANNUAL


# ── Team product ─────────────────────────────────────────────────────────────

stripe products create \
  --name="Confire Team" \
  --description="Org-wide agent security policies, audit logs, and fleet controls for engineering teams." \
  --metadata[plan_id]=team

# → copy the product ID: prod_TEAM

stripe prices create \
  --product=prod_TEAM \
  --currency=usd \
  --unit-amount=24900 \
  --recurring[interval]=month \
  --nickname="Team Monthly (up to 10 seats)" \
  --metadata[plan_id]=team \
  --metadata[billing_interval]=monthly \
  --metadata[seats_included]=10

# → copy the price ID: price_TEAM_MONTHLY
```

---

## Step 2 — Supabase schema migration

Run in Supabase SQL Editor. This widens the `profiles.plan` CHECK constraint
to include `team` and drops the old optimizer-era plan IDs.

```sql
-- Widen the plan CHECK constraint.
-- Any existing rows with 'dev_annual', 'pro', 'pro_annual' become 'dev' or 'free'.
-- Run the data migration first, then drop and recreate the constraint.

-- 1. Migrate stale plan IDs on existing profiles
UPDATE profiles SET plan = 'dev'  WHERE plan IN ('dev_annual', 'pro', 'pro_annual');
UPDATE profiles SET plan = 'free' WHERE plan NOT IN ('free', 'dev', 'team', 'enterprise');

-- 2. Drop old constraint
ALTER TABLE profiles DROP CONSTRAINT IF EXISTS profiles_plan_check;

-- 3. Add new constraint
ALTER TABLE profiles ADD CONSTRAINT profiles_plan_check
  CHECK (plan IN ('free', 'dev', 'team', 'enterprise'));
```

---

## Step 3 — Supabase plan rows

Replace all plan rows. Run in Supabase SQL Editor.
Substitute the real Stripe IDs from Step 1 before running.

```sql
-- Remove all old plans
DELETE FROM plans;

-- ── Free ─────────────────────────────────────────────────────────────────────
INSERT INTO plans (id, status, config) VALUES (
  'free', 'active',
  '{
    "id": "free",
    "name": "Free",
    "tagline": "Full local firewall for individual developers. No account, works offline.",
    "status": "active",
    "audience": "individual",
    "billingMode": "free",
    "interval": null,
    "stripe": {
      "productId": null,
      "priceId": null,
      "prices": {},
      "checkoutMode": null
    },
    "pricing": {
      "amountCents": 0,
      "currency": "usd",
      "displayPrice": "$0"
    },
    "limits": {
      "maxPayloadBytes": 3145728,
      "retainedHistoryDays": 7,
      "cliSessions": 3,
      "maxCustomRules": 0
    },
    "features": {
      "firewallEnabled": false,
      "customRules": false,
      "policySync": false,
      "usageDashboard": false,
      "advancedUsageDashboard": false,
      "cliSessionManagement": false,
      "exportData": false,
      "ssoSaml": false,
      "firewallGroupToggles": false,
      "securityEventHistory": false,
      "provenanceTracking": true,
      "sharedPolicies": false,
      "teamDashboard": false,
      "auditLogs": false,
      "agentInventory": false,
      "fleetControls": false
    },
    "telemetry": {
      "requiredUsageMetering": true,
      "optionalProductAnalyticsDefault": false
    }
  }'::jsonb
);

-- ── Dev ──────────────────────────────────────────────────────────────────────
-- Replace prod_DEV, price_DEV_MONTHLY, price_DEV_ANNUAL with real IDs from Step 1.
INSERT INTO plans (id, status, config) VALUES (
  'dev', 'active',
  '{
    "id": "dev",
    "name": "Dev",
    "tagline": "Custom guardrails, firewall history, and dashboard for developers who want control.",
    "status": "active",
    "audience": "individual",
    "billingMode": "subscription",
    "interval": null,
    "stripe": {
      "productId": "prod_DEV",
      "priceId": "price_DEV_MONTHLY",
      "prices": {
        "monthly": "price_DEV_MONTHLY",
        "annual": "price_DEV_ANNUAL"
      },
      "checkoutMode": "subscription"
    },
    "pricing": {
      "amountCents": 1200,
      "currency": "usd",
      "displayPrice": "$12/mo"
    },
    "limits": {
      "maxPayloadBytes": 10485760,
      "retainedHistoryDays": 90,
      "cliSessions": 10,
      "maxCustomRules": 100
    },
    "features": {
      "firewallEnabled": true,
      "customRules": true,
      "policySync": true,
      "usageDashboard": true,
      "advancedUsageDashboard": false,
      "cliSessionManagement": true,
      "exportData": true,
      "ssoSaml": false,
      "firewallGroupToggles": true,
      "securityEventHistory": true,
      "provenanceTracking": true,
      "sharedPolicies": false,
      "teamDashboard": false,
      "auditLogs": false,
      "agentInventory": false,
      "fleetControls": false
    },
    "telemetry": {
      "requiredUsageMetering": true,
      "optionalProductAnalyticsDefault": true
    }
  }'::jsonb
);

-- ── Team (hidden — early access, not yet purchasable) ─────────────────────────
-- Replace prod_TEAM, price_TEAM_MONTHLY with real IDs from Step 1.
INSERT INTO plans (id, status, config) VALUES (
  'team', 'hidden',
  '{
    "id": "team",
    "name": "Team",
    "tagline": "Org-wide agent security policies, audit logs, and fleet controls for engineering teams.",
    "status": "hidden",
    "audience": "team",
    "billingMode": "subscription",
    "interval": null,
    "stripe": {
      "productId": "prod_TEAM",
      "priceId": "price_TEAM_MONTHLY",
      "prices": {
        "monthly": "price_TEAM_MONTHLY"
      },
      "checkoutMode": "subscription"
    },
    "pricing": {
      "amountCents": 24900,
      "currency": "usd",
      "displayPrice": "$249/mo"
    },
    "limits": {
      "maxPayloadBytes": 52428800,
      "retainedHistoryDays": 365,
      "cliSessions": -1,
      "maxCustomRules": -1
    },
    "features": {
      "firewallEnabled": true,
      "customRules": true,
      "policySync": true,
      "usageDashboard": true,
      "advancedUsageDashboard": true,
      "cliSessionManagement": true,
      "exportData": true,
      "ssoSaml": false,
      "firewallGroupToggles": true,
      "securityEventHistory": true,
      "provenanceTracking": true,
      "sharedPolicies": true,
      "teamDashboard": true,
      "auditLogs": true,
      "agentInventory": true,
      "fleetControls": false
    },
    "telemetry": {
      "requiredUsageMetering": true,
      "optionalProductAnalyticsDefault": true
    }
  }'::jsonb
);
```

---

## Step 4 — Force-sync KV

After the SQL runs, bust the KV cache so Workers pick up the new plans immediately.

```bash
curl -X POST https://api.confire.dev/admin/plans/sync \
  -H "Authorization: Bearer $ADMIN_SECRET"
```

Or via wrangler if you prefer:
```bash
wrangler kv key delete plans_v1 --namespace-id <YOUR_KV_ID>
# Workers will self-bootstrap from Supabase on next request.
```

---

## Step 5 — Verify

```bash
# Plans in KV match what you inserted
curl https://api.confire.dev/health

# Spot-check dev plan
curl https://api.confire.dev/admin/plans/sync \
  -H "Authorization: Bearer $ADMIN_SECRET" | jq '.dev.pricing'
# → { "amountCents": 1200, "displayPrice": "$12/mo", ... }

# Stripe: confirm prices exist
stripe prices list --product=prod_DEV
stripe prices list --product=prod_TEAM
```

---

## Limits reference

| Field | Free | Dev | Team |
|---|---|---|---|
| `maxPayloadBytes` | 3 MB | 10 MB | 50 MB |
| `retainedHistoryDays` | 7 | 90 | 365 |
| `cliSessions` | 3 | 10 | unlimited (−1) |
| `maxCustomRules` | 0 | 100 | unlimited (−1) |

---

## Pricing page display values

Annual toggle pre-selected. Show Team first for anchoring.

| | Team | Dev | Free |
|---|---|---|---|
| Price | $249/mo | $8.25/mo | $0 |
| Subtext | up to 10 seats | billed $99/yr | no account needed |
| CTA | Join waitlist | Get early access | Start for free |

---

## When to flip Team to active

Change `status` from `hidden` → `active` when:
- Shared policy sync ships (v1.3)
- Team dashboard is live
- At least one design partner on the plan

```sql
UPDATE plans
SET config = jsonb_set(config, '{status}', '"active"'),
    status = 'active'
WHERE id = 'team';
```

KV syncs automatically via DB webhook within ~1 second.

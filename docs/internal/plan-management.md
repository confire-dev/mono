# Plan Management

Plans are stored in the `plans` table in Supabase.
Cloudflare Workers read from KV (~1ms). KV is updated automatically via DB webhook when you change a plan.

## How it works

```
You edit a plan in Supabase
    ↓ (Supabase DB webhook, <1s)
POST /webhooks/supabase → Worker
    ↓
Worker re-syncs all plans → KV
    ↓
All Workers see the new plan on next request
```

On first Worker request ever (KV empty): Worker bootstraps itself by loading from Supabase → KV. No manual step needed.

---

## Update a plan limit

Go to **Supabase Dashboard → SQL Editor** and run:

```sql
-- Update a single limit (e.g. free plan monthly optimizations)
UPDATE plans
SET config = jsonb_set(config, '{limits,cloudOptimizationsMonthly}', '1000')
WHERE id = 'free';

-- Update multiple limits at once
UPDATE plans
SET config = config || '{
  "limits": {
    "cloudOptimizationsMonthly": 1000,
    "cloudTokensMonthly": 10000000,
    "maxPayloadBytes": 3000000
  }
}'::jsonb
WHERE id = 'free';

-- Update a feature flag
UPDATE plans
SET config = jsonb_set(config, '{features,exportData}', 'true')
WHERE id = 'dev';

-- Update plan price (for display only — Stripe is the source of truth for billing)
UPDATE plans
SET config = jsonb_set(config, '{pricing,displayPrice}', '"$7/mo"')
WHERE id = 'dev';
```

The DB webhook fires automatically. Workers pick up the change within ~1 second.

---

## Check what's currently in KV

```bash
# See the plans currently cached in KV
wrangler kv key get plans_v1 --namespace-id <YOUR_KV_ID>
```

Or call the health endpoint:
```bash
curl https://api.confire.dev/health
```

---

## Force-sync plans (if webhook failed)

```bash
curl -X POST https://api.confire.dev/admin/plans/sync \
  -H "Authorization: Bearer $ADMIN_SECRET"
```

This directly reads Supabase → writes KV. Use when:
- Webhook is misconfigured
- You made many rapid changes and want to confirm KV is consistent
- After initial deployment (KV is empty)

---

## Add a new plan

1. Insert a row in the `plans` table:

```sql
INSERT INTO plans (id, status, config) VALUES (
  'pro_yearly',
  'active',
  '{
    "id": "pro_yearly",
    "name": "Pro (Annual)",
    "tagline": "Pro features, billed annually.",
    "status": "active",
    "audience": "individual",
    "billingMode": "subscription",
    "interval": "year",
    "stripe": {
      "productId": "prod_xxx",
      "priceId": "price_xxx",
      "checkoutMode": "subscription"
    },
    "pricing": {
      "amountCents": 8100,
      "currency": "usd",
      "displayPrice": "$81/yr"
    },
    "limits": {
      "cloudOptimizationsMonthly": 5000,
      "cloudTokensMonthly": 50000000,
      "maxRawTokensPerOptimization": 750000,
      "maxPayloadBytes": 10000000,
      "retainedHistoryDays": 90,
      "cliSessions": 10
    },
    "credits": {
      "includedMonthly": 5000,
      "rollover": false,
      "allowManualGrants": true,
      "allowPurchases": true
    },
    "features": {
      "localOptimization": true,
      "remoteOptimization": true,
      "usageDashboard": true,
      "advancedUsageDashboard": true,
      "cliSessionManagement": true,
      "payloadCapture": false,
      "exportData": true,
      "priorityOptimizerUpdates": true,
      "customOptimizers": false,
      "ssoSaml": false
    },
    "optimizers": {
      "local": ["generic","bash","read","webfetch"],
      "remote": ["figma","github_pr","jira","confluence","clickup","slack","notion","amplitude","fireflies","playwright","zapier","google_drive"]
    },
    "telemetry": {
      "requiredUsageMetering": true,
      "optionalProductAnalyticsDefault": true
    }
  }'::jsonb
);
```

2. DB webhook fires → KV updated → Workers serve the new plan.

3. If you want users to see it in the pricing page, add it to the platform's pricing page component.

---

## Deprecate a plan

```sql
-- Hide from public (existing users keep it, new signups can't select it)
UPDATE plans SET status = 'hidden' WHERE id = 'dev';

-- Fully deprecated (show migration notice in dashboard)
UPDATE plans SET status = 'deprecated' WHERE id = 'dev';
```

Existing users with `plan_id = 'dev'` keep working — `getPlan()` still returns the plan object, it just won't appear in pricing UI.

---

## Plan config schema reference

Every plan's `config` column is a JSONB object matching this shape:

```typescript
{
  id: string                    // 'free' | 'dev' | 'pro' | 'enterprise'
  name: string                  // display name
  tagline: string               // one-line description for pricing page
  status: 'active' | 'hidden' | 'deprecated'
  audience: 'individual' | 'team' | 'enterprise'
  billingMode: 'free' | 'subscription' | 'enterprise'
  interval: 'month' | 'year' | null
  stripe: {
    productId: string | null    // Stripe product ID (e.g. "prod_xxx")
    priceId:   string | null    // Stripe price ID (e.g. "price_xxx")
    checkoutMode: 'subscription' | 'payment' | null
  }
  pricing: {
    amountCents: number         // 0 = free, 900 = $9, 8100 = $81/yr
    currency: 'usd'
    displayPrice: string        // shown in UI: "$0", "$9/mo", "Contact us"
  }
  limits: {
    cloudOptimizationsMonthly: number   // remote optimizer calls per month
    cloudTokensMonthly: number          // raw tokens processed remotely per month
    maxRawTokensPerOptimization: number // per-call limit (abuse protection)
    maxPayloadBytes: number             // per-call bytes limit
    retainedHistoryDays: number         // usage_periods / tool_call_summaries TTL
    cliSessions: number                 // max concurrent CLI sessions
  }
  credits: {
    includedMonthly: number     // credits granted each period
    rollover: boolean           // carry unused credits forward?
    allowManualGrants: boolean  // can founder grant bonus credits?
    allowPurchases: boolean     // can user buy extra credits?
  }
  features: {
    localOptimization: boolean
    remoteOptimization: boolean
    usageDashboard: boolean
    advancedUsageDashboard: boolean
    cliSessionManagement: boolean
    payloadCapture: boolean
    exportData: boolean
    priorityOptimizerUpdates: boolean
    customOptimizers: boolean       // enterprise only
    ssoSaml: boolean                // enterprise only
    optimizationHistory: boolean    // dev+
    earlyAccessAdapters: boolean    // dev+
    sessionMemoryGuard: boolean     // pro+ (not yet implemented — feature flag only)
    preCompactOptimizer: boolean    // pro+ (not yet implemented — feature flag only)
    localMemoryPacks: boolean       // pro+ (not yet implemented — feature flag only)
  }
  optimizers: {
    local: string[]   // local optimizer names (unlimited, never counted)
    remote: string[]  // remote optimizer names (counted against limits)
  }
  telemetry: {
    requiredUsageMetering: boolean            // always true
    optionalProductAnalyticsDefault: boolean  // Amplitude default for new users
  }
}
```

---

## Rate limits

Per-minute rate limits are enforced by the Cloudflare Workers Rate Limiting API.
They are **not** stored in the `plans` table — they live in `wrangler.toml` bindings
and are fixed at deploy time.

| Plan tier          | Requests per minute | Binding   |
|--------------------|---------------------|-----------|
| free               | 20                  | `RL_FREE` |
| dev / dev_annual   | 60                  | `RL_DEV`  |
| pro / pro_annual / enterprise | 200    | `RL_PRO`  |

To change rate limits, update `wrangler.toml` and redeploy. No Supabase change needed.

To create the Cloudflare rate limit namespaces for the first time:
```bash
wrangler rate-limit create rl-free   # note the namespace_id, put in wrangler.toml
wrangler rate-limit create rl-dev
wrangler rate-limit create rl-pro
```

See `docs/rate-limits.md` for the user-facing explanation.

---

## Set up the Supabase webhook (first time)

1. Go to **Supabase Dashboard → Database → Webhooks → Create a new hook**
2. Fill in:
   - **Name**: `sync-plans-to-kv`
   - **Table**: `plans`
   - **Events**: ✓ INSERT  ✓ UPDATE  ✓ DELETE
   - **Type**: HTTP Request
   - **URL**: `https://api.confire.dev/webhooks/supabase`
   - **HTTP Headers**: `Authorization: Bearer <SUPABASE_WEBHOOK_SECRET>`
3. Set the secret: `wrangler secret put SUPABASE_WEBHOOK_SECRET`
   (Use the same value in both places.)
4. Save the webhook.
5. Test: update any plan row → check Worker logs (`wrangler tail`) for the sync confirmation.

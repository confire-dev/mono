# Billing Operations

Stripe is the source of truth for payments. Supabase stores the product-facing billing state.
Manual edge cases are handled via Stripe Dashboard + Supabase Dashboard. No custom admin UI in v1.

---

## Grant bonus credits to a user

For support cases, beta testers, refund alternatives:

```sql
-- Grant 500 bonus credits to a specific user
SELECT grant_credits(
  p_user_id  := '<user_uuid>',
  p_bonus    := 500,
  p_source   := 'manual'
);

-- Grant with a note (for audit trail)
INSERT INTO credit_ledger (user_id, event_type, amount, source, metadata)
VALUES (
  '<user_uuid>',
  'adjustment',
  500,
  'manual',
  '{"reason": "support credit, ticket #123", "granted_by": "efe@efebehar.dev"}'::jsonb
);

UPDATE credit_balances
SET bonus_credits = bonus_credits + 500, updated_at = now()
WHERE user_id = '<user_uuid>';
```

---

## Manually change a user's plan

When a user's Stripe subscription doesn't sync correctly, or you're manually upgrading a beta tester:

```sql
-- Find the user
SELECT id, email, plan, subscription_status
FROM profiles
WHERE email = 'user@example.com';

-- Change plan
UPDATE profiles
SET
  plan    = 'pro',
  plan_id = 'pro',
  subscription_status = 'active',
  updated_at = now()
WHERE email = 'user@example.com';
```

> ⚠️ If upgrading, also grant the included credits for the new plan:
```sql
SELECT grant_credits(
  p_user_id  := '<user_uuid>',
  p_included := 5000,  -- pro plan limit
  p_source   := 'manual'
);
```

---

## Check a user's current usage

```sql
-- Current period usage
SELECT
  up.plan_id,
  up.period_start,
  up.cloud_optimizations_used,
  up.cloud_tokens_used,
  up.saved_tokens,
  p.config->'limits'->>'cloudOptimizationsMonthly' AS plan_limit
FROM usage_periods up
JOIN plans p ON p.id = up.plan_id
WHERE up.user_id = '<user_uuid>'
  AND up.period_end IS NULL
ORDER BY up.period_start DESC
LIMIT 1;

-- Credit balance
SELECT included_credits, bonus_credits, purchased_credits, total_credits
FROM credit_balances
WHERE user_id = '<user_uuid>';

-- Last 20 tool calls
SELECT tool_type, raw_bytes, optimized_bytes, reduction_ratio, created_at
FROM tool_call_summaries
WHERE user_id = '<user_uuid>'
ORDER BY created_at DESC
LIMIT 20;
```

---

## Stripe webhook events handled

| Event | What the Worker does |
|---|---|
| `customer.subscription.created` | Updates profile billing state, grants included credits |
| `customer.subscription.updated` | Updates status, period end, credits on renewal |
| `customer.subscription.deleted` | Sets `plan = 'free'`, `status = 'canceled'` |
| `invoice.paid` | Audit log |
| `invoice.payment_failed` | Audit log (subscription.updated with `past_due` follows) |

Events NOT handled (manual via Stripe Dashboard):
- Refunds
- Disputes
- Failed payment dunning
- Subscription pauses

---

## Reset a user's monthly usage

If there's been a billing error and you need to reset their counter:

```sql
-- Close the current period early and let the next request create a fresh one
UPDATE usage_periods
SET period_end = now()
WHERE user_id = '<user_uuid>'
  AND period_end IS NULL;

-- Optionally: restore their credit balance to plan limit
UPDATE credit_balances
SET included_credits = 500,  -- or the plan's includedMonthly
    updated_at       = now()
WHERE user_id = '<user_uuid>';
```

---

## Credit ledger audit trail

Every credit change is logged in `credit_ledger`:

```sql
SELECT
  event_type,
  amount,
  source,
  stripe_event_id,
  metadata,
  created_at
FROM credit_ledger
WHERE user_id = '<user_uuid>'
ORDER BY created_at DESC
LIMIT 50;
```

---

## Find users approaching their limit (send nudge)

```sql
SELECT
  p.email,
  p.plan,
  up.cloud_optimizations_used,
  (plns.config->'limits'->>'cloudOptimizationsMonthly')::int AS plan_limit,
  ROUND(
    up.cloud_optimizations_used::numeric /
    (plns.config->'limits'->>'cloudOptimizationsMonthly')::int * 100
  ) AS pct_used
FROM usage_periods up
JOIN profiles p    ON p.id = up.user_id
JOIN plans plns    ON plns.id = up.plan_id
WHERE up.period_end IS NULL
  AND up.cloud_optimizations_used::numeric /
      (plns.config->'limits'->>'cloudOptimizationsMonthly')::int >= 0.8
ORDER BY pct_used DESC;
```

---

## Promotions / coupons

Create a promo code for a launch discount:

```sql
INSERT INTO promotions (
  code, description, discount_type, discount_value,
  stripe_coupon_id, max_redemptions, valid_until, first_period_only
) VALUES (
  'LAUNCH50',
  '50% off first month for launch users',
  'percent',
  50,
  'coup_xxx',    -- Stripe coupon ID (create in Stripe first)
  500,           -- max 500 redemptions
  '2025-08-01',  -- expires Aug 1
  true           -- first period only
);
```

Create a beta tester credit grant promo:

```sql
INSERT INTO promotions (
  code, description, discount_type, discount_value,
  max_redemptions, valid_until
) VALUES (
  'BETACREDITS',
  '1000 bonus credits for beta testers',
  'fixed_credits',
  1000,
  100,
  '2025-07-01'
);
```

Check redemptions:

```sql
SELECT
  pr.code,
  pr.redemption_count,
  pr.max_redemptions,
  u.email,
  r.redeemed_at
FROM promotion_redemptions r
JOIN promotions pr ON pr.id = r.promotion_id
JOIN auth.users u  ON u.id  = r.user_id
WHERE pr.code = 'LAUNCH50'
ORDER BY r.redeemed_at DESC;
```

-- Migration 003 — Billing v2: billing_items, stripe idempotency, interval support
-- Run in: Supabase Dashboard → SQL Editor
-- Apply to: dev project AND production project separately

-- ── 1. billing_items — top-up and future one-off products ────────────────────
-- Keeps non-subscription billing configs out of the plans table.

CREATE TABLE IF NOT EXISTS billing_items (
  id         text        PRIMARY KEY,
  type       text        NOT NULL CHECK (type IN ('topup')),
  config     jsonb       NOT NULL,
  active     boolean     NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TRIGGER billing_items_updated_at
  BEFORE UPDATE ON billing_items
  FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

ALTER TABLE billing_items ENABLE ROW LEVEL SECURITY;
CREATE POLICY "billing_items_public_read" ON billing_items FOR SELECT USING (active = true);

-- ── 2. stripe_webhook_events — dispatch-level idempotency ────────────────────
-- One row per processed Stripe event. Prevents double-processing on retries.

CREATE TABLE IF NOT EXISTS stripe_webhook_events (
  id           text        PRIMARY KEY,  -- Stripe event ID
  type         text        NOT NULL,
  processed_at timestamptz NOT NULL DEFAULT now()
);

-- ── 3. billing_interval on profiles ─────────────────────────────────────────

ALTER TABLE profiles
  ADD COLUMN IF NOT EXISTS billing_interval text
    CHECK (billing_interval IN ('monthly', 'annual'));

-- ── 4. Extra columns on credit_ledger ────────────────────────────────────────

ALTER TABLE credit_ledger
  ADD COLUMN IF NOT EXISTS idempotency_key          text,
  ADD COLUMN IF NOT EXISTS stripe_session_id        text,
  ADD COLUMN IF NOT EXISTS stripe_invoice_id        text,
  ADD COLUMN IF NOT EXISTS stripe_payment_intent_id text;

CREATE UNIQUE INDEX IF NOT EXISTS credit_ledger_idempotency_key_idx
  ON credit_ledger (idempotency_key)
  WHERE idempotency_key IS NOT NULL;

-- ── 5. Update dev plan — add stripe.prices + credits.annual ─────────────────

UPDATE plans
SET config =
  config
  || jsonb_build_object(
    'stripe', (config->'stripe')
      || jsonb_build_object(
           'productId', 'prod_UdYngLggCJUx5B',
           'priceId',   'price_1TeHfTCjsgbilsT2kKli4wU5',
           'prices',    jsonb_build_object(
             'monthly', 'price_1TeHfTCjsgbilsT2kKli4wU5',
             'annual',  'price_1TeHggCjsgbilsT2a56012vt'
           )
         ),
    'credits', (config->'credits')
      || jsonb_build_object('annual', 60000)
  )
WHERE id = 'dev';

UPDATE plans
SET config =
  config
  || jsonb_build_object(
    'stripe', (config->'stripe')
      || jsonb_build_object(
           'productId', 'prod_UdYow0riWjmnFH',
           'priceId',   'price_1TeHggCjsgbilsT2a56012vt'
         )
  )
WHERE id = 'dev_annual';

-- ── 6. Seed top-up billing item ──────────────────────────────────────────────

INSERT INTO billing_items (id, type, config, active)
VALUES (
  'topup_remote_optimization_pack',
  'topup',
  '{
    "name": "Confire Credit Pack",
    "description": "One-time top-up for additional remote optimization credits.",
    "creditsPerUnit": 5000,
    "maxQuantity": 20,
    "creditType": "remote_optimization",
    "stripe": {
      "productId": "prod_UdYvfOnfvYjr9H",
      "priceId":   "price_1TeHo1CjsgbilsT2qq4wmI2W"
    }
  }'::jsonb,
  true
)
ON CONFLICT (id) DO UPDATE SET
  config     = EXCLUDED.config,
  active     = EXCLUDED.active,
  updated_at = now();

-- ── 7. Update seed trigger to support both credit field names ────────────────

CREATE OR REPLACE FUNCTION seed_free_plan_credits()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
DECLARE
  v_credits integer;
BEGIN
  SELECT COALESCE(
    (config->'credits'->>'includedMonthly')::integer,
    (config->'credits'->>'monthly')::integer,
    500
  )
  INTO v_credits
  FROM plans WHERE id = 'free';

  INSERT INTO credit_balances (user_id, included_credits)
  VALUES (NEW.id, COALESCE(v_credits, 500))
  ON CONFLICT (user_id) DO NOTHING;

  INSERT INTO credit_ledger (user_id, event_type, amount, source)
  VALUES (NEW.id, 'grant', COALESCE(v_credits, 500), 'system')
  ON CONFLICT DO NOTHING;

  RETURN NEW;
END;
$$;

-- ── Verify ───────────────────────────────────────────────────────────────────

SELECT
  id,
  config->'stripe'->>'productId'        AS product_id,
  config->'stripe'->>'priceId'          AS price_id,
  config->'stripe'->'prices'->>'monthly' AS price_monthly,
  config->'stripe'->'prices'->>'annual'  AS price_annual,
  config->'credits'->>'includedMonthly'  AS credits_monthly,
  config->'credits'->>'annual'           AS credits_annual
FROM plans
WHERE id IN ('dev', 'dev_annual');

SELECT id, type, config->'stripe'->>'priceId' AS stripe_price_id
FROM billing_items;

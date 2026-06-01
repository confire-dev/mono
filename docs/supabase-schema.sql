-- ============================================================
-- Confire v1 — Supabase Schema
-- Author: Efe <efe@efebehar.dev>
--
-- Principles:
--   • Supabase = source of truth for everything user-visible,
--     billable, auditable, or plan-gated.
--   • Stripe = source of truth for payments, invoices, renewals.
--   • Amplitude = behavioral analytics only (never used for
--     billing decisions or the user's dashboard).
--   • Individual billing only in v1. No orgs/teams/seats yet.
--     owner_type/owner_id pattern reserved for v2 if needed.
--
-- Apply in Supabase dashboard: SQL Editor → New Query → Run.
-- ============================================================


-- ── Extensions ───────────────────────────────────────────────────────────────
CREATE EXTENSION IF NOT EXISTS "pgcrypto";  -- gen_random_uuid()


-- ── profiles ─────────────────────────────────────────────────────────────────
-- One row per registered user. Extends auth.users with product and billing state.
-- Stripe is the source of truth for payment state; this table stores the
-- product-side view needed to decide what the user can access.

CREATE TABLE profiles (
  id                              uuid PRIMARY KEY REFERENCES auth.users(id) ON DELETE CASCADE,
  email                           text NOT NULL,
  name                            text,

  -- Stripe billing (Stripe is authoritative — update via webhooks only)
  stripe_customer_id              text UNIQUE,
  stripe_subscription_id          text UNIQUE,
  plan                            text NOT NULL DEFAULT 'free'
                                    CHECK (plan IN ('free', 'dev', 'dev_annual', 'pro', 'pro_annual', 'enterprise')),
  subscription_status             text NOT NULL DEFAULT 'none'
                                    CHECK (subscription_status IN (
                                      'none', 'trialing', 'active', 'past_due', 'canceled', 'incomplete'
                                    )),
  subscription_current_period_end timestamptz,

  -- Soft delete / ban flag (manual ops via Supabase dashboard)
  is_banned                       boolean NOT NULL DEFAULT false,

  created_at  timestamptz NOT NULL DEFAULT now(),
  updated_at  timestamptz NOT NULL DEFAULT now()
);

-- Keep updated_at current automatically.
CREATE OR REPLACE FUNCTION touch_updated_at()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN NEW.updated_at = now(); RETURN NEW; END;
$$;

CREATE TRIGGER profiles_updated_at
  BEFORE UPDATE ON profiles
  FOR EACH ROW EXECUTE FUNCTION touch_updated_at();


-- ── credit_balances ───────────────────────────────────────────────────────────
-- Current usable credit state per user. Derived from credit_ledger (events).
-- Updated by: subscription webhooks (included), manual grants (bonus),
-- one-off purchases (purchased). Decremented by usage.
-- Do NOT write to this table directly — use credit_ledger + update triggers.

CREATE TABLE credit_balances (
  user_id           uuid PRIMARY KEY REFERENCES auth.users(id) ON DELETE CASCADE,
  included_credits  integer NOT NULL DEFAULT 0 CHECK (included_credits >= 0),
  bonus_credits     integer NOT NULL DEFAULT 0 CHECK (bonus_credits >= 0),
  purchased_credits integer NOT NULL DEFAULT 0 CHECK (purchased_credits >= 0),
  updated_at        timestamptz NOT NULL DEFAULT now()
);

-- Computed helper: total available credits.
ALTER TABLE credit_balances ADD COLUMN IF NOT EXISTS
  total_credits integer GENERATED ALWAYS AS
    (included_credits + bonus_credits + purchased_credits) STORED;


-- ── credit_ledger ─────────────────────────────────────────────────────────────
-- Immutable append-only audit trail of every credit change.
-- amount > 0 = credits added (grant, purchase, adjustment+)
-- amount < 0 = credits consumed (usage, adjustment-, expiration)

CREATE TABLE credit_ledger (
  id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id         uuid NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
  event_type      text NOT NULL
                    CHECK (event_type IN (
                      'grant',        -- subscription included credits
                      'usage',        -- tool call optimized
                      'adjustment',   -- manual correction
                      'purchase',     -- one-off credit purchase
                      'expiration'    -- credits expired at period end
                    )),
  amount          integer NOT NULL,   -- +N or -N
  source          text NOT NULL
                    CHECK (source IN ('subscription', 'manual', 'stripe', 'system')),
  stripe_event_id text,               -- idempotency key for Stripe webhooks
  metadata        jsonb NOT NULL DEFAULT '{}',
  created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX credit_ledger_user_id_idx ON credit_ledger (user_id, created_at DESC);
CREATE UNIQUE INDEX credit_ledger_stripe_event_idx ON credit_ledger (stripe_event_id)
  WHERE stripe_event_id IS NOT NULL;


-- ── promotions ────────────────────────────────────────────────────────────────
-- Coupon / promotion system. Supports:
--   • Percent discount off subscription price   (e.g. "99% off first month")
--   • Fixed bonus credits on sign-up             (e.g. "500 extra credits")
--   • Trial extension                            (e.g. "30 days free trial")
-- Works with Stripe coupons for pricing discounts;
-- internal-only for credit grants.

CREATE TABLE promotions (
  id                 uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  code               text UNIQUE NOT NULL,               -- human code, e.g. "LAUNCH50"
  description        text,
  discount_type      text NOT NULL
                       CHECK (discount_type IN (
                         'percent',        -- off Stripe subscription price
                         'fixed_credits',  -- one-time bonus credit grant
                         'trial_days'      -- extend trial by N days
                       )),
  discount_value     integer NOT NULL CHECK (discount_value > 0),
                     -- percent: 1-100 | fixed_credits: N | trial_days: N
  stripe_coupon_id   text,                               -- linked Stripe coupon (if any)
  max_redemptions    integer,                            -- null = unlimited
  redemption_count   integer NOT NULL DEFAULT 0,
  valid_from         timestamptz NOT NULL DEFAULT now(),
  valid_until        timestamptz,                        -- null = no expiry
  plans_applicable   text[] NOT NULL DEFAULT ARRAY['free', 'pro'],
  first_period_only  boolean NOT NULL DEFAULT true,      -- for "0.99 first month" promos
  metadata           jsonb NOT NULL DEFAULT '{}',
  created_at         timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE promotion_redemptions (
  id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id         uuid NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
  promotion_id    uuid NOT NULL REFERENCES promotions(id),
  stripe_event_id text,
  redeemed_at     timestamptz NOT NULL DEFAULT now(),
  UNIQUE (user_id, promotion_id)
);


-- ── api_keys ──────────────────────────────────────────────────────────────────
-- CLI authentication. One key per device typically; a user can have multiple.
-- key_hash = SHA-256(raw_key). Never store the raw key.

CREATE TABLE api_keys (
  id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id      uuid NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
  key_hash     text UNIQUE NOT NULL,
  key_prefix   text NOT NULL,                           -- first 12 chars, for display
  device_id    text,                                    -- from client's keychain
  last_used_at timestamptz,
  revoked_at   timestamptz,                             -- non-null = revoked
  created_at   timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX api_keys_user_id_idx ON api_keys (user_id) WHERE revoked_at IS NULL;


-- ── cli_sessions ──────────────────────────────────────────────────────────────
-- One row per Claude Code session (SessionStart → SessionEnd).
-- session_id comes from Claude Code's hook payload (uuid).

CREATE TABLE cli_sessions (
  id                text PRIMARY KEY,                   -- Claude Code session_id
  user_id           uuid NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
  device_id         text,
  cli_version       text,
  integration       text NOT NULL DEFAULT 'claude_code',
  started_at        timestamptz NOT NULL DEFAULT now(),
  ended_at          timestamptz,
  total_tool_calls  integer NOT NULL DEFAULT 0,
  optimized_calls   integer NOT NULL DEFAULT 0,
  raw_bytes         bigint NOT NULL DEFAULT 0,
  optimized_bytes   bigint NOT NULL DEFAULT 0
);

CREATE INDEX cli_sessions_user_id_idx ON cli_sessions (user_id, started_at DESC);


-- ── tool_call_summaries ───────────────────────────────────────────────────────
-- Per-call detail record written by the Worker on every optimization.
-- Powers the user dashboard: savings today/month, top tools, credits used.
-- NOTE: raw content is never stored here — only metadata/metrics.

CREATE TABLE tool_call_summaries (
  id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id         uuid NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
  cli_session_id  text REFERENCES cli_sessions(id) ON DELETE SET NULL,
  tool_type       text NOT NULL,                        -- "bash", "figma", "github_pr", etc.
  integration     text NOT NULL DEFAULT 'claude_code',
  optimizer       text,                                 -- which optimizer handled it
  raw_bytes       integer NOT NULL DEFAULT 0,
  optimized_bytes integer NOT NULL DEFAULT 0,
  saved_bytes     integer GENERATED ALWAYS AS (raw_bytes - optimized_bytes) STORED,
  reduction_ratio numeric(5,4) GENERATED ALWAYS AS (
    CASE WHEN raw_bytes > 0
    THEN ROUND((raw_bytes - optimized_bytes)::numeric / raw_bytes, 4)
    ELSE 0 END
  ) STORED,
  duration_ms     integer,
  was_cached      boolean NOT NULL DEFAULT false,
  credits_used    integer NOT NULL DEFAULT 1,           -- credits charged for this call
  created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX tool_call_summaries_user_month_idx
  ON tool_call_summaries (user_id, created_at DESC);


-- ── monthly_usage ─────────────────────────────────────────────────────────────
-- Fast rollup table for billing checks and the dashboard.
-- Updated atomically by the increment_usage() function.

CREATE TABLE monthly_usage (
  user_id       uuid REFERENCES auth.users(id) ON DELETE CASCADE,
  month         text NOT NULL,                          -- "2025-05"
  request_count integer NOT NULL DEFAULT 0,
  saved_bytes   bigint NOT NULL DEFAULT 0,
  CONSTRAINT pk_monthly_usage PRIMARY KEY (user_id, month)
);


-- ── audit_events ──────────────────────────────────────────────────────────────
-- Security-sensitive actions: API key generation, revocation, login, device
-- changes, billing state transitions. For manual review via Supabase dashboard.

CREATE TABLE audit_events (
  id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     uuid REFERENCES auth.users(id) ON DELETE SET NULL,
  event_type  text NOT NULL,
  device_id   text,
  ip_address  inet,
  metadata    jsonb NOT NULL DEFAULT '{}',
  created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX audit_events_user_id_idx ON audit_events (user_id, created_at DESC);


-- ── Stored procedures ─────────────────────────────────────────────────────────

-- increment_usage: atomic upsert for monthly usage rollup.
-- Called by the Worker on every successful optimization.
CREATE OR REPLACE FUNCTION increment_usage(
  p_user_id     uuid,
  p_month       text,
  p_bytes_saved bigint DEFAULT 0
) RETURNS void LANGUAGE plpgsql AS $$
BEGIN
  INSERT INTO monthly_usage (user_id, month, request_count, saved_bytes)
  VALUES (p_user_id, p_month, 1, p_bytes_saved)
  ON CONFLICT (user_id, month) DO UPDATE SET
    request_count = monthly_usage.request_count + 1,
    saved_bytes   = monthly_usage.saved_bytes + p_bytes_saved;
END;
$$;

-- consume_credits: decrement credits and write a ledger entry atomically.
-- Returns true if credits were available and consumed, false if insufficient.
CREATE OR REPLACE FUNCTION consume_credits(
  p_user_id    uuid,
  p_amount     integer DEFAULT 1,
  p_session_id text    DEFAULT NULL
) RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE
  v_total integer;
BEGIN
  SELECT total_credits INTO v_total
  FROM credit_balances
  WHERE user_id = p_user_id
  FOR UPDATE;

  IF v_total IS NULL OR v_total < p_amount THEN
    RETURN false;
  END IF;

  -- Deduct from included first, then bonus, then purchased.
  UPDATE credit_balances SET
    included_credits = GREATEST(0, included_credits - p_amount),
    bonus_credits    = GREATEST(0,
      bonus_credits - GREATEST(0, p_amount - included_credits)),
    purchased_credits = GREATEST(0,
      purchased_credits - GREATEST(0,
        p_amount - included_credits - bonus_credits)),
    updated_at = now()
  WHERE user_id = p_user_id;

  INSERT INTO credit_ledger (user_id, event_type, amount, source, metadata)
  VALUES (p_user_id, 'usage', -p_amount, 'system',
    jsonb_build_object('session_id', p_session_id));

  RETURN true;
END;
$$;

-- consume_purchased_credits: deduct from the purchased top-up pool only.
-- Used when monthly plan usage is exhausted but purchased credits remain.
CREATE OR REPLACE FUNCTION consume_purchased_credits(
  p_user_id    uuid,
  p_amount     integer DEFAULT 1,
  p_session_id text    DEFAULT NULL
) RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE
  v_purchased integer;
BEGIN
  SELECT purchased_credits INTO v_purchased
  FROM credit_balances
  WHERE user_id = p_user_id
  FOR UPDATE;

  IF v_purchased IS NULL OR v_purchased < p_amount THEN
    RETURN false;
  END IF;

  UPDATE credit_balances SET
    purchased_credits = purchased_credits - p_amount,
    updated_at = now()
  WHERE user_id = p_user_id;

  INSERT INTO credit_ledger (user_id, event_type, amount, source, metadata)
  VALUES (p_user_id, 'usage', -p_amount, 'stripe_topup',
    jsonb_build_object('session_id', p_session_id, 'pool', 'purchased'));

  RETURN true;
END;
$$;

-- grant_credits: add credits (from subscription refresh, manual grant, purchase).
CREATE OR REPLACE FUNCTION grant_credits(
  p_user_id        uuid,
  p_included        integer DEFAULT 0,
  p_bonus           integer DEFAULT 0,
  p_purchased       integer DEFAULT 0,
  p_source          text    DEFAULT 'system',
  p_stripe_event_id text    DEFAULT NULL
) RETURNS void LANGUAGE plpgsql AS $$
BEGIN
  -- Idempotency: skip if we've already processed this Stripe event.
  IF p_stripe_event_id IS NOT NULL AND
     EXISTS (SELECT 1 FROM credit_ledger WHERE stripe_event_id = p_stripe_event_id) THEN
    RETURN;
  END IF;

  INSERT INTO credit_balances (user_id, included_credits, bonus_credits, purchased_credits)
  VALUES (p_user_id, p_included, p_bonus, p_purchased)
  ON CONFLICT (user_id) DO UPDATE SET
    included_credits  = credit_balances.included_credits + p_included,
    bonus_credits     = credit_balances.bonus_credits + p_bonus,
    purchased_credits = credit_balances.purchased_credits + p_purchased,
    updated_at        = now();

  IF (p_included + p_bonus + p_purchased) > 0 THEN
    INSERT INTO credit_ledger
      (user_id, event_type, amount, source, stripe_event_id)
    VALUES
      (p_user_id, 'grant', p_included + p_bonus + p_purchased,
       p_source, p_stripe_event_id);
  END IF;
END;
$$;


-- ── Auto-seed free credits on profile creation ───────────────────────────────
-- Fires on every INSERT into profiles (platform signup OR CLI first login).
-- Reads the free plan's includedMonthly from the plans table so no hardcoding.
-- ON CONFLICT DO NOTHING makes it safe to call multiple times.

CREATE OR REPLACE FUNCTION seed_free_plan_credits()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
DECLARE
  v_credits integer;
BEGIN
  SELECT COALESCE((config->'credits'->>'includedMonthly')::integer, 500)
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

CREATE TRIGGER profiles_seed_free_credits
  AFTER INSERT ON profiles
  FOR EACH ROW EXECUTE FUNCTION seed_free_plan_credits();


-- ── Row-Level Security ────────────────────────────────────────────────────────
-- Users can only read their own rows.
-- The Worker uses the SERVICE ROLE KEY which bypasses RLS entirely.

ALTER TABLE profiles              ENABLE ROW LEVEL SECURITY;
ALTER TABLE credit_balances       ENABLE ROW LEVEL SECURITY;
ALTER TABLE credit_ledger         ENABLE ROW LEVEL SECURITY;
ALTER TABLE api_keys              ENABLE ROW LEVEL SECURITY;
ALTER TABLE cli_sessions          ENABLE ROW LEVEL SECURITY;
ALTER TABLE tool_call_summaries   ENABLE ROW LEVEL SECURITY;
ALTER TABLE monthly_usage         ENABLE ROW LEVEL SECURITY;
ALTER TABLE promotion_redemptions ENABLE ROW LEVEL SECURITY;
-- promotions: public read so the CLI can validate promo codes
-- audit_events: no user-facing RLS (internal only)

CREATE POLICY "users_own_profile"      ON profiles              FOR ALL  USING (id = auth.uid());
CREATE POLICY "users_own_credits"      ON credit_balances       FOR ALL  USING (user_id = auth.uid());
CREATE POLICY "users_own_ledger"       ON credit_ledger         FOR SELECT USING (user_id = auth.uid());
CREATE POLICY "users_own_keys"         ON api_keys              FOR ALL  USING (user_id = auth.uid());
CREATE POLICY "users_own_sessions"     ON cli_sessions          FOR ALL  USING (user_id = auth.uid());
CREATE POLICY "users_own_summaries"    ON tool_call_summaries   FOR ALL  USING (user_id = auth.uid());
CREATE POLICY "users_own_usage"        ON monthly_usage         FOR ALL  USING (user_id = auth.uid());
CREATE POLICY "users_own_redemptions"  ON promotion_redemptions FOR ALL  USING (user_id = auth.uid());
CREATE POLICY "promos_public_read"     ON promotions            FOR SELECT USING (true);


-- ── Seed data: free plan limits ───────────────────────────────────────────────
-- Plan credit amounts are enforced in application logic (Worker).
-- Free: 500 credits/month. Pro: 10,000 credits/month.
-- Change these in Worker config without touching the schema.

COMMENT ON TABLE profiles IS
  'v1 individual billing only. Plan limits: free=500/month, pro=10000/month.';
COMMENT ON COLUMN credit_ledger.stripe_event_id IS
  'Stripe webhook event ID used as idempotency key. Unique when non-null.';
COMMENT ON TABLE tool_call_summaries IS
  'Raw content is never stored. Metrics only. Drives user dashboard.';


-- ── user_preferences ─────────────────────────────────────────────────────────
-- Per-user configurable preferences. Authoritative source: Supabase.
-- CLI reads these at session start; falls back to ~/.confire/config.json offline.
-- Synced via: confire config set <key>=<value>

CREATE TABLE user_preferences (
  user_id                 uuid PRIMARY KEY REFERENCES auth.users(id) ON DELETE CASCADE,

  -- Notification preferences (mirrors cli NotificationConfig)
  notifications_enabled   boolean     NOT NULL DEFAULT true,
  notification_style      text        NOT NULL DEFAULT 'brand'
                            CHECK (notification_style IN ('brand', 'minimal', 'off')),
  min_saved_tokens        integer     NOT NULL DEFAULT 5000,
  big_save_tokens         integer     NOT NULL DEFAULT 50000,

  -- Analytics preference
  -- false = skip Amplitude; usage accounting for billing/dashboard remains active.
  analytics_enabled       boolean     NOT NULL DEFAULT true,

  updated_at              timestamptz NOT NULL DEFAULT now()
);

CREATE TRIGGER user_preferences_updated_at
  BEFORE UPDATE ON user_preferences
  FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

ALTER TABLE user_preferences ENABLE ROW LEVEL SECURITY;
CREATE POLICY "users_own_preferences" ON user_preferences FOR ALL USING (user_id = auth.uid());

COMMENT ON TABLE user_preferences IS
  'Per-user CLI and product preferences. Synced from CLI via confire config set.';
COMMENT ON COLUMN user_preferences.analytics_enabled IS
  'Controls Amplitude forwarding only. Does NOT disable usage accounting (required for billing).';


-- ── Schema additions for v1 plan system ──────────────────────────────────────
-- Run this section AFTER the initial schema above.
-- Adds: usage_periods (replaces monthly_usage), updates profiles.

-- Add missing columns to profiles (if not already present from initial migration)
ALTER TABLE profiles
  ADD COLUMN IF NOT EXISTS plan_id                           text NOT NULL DEFAULT 'free',
  ADD COLUMN IF NOT EXISTS subscription_current_period_start timestamptz,
  ADD COLUMN IF NOT EXISTS cancel_at_period_end              boolean NOT NULL DEFAULT false;

-- Keep backward compat: plan_id mirrors plan column until migration complete
-- TODO: rename 'plan' → 'plan_id' and drop old column after apps are updated

-- ── usage_periods ──────────────────────────────────────────────────────────
-- Replaces monthly_usage. Tracks both optimization count AND token volume
-- so the plan system can gate on either dimension.
-- Period = one subscription cycle (monthly for now, could be annual later).

CREATE TABLE IF NOT EXISTS usage_periods (
  id                           uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id                      uuid        NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
  plan_id                      text        NOT NULL DEFAULT 'free',
  period_start                 timestamptz NOT NULL,
  period_end                   timestamptz,            -- null = currently active
  cloud_optimizations_used     integer     NOT NULL DEFAULT 0,
  cloud_tokens_used            bigint      NOT NULL DEFAULT 0,
  local_optimizations_count    integer     NOT NULL DEFAULT 0,
  saved_tokens                 bigint      NOT NULL DEFAULT 0,
  created_at                   timestamptz NOT NULL DEFAULT now(),
  updated_at                   timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS usage_periods_user_active_idx
  ON usage_periods (user_id, period_start DESC)
  WHERE (period_end IS NULL);

CREATE INDEX IF NOT EXISTS usage_periods_user_id_idx
  ON usage_periods (user_id, period_start DESC);

CREATE TRIGGER usage_periods_updated_at
  BEFORE UPDATE ON usage_periods
  FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

ALTER TABLE usage_periods ENABLE ROW LEVEL SECURITY;
CREATE POLICY "users_own_periods" ON usage_periods FOR ALL USING (user_id = auth.uid());

-- Add period_id FK to credit_ledger for audit trail
ALTER TABLE credit_ledger
  ADD COLUMN IF NOT EXISTS period_id uuid REFERENCES usage_periods(id);


-- ── Stored procedures for usage_periods ─────────────────────────────────────

-- get_or_create_active_period: finds the current period or creates one.
-- Returns the period UUID. Called by the Worker before each optimization.
CREATE OR REPLACE FUNCTION get_or_create_active_period(
  p_user_id uuid,
  p_plan_id text DEFAULT 'free'
) RETURNS uuid LANGUAGE plpgsql AS $$
DECLARE
  v_period_id    uuid;
  v_period_start timestamptz;
  v_sub_start    timestamptz;
  v_sub_end      timestamptz;
BEGIN
  -- Fetch Stripe's authoritative period boundaries (null for free/no-sub users).
  -- Annual subscribers get a 12-month window; monthly get 1-month.
  -- This is the single source of truth — we don't hardcode interval logic here.
  SELECT subscription_current_period_start, subscription_current_period_end
  INTO v_sub_start, v_sub_end
  FROM profiles
  WHERE id = p_user_id;

  -- Check for an open period (no end date = still active)
  SELECT id INTO v_period_id
  FROM usage_periods
  WHERE user_id = p_user_id
    AND period_end IS NULL
  ORDER BY period_start DESC
  LIMIT 1;

  -- Close the open period if it has expired
  IF v_period_id IS NOT NULL THEN
    DECLARE v_start timestamptz;
    BEGIN
      SELECT period_start INTO v_start FROM usage_periods WHERE id = v_period_id;

      -- Subscribers: expired when now() is past Stripe's period_end
      -- Free users: expired when we've rolled into a new calendar month
      IF (v_sub_end IS NOT NULL AND now() > v_sub_end)
         OR (v_sub_end IS NULL AND date_trunc('month', v_start) < date_trunc('month', now())) THEN
        UPDATE usage_periods
        SET period_end = COALESCE(v_sub_end, date_trunc('month', now()) - interval '1 microsecond')
        WHERE id = v_period_id;
        v_period_id := NULL;
      END IF;
    END;
  END IF;

  IF v_period_id IS NOT NULL THEN
    RETURN v_period_id;
  END IF;

  -- Determine new period start:
  --   Subscribers: use Stripe's current period start (exact billing date)
  --   Free users:  start of current calendar month
  IF v_sub_start IS NOT NULL AND v_sub_end IS NOT NULL AND now() <= v_sub_end THEN
    v_period_start := v_sub_start;
  ELSE
    v_period_start := date_trunc('month', now());
  END IF;

  INSERT INTO usage_periods (user_id, plan_id, period_start)
  VALUES (p_user_id, p_plan_id, v_period_start)
  RETURNING id INTO v_period_id;

  RETURN v_period_id;
END;
$$;


-- increment_usage_period: atomically increments the active period counters.
-- Called after each successful cloud optimization.
CREATE OR REPLACE FUNCTION increment_usage_period(
  p_user_id       uuid,
  p_plan_id       text    DEFAULT 'free',
  p_cloud_opts    integer DEFAULT 1,
  p_cloud_tokens  bigint  DEFAULT 0,
  p_saved_tokens  bigint  DEFAULT 0
) RETURNS void LANGUAGE plpgsql AS $$
DECLARE
  v_period_id uuid;
BEGIN
  v_period_id := get_or_create_active_period(p_user_id, p_plan_id);

  UPDATE usage_periods SET
    cloud_optimizations_used  = cloud_optimizations_used  + p_cloud_opts,
    cloud_tokens_used         = cloud_tokens_used         + p_cloud_tokens,
    saved_tokens              = saved_tokens              + p_saved_tokens,
    plan_id                   = p_plan_id   -- update plan if it changed
  WHERE id = v_period_id;
END;
$$;



-- ── plans table (DB is the source of truth for plan limits) ────────────────
-- Worker loads plans from this table at startup (cached in KV, 5-min TTL).
-- To update a limit: UPDATE plans SET config = config || '{"limits":{"cloudOptimizationsMonthly":1000}}' WHERE id = 'free';
-- No Worker redeploy needed.

CREATE TABLE IF NOT EXISTS plans (
  id          text        PRIMARY KEY,   -- 'free', 'dev', 'pro', 'enterprise'
  config      jsonb       NOT NULL,      -- full Plan object (matches plans.ts interface)
  status      text        NOT NULL DEFAULT 'active'
                CHECK (status IN ('active', 'hidden', 'deprecated')),
  created_at  timestamptz NOT NULL DEFAULT now(),
  updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE TRIGGER plans_updated_at
  BEFORE UPDATE ON plans
  FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

-- Public read (Worker needs this without auth, via anon key in some flows)
ALTER TABLE plans ENABLE ROW LEVEL SECURITY;
CREATE POLICY "plans_public_read" ON plans FOR SELECT USING (status = 'active');

-- ── Seed data — initial plan definitions ─────────────────────────────────────
-- These match the FALLBACK_* constants in worker/src/lib/plans.ts.
-- Edit these rows to change plan limits — no code changes needed.

INSERT INTO plans (id, status, config) VALUES

('free', 'active', '{
  "id": "free",
  "name": "Free",
  "tagline": "Local context cleanup for every developer.",
  "status": "active",
  "audience": "individual",
  "billingMode": "free",
  "interval": null,
  "stripe": { "productId": null, "priceId": null, "checkoutMode": null },
  "pricing": { "amountCents": 0, "currency": "usd", "displayPrice": "$0" },
  "limits": {
    "cloudOptimizationsMonthly": 500,
    "cloudTokensMonthly": 5000000,
    "maxRawTokensPerOptimization": 250000,
    "maxPayloadBytes": 2000000,
    "retainedHistoryDays": 30,
    "cliSessions": 3
  },
  "credits": { "includedMonthly": 500, "rollover": false, "allowManualGrants": true, "allowPurchases": false },
  "features": {
    "localOptimization": true, "remoteOptimization": true,
    "usageDashboard": true, "advancedUsageDashboard": false,
    "cliSessionManagement": true, "payloadCapture": false,
    "exportData": false, "priorityOptimizerUpdates": false,
    "customOptimizers": false, "ssoSaml": false,
    "optimizationHistory": false, "earlyAccessAdapters": false,
    "sessionMemoryGuard": false, "preCompactOptimizer": false, "localMemoryPacks": false
  },
  "optimizers": {
    "local": ["generic","bash","read","webfetch"],
    "remote": ["figma","github_pr","jira","confluence","clickup","slack","notion","amplitude","fireflies","playwright","zapier","google_drive"]
  },
  "telemetry": { "requiredUsageMetering": true, "optionalProductAnalyticsDefault": true }
}'),

('dev', 'active', '{
  "id": "dev",
  "name": "Dev",
  "tagline": "For developers who use AI agents daily.",
  "status": "active",
  "audience": "individual",
  "billingMode": "subscription",
  "interval": "month",
  "stripe": { "productId": null, "priceId": null, "checkoutMode": "subscription" },
  "pricing": { "amountCents": 1000, "currency": "usd", "displayPrice": "$10/mo" },
  "limits": {
    "cloudOptimizationsMonthly": 5000,
    "cloudTokensMonthly": 50000000,
    "maxRawTokensPerOptimization": 500000,
    "maxPayloadBytes": 5000000,
    "retainedHistoryDays": 60,
    "cliSessions": 5
  },
  "credits": { "includedMonthly": 5000, "rollover": false, "allowManualGrants": true, "allowPurchases": true },
  "features": {
    "localOptimization": true, "remoteOptimization": true,
    "usageDashboard": true, "advancedUsageDashboard": true,
    "cliSessionManagement": true, "payloadCapture": false,
    "exportData": false, "priorityOptimizerUpdates": false,
    "customOptimizers": false, "ssoSaml": false,
    "optimizationHistory": true, "earlyAccessAdapters": true,
    "sessionMemoryGuard": false, "preCompactOptimizer": false, "localMemoryPacks": false
  },
  "optimizers": {
    "local": ["generic","bash","read","webfetch"],
    "remote": ["figma","github_pr","jira","confluence","clickup","slack","notion","amplitude","fireflies","playwright","zapier","google_drive"]
  },
  "telemetry": { "requiredUsageMetering": true, "optionalProductAnalyticsDefault": true }
}'),

('pro', 'active', '{
  "id": "pro",
  "name": "Pro",
  "tagline": "For power users who optimize every session.",
  "status": "active",
  "audience": "individual",
  "billingMode": "subscription",
  "interval": "month",
  "stripe": { "productId": null, "priceId": null, "checkoutMode": "subscription" },
  "pricing": { "amountCents": 2000, "currency": "usd", "displayPrice": "$20/mo" },
  "limits": {
    "cloudOptimizationsMonthly": 999999999,
    "cloudTokensMonthly": 500000000,
    "maxRawTokensPerOptimization": 750000,
    "maxPayloadBytes": 10000000,
    "retainedHistoryDays": 90,
    "cliSessions": 10
  },
  "credits": { "includedMonthly": 999999, "rollover": false, "allowManualGrants": true, "allowPurchases": true },
  "features": {
    "localOptimization": true, "remoteOptimization": true,
    "usageDashboard": true, "advancedUsageDashboard": true,
    "cliSessionManagement": true, "payloadCapture": false,
    "exportData": true, "priorityOptimizerUpdates": true,
    "customOptimizers": false, "ssoSaml": false,
    "optimizationHistory": true, "earlyAccessAdapters": true,
    "sessionMemoryGuard": true, "preCompactOptimizer": true, "localMemoryPacks": true
  },
  "optimizers": {
    "local": ["generic","bash","read","webfetch"],
    "remote": ["figma","github_pr","jira","confluence","clickup","slack","notion","amplitude","fireflies","playwright","zapier","google_drive"]
  },
  "telemetry": { "requiredUsageMetering": true, "optionalProductAnalyticsDefault": true }
}'),

('dev_annual', 'active', '{
  "id": "dev_annual",
  "name": "Dev (Annual)",
  "tagline": "For developers who use AI agents daily.",
  "status": "active",
  "audience": "individual",
  "billingMode": "subscription",
  "interval": "year",
  "stripe": { "productId": null, "priceId": null, "checkoutMode": "subscription" },
  "pricing": { "amountCents": 9500, "currency": "usd", "displayPrice": "$95/yr" },
  "limits": {
    "cloudOptimizationsMonthly": 5000,
    "cloudTokensMonthly": 50000000,
    "maxRawTokensPerOptimization": 500000,
    "maxPayloadBytes": 5000000,
    "retainedHistoryDays": 60,
    "cliSessions": 5
  },
  "credits": { "includedMonthly": 5000, "rollover": false, "allowManualGrants": true, "allowPurchases": true },
  "features": {
    "localOptimization": true, "remoteOptimization": true,
    "usageDashboard": true, "advancedUsageDashboard": true,
    "cliSessionManagement": true, "payloadCapture": false,
    "exportData": false, "priorityOptimizerUpdates": false,
    "customOptimizers": false, "ssoSaml": false,
    "optimizationHistory": true, "earlyAccessAdapters": true,
    "sessionMemoryGuard": false, "preCompactOptimizer": false, "localMemoryPacks": false
  },
  "optimizers": {
    "local": ["generic","bash","read","webfetch"],
    "remote": ["figma","github_pr","jira","confluence","clickup","slack","notion","amplitude","fireflies","playwright","zapier","google_drive"]
  },
  "telemetry": { "requiredUsageMetering": true, "optionalProductAnalyticsDefault": true }
}'),

('pro_annual', 'active', '{
  "id": "pro_annual",
  "name": "Pro (Annual)",
  "tagline": "For power users who optimize every session.",
  "status": "active",
  "audience": "individual",
  "billingMode": "subscription",
  "interval": "year",
  "stripe": { "productId": null, "priceId": null, "checkoutMode": "subscription" },
  "pricing": { "amountCents": 19500, "currency": "usd", "displayPrice": "$195/yr" },
  "limits": {
    "cloudOptimizationsMonthly": 999999999,
    "cloudTokensMonthly": 500000000,
    "maxRawTokensPerOptimization": 750000,
    "maxPayloadBytes": 10000000,
    "retainedHistoryDays": 90,
    "cliSessions": 10
  },
  "credits": { "includedMonthly": 999999, "rollover": false, "allowManualGrants": true, "allowPurchases": true },
  "features": {
    "localOptimization": true, "remoteOptimization": true,
    "usageDashboard": true, "advancedUsageDashboard": true,
    "cliSessionManagement": true, "payloadCapture": false,
    "exportData": true, "priorityOptimizerUpdates": true,
    "customOptimizers": false, "ssoSaml": false,
    "optimizationHistory": true, "earlyAccessAdapters": true,
    "sessionMemoryGuard": true, "preCompactOptimizer": true, "localMemoryPacks": true
  },
  "optimizers": {
    "local": ["generic","bash","read","webfetch"],
    "remote": ["figma","github_pr","jira","confluence","clickup","slack","notion","amplitude","fireflies","playwright","zapier","google_drive"]
  },
  "telemetry": { "requiredUsageMetering": true, "optionalProductAnalyticsDefault": true }
}'),

('enterprise', 'hidden', '{
  "id": "enterprise",
  "name": "Enterprise",
  "tagline": "Custom limits, SSO, dedicated support.",
  "status": "hidden",
  "audience": "enterprise",
  "billingMode": "enterprise",
  "interval": null,
  "stripe": { "productId": null, "priceId": null, "checkoutMode": null },
  "pricing": { "amountCents": 0, "currency": "usd", "displayPrice": "Contact us" },
  "limits": {
    "cloudOptimizationsMonthly": 999999999,
    "cloudTokensMonthly": 999999999999,
    "maxRawTokensPerOptimization": 2000000,
    "maxPayloadBytes": 50000000,
    "retainedHistoryDays": 365,
    "cliSessions": 999
  },
  "credits": { "includedMonthly": 999999999, "rollover": true, "allowManualGrants": true, "allowPurchases": true },
  "features": {
    "localOptimization": true, "remoteOptimization": true,
    "usageDashboard": true, "advancedUsageDashboard": true,
    "cliSessionManagement": true, "payloadCapture": true,
    "exportData": true, "priorityOptimizerUpdates": true,
    "customOptimizers": true, "ssoSaml": true,
    "optimizationHistory": true, "earlyAccessAdapters": true,
    "sessionMemoryGuard": true, "preCompactOptimizer": true, "localMemoryPacks": true
  },
  "optimizers": {
    "local": ["generic","bash","read","webfetch"],
    "remote": ["figma","github_pr","jira","confluence","clickup","slack","notion","amplitude","fireflies","playwright","zapier","google_drive"]
  },
  "telemetry": { "requiredUsageMetering": true, "optionalProductAnalyticsDefault": false }
}')

ON CONFLICT (id) DO NOTHING;  -- idempotent seed

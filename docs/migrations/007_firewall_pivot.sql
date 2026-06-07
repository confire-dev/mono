-- Migration 007: Firewall pivot
-- Removes optimizer columns, adds security/provenance/bypass tables.

-- ── Drop optimizer columns ────────────────────────────────────────────────────

ALTER TABLE tool_call_summaries
  DROP COLUMN IF EXISTS raw_bytes,
  DROP COLUMN IF EXISTS optimized_bytes,
  DROP COLUMN IF EXISTS optimizer,
  DROP COLUMN IF EXISTS was_cached,
  DROP COLUMN IF EXISTS credits_used;

ALTER TABLE usage_periods
  DROP COLUMN IF EXISTS saved_tokens,
  DROP COLUMN IF EXISTS cloud_tokens_used,
  DROP COLUMN IF EXISTS cloud_optimizations_used,
  DROP COLUMN IF EXISTS local_optimizations_count;

-- ── New tables ────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS security_events (
  id              UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  user_id         UUID REFERENCES profiles(id) ON DELETE CASCADE,
  session_id      TEXT,
  tool_name       TEXT NOT NULL,
  event_type      TEXT NOT NULL,
  -- PROMPT_INJECTION, HIDDEN_TEXT, SECRET_REDACTED, CREDENTIAL_LURE,
  -- DESTRUCTIVE_CMD, SECRET_FILE_ACCESS, MUTATING_MCP, SOCIAL_ENGINEERING
  risk_level      TEXT NOT NULL,
  -- NONE, LOW, MEDIUM, HIGH, CRITICAL
  action_taken    TEXT NOT NULL,
  -- ALLOWED, WARNED, REVIEW_REQUIRED, BLOCKED, SANITIZED
  pattern_matched TEXT,
  bypassed        BOOLEAN NOT NULL DEFAULT FALSE,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS security_events_user_id_idx ON security_events(user_id);
CREATE INDEX IF NOT EXISTS security_events_created_at_idx ON security_events(created_at DESC);

CREATE TABLE IF NOT EXISTS provenance_events (
  id              UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  user_id         UUID REFERENCES profiles(id) ON DELETE CASCADE,
  session_id      TEXT,
  tool_name       TEXT NOT NULL,
  trust_level     TEXT NOT NULL,
  -- internal, external_trusted, external_untrusted
  sanitized       BOOLEAN NOT NULL DEFAULT FALSE,
  redaction_count INTEGER NOT NULL DEFAULT 0,
  flags           TEXT[],
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS provenance_events_user_id_idx ON provenance_events(user_id);

CREATE TABLE IF NOT EXISTS bypass_events (
  id         UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  user_id    UUID REFERENCES profiles(id) ON DELETE CASCADE,
  session_id TEXT,
  tool_name  TEXT,
  reason     TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ── RLS policies ──────────────────────────────────────────────────────────────

ALTER TABLE security_events ENABLE ROW LEVEL SECURITY;
CREATE POLICY "Users see own security events" ON security_events
  FOR SELECT USING (auth.uid() = user_id);

ALTER TABLE provenance_events ENABLE ROW LEVEL SECURITY;
CREATE POLICY "Users see own provenance events" ON provenance_events
  FOR SELECT USING (auth.uid() = user_id);

ALTER TABLE bypass_events ENABLE ROW LEVEL SECURITY;
CREATE POLICY "Users see own bypass events" ON bypass_events
  FOR SELECT USING (auth.uid() = user_id);

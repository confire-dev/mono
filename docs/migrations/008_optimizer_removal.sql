-- Migration 008: Optimizer removal + schema cleanup
--
-- Drops remaining optimizer-only columns and stored procedures that survived
-- migration 007. Adds mcp_server and origin_domain to provenance_events
-- (required by recordProvenanceEvent() in the Worker).

-- ── Drop remaining optimizer columns ─────────────────────────────────────────

-- tool_call_summaries: drop computed columns that depended on dropped raw_bytes/optimized_bytes
ALTER TABLE tool_call_summaries
  DROP COLUMN IF EXISTS saved_bytes,
  DROP COLUMN IF EXISTS reduction_ratio;

-- cli_sessions: drop optimizer tracking columns
ALTER TABLE cli_sessions
  DROP COLUMN IF EXISTS optimized_calls,
  DROP COLUMN IF EXISTS raw_bytes,
  DROP COLUMN IF EXISTS optimized_bytes;

-- monthly_usage: drop optimizer metric column
ALTER TABLE monthly_usage
  DROP COLUMN IF EXISTS saved_bytes;

-- ── Drop optimizer-only stored procedures ─────────────────────────────────────

DROP FUNCTION IF EXISTS increment_usage_period(uuid, integer, bigint, bigint);
DROP FUNCTION IF EXISTS get_or_create_active_period(uuid, text);
DROP FUNCTION IF EXISTS increment_usage(uuid, integer, integer);

-- ── Extend provenance_events for full provenance label ───────────────────────

-- Add columns written by recordProvenanceEvent() in the Worker
ALTER TABLE provenance_events
  ADD COLUMN IF NOT EXISTS mcp_server    TEXT,
  ADD COLUMN IF NOT EXISTS origin_domain TEXT;

-- Migration 006 — Add early_access_requests table
-- Run in: Supabase Dashboard → SQL Editor
-- Apply to: dev project AND production project separately

CREATE TABLE IF NOT EXISTS early_access_requests (
    id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    email          TEXT        NOT NULL,
    name           TEXT,
    company        TEXT,
    client         TEXT,
    use_case       TEXT,
    plan_interest  TEXT        NOT NULL DEFAULT 'dev',
    source         TEXT        NOT NULL DEFAULT 'pricing_page',
    status         TEXT        NOT NULL DEFAULT 'new',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS early_access_requests_email_plan_idx
    ON early_access_requests (lower(email), plan_interest);

-- Public, unauthenticated inserts only — no read access for anon/authenticated.
ALTER TABLE early_access_requests ENABLE ROW LEVEL SECURITY;

CREATE POLICY "public_insert" ON early_access_requests
    FOR INSERT WITH CHECK (true);

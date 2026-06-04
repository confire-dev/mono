-- Migration 005 — Add cancel_at_period_end to profiles
-- Run in: Supabase Dashboard → SQL Editor
-- Apply to: dev project AND production project separately

ALTER TABLE profiles
  ADD COLUMN IF NOT EXISTS cancel_at_period_end boolean NOT NULL DEFAULT false;

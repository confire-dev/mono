# User Management

All user operations are done via Supabase Dashboard SQL or Supabase Table Editor.
No custom admin panel in v1.

---

## Find a user

```sql
-- By email
SELECT id, email, name, plan, subscription_status, created_at, is_banned
FROM profiles
WHERE email = 'user@example.com';

-- By API key prefix (from user's `confire whoami`)
SELECT p.id, p.email, p.plan, k.key_prefix, k.last_used_at, k.device_id
FROM api_keys k
JOIN profiles p ON p.id = k.user_id
WHERE k.key_prefix LIKE 'cf_live_xxxx%';  -- first 12 chars of their key
```

---

## Ban a user

```sql
UPDATE profiles
SET is_banned = true, updated_at = now()
WHERE email = 'bad-actor@example.com';
```

The Worker checks `is_banned` in `authenticate()`. Banned users get 401 on every request.

To unban:
```sql
UPDATE profiles SET is_banned = false, updated_at = now() WHERE email = '...';
```

---

## Revoke a specific API key (device)

Use when a user's device is compromised, or they ask you to invalidate a key:

```sql
-- Revoke by key prefix (user provides the first 12 chars from `confire whoami`)
UPDATE api_keys
SET revoked_at = now()
WHERE key_prefix = 'cf_live_xxxx';

-- Revoke all keys for a user (force re-login on all devices)
UPDATE api_keys
SET revoked_at = now()
WHERE user_id = '<user_uuid>'
  AND revoked_at IS NULL;
```

---

## See all active sessions for a user

```sql
SELECT
  id,
  device_id,
  cli_version,
  integration,
  started_at,
  ended_at,
  optimized_calls,
  saved_tokens
FROM cli_sessions
WHERE user_id = '<user_uuid>'
ORDER BY started_at DESC
LIMIT 20;
```

---

## Delete a user (GDPR / account deletion)

Supabase's `ON DELETE CASCADE` on `auth.users` handles all child records automatically:
- profiles, credit_balances, credit_ledger, api_keys, cli_sessions, tool_call_summaries, usage_periods, audit_events, user_preferences

```sql
-- This cascades to all user data (profiles, sessions, credits, etc.)
DELETE FROM auth.users WHERE email = 'user@example.com';
```

> ⚠️ This is irreversible. Confirm with the user before doing this.

For soft deletion (keep records for billing reconciliation):
```sql
-- Anonymise the profile without deleting records
UPDATE profiles SET email = 'deleted_' || id || '@deleted.invalid', name = null WHERE id = '<uuid>';
UPDATE auth.users SET email = 'deleted_' || id || '@deleted.invalid' WHERE id = '<uuid>';
```

---

## Audit log for a user

All security-sensitive actions are logged in `audit_events`:

```sql
SELECT event_type, device_id, metadata, created_at
FROM audit_events
WHERE user_id = '<user_uuid>'
ORDER BY created_at DESC
LIMIT 50;
```

Common event types: `api_key_generated`, `hook_installed`, `api_key_revoked`

---

## User stats snapshot

```sql
SELECT
  p.email,
  p.plan,
  p.subscription_status,
  p.created_at                          AS joined_at,
  cb.total_credits,
  up.cloud_optimizations_used           AS opts_this_period,
  up.cloud_tokens_used                  AS tokens_this_period,
  up.saved_tokens                       AS tokens_saved_this_period,
  (SELECT COUNT(*) FROM api_keys WHERE user_id = p.id AND revoked_at IS NULL) AS active_keys
FROM profiles p
LEFT JOIN credit_balances cb ON cb.user_id = p.id
LEFT JOIN usage_periods up   ON up.user_id = p.id AND up.period_end IS NULL
WHERE p.email = 'user@example.com';
```

# Incident Runbook

Quick reference for when things break. Each scenario has: symptoms, diagnosis, fix.

---

## Worker returning 5xx on all requests

**Symptoms:** Users report `confire hook` not working; `wrangler tail` shows 5xx.

**Diagnose:**
```bash
wrangler tail --format=pretty | head -50
curl https://api.confire.dev/health
```

**Common causes and fixes:**

1. **Secrets missing after deploy**
   ```bash
   wrangler secret list
   # If SUPABASE_URL etc. are missing:
   wrangler secret put SUPABASE_URL
   wrangler secret put SUPABASE_SERVICE_KEY
   ```

2. **Broken deploy — roll back**
   ```bash
   wrangler deployments list
   wrangler rollback <last-good-deployment-id>
   ```

3. **Supabase down** — check https://status.supabase.com
   - Worker will fall back to passthrough (returns unoptimized responses)
   - Users see "Using local optimizer" in Claude context
   - No action needed; auto-recovers when Supabase comes back

---

## Plans not loading (KV empty)

**Symptoms:** Worker logs `Plans not found in KV`. Optimizations still work (Worker throws, falls back to passthrough).

**Fix:**
```bash
curl -X POST https://api.confire.dev/admin/plans/sync \
  -H "Authorization: Bearer $ADMIN_SECRET"
```

This reads plans from Supabase and writes to KV. Should take < 2 seconds.

**Why it happened:**
- First deploy (KV always empty until first request auto-bootstraps)
- KV namespace was deleted or recreated
- Supabase webhook is misconfigured (plans changed in DB but webhook didn't fire)

**Check webhook:**
Go to Supabase Dashboard → Database → Webhooks → check `sync-plans-to-kv` logs.

---

## Users can't log in (CLI login fails)

**Symptoms:** `confire login` fails; browser shows error after OAuth.

**Diagnose:**
```bash
# Check Worker auth endpoint
curl https://api.confire.dev/health

# Check Supabase Auth settings
# Supabase Dashboard → Authentication → URL Configuration
# Site URL should be: https://confire.dev
# Redirect URLs should include: https://confire.dev/auth/callback
```

**Common causes:**
1. OAuth provider credentials expired → update in Supabase Auth providers
2. Redirect URL mismatch → add correct URL to Supabase allowed list
3. Worker `SUPABASE_ANON_KEY` is wrong → `wrangler secret put SUPABASE_ANON_KEY`

---

## Stripe webhook not updating user plans

**Symptoms:** User paid but still on free plan.

**Diagnose:**
```bash
# Check Stripe dashboard → Webhooks → recent deliveries
# Look for failed attempts on subscription events
```

**Fix (manual):**
```sql
UPDATE profiles
SET
  plan                            = 'pro',
  plan_id                         = 'pro',
  subscription_status             = 'active',
  stripe_subscription_id          = 'sub_xxx',
  subscription_current_period_end = '2025-07-01',
  updated_at                      = now()
WHERE email = 'user@example.com';

-- Grant the pro plan credits
SELECT grant_credits(
  p_user_id  := (SELECT id FROM profiles WHERE email = 'user@example.com'),
  p_included := 5000,
  p_source   := 'manual'
);
```

Then in Stripe Dashboard → Webhooks → resend the `customer.subscription.updated` event.

---

## User reports optimizations stopped working mid-month

**Symptoms:** User says hook is silent (no `🔥` lines); optimizations not happening.

**Diagnose:**
```sql
-- Check their current usage
SELECT up.cloud_optimizations_used, (p.config->'limits'->>'cloudOptimizationsMonthly')::int AS limit
FROM usage_periods up
JOIN plans p ON p.id = up.plan_id
WHERE up.user_id = (SELECT id FROM profiles WHERE email = 'user@example.com')
  AND up.period_end IS NULL;
```

**If they hit their limit:**
- Expected behavior — local optimization still works
- Offer upgrade or grant bonus credits (see billing-ops.md)

**If they haven't hit their limit:**
- Check if their API key was revoked: `SELECT * FROM api_keys WHERE user_id = '...' AND revoked_at IS NULL`
- Check if they're banned: `SELECT is_banned FROM profiles WHERE email = '...'`
- Ask them to run `confire status` and share the output

---

## Database connection pool exhausted

**Symptoms:** Supabase returns 503; Worker logs DB errors.

**Supabase Dashboard → Database → Connection Pooling:**
- Default: 15 direct connections, Supabase connection pooler handles more
- Use the connection pooler URL, not the direct URL, in `SUPABASE_URL`

---

## Rollback a plan change that broke something

If you updated a plan in the DB and it caused issues:

```sql
-- View the updated_at to confirm when it changed
SELECT id, config, updated_at FROM plans WHERE id = 'pro';

-- Restore limits to previous values
UPDATE plans
SET config = jsonb_set(config, '{limits,cloudOptimizationsMonthly}', '5000')
WHERE id = 'pro';
-- → DB webhook fires → KV updated → Workers pick up change within ~1s
```

---

## Emergency: stop all optimizations

If there's a critical bug in an optimizer:

```sql
-- Disable remote optimization for all plans (Workers return passthrough)
UPDATE plans
SET config = jsonb_set(config, '{features,remoteOptimization}', 'false');
-- → webhook fires → KV updated
```

Re-enable:
```sql
UPDATE plans
SET config = jsonb_set(config, '{features,remoteOptimization}', 'true');
```

---

## Contact & escalation

| Situation | Who |
|---|---|
| Stripe billing dispute | Stripe Dashboard → handle directly |
| Supabase outage | https://status.supabase.com |
| Cloudflare Workers outage | https://www.cloudflarestatus.com |
| User data deletion request | Delete from `auth.users` (cascades) |

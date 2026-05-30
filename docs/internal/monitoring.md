# Monitoring

---

## Live Worker logs

```bash
# All logs
wrangler tail

# Errors only
wrangler tail --format=pretty | grep -iE "error|failed|500|exception"

# Specific route
wrangler tail --format=pretty | grep "/optimize"
```

---

## Key metrics to watch

### Cloudflare Analytics Engine

Query via the Cloudflare dashboard (Workers & Pages → Analytics Engine) or REST API.

| Metric | Description | Alert if… |
|---|---|---|
| `tool_optimized` events/min | Optimization throughput | Drops to 0 unexpectedly |
| Error rate on `/optimize` | Worker errors | > 1% |
| p95 latency on `/optimize` | Response time | > 500ms |
| KV read rate | Cache hit rate | Spikes (cold starts) |

### Amplitude — product analytics events

These events flow to Amplitude for funnel analysis:

| Event | Meaning |
|---|---|
| `api_key_generated` | User completed login + authorized CLI |
| `hook_installed` | User ran `confire setup` |
| `tool_call_optimized` | Optimization happened |
| `cli_session_started` | Claude Code session began |
| `cli_session_ended` | Session ended (check session stats) |

Useful Amplitude funnels:
1. `api_key_generated` → `hook_installed` → `tool_call_optimized` (activation funnel)
2. `cli_session_started` → `tool_call_optimized` (active user signal)

---

## Supabase — useful queries

### Active users (last 7 days)

```sql
SELECT COUNT(DISTINCT user_id) AS active_users_7d
FROM cli_sessions
WHERE started_at >= now() - interval '7 days';
```

### Daily optimizations

```sql
SELECT
  DATE_TRUNC('day', created_at) AS day,
  COUNT(*) AS optimizations,
  SUM(saved_tokens) AS tokens_saved
FROM tool_call_summaries
WHERE created_at >= now() - interval '30 days'
GROUP BY 1
ORDER BY 1 DESC;
```

### Top tools being optimized

```sql
SELECT
  tool_type,
  COUNT(*) AS calls,
  AVG(reduction_ratio) AS avg_reduction,
  SUM(saved_tokens) AS total_tokens_saved
FROM tool_call_summaries
WHERE created_at >= now() - interval '30 days'
GROUP BY tool_type
ORDER BY calls DESC;
```

### Users hitting rate limits

```sql
SELECT
  p.email,
  p.plan,
  up.cloud_optimizations_used,
  (plns.config->'limits'->>'cloudOptimizationsMonthly')::int AS limit
FROM usage_periods up
JOIN profiles p ON p.id = up.user_id
JOIN plans plns ON plns.id = up.plan_id
WHERE up.period_end IS NULL
  AND up.cloud_optimizations_used >=
      (plns.config->'limits'->>'cloudOptimizationsMonthly')::int
ORDER BY up.cloud_optimizations_used DESC;
```

### New signups today

```sql
SELECT email, created_at FROM profiles
WHERE created_at >= CURRENT_DATE
ORDER BY created_at DESC;
```

### Revenue (paid users)

```sql
SELECT
  plan,
  COUNT(*) AS users,
  COUNT(*) FILTER (WHERE subscription_status = 'active') AS active
FROM profiles
WHERE plan != 'free'
GROUP BY plan;
```

---

## Error patterns to know

| Log message | Cause | Fix |
|---|---|---|
| `Plans not found in KV` | KV empty, DB unreachable | Check Supabase URL/key; run `/admin/plans/sync` |
| `supabaseUrl is required` | Env vars not set in Worker | `wrangler secret put SUPABASE_URL` |
| `invalid API key` | User's key was revoked or never created | Ask user to run `confire login` |
| `Payload too large` | Tool response > plan limit | Expected; user sees a helpful message |
| `monthly_optimizations_exceeded` | User hit their plan limit | Expected; local optimization still works |
| `DB 500` from Supabase | Supabase outage or query error | Check Supabase status page |

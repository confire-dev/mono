# Confire — Internal Operations

This directory documents operational tasks for the Confire team.
Everything here is internal — not customer-facing docs.

## Index

| Doc | What it covers |
|---|---|
| [plan-management.md](./plan-management.md) | Updating plan limits, adding plans, the webhook sync system |
| [billing-ops.md](./billing-ops.md) | Stripe webhooks, manual credit grants, subscription edge cases |
| [user-management.md](./user-management.md) | Banning users, manual plan changes, API key revocation |
| [deploy.md](./deploy.md) | Deploy Worker, secrets, KV setup, wrangler commands |
| [monitoring.md](./monitoring.md) | Analytics Engine queries, Amplitude events, error patterns |
| [incident-runbook.md](./incident-runbook.md) | What to do when things break |

## Quick reference

```bash
# Deploy Worker
cd worker && wrangler deploy

# Sync plans from Supabase → KV (normally done by DB webhook automatically)
curl -X POST https://api.confire.dev/admin/plans/sync \
  -H "Authorization: Bearer $ADMIN_SECRET"

# Check Worker health
curl https://api.confire.dev/health

# Tail Worker logs live
wrangler tail
```

## Key URLs

| Thing | URL |
|---|---|
| Supabase dashboard | https://supabase.com/dashboard |
| Cloudflare Workers | https://dash.cloudflare.com |
| Stripe dashboard | https://dashboard.stripe.com |
| Amplitude | https://analytics.amplitude.com |

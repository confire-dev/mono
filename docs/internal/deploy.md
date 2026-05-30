# Deploy & Infrastructure

---

## First deploy (new environment)

### 1. Create Cloudflare resources

```bash
# KV namespace for plan cache + content-hash cache
wrangler kv namespace create confire-cache
wrangler kv namespace create confire-cache --preview
# → copy the IDs into worker/wrangler.toml

# Analytics Engine dataset
wrangler analytics-engine dataset create confire_events
```

### 2. Set secrets

```bash
wrangler secret put SUPABASE_URL
wrangler secret put SUPABASE_ANON_KEY
wrangler secret put SUPABASE_SERVICE_KEY
wrangler secret put STRIPE_WEBHOOK_SECRET
wrangler secret put SUPABASE_WEBHOOK_SECRET
wrangler secret put AMPLITUDE_KEY
```

### 3. Apply Supabase schema

Go to **Supabase Dashboard → SQL Editor** and run `docs/supabase-schema.sql`.
This creates all tables, stored procedures, RLS policies, and seeds the `plans` table.

### 4. Configure Supabase Auth providers

**Supabase Dashboard → Authentication → Providers:**
- Enable GitHub: add Client ID + Secret from GitHub OAuth App
- Enable Google: add Client ID + Secret from Google Cloud Console

**Site URL:** `https://confire.dev`  
**Redirect URLs:** `https://confire.dev/auth/callback`

### 5. Configure Supabase DB webhook

**Supabase Dashboard → Database → Webhooks → Create:**
- Table: `plans`, Events: INSERT, UPDATE, DELETE
- URL: `https://api.confire.dev/webhooks/supabase`
- Header: `Authorization: Bearer <SUPABASE_WEBHOOK_SECRET>`

### 6. Deploy the Worker

```bash
cd worker
wrangler deploy
```

### 7. Verify

```bash
# Health check
curl https://api.confire.dev/health

# Plans loaded from DB into KV (first request auto-bootstraps)
curl https://api.confire.dev/health
# Worker logs: "[confire] Plans not in KV, loading from Supabase..."
```

---

## Routine deploy (update)

```bash
cd worker
wrangler deploy
```

Plans are in KV — they survive Worker redeploys unchanged.
If you updated plan config in the DB, the webhook handles KV sync automatically.

---

## Rollback

Cloudflare keeps the last deployment:

```bash
# List recent deployments
wrangler deployments list

# Roll back to a specific version
wrangler rollback <deployment-id>
```

---

## Environment variables reference

| Secret | Where to get it |
|---|---|
| `SUPABASE_URL` | Supabase Dashboard → Project Settings → API → Project URL |
| `SUPABASE_ANON_KEY` | Supabase Dashboard → Project Settings → API → anon/public |
| `SUPABASE_SERVICE_KEY` | Supabase Dashboard → Project Settings → API → service_role |
| `STRIPE_WEBHOOK_SECRET` | Stripe Dashboard → Webhooks → endpoint → Signing secret |
| `SUPABASE_WEBHOOK_SECRET` | Generate: `openssl rand -base64 32` — use same value in Supabase webhook header |
| `AMPLITUDE_KEY` | Amplitude → Settings → Projects → API Key |

---

## CLI binary release

```bash
cd cli

# Dev build (with symbols, fast)
make dev

# Release build (garble obfuscation + stripped symbols)
make release

# Install to PATH
make install
```

The release binary uses `garble` for obfuscation. Install it first:
```bash
go install mvdan.cc/garble@latest
```

---

## Platform (Astro)

The platform deploys to Cloudflare Pages.

```bash
cd platform
pnpm build

# Deploy to Cloudflare Pages (configure in Cloudflare dashboard)
# Connect GitHub repo → auto-deploy on push to main
```

Environment variables for the platform (set in Cloudflare Pages settings):
```
PUBLIC_SUPABASE_URL=https://xxx.supabase.co
PUBLIC_SUPABASE_ANON_KEY=eyJ...
PUBLIC_WORKER_URL=https://api.confire.dev
PUBLIC_STRIPE_PUBLISHABLE_KEY=pk_live_...
```

---

## Monitoring

```bash
# Live Worker logs
wrangler tail

# Filter to errors only
wrangler tail --format=pretty | grep -i "error\|failed\|500"

# Analytics Engine — query via Cloudflare dashboard or API
# https://dash.cloudflare.com → Workers → Analytics Engine
```

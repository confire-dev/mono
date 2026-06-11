# Deploy & Infrastructure

---

## Environments

| Environment | Worker | Platform | Docs |
|---|---|---|---|
| **dev** | `confire-worker-dev` · `api.dev.confire.dev` | `dev.confire.dev` | `docs.dev.confire.dev` (if configured) |
| **production** | `confire-worker` · `api.confire.dev` | `confire.dev` | `docs.confire.dev` |

Dev uses a separate Supabase project and Stripe test-mode keys.
Production uses the live Supabase project and Stripe live keys.

---

## Deploy triggers

| What | Dev | Production |
|---|---|---|
| CLI binary | Auto — push to `develop` touching `cli/` | Tag `v*` → `release.yml` |
| Worker | **Manual** — `wrangler deploy --env dev` | **Manual** — `wrangler deploy --env production` |
| Platform | Auto — Cloudflare Pages git integration (`develop` branch) | Manual — `wrangler pages deploy dist` or workflow dispatch |
| Docs site | Auto — Cloudflare Pages git integration (`develop` branch) | Auto — same |

---

## First deploy (new environment)

### 1. Create Cloudflare KV namespaces

```bash
# dev
wrangler kv namespace create confire_cache_dev
wrangler kv namespace create confire_cache_dev --preview

# production
wrangler kv namespace create confire_cache
wrangler kv namespace create confire_cache --preview
```

Copy the returned IDs into `worker/wrangler.toml`.

**Analytics Engine:** no setup needed — the `AE` binding is declared in `wrangler.toml` and the dataset
is created automatically on the first `writeDataPoint()` call.

**Rate limiting:** no setup needed — `[[ratelimits]]` bindings activate on deploy.

### 2. Set Worker secrets

Run from `worker/`:

```bash
# dev
wrangler secret put SUPABASE_URL             --env dev
wrangler secret put SUPABASE_ANON_KEY        --env dev
wrangler secret put SUPABASE_SERVICE_KEY     --env dev
wrangler secret put STRIPE_SECRET_KEY        --env dev   # sk_test_...
wrangler secret put STRIPE_WEBHOOK_SECRET    --env dev
wrangler secret put SUPABASE_WEBHOOK_SECRET  --env dev   # openssl rand -base64 32
wrangler secret put AMPLITUDE_KEY            --env dev

# production
wrangler secret put SUPABASE_URL             --env production
wrangler secret put SUPABASE_ANON_KEY        --env production
wrangler secret put SUPABASE_SERVICE_KEY     --env production
wrangler secret put STRIPE_SECRET_KEY        --env production   # sk_live_...
wrangler secret put STRIPE_WEBHOOK_SECRET    --env production
wrangler secret put SUPABASE_WEBHOOK_SECRET  --env production
wrangler secret put AMPLITUDE_KEY            --env production
```

### 3. Apply Supabase migrations

Run all migration files in order in **Supabase Dashboard → SQL Editor** for each project:

```
docs/migrations/002_stripe_dev_plan_prices.sql
docs/migrations/003_billing_v2.sql
docs/migrations/004_handle_new_user_trigger.sql
docs/migrations/005_cancel_at_period_end.sql
docs/migrations/006_early_access_requests.sql
docs/migrations/007_firewall_pivot.sql
docs/migrations/008_optimizer_removal.sql
```

Migration 008 is required before deploying the current Worker — it adds `mcp_server` and `origin_domain`
columns to `provenance_events` that `recordProvenanceEvent()` writes.

### 4. Seed plan data

After running migrations, the `plans` table needs plan configs. If starting from scratch, insert the
free and dev plans directly via the Supabase SQL editor or admin API.

For dev plan Stripe prices (test mode), update after creating products in Stripe:

```sql
UPDATE plans
SET config = jsonb_set(jsonb_set(jsonb_set(config,
  '{stripe,productId}', '"prod_YOUR_TEST_PRODUCT"'),
  '{stripe,prices,monthly}', '"price_YOUR_MONTHLY"'),
  '{stripe,prices,annual}',  '"price_YOUR_ANNUAL"')
WHERE id = 'dev';
```

See `docs/internal/stripe-metadata.md` for the full Stripe setup including webhook events and test-mode IDs.

### 5. Configure Supabase Auth providers

**Supabase Dashboard → Authentication → Providers** for each project:

- Enable GitHub: Client ID + Secret
- Enable Google: Client ID + Secret

| Project | Site URL | Redirect URL |
|---|---|---|
| dev | `https://dev.confire.dev` | `https://dev.confire.dev/auth/callback` |
| production | `https://confire.dev` | `https://confire.dev/auth/callback` |

### 6. Configure Supabase DB webhook (plans → KV sync)

**Supabase Dashboard → Database → Webhooks → Create a new hook** for each project:

| Field | Value |
|---|---|
| Name | `sync-plans-to-worker` |
| Table | `public` · `plans` |
| Events | `INSERT` `UPDATE` `DELETE` |
| Type | `HTTP Request` · `POST` |
| URL (dev) | `https://api.dev.confire.dev/webhooks/supabase` |
| URL (prod) | `https://api.confire.dev/webhooks/supabase` |
| HTTP Headers | `Authorization: Bearer <SUPABASE_WEBHOOK_SECRET>` |
| Retry | disabled |

The `SUPABASE_WEBHOOK_SECRET` value must match what you set via `wrangler secret put`.

### 7. First Worker deploy

```bash
cd worker
wrangler deploy --env dev
wrangler deploy --env production
```

### 8. Configure Cloudflare Pages (Platform)

**Dev platform (`dev.confire.dev`):**

1. Cloudflare Dashboard → Workers & Pages → Create → Pages → Connect to Git
2. Repo: this repo · branch: `develop` · project name: `confire-platform-dev`
3. Build command: `pnpm build` · output: `dist` · root: `platform`
4. Environment variables:
   ```
   PUBLIC_SUPABASE_URL=https://<dev-project>.supabase.co
   PUBLIC_SUPABASE_ANON_KEY=eyJ...
   PUBLIC_WORKER_URL=https://api.dev.confire.dev
   PUBLIC_STRIPE_PUBLISHABLE_KEY=pk_test_...
   ```
5. Custom domain: `dev.confire.dev`

**Prod platform (`confire.dev`):**

1. Cloudflare Dashboard → Workers & Pages → Create → Pages → Direct Upload
2. Project name: `confire-platform` · Custom domain: `confire.dev`
3. Add GitHub Actions secrets (repo Settings → Secrets → Actions):
   - `PROD_PUBLIC_SUPABASE_URL`
   - `PROD_PUBLIC_SUPABASE_ANON_KEY`
   - `PROD_PUBLIC_STRIPE_PUBLISHABLE_KEY`

### 9. Verify

```bash
curl https://api.dev.confire.dev/health
curl https://api.confire.dev/health
```

---

## Routine deploys

### Worker (manual — no CI)

```bash
cd worker
wrangler deploy --env dev
wrangler deploy --env production
```

Plans in KV survive redeploys. If you updated plan config in the DB, the Supabase webhook syncs KV
automatically — no redeploy needed.

### Platform

Dev: auto-deploys from `develop` via Cloudflare Pages git integration.

Production (manual):
```bash
cd platform && pnpm build
wrangler pages deploy dist --project-name confire-platform
```
Or trigger via GitHub Actions → **Deploy Platform (production)**.

### Docs site

Both envs auto-deploy from `develop` via Cloudflare Pages git integration.

### Distribution worker (`get.confire.dev`)

```bash
pnpm distribution:deploy
```

Only needed when `distribution/` changes. No dev/prod split.

---

## GitHub Actions secrets

| Secret | Used by | Purpose |
|---|---|---|
| `CLOUDFLARE_API_TOKEN` | all deploy workflows + release | Workers Edit + R2 Edit permissions |
| `CLOUDFLARE_ACCOUNT_ID` | all deploy workflows + release | Cloudflare account ID |
| `PROD_PUBLIC_SUPABASE_URL` | `deploy-platform-prod` | Platform build env var |
| `PROD_PUBLIC_SUPABASE_ANON_KEY` | `deploy-platform-prod` | Platform build env var |
| `PROD_PUBLIC_STRIPE_PUBLISHABLE_KEY` | `deploy-platform-prod` | `pk_live_...` |
| `HOMEBREW_TAP_TOKEN` | `release` | *(optional)* PAT with write access to `confire-ai/homebrew-confire` |

---

## Rollback

```bash
wrangler deployments list --env dev
wrangler rollback <deployment-id> --env dev

wrangler deployments list --env production
wrangler rollback <deployment-id> --env production
```

---

## Worker secrets reference

| Secret | Where to get it |
|---|---|
| `SUPABASE_URL` | Supabase Dashboard → Project Settings → API → Project URL |
| `SUPABASE_ANON_KEY` | Supabase Dashboard → Project Settings → API → anon/public key |
| `SUPABASE_SERVICE_KEY` | Supabase Dashboard → Project Settings → API → service_role key |
| `STRIPE_SECRET_KEY` | Stripe Dashboard → Developers → API keys → Secret key |
| `STRIPE_WEBHOOK_SECRET` | Stripe Dashboard → Webhooks → endpoint → Signing secret |
| `SUPABASE_WEBHOOK_SECRET` | Generate: `openssl rand -base64 32` — use same value in the Supabase webhook header |
| `AMPLITUDE_KEY` | Amplitude → Settings → Projects → API Key |

---

## Monitoring

```bash
# Live Worker logs
wrangler tail --env dev
wrangler tail --env production

# Errors only
wrangler tail --env production --format=pretty | grep -i "error\|failed\|500"
```

Cloudflare Analytics Engine: Dashboard → Workers → Analytics Engine → `confire_events` dataset.

---

## CLI

### Local builds

```bash
cd cli

# confire — prod APIs
make dev          # build with symbols (fast)
make install      # install to GOPATH/bin

# confire-dev — dev APIs (api.dev.confire.dev / dev.confire.dev)
make build-dev-env    # build ./confire-dev
make install-dev-env  # install to GOPATH/bin/confire-dev
```

### Tagged release (CI)

The `bump-version.yml` workflow runs on every push to `main`, reads conventional commits since the
last tag, bumps semver, updates `CHANGELOG.md`, commits, and creates a tag.
The tag triggers `release.yml`:

| Job | What it does |
|---|---|
| `build` | Cross-compile `confire_{darwin,linux}_{amd64,arm64}` |
| `release` | Create GitHub Release + upload binaries and checksums |
| `publish-r2` | Upload to R2, update `latest.json` and `install.sh` |

To cut a release manually:

```bash
git tag v1.2.3
git push origin v1.2.3
```

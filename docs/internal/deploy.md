# Deploy & Infrastructure

---

## Environments

| Environment | Worker | Platform | Deploy trigger |
|---|---|---|---|
| **dev** | `confire-worker-dev` at `api-dev.confire.dev` | `dev.confire.dev` (Pages project `confire-platform-dev`) | Push to `main` (CI auto) |
| **production** | `confire-worker` at `api.confire.dev` | `confire.dev` (Pages project `confire-platform`) | Manual workflow dispatch only |

Dev uses separate Supabase project + Stripe test-mode keys.
Production uses live Supabase project + Stripe live keys.

---

## First deploy (new environment)

Run these steps once for each environment (`dev` or `production`).
All `wrangler` commands below take `--env dev` or `--env production`.

### 1. Create Cloudflare resources

```bash
# KV namespace
wrangler kv namespace create confire-cache-dev   # dev
wrangler kv namespace create confire-cache-dev --preview
wrangler kv namespace create confire-cache       # production
wrangler kv namespace create confire-cache --preview
# Copy the IDs into worker/wrangler.toml under [env.dev] and [env.production]

# Analytics Engine (production only — dev reuses the dev dataset name)
wrangler analytics-engine dataset create confire_events

# Rate limiting namespaces (one per plan tier, per environment)
wrangler rate-limit create rl-free-dev   # dev
wrangler rate-limit create rl-dev-dev
wrangler rate-limit create rl-pro-dev
wrangler rate-limit create rl-free      # production
wrangler rate-limit create rl-dev
wrangler rate-limit create rl-pro
# Copy namespace_ids into worker/wrangler.toml
```

### 2. Set worker secrets

Run from `worker/`:

```bash
# Dev secrets (Supabase dev project, Stripe test keys)
wrangler secret put SUPABASE_URL          --env dev
wrangler secret put SUPABASE_ANON_KEY     --env dev
wrangler secret put SUPABASE_SERVICE_KEY  --env dev
wrangler secret put STRIPE_SECRET_KEY     --env dev   # sk_test_...
wrangler secret put STRIPE_WEBHOOK_SECRET --env dev
wrangler secret put STRIPE_TOPUP_PRICE_ID --env dev
wrangler secret put SUPABASE_WEBHOOK_SECRET --env dev
wrangler secret put AMPLITUDE_KEY         --env dev

# Production secrets (Supabase prod project, Stripe live keys)
wrangler secret put SUPABASE_URL          --env production
wrangler secret put SUPABASE_ANON_KEY     --env production
wrangler secret put SUPABASE_SERVICE_KEY  --env production
wrangler secret put STRIPE_SECRET_KEY     --env production   # sk_live_...
wrangler secret put STRIPE_WEBHOOK_SECRET --env production
wrangler secret put STRIPE_TOPUP_PRICE_ID --env production
wrangler secret put SUPABASE_WEBHOOK_SECRET --env production
wrangler secret put AMPLITUDE_KEY         --env production
```

### 3. Apply Supabase schema

Go to **Supabase Dashboard → SQL Editor** for each project and run `docs/supabase-schema.sql`.

### 4. Configure Supabase Auth providers

**Supabase Dashboard → Authentication → Providers** for each project:
- Enable GitHub: Client ID + Secret
- Enable Google: Client ID + Secret

| Project | Site URL | Redirect URL |
|---|---|---|
| dev | `https://dev.confire.dev` | `https://dev.confire.dev/auth/callback` |
| production | `https://confire.dev` | `https://confire.dev/auth/callback` |

### 5. Configure Supabase DB webhook

**Supabase Dashboard → Database → Webhooks → Create** for each project:
- Table: `plans`, Events: INSERT, UPDATE, DELETE
- Header: `Authorization: Bearer <SUPABASE_WEBHOOK_SECRET>`

| Project | Webhook URL |
|---|---|
| dev | `https://api-dev.confire.dev/webhooks/supabase` |
| production | `https://api.confire.dev/webhooks/supabase` |

### 6. First worker deploy

```bash
cd worker
wrangler deploy --env dev         # dev
wrangler deploy --env production  # production
```

### 7. Set up Cloudflare Pages

**Dev platform (`dev.confire.dev`):**
1. Cloudflare Dashboard → Pages → Create project
2. Connect GitHub repo, branch: `main`, project name: `confire-platform-dev`
3. Build command: `pnpm build`, output: `dist`, root: `platform`
4. Set env vars (under Settings → Environment variables):
   ```
   PUBLIC_SUPABASE_URL=https://<dev-project>.supabase.co
   PUBLIC_SUPABASE_ANON_KEY=eyJ...
   PUBLIC_WORKER_URL=https://api-dev.confire.dev
   PUBLIC_STRIPE_PUBLISHABLE_KEY=pk_test_...
   ```
5. Custom domain: `dev.confire.dev`

**Prod platform (`confire.dev`):**
1. Cloudflare Dashboard → Pages → Create project
2. Name it `confire-platform`. Do NOT connect git (deployed via CI only).
3. Custom domain: `confire.dev`
4. Add GitHub Actions secrets (see "GitHub Actions secrets" section below):
   - `PROD_PUBLIC_SUPABASE_URL`
   - `PROD_PUBLIC_SUPABASE_ANON_KEY`
   - `PROD_PUBLIC_STRIPE_PUBLISHABLE_KEY`

### 8. Verify

```bash
curl https://api-dev.confire.dev/health
curl https://api.confire.dev/health
```

---

## Routine deploys

### Worker

Dev deploys automatically on every push to `main` that touches `worker/`.
Production deploys are manual — trigger via GitHub Actions → Deploy Worker (production).

```bash
# Manual deploy (local, any environment)
cd worker
wrangler deploy --env dev
wrangler deploy --env production
```

Plans are in KV and survive redeploys. If you updated plan config in the DB, the webhook syncs KV automatically.

### Platform

Dev platform auto-deploys from `main` via Cloudflare Pages git integration.

Production platform is deployed manually:
- GitHub Actions → Deploy Platform (production), or
- locally: `cd platform && pnpm build && wrangler pages deploy dist --project-name confire-platform`

Distribution worker: `pnpm distribution:deploy` (only when `distribution/` changes — no dev/prod split needed).

---

## GitHub Actions secrets

Required in repo Settings → Secrets (and for prod deploys, in the `production` GitHub Environment):

| Secret | Used by | Purpose |
|---|---|---|
| `CLOUDFLARE_API_TOKEN` | all deploy workflows + release | Workers deploy + R2 upload. Needs Account → Workers → Edit and Account → R2 → Edit |
| `CLOUDFLARE_ACCOUNT_ID` | all deploy workflows + release | Cloudflare account ID |
| `PROD_PUBLIC_SUPABASE_URL` | deploy-platform-prod | Platform build env var |
| `PROD_PUBLIC_SUPABASE_ANON_KEY` | deploy-platform-prod | Platform build env var |
| `PROD_PUBLIC_STRIPE_PUBLISHABLE_KEY` | deploy-platform-prod | Platform build env var (pk_live_...) |
| `HOMEBREW_TAP_TOKEN` | release | *(optional)* PAT for homebrew tap update |

The `production` GitHub Environment (Settings → Environments) gates the two manual workflows.
Add the same secrets there as environment-level overrides if you want environment-specific tokens.

---

## Rollback

```bash
wrangler deployments list --env dev
wrangler rollback <deployment-id> --env dev

wrangler deployments list --env production
wrangler rollback <deployment-id> --env production
```

---

## Environment variables reference

| Secret | Where to get it |
|---|---|
| `SUPABASE_URL` | Supabase Dashboard → Project Settings → API → Project URL |
| `SUPABASE_ANON_KEY` | Supabase Dashboard → Project Settings → API → anon/public |
| `SUPABASE_SERVICE_KEY` | Supabase Dashboard → Project Settings → API → service_role |
| `STRIPE_SECRET_KEY` | Stripe Dashboard → Developers → API keys → Secret key |
| `STRIPE_WEBHOOK_SECRET` | Stripe Dashboard → Webhooks → endpoint → Signing secret |
| `STRIPE_TOPUP_PRICE_ID` | Stripe Dashboard → Products → top-up pack → Price ID (one-time $5 / 5,000 credits) |
| `SUPABASE_WEBHOOK_SECRET` | Generate: `openssl rand -base64 32` — use same value in Supabase webhook header |
| `AMPLITUDE_KEY` | Amplitude → Settings → Projects → API Key |

## Feature flags

Feature flags are `[vars]` in `wrangler.toml`. They can also be overridden at runtime
without a redeploy using `wrangler secret put <FLAG> --env <dev|production>`.

| Flag | Default | Effect |
|---|---|---|
| `OPTIMIZER_API_ENABLED` | `"false"` | `"true"` enables `POST /v1/optimize` (standalone optimizer API). Returns 404 while disabled. |

---

## CLI binary release

### Local builds

```bash
cd cli

# confire — prod APIs (for testing against production)
make dev
make install

# confire-dev — dev APIs (api-dev.confire.dev / dev.confire.dev)
make build-dev-env
make install-dev-env

# Release build (garble obfuscation + stripped symbols)
make release

# Install garble first if not already installed
go install mvdan.cc/garble@latest
```

### Tagged release (CI)

Push a semver tag to trigger `.github/workflows/release.yml`:

```bash
git tag v0.3.0
git push origin v0.3.0
```

After the build matrix finishes, `release` and `publish-r2` run in parallel:

| Job | What it does |
|-----|----------------|
| `build` | Cross-compile `confire_{darwin,linux}_{amd64,arm64}` |
| `release` | Create GitHub Release + upload binaries and checksums |
| `publish-r2` | Upload binaries, `latest.json`, and `install.sh` to R2 |

If `HOMEBREW_TAP_TOKEN` is set, the `release` job also bumps `homebrew/confire.rb`.

---

## CLI distribution (get.confire.dev / releases.confire.dev)

Public install URLs. Binaries and `install.sh` live in R2; a small Worker serves them on custom subdomains.

| URL | Purpose |
|-----|---------|
| `https://get.confire.dev` | `install.sh` (curl pipe target) |
| `https://get.confire.dev/latest.json` | Version manifest |
| `https://releases.confire.dev/vX.Y.Z/confire_*` | Platform binaries + checksums |

### One-time setup

```bash
cd distribution
pnpm exec wrangler r2 bucket create confire-releases
pnpm deploy
# or from repo root: pnpm distribution:deploy
```

### Bootstrap first release

```bash
chmod +x scripts/publish-release-r2.sh
./scripts/publish-release-r2.sh v0.3.0 ./dist
```

```bash
curl -fsSL https://get.confire.dev/latest.json
curl -fsSL https://releases.confire.dev/v0.3.0/confire_checksums.txt
```

User install command:

```bash
curl -fsSL https://get.confire.dev | sh
```

---

## Monitoring

```bash
# Live Worker logs (pick environment)
wrangler tail --env dev
wrangler tail --env production

# Filter to errors
wrangler tail --env production --format=pretty | grep -i "error\|failed\|500"

# Analytics Engine — Cloudflare Dashboard → Workers → Analytics Engine
```

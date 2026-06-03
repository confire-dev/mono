# Deploy & Infrastructure

---

## Environments

| Environment | Worker | Platform | Deploy trigger |
|---|---|---|---|
| **dev** | `confire-worker-dev` at `api-dev.confire.dev` | `dev.confire.dev` (Pages: `confire-platform-dev`) | Auto on push to `main` |
| **production** | `confire-worker` at `api.confire.dev` | `confire.dev` (Pages: `confire-platform`) | Manual workflow dispatch only |

Dev uses a separate Supabase project and Stripe test-mode keys.
Production uses the live Supabase project and Stripe live keys.

---

## First deploy (new environment)

### 1. Create Cloudflare KV namespaces

Run from the repo root (or `worker/`):

```bash
# dev
wrangler kv namespace create confire_cache_dev
wrangler kv namespace create confire_cache_dev --preview

# production
wrangler kv namespace create confire_cache
wrangler kv namespace create confire_cache --preview
```

Copy the returned IDs into `worker/wrangler.toml` — replace `REPLACE_WITH_DEV_KV_ID` etc.

**Analytics Engine:** no setup needed. The `[[analytics_engine_datasets]]` binding is already declared in `wrangler.toml`. The dataset is created automatically on the first `env.AE.writeDataPoint()` call.

**Rate limiting:** no setup needed. `[[ratelimits]]` bindings are declared in `wrangler.toml` with fixed `namespace_id` numbers. They activate on deploy and are used for burst protection only — monthly quota is enforced in Supabase.

### 2. Set worker secrets

Run from `worker/`:

```bash
# dev (Supabase dev project, Stripe test keys)
wrangler secret put SUPABASE_URL            --env dev
wrangler secret put SUPABASE_ANON_KEY       --env dev
wrangler secret put SUPABASE_SERVICE_KEY    --env dev
wrangler secret put STRIPE_SECRET_KEY       --env dev   # sk_test_...
wrangler secret put STRIPE_WEBHOOK_SECRET   --env dev
wrangler secret put STRIPE_TOPUP_PRICE_ID   --env dev
wrangler secret put SUPABASE_WEBHOOK_SECRET --env dev
wrangler secret put AMPLITUDE_KEY           --env dev

# production (Supabase prod project, Stripe live keys)
wrangler secret put SUPABASE_URL            --env production
wrangler secret put SUPABASE_ANON_KEY       --env production
wrangler secret put SUPABASE_SERVICE_KEY    --env production
wrangler secret put STRIPE_SECRET_KEY       --env production   # sk_live_...
wrangler secret put STRIPE_WEBHOOK_SECRET   --env production
wrangler secret put STRIPE_TOPUP_PRICE_ID   --env production
wrangler secret put SUPABASE_WEBHOOK_SECRET --env production
wrangler secret put AMPLITUDE_KEY           --env production
```

### 3. Apply Supabase schema

Go to **Supabase Dashboard → SQL Editor** for each project and run `docs/supabase-schema.sql`.

### 4. Configure Supabase Auth providers

**Supabase Dashboard → Authentication → Providers** for each project:
- Enable GitHub: add Client ID + Secret
- Enable Google: add Client ID + Secret

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
wrangler deploy --env dev
wrangler deploy --env production
```

### 7. Set up Cloudflare Pages

**Dev platform (`dev.confire.dev`):**
1. Cloudflare Dashboard → Workers & Pages → Create → Pages → Connect to Git
2. Repo: this repo, branch: `main`, project name: `confire-platform-dev`
3. Build command: `pnpm build`, output directory: `dist`, root directory: `platform`
4. Environment variables (Settings → Environment variables):
   ```
   PUBLIC_SUPABASE_URL=https://<dev-project>.supabase.co
   PUBLIC_SUPABASE_ANON_KEY=eyJ...
   PUBLIC_WORKER_URL=https://api-dev.confire.dev
   PUBLIC_STRIPE_PUBLISHABLE_KEY=pk_test_...
   ```
5. Custom domain: `dev.confire.dev`

**Prod platform (`confire.dev`):**
1. Cloudflare Dashboard → Workers & Pages → Create → Pages → Direct Upload
2. Project name: `confire-platform` (no git connection — CI deploys via `wrangler pages deploy`)
3. Custom domain: `confire.dev`
4. Add these GitHub Actions secrets (repo Settings → Secrets and variables → Actions):
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

Dev deploys automatically on every push to `main` that touches `worker/` (via `deploy-worker-dev.yml`).
Production deploys are manual only — trigger via GitHub Actions → **Deploy Worker (production)**.

```bash
# Local manual deploy
cd worker
wrangler deploy --env dev
wrangler deploy --env production
```

Plans are in KV and survive redeploys. If you updated plan config in the DB, the Supabase webhook syncs KV automatically.

### Platform

Dev platform auto-deploys from `main` via Cloudflare Pages git integration — no action needed.

Production platform is manual only:
- GitHub Actions → **Deploy Platform (production)**, or
- Locally: `cd platform && pnpm build && wrangler pages deploy dist --project-name confire-platform`

Distribution worker (`get.confire.dev` / `releases.confire.dev`): `pnpm distribution:deploy` — only needed when `distribution/` changes. No dev/prod split.

---

## GitHub Actions secrets

| Secret | Used by | Purpose |
|---|---|---|
| `CLOUDFLARE_API_TOKEN` | all deploy workflows + release | Workers deploy + R2 upload. Needs Workers Edit + R2 Edit permissions. |
| `CLOUDFLARE_ACCOUNT_ID` | all deploy workflows + release | Cloudflare account ID |
| `PROD_PUBLIC_SUPABASE_URL` | `deploy-platform-prod` | Platform build env var |
| `PROD_PUBLIC_SUPABASE_ANON_KEY` | `deploy-platform-prod` | Platform build env var |
| `PROD_PUBLIC_STRIPE_PUBLISHABLE_KEY` | `deploy-platform-prod` | `pk_live_...` |
| `HOMEBREW_TAP_TOKEN` | `release` | *(optional)* PAT with write access to `confire-ai/homebrew-confire` |

Create a `production` GitHub Environment (Settings → Environments) to gate the two manual workflows. Add environment-level secrets there if you want separate tokens for prod.

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
| `STRIPE_TOPUP_PRICE_ID` | Stripe Dashboard → Products → top-up price ID (one-time $5 / 5,000 credits) |
| `SUPABASE_WEBHOOK_SECRET` | Generate: `openssl rand -base64 32` — use same value in the Supabase webhook header |
| `AMPLITUDE_KEY` | Amplitude → Settings → Projects → API Key |

---

## Feature flags

Feature flags live in `[vars]` in `wrangler.toml`. They can be overridden live without a redeploy:

```bash
wrangler secret put OPTIMIZER_API_ENABLED --env production
# enter: true
```

| Flag | Default | Effect |
|---|---|---|
| `OPTIMIZER_API_ENABLED` | `"false"` | Enables `POST /v1/optimize` (standalone optimizer API). Returns 404 while disabled. |

---

## CLI

### Local builds

```bash
cd cli

# confire — prod APIs
make dev          # build with symbols (fast)
make install      # install to GOPATH/bin

# confire-dev — dev APIs (api-dev.confire.dev / dev.confire.dev)
make build-dev-env    # build ./confire-dev
make install-dev-env  # install to GOPATH/bin/confire-dev

# Release build (garble obfuscation + stripped symbols)
make release
# Requires garble: go install mvdan.cc/garble@latest
```

### Self-update

```bash
confire update
```

Fetches `get.confire.dev/latest.json`, downloads the binary for the current platform, stops the daemon, atomically replaces the executable, and restarts the daemon. No-ops on `confire-dev` builds (no release channel).

### Tagged release (CI)

The `bump-version.yml` workflow runs on every push to `main`, reads conventional commits since the last tag, bumps semver, updates `CHANGELOG.md`, commits, and pushes a new tag. The tag triggers `release.yml`:

| Job | What it does |
|---|---|
| `build` | Cross-compile `confire_{darwin,linux}_{amd64,arm64}` |
| `release` | Create GitHub Release + upload binaries and checksums |
| `publish-r2` | Upload binaries, `latest.json`, and `install.sh` to R2 |

If `HOMEBREW_TAP_TOKEN` is set, the `release` job also bumps `homebrew/confire.rb`.

To cut a release manually:

```bash
git tag v1.2.3
git push origin v1.2.3
```

---

## CLI distribution (get.confire.dev / releases.confire.dev)

| URL | Purpose |
|---|---|
| `https://get.confire.dev` | `install.sh` (curl pipe target) |
| `https://get.confire.dev/latest.json` | Version manifest (used by `confire update`) |
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

Verify:

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
# Live Worker logs
wrangler tail --env dev
wrangler tail --env production

# Errors only
wrangler tail --env production --format=pretty | grep -i "error\|failed\|500"

# Analytics Engine — Cloudflare Dashboard → Workers → Analytics Engine
```

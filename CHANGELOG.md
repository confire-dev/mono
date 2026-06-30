## v0.14.0 — 2026-06-30

### Fixes
- fix: gate security_events and provenance_events writes behind plan flags (3824caf)


---

## v0.13.0 — 2026-06-07

### Fixes
- fix: remove duplicate recordSecurityEvent from supabase.ts (218db63)
- fix(docs): correct all inline relative links (a2c827c)
- fix(platform): use user session client for key gen, drop service key dependency (047ace9)
- fix(docs): real flame logo, fix broken links, remove dark/light toggle (d9b9951)
- fix(docs): match platform colors exactly, fix favicon with confire flame SVG (518234b)
- fix(platform): pass correct DOCS_SITE_URL to docs build for dev and prod (649fbc8)
- fix(platform): redirect /docs to docs.confire.dev (68b7bda)
- fix(platform): callback via navigation not fetch, inline key gen, fix /docs crash (543c75f)
- fix(platform): use CF service binding for worker calls to avoid same-zone 1003 error (e2d311a)
- fix(platform): use runtime WORKER_URL for server-side auth, update CLI authorize icons (0441e18)
- fix(benchmark): skip tsc in build to unblock monorepo CI (64fd30e)
- fix(platform): make Supabase client null-safe to prevent Worker 1101 on missing env vars (799c194)
- fix(install): reconnect /dev/tty for setup when piped, style start message (b6e90b1)
- fix(platform): disable Cloudflare Images runtime binding to prevent potential 1101 (464f79a)
- fix(install): use $BINARY var in post-install instead of hardcoded 'confire' (8a355cc)
- fix(ci): name dev archives confire-dev_* to match install script BINARY name (ac9227a)
- fix(ci): build confire-dev binary for dev installs, fix BINARY name in dev install script (83c3240)
- fix(install): fix dev releases URL path, shorten license notice, terms points to platform (0fd323d)
- fix(platform): fix code block top alignment, force Shiki dark theme colors for contrast (95986f2)
- fix(platform): replace custom regex highlighter with Kumo CodeHighlighted (Shiki) (ee90da6)
- fix(platform): skip comment highlighting inside URLs (https://) (778e692)
- fix(platform): disable Cloudflare KV auto-session to fix Worker 1101 error (caa8512)


---

## v0.12.0 — 2026-06-04


---

## v0.11.0 — 2026-06-04

### Fixes
- fix(ci): add --remote flag to wrangler R2 uploads (5dba43c)
- fix(ci): add buildBinaryName=confire-dev to dev CLI build (22a7c3c)


---

## v0.10.0 — 2026-06-04

### Fixes
- fix(release): use tar.gz format, fix ldflags pkg path, add buildBuiltBy, fix installer URL (760441b)
- fix(platform): keep assets.directory in wrangler.json, only remove reserved binding name (0dd5661)
- fix(platform): strip reserved ASSETS binding from generated wrangler.json after build (8b822aa)
- fix(platform): use pages_build_output_dir, switch deploy to wrangler pages deploy (9c2b9d3)
- fix(worker): restrict CORS to own origins, disable unauthenticated admin sync route (13778d7)
- fix(platform): switch to Cloudflare adapter for Workers compatibility (d5235b5)


---

## v0.9.4 — 2026-06-03


---

## v0.9.3 — 2026-06-03


---

## v0.9.2 — 2026-06-03


---

## v0.9.1 — 2026-06-03


---

## v0.9.0 — 2026-06-03


---

## v0.8.3 — 2026-06-03


---

## v0.8.2 — 2026-06-03

### Fixes
- fix(worker): rename binding to name in env.dev and env.production ratelimits (fbba9c1)


---

## v0.8.1 — 2026-06-03

### Fixes
- fix(worker): ratelimits use name (not binding) and string namespace_id (e379d93)


---

## v0.8.0 — 2026-06-03


---

## v0.7.7 — 2026-06-03


---

## v0.7.6 — 2026-06-03

### Fixes
- fix(worker): use [[ratelimits]] (correct TOML key) with chosen namespace_ids (59bb03f)


---

## v0.7.5 — 2026-06-03


---

## v0.7.4 — 2026-06-03


---

## v0.7.3 — 2026-06-03

### Fixes
- fix(platform): revert favicon to transparent orange logo (ebe7074)


---

## v0.7.2 — 2026-06-03

### Fixes
- fix(platform): use black bg white logo PNG as favicon (3094ddc)
- fix(platform): use confire orange PNG as favicon (e429418)


---

## v0.7.1 — 2026-06-03

### Fixes
- fix(platform): use PNG with 2x srcSet for ConfireMark instead of SVG (ee4aa4f)


---

## v0.7.0 — 2026-06-03


---

## v0.6.0 — 2026-06-03


---

## v0.5.0 — 2026-06-03


---

## v0.4.0 — 2026-06-03


---

## v0.3.6 — 2026-06-03


---

## v0.3.5 — 2026-06-02


---

## v0.3.4 — 2026-06-02

### Fixes
- fix: dashboard Activity icon + honest home page copy (ac3d94c)


---

## v0.3.3 — 2026-06-02


---

## v0.3.2 — 2026-06-02


---

## v0.3.1 — 2026-06-02


---

## v0.3.0 — 2026-06-02


---

## v0.2.0 — 2026-06-02


---

## v0.1.1 — 2026-06-01


---

# Changelog

## v0.1.0 — 2026-06-01

### Fixes
- fix(cli): stop SessionStart status banner on healthy sessions (181ee09)
- fix(db): hide pro_annual plan from checkout (091a2f2)
- fix(ci): bump-version changelog when no git tags exist (fb56a70)
- fix(cli,docs): .env.example exception, precise privacy model, tighter docs (f4eee61)
- fix(cli,docs): review UX, silent SessionStart, accurate privacy model (8c3781b)
- fix(cli): handle object tool_response shapes for Bash, Read, WebFetch (04b4865)
- fix(cli,worker,platform): local stats, Write passthrough, CLI auth redirect (6551e57)
- fix(worker): two-step api key lookup (FK to auth.users, not profiles) (67f6307)



# Confire

Context firewall for AI coding agents. Reviews risky tool calls before they run and sanitizes tool output before it reaches the model.

## Monorepo layout

```
cli/        Go CLI + daemon — runs on the developer's machine
worker/     Cloudflare Worker — remote optimizer + account API (TypeScript)
platform/   Astro web app — dashboard, login, billing, pricing
docs/       Architecture, decisions, hook contract, internal ops
```

## Prerequisites

- Go 1.23+
- Node 22+ and pnpm 10+
- [wrangler](https://developers.cloudflare.com/workers/wrangler/install-and-update/) — `npm i -g wrangler`
- [garble](https://github.com/burrowers/garble) — `go install mvdan.cc/garble@latest` (release builds only)

## Local development

### 1. Install dependencies

```bash
pnpm install
```

### 2. Configure environment

```bash
# Worker
cp worker/.dev.vars.example worker/.dev.vars
# Fill in worker/.dev.vars with your Supabase + Stripe secrets

# Platform
cp platform/.env.example platform/.env.local
# Fill in platform/.env.local
```

### 3. Start the Worker locally

```bash
cd worker
pnpm dev
# → http://localhost:8787
```

### 4. Start the platform

```bash
cd platform
pnpm dev
# → http://localhost:4321
```

### 5. Build and run the CLI

```bash
cd cli
make dev                    # build ./confire binary
./confire setup             # detect AI tools, install hook
./confire login             # connect account (opens browser)
./confire daemon &          # start daemon (background)
./confire status            # verify everything is running
```

### 6. Run tests

```bash
# Go: unit + E2E tests
cd cli && go test ./...

# TypeScript: type-check
pnpm --filter worker run type-check
pnpm --filter platform run type-check

# Build check
cd cli && make dev
pnpm --filter worker run build
pnpm --filter platform run build
```

## How it works

Two pillars — Tool Firewall (PreToolUse) and Context Firewall (PostToolUse):

```
Claude Code is about to run a tool call
  ↓ Claude Code calls `confire hook` (stdin: PreToolUse JSON)
  ↓ hook → daemon — evaluates built-in + custom policy rules
  ↓ Risky action? → block / require review / warn
  ↓ daemon → hook → Claude Code (stdout: decision)

Claude Code finishes a tool call
  ↓ Claude Code calls `confire hook` (stdin: PostToolUse JSON)
  ↓ hook → daemon — redact secrets, strip injections, trim noise
  ↓ Labels each output with a provenance trust level
  ↓ Detects cross-tool flow risks (e.g. untrusted read → shell exec)
  ↓ daemon → hook → Claude Code (stdout: sanitized output)
Claude sees clean, safe context instead of raw tool output
```

Local (no API key): full tool firewall + context sanitization — free, offline, unlimited.  
Remote (API key + daemon): custom rules synced from the dashboard, security event history, and team policy.

## Project docs

- [Architecture](docs/architecture.md)
- [Decisions log](docs/decisions.md)
- [Claude Code hook contract](docs/claude-code-hooks.md) — verified from binary
- [Supabase schema](docs/supabase-schema.sql)
- [Internal operations](docs/internal/README.md)

## Deploy

See [docs/internal/deploy.md](docs/internal/deploy.md) for the full deployment checklist.

Quick deploy:
```bash
cd worker && wrangler deploy
```

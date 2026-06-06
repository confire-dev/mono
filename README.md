# Confire

AI agent context optimizer. Intercepts tool call responses before they reach the model, strips noise, returns lean signal.

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

```
Claude Code finishes a tool call
  ↓ Claude Code calls `confire hook` (stdin: tool response JSON)
  ↓ hook → daemon (unix socket)
  ↓ daemon → Worker (HTTP/2, if API key present)
  ↓ Worker optimizes: Figma 98%, GitHub PR 87%, generic fallback
  ↓ daemon → hook → Claude Code (stdout: hookSpecificOutput.updatedToolOutput)
Claude sees lean context instead of noisy raw tool output
```

Local (no API key): Bash/Read/WebFetch/Generic optimizers only — free, offline, unlimited.  
Remote (API key + daemon): all platform optimizers — Figma, GitHub, Jira, Slack, etc.

## Project docs

- [Architecture](docs/architecture.md)
- [Decisions log](docs/decisions.md)
- [Claude Code hook contract](docs/claude-code-hooks.md) — verified from binary
- [Supabase schema](docs/supabase-schema.sql)
- [Internal operations](docs/internal/README.md)

## Legal & security

- [CONTRIBUTING.md](CONTRIBUTING.md)
- [SECURITY.md](SECURITY.md) — vulnerability disclosure
- [PRIVACY.md](PRIVACY.md) — what data we collect and why

## Deploy

See [docs/internal/deploy.md](docs/internal/deploy.md) for the full deployment checklist.

Quick deploy:
```bash
cd worker && wrangler deploy
```

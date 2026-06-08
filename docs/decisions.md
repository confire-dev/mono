# Architecture Decision Records

## ADR-001: PostToolUse hook over MCP proxy

**Decision:** Use Claude Code's PostToolUse hook as the primary integration mechanism, not an MCP stdio/HTTP/SSE proxy.

**Rationale:**
- Hook fires on ALL tools (Bash, Read, Edit, WebFetch, all MCP) — the proxy only sees MCP tools.
- Hook handles both output (PostToolUse) and input (PreToolUse / `updatedInput`) — proxy only handles responses.
- Hook is a dead-simple stdin/stdout protocol — proxy fights transport complexity (SSE endpoint rewrites, session correlation, port management).
- The Go proxy code is kept in `reference/` as the `mcp-proxy` strategy for non-Claude-Code hosts.

## ADR-002: Go for CLI + daemon, TypeScript for Worker

**Decision:** Go local, TypeScript on Cloudflare.

**Rationale:**
- Go: ~1-5ms cold start (critical for a per-call hook process), static binary, native OS keychain (`go-keyring`), excellent long-lived daemon.
- TypeScript: runs natively in Cloudflare's V8 isolate (no WASM overhead), same language as leanmcp's 12 existing optimizers we're reusing.
- Optimizer logic lives in both: TS Worker = full/server-updatable suite; Go = thin offline fallback.
- Shared contract = `InterceptEvent` JSON — language-agnostic.

## ADR-003: Canonical phase abstraction

**Decision:** Define host-agnostic canonical phases (`tool.post`, `tool.pre`, `session.start`, …) that adapters map native events onto.

**Rationale:**
- A Figma optimizer registered for `tool.post` works for Claude Code, Cursor, Cline, or any future host.
- New hosts = one adapter + one capability-matrix entry. Zero core changes.
- Handlers declare which phases they want; the Engine only calls them when matched.

## ADR-004: pre-compact disabled in v1 — ships dark

**Decision:** The `context.pre-compact` phase slot exists and is mapped, but no Handler runs for it in v1. Setup shows it as "Soon (disabled)".

**Rationale:**
- PreCompact replaces transcript state. A bad compaction = invisible context loss / session corruption.
- Bounded blast radius (one tool result wrong → recoverable next call) vs unbounded (transcript corrupted → session broken).
- Users forgive "not improved yet", not "made it worse".
- Enable week 2 after output/input phases are proven stable.

## ADR-005: Transport is a pluggable interface

**Decision:** `Transport { Send(InterceptEvent) InterceptResult }` with `LocalTransport`, `WorkerTransport`, `FallbackTransport`.

**Rationale:**
- v1 default = `LocalTransport` (in-process, offline, instant, testable with zero infra).
- Cloud upgrade = config flip to `FallbackTransport(Worker, Local)`. Same envelope, no code change.
- `FallbackTransport` ensures we never return raw output on network failure.
- Future open-source seam: third parties publish `Transport` implementations.

## ADR-006: OAuth via daemon, OS keychain

**Decision:** `confire login` opens the browser for GitHub OAuth / email magic link → callback → `go-keyring` stores the key.

**Rationale:**
- Zero copy-paste, zero env vars, zero friction.
- OS keychain survives reboots, is OS-encrypted, never in config files.
- The daemon owns the token; the hook shim stays dumb.

## ADR-007: Free tier never blocks

**Decision:** Local firewall and context passes always run, regardless of account state. Never degrade the developer's flow.

**Rationale:**
- A tool that breaks the developer's flow will be uninstalled immediately.
- All security passes (secret redaction, injection detection, unicode stripping) and local context passes run free and offline.
- Cloud policy sync, custom rules, and the security event dashboard require a paid plan — but these are additive, never blockers.

## ADR-008: SessionStart does three things in one call

**Decision:** SessionStart hook → daemon POST `/session/start` → returns `{notification, rulesVersion, pullRules}`.

**Rationale:**
- Warms the HTTP/2 connection (first call of the session; stays warm for all subsequent calls).
- Delivers notification to Claude as `additionalContext` (one line max — must not add noise).
- Checks for firewall rule updates (`pullRules → fetch latest from Worker`).
- Three birds, one stone, one network round-trip at session start.

## ADR-009: Context processing boundary — Local (CLI) vs Remote (Worker)

**Decision:** Two explicit processing layers; the split is WHERE, not WHO PAYS.

```
cli/optimizer/      Local context passes — infrastructure tools, always free
  bash.go           Shell command output: trim logs, keep failures
  read.go           File reads: emergency cap only
  webfetch.go       Web pages: strip HTML/CSS/JS boilerplate
  generic.go        Universal JSON noise stripping (fallback)

worker/             Cloud backend — policy sync, telemetry, custom rules
  handlers/         No optimizer logic post-pivot; firewall telemetry only
```

**Rationale:** After the security pivot, the Worker is the policy sync and
telemetry backend, not an optimization engine. Local context passes remain
in the CLI daemon as a secondary benefit of the Context Firewall — security
passes always run first, context cleanup runs after.

**Processing decision logic:**
```
Tool response arrives at daemon
  ↓
Security passes (MCP only): redaction → unicode strip → injection detect
  ↓
Local context passes (all tools): bash/read/webfetch/generic noise trim
  ↓
Return processed output to hook
```

**Product boundary:**
- Free/local: full security pipeline + local context passes
- Paid/cloud: custom firewall rules, security event history, provenance tracking, policy sync

## ADR-010: Repo layout — cli/, worker/, apps/, docs/

**Decision:** Rename `confire/` → `cli/` to make the product model obvious from the directory structure.

```
cli/        Local CLI + daemon (Go). Runs on the developer's machine.
worker/     Cloud policy sync + telemetry API (TypeScript on Cloudflare).
apps/web/   Dashboard/login/billing (future).
docs/       Architecture and decisions.
```

The Go module path remains `github.com/confire-dev/confire` — only the directory
name changed. All internal imports are unaffected.

## ADR-011: Supabase = truth, Amplitude = insight, Worker = boundary

**Decision:** Three-layer data architecture with explicit, non-overlapping responsibilities.

```
CLI daemon
  ↓  POST /v1/events (validated, signed)
Cloudflare Worker
  ↓ writes exact values          ↓ forwards bucketed, sanitized values
Supabase                         Amplitude
(billing truth,                  (behavioral analytics,
 user dashboard,                  funnel analysis,
 credits,                         cohort studies,
 sessions,                        feature adoption)
 audit logs)
```

**Rules:**
- Supabase stores anything the user can see, pay for, dispute, or rely on.
- Amplitude stores behavioral/product funnel events only — never raw values.
- The Worker sanitizes and buckets before forwarding to Amplitude.
- `telemetry = false` disables Amplitude forwarding. It does NOT disable
  Supabase writes (required for billing, dashboard, limits).
- Never use Amplitude as the source for user-facing usage data.

**`analytics_consented` flag:** Sent in every `/v1/events` payload.
  - `true`  → Worker writes Supabase + forwards to Amplitude
  - `false` → Worker writes Supabase only (usage accounting still required)

## ADR-012: Cloudflare Workers Rate Limiting API for per-user throttling

**Decision:** Use the Cloudflare Workers Rate Limiting API (one binding per plan tier) for per-minute
request throttling. Do not store rate limit state in Supabase or KV.

**Rationale:**
- Workers Rate Limiting runs at the Cloudflare edge in ~1ms with no external calls — adding it
  between auth and the Supabase usage query adds zero meaningful latency.
- One fixed binding per tier (`RL_FREE`, `RL_DEV`, `RL_PRO`) maps cleanly to plan groups.
  The plan_id→tier mapping lives in code (two lines); no DB lookup needed.
- Keeping rate limits in `wrangler.toml` (not the plans DB) separates two concerns:
  the DB controls what users can access; wrangler controls how fast they can access it.
  Changing rate limits is a deploy, not a DB edit — intentionally requires a code review.
- Alternatives considered:
  - KV sliding window: adds ~2ms write, complex rollover logic.
  - Durable Object: precise but heavyweight for per-minute limits; overkill.
  - IP-based Cloudflare firewall rules: don't distinguish plans; bypass-able behind proxies.

**Rate limits (v1):**

| Plan tier               | Requests per minute |
|-------------------------|---------------------|
| Free                    | 20                  |
| Dev / Dev Annual        | 60                  |
| Pro / Pro Annual / Enterprise | 200           |

**Local optimizer is never rate-limited** — only remote calls to `POST /api/optimize` are counted.

## ADR-013: Standalone Optimizer API as a thin adapter over the existing engine

**Decision:** Expose `POST /v1/optimize` as a general-purpose pre-LLM context reduction API.
The endpoint is a thin adapter — it wraps the caller's content in a synthetic `InterceptEvent`
and calls `handle()`, the same function used by the Claude Code hook path.

**Rationale:**
- The optimizer functions (`optimizeFigma`, `optimizeGeneric`, `optimizeWebFetch`, …) already
  take a raw string. They have no dependency on Claude Code specifics — the hook pipeline is
  just one way to feed them content.
- A `type` hint (`"json"`, `"html"`, `"github"`, …) maps to a tool name, which routes the
  synthetic event through `dispatch()` exactly as if it came from a real hook.
- Zero new optimizer code. Zero new plan/entitlement code. Zero new rate limit code.
  The only addition is the handler (~140 lines) and the new request/response types.
- Sharing the same quota counter as hook-based calls in v1 keeps the billing system simple.
  A separate `apiOptimizationsMonthly` limit can be added if use cases diverge.

**Product angle:** "Call us before calling Groq. Pay us once, pay the LLM less every time."
The endpoint is meaningful at every tier: Free users get 500 pre-LLM optimizations/month;
Pro users get effectively unlimited. See `docs/internal/optimizer-api.md` for full details.

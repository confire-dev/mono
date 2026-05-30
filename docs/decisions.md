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

**Decision:** At 100% of free tier, fall back to bundled local optimizer (reduced coverage) + show upgrade nudge. Never block.

**Rationale:**
- A tool that breaks the developer's flow will be uninstalled immediately.
- Reduced optimization is acceptable; zero optimization is not; broken flow is unacceptable.
- Upgrade warning at 80% (one line), at 100% (one line + local fallback). Always shown regardless of notification settings.

## ADR-008: SessionStart does three things in one call

**Decision:** SessionStart hook → daemon POST `/session/start` → returns `{notification, rulesVersion, pullRules}`.

**Rationale:**
- Warms the HTTP/2 connection (first call of the session; stays warm for all subsequent calls).
- Delivers notification to Claude as `additionalContext` (one line max — confire is an optimizer, must not add noise).
- Checks for optimizer rule updates (`pullRules → fetch latest from Worker`).
- Three birds, one stone, one network round-trip at session start.

## ADR-009: Optimizer boundary — Local (CLI) vs Remote (Worker)

**Decision:** Two explicit optimizer layers; the split is WHERE, not WHO PAYS.

```
cli/optimizer/      Local optimizer — infrastructure tools, free tier
  bash.go           Shell command output: trim logs, keep failures
  read.go           File reads: cap at 500 lines before read, dedupe
  webfetch.go       Web pages: strip HTML/CSS/JS boilerplate
  generic.go        Universal JSON noise stripping (fallback)

worker/src/optimizers/   Remote optimizer — paid cloud tier
  figma.ts          JSX → section map (98% reduction)
  github.ts         PR objects, diffs, bot comment stripping (87%)
  atlassian.ts      Jira ADF → markdown, Confluence (70%/96%)
  slack.ts, notion.ts, clickup.ts, amplitude.ts, …
```

**Rationale:** Platform-specific optimizers (Figma, GitHub, Jira, Slack, etc.) are the
primary paid value proposition. Keeping them in the Worker means:
  1. They update server-side without a CLI release.
  2. The binary doesn't contain paid logic that could be extracted.
  3. The free tier has real utility (log trimming, file capping) without
     giving away the paid optimizers.

**Naming rule:** Always use `LocalOptimizer`/`RemoteOptimizer` in code —
never `FreeOptimizer`/`PaidOptimizer`. Entitlement (who can use Remote) is
decided by the daemon checking the keychain for a valid API key. The engine
is completely mode-agnostic. An enterprise self-hosting the Worker gets
Remote optimization at any price point.

**Transport decision logic:**
```
Tool response arrives at daemon
  ↓
Is there a LocalOptimizer for this tool? (Bash/Read/WebFetch/Generic)
  → yes: optimize locally, return immediately
  → no (platform-specific MCP tool):
      Is user logged in (API key in keychain)?
        → yes: send to WorkerTransport → FallbackTransport
        → no:  return passthrough (no optimization, no error)
```

**Product boundary message:**
- Free/local: basic context cleanup for shell, files, web output
- Paid/remote: high-impact platform-specific context optimization

## ADR-010: Repo layout — cli/, worker/, apps/, docs/

**Decision:** Rename `confire/` → `cli/` to make the product model obvious from the directory structure.

```
cli/        Local CLI + daemon (Go). Runs on the developer's machine.
worker/     Paid cloud optimizer + account API (TypeScript on Cloudflare).
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

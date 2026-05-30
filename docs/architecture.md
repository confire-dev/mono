# Confire Architecture

## Three-axis abstraction

Confire's core is blind to **how** it's invoked, **which host** sent the event, and **which mechanism** the host uses. Those are three independent axes:

### Axis 1 — Host
Which AI coding agent is being optimized. `claude-code` today; `cursor`, `cline`, `windsurf`... are new entries in the capability matrix. Each declares which strategies it supports.

### Axis 2 — Strategy  
How events are extracted from the host:
- **`hooks`** — Native lifecycle hooks. Claude Code only (so far). Richest: all 29 phases, rewrites both output (`updatedToolOutput`) and input (`updatedInput`).
- **`mcp-proxy`** — Intercept at the MCP transport layer. Universal: any MCP-speaking host. Limited to tool pre/post phases.
- Future: editor-extension API, LSP middleware.

### Axis 3 — Phase (canonical, host-agnostic)
Adapters map native events onto these names. Optimizers register against phases, not native event names.

| Canonical phase       | Claude Code event     | MCP-proxy  |
|-----------------------|-----------------------|------------|
| `session.start`       | SessionStart          | initialize |
| `session.end`         | SessionEnd            | close      |
| `tool.pre`            | PreToolUse            | request    |
| `tool.post`           | PostToolUse           | response   |
| `tool.batch.post`     | PostToolBatch         | —          |
| `turn.stop`           | Stop                  | —          |
| `context.pre-compact` | PreCompact (disabled) | —          |
| `prompt.submit`       | UserPromptSubmit      | —          |

## Envelope contract

The `InterceptEvent` / `InterceptResult` JSON pair is the **only thing shared** between the Go daemon and the TS Worker. Changing the shape requires updating both `confire/intercept/types.go` and `worker/src/types.ts`.

```
InterceptEvent  { host, strategy, phase, session{id,cwd,transcriptPath},
                  tool?{name,input,output,useId,isMcp,mcpServer,durationMs}, prompt?, raw }

InterceptResult { kind: passthrough|replace-output|replace-input|add-context|compact,
                  toolOutput?, toolInput?, context?, stats?{beforeBytes,afterBytes,optimizer} }
```

## Handler registration

```
Handler { ID(); Phases() []Phase; Matches(InterceptEvent) bool; Run(InterceptEvent) InterceptResult }
```

A Figma optimizer: `Phases: ["tool.post"], Matches: name contains "figma"`.  
Adding a new optimizer = one new file + one registry entry. No core changes.

## Transport (pluggable)

```
Transport { Send(InterceptEvent) → InterceptResult }
```

- `LocalTransport` — runs bundled Go core in-process. Zero network. ~1ms. Used in v1 default, tests, offline fallback.
- `WorkerTransport` — HTTP/2 POST to the Cloudflare Worker. Warm, multiplexed. Owned by the daemon.
- `FallbackTransport(primary, secondary)` — tries primary, falls back to secondary on error. Never returns raw output.

## Polyglot split

```
Go (CLI + daemon):       local hot path, ~1-5ms cold start, native keychain
TypeScript (Worker):     full optimizer suite, server-updatable, V8-native on Cloudflare
```

Optimizer logic lives in both: TS Worker = full/primary suite; Go local = thin offline fallback (figma + generic + bash). Same `InterceptEvent` JSON is the wire format.

## Monorepo structure

```
confire-mono/
├── confire/          Go CLI + daemon
│   ├── cmd/          confire setup, hook, login, status, …
│   ├── intercept/    canonical types: InterceptEvent, InterceptResult, Phase, Handler, Engine
│   ├── hosts/        ClaudeCodeHooksAdapter; mcp-proxy iface
│   ├── transport/    LocalTransport, WorkerTransport (http2), FallbackTransport
│   ├── optimizer/    Go fallback: figma, generic, helpers, registry
│   └── auth/         OAuth callback server, OS keychain (go-keyring)
├── worker/           Cloudflare Worker (TypeScript)
│   └── src/
│       ├── index.ts          router
│       ├── engine.ts         phase router
│       ├── types.ts          InterceptEvent / InterceptResult (TS mirror)
│       └── optimizers/       figma, generic, bash, read, webfetch + registry
├── apps/             FUTURE: landing/, dashboard/
└── docs/             this file; decisions.md; claude-code-hooks.md
```

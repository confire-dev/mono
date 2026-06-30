# Client Host Feature Set

Ground truth for what Confire actually does per supported client. Update when capabilities change.

**Last updated:** June 2026

---

## Integration strategies

| Strategy | Mechanism | Clients |
|----------|-----------|---------|
| `hooks` | Client lifecycle hooks (PreToolUse, PostToolUse, SessionStart, SessionEnd) | Claude Code, Cursor, VS Code |
| `mcp-proxy` | Intercepts at MCP transport layer; only tools routed through Confire's MCP server | Cline*, Windsurf*, Codex CLI* |

\* Coming soon

---

## Per-client capability table

| Capability | Claude Code | Cursor | VS Code | Cline | Windsurf | Codex CLI |
|------------|:-----------:|:------:|:-------:|:-----:|:--------:|:---------:|
| **Status** | Live | Live | Live | Coming soon | Coming soon | Coming soon |
| **Strategy** | hooks | hooks | hooks | mcp-proxy | mcp-proxy | mcp-proxy |
| **PreToolUse firewall** (block/review before tool runs) | ✅ | ✅ | ⚠️ unverified | ❌ | ❌ | ❌ |
| **PostToolUse native output replacement** (shrink/sanitize actual tool output) | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **OptimizeNative** (local tools: Bash, Read, WebFetch) | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **OptimizeMCP** (MCP-routed tool output) | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **PostToolSteer** (security/guidance hints via `additional_context`) | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **SessionStart context injection** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **SessionEnd cleanup** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |

### PreToolUse block behavior by client

**Claude Code** — confirmed hard block. `hook.go` calls `os.Exit(2)` after writing `{"decision":"block"}` to stdout.

**Cursor** — confirmed from code: `EncodeCursorPreToolResult` returns `{permission: "deny"}` and `hook.go` calls `os.Exit(2)` when `shouldBlock=true`. **Needs one live E2E test** to confirm Cursor actually halts on exit 2.

**VS Code** — code path sets `permissionDecision: "deny"` in `hookSpecificOutput` but always writes `Continue: true`. No `os.Exit(2)` is called. **Whether VS Code Copilot treats `continue: true` + `permissionDecision: "deny"` as a halt is unconfirmed from code alone.** A live test is required before claiming VS Code as a supported block surface in user-facing docs.

### Notes on Cursor / VS Code native tool coverage

The hooks strategy fires on all registered PreToolUse events regardless of whether the tool is MCP-routed or native. However, native tools (Bash, shell commands, file reads) run inside the client — Confire intercepts the PreToolUse decision but does not replace PostToolUse output for these tools. PostToolSteer injects `additional_context` security hints on supported hooks.

---

## Hook event coverage (Claude Code only)

Claude Code exposes 29 hook event types. Confire registers 4:

| Hook | Registered | What Confire does |
|------|-----------|-------------------|
| `PreToolUse` | ✅ | Firewall: review/block risky tool calls before they run |
| `PostToolUse` | ✅ | Optimize + sanitize tool output; inject PostToolSteer hints |
| `SessionStart` | ✅ | Silent health check; inject system context if needed |
| `SessionEnd` | ✅ | Cleanup hook (no-op on healthy sessions) |
| `Stop` / `StopFailure` | ❌ not registered | Mapped in `MapPhase` if received; not actively installed |
| All others (27) | ❌ | Not handled |

---

## Observations / future work

- **PreToolUse for Cursor/VS Code:** Not possible with mcp-proxy — clients don't expose a pre-call hook at the MCP transport layer. Would require client-side SDK support or a dedicated sidecar.
- **Native tool coverage for Cursor/VS Code:** No path to intercept Bash/shell/read output in these clients today. PostToolSteer is the only security lever. Flag this gap clearly in user-facing docs.
- **Stop/StopFailure hooks (Claude Code):** `MapPhase` handles them but they are not registered in `settings.json`. Consider registering for turn-level analytics or guardrail summaries.
- **SubagentStop / PostToolBatch (Claude Code):** Not handled. Could be valuable for subagent-level firewall summaries.
- **Cline / Windsurf / Codex CLI:** All in registry as `ComingSoon`. MCP proxy strategy already defined; needs install/detect implementation and E2E tests.
- **UserPromptSubmit hook (Claude Code):** Could enable prompt-level injection detection before any tool runs — not implemented.
- **PreCompact hook (Claude Code):** Reserved in `intercept/types.go` as `PhasePreCompact` but explicitly marked DISABLED in v1. Re-evaluate for context budget guardrails.

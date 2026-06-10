# Client Host Feature Set

Ground truth for what Confire actually does per supported client. Update when capabilities change.

**Last updated:** June 2026

---

## Integration strategies

| Strategy | Mechanism | Clients |
|----------|-----------|---------|
| `hooks` | Native lifecycle hooks (PreToolUse, PostToolUse, SessionStart, SessionEnd) | All supported clients |
| `mcp-proxy` | Intercepts at MCP transport layer; only tools routed through Confire's MCP server | Reserved — no active clients |

All clients now use `StrategyHooks`. The `mcp-proxy` strategy is reserved for
future MCP-only clients that do not expose a native hook API.

---

## Per-client capability table

| Capability | Claude Code | Cursor | VS Code | Windsurf | Codex CLI | Cline | OpenCode | OpenClaw |
|------------|:-----------:|:------:|:-------:|:--------:|:---------:|:-----:|:--------:|:--------:|
| **Status** | Live | Live | Live | Live | Live | Live | Live | Live |
| **Strategy** | hooks | hooks | hooks | hooks | hooks | hooks (JS bridge) | hooks (JS bridge) | hooks (JS bridge) |
| **PreToolUse firewall** (block/review before tool runs) | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **PostToolUse native output replacement** | ✅ | MCP only | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **PostToolSteer** (security context via additionalContext) | ✅ | ✅ | ✅ | ❌ | ✅ | ✅ | ✅ | ✅ |
| **SessionStart context injection** | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ |
| **SessionEnd cleanup** | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ |

---

## Hook config file locations

| Client | Global | Local |
|--------|--------|-------|
| Claude Code | `~/.claude/settings.json` | `.claude/settings.json` |
| Cursor | `~/.cursor/hooks.json` | `.cursor/hooks.json` |
| VS Code | `~/.copilot/hooks/confire.json` | `.github/hooks/confire.json` |
| Windsurf | `~/.windsurf/hooks.json` | `.windsurf/hooks.json` |
| Codex CLI | `~/.codex/hooks.json` | `.codex/hooks.json` |
| Cline | `~/.../globalStorage/saoudrizwan.claude-dev/plugins/confire.js` | — |
| OpenCode | `~/.config/opencode/plugins/confire.js` | `.opencode/plugins/confire.js` |
| OpenClaw | `~/.openclaw/plugins/confire.js` | `.openclaw/plugins/confire.js` |

---

## Hook detection in `confire hook`

The `confire hook` command dispatches to the correct host handler using the
`--host` flag (registered in each client's config). Existing Claude Code,
Cursor, and VS Code installs use backward-compatible payload auto-detection.

| Method | Clients |
|--------|---------|
| `--host windsurf` flag | Windsurf |
| `--host codex` flag | Codex CLI |
| `--host cline` flag | Cline (via JS bridge) |
| `--host opencode` flag | OpenCode (via JS bridge) |
| `--host openclaw` flag | OpenClaw (via JS bridge) |
| `cursor_version` field in payload | Cursor (backward compat) |
| `sessionId` camelCase + no `session_id` + no `conversation_id` | VS Code (backward compat) |
| Default | Claude Code |

---

## JS bridge clients (Cline, OpenCode, OpenClaw)

These clients use a TypeScript/JS plugin SDK rather than a CLI hook binary.
Confire installs a small bridge plugin (`confire.js`) that:

1. Registers the client's native hook events.
2. Calls `confire hook --host <client>` as a subprocess with the event JSON on stdin.
3. Returns the result in the client's expected format.

The bridge plugin is embedded in the Confire binary as a string constant and
written to the client's plugin directory by `confire setup` / `confire install`.

**Limitation:** JS bridge clients do not support SessionStart/SessionEnd because
those events are not mapped in the bridge plugin. The tool firewall and post-tool
steer (additionalContext) work normally.

---

## Observations / future work

- **Windsurf PostToolSteer:** Current PostToolUse response only supports `decision/reason`. If Windsurf adds `additionalContext` support, update `EncodeWindsurfResult` and set `PostToolSteer: true` in `capabilities.go`.
- **Codex `updatedToolOutput`:** Codex may support output replacement — verify against live binary. If confirmed, set `NativeOutputReplaceable: true` for `codex` in `capabilities.go`.
- **Cline `beforeTool`/`afterTool` context shape:** Verify exact field names against the Cline SDK before 1.0 release — the bridge plugin uses `context.sessionId`, `context.cwd`, `context.tool`, `context.input`, `context.output`, `context.additionalContext`.
- **OpenClaw plugin registration:** Verify the exact plugin directory path and whether `api.on` is the correct registration method.
- **UserPromptSubmit hook (all clients):** Could enable prompt-level injection detection. Not implemented.
- **Windsurf config path:** The `docs.devin.ai` format uses `.devin/hooks.v1.json`. If Windsurf ships its own config path, update `windsurfHooksPath()` in `windsurf.go`.

---
title: Client support
description: Confire supports Claude Code, Cursor, VS Code, Windsurf, Codex CLI, Cline, OpenCode, and OpenClaw.
---

Confire integrates with supported AI coding agents using each client's native
hook system. All supported clients get pre-tool firewall evaluation and
post-tool result inspection. Capabilities vary by client.

## Supported clients

| Client | Status |
|---|---|
| Claude Code | Supported |
| Cursor | Supported |
| VS Code (Copilot) | Supported |
| Windsurf | Supported |
| Codex CLI | Supported |
| Cline | Supported |
| OpenCode | Supported |
| OpenClaw | Supported |

## Capability comparison

| Capability | Claude Code | Cursor | VS Code | Windsurf | Codex CLI | Cline | OpenCode | OpenClaw |
|---|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| Pre-tool firewall (block before execution) | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Post-tool result inspection | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Security context returned to agent | ✅ | ✅ | ✅ | ✗ | ✅ | ✅ | ✅ | ✅ |
| Native tool output replacement | ✅ | MCP only | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ |
| Session start context injection | ✅ | ✅ | ✅ | ✅ | ✅ | ✗ | ✗ | ✗ |
| Integration mechanism | hooks | hooks | hooks | hooks | hooks | plugin | plugin | plugin |

**Security context returned to agent** means Confire can include a brief
explanation alongside a review or block decision. For example, describing
why a command was flagged and how to proceed. Windsurf's hook API is
one-directional: Confire can block or allow a tool call but cannot inject
additional text back into the conversation. The firewall still works
fully; the agent just won't see a reason message.

**Session start context injection** means Confire sends a short system
message at the start of each session telling the agent that the firewall
is active and what mode is in effect. This helps the agent handle review
and block decisions correctly. Plugin-based clients (Cline, OpenCode,
OpenClaw) don't expose a session-start hook, so they skip this step.
The firewall still evaluates all tool calls normally.

**Native tool output replacement** means Confire can modify the actual content
the agent receives from a tool result. Claude Code supports this for all tools.
Cursor supports it for MCP-routed tools only. Other clients receive security
context alongside the original output instead.

**Plugin clients** (Cline, OpenCode, OpenClaw) use a small bridge plugin
installed by `confire setup`. The plugin proxies hook events to the Confire
daemon using the same evaluation engine as all other clients.

## Check your setup

```bash
confire status
```

The **Agents** section shows which clients were detected and whether Confire
is connected.

## How `confire setup` works

`confire setup` detects supported clients and installs the right integration
for each one. No manual configuration needed.

```bash
confire setup
```

For hook-based clients (Claude Code, Cursor, VS Code, Windsurf, Codex),
Confire writes an entry to the client's hooks config file. For plugin clients
(Cline, OpenCode, OpenClaw), Confire writes a bridge plugin to the client's
plugin directory.

All hook evaluation runs locally on your machine. The Confire daemon handles
policy decisions; the client receives the result.

---
title: Per-agent settings
description: >-
  Override Confire's default behavior for a specific
  AI coding agent.
---

Confire's default behavior differs per agent based on what each
agent's hook API supports. You can override these defaults in
your config file using the `hosts.<agent-id>.*` keys.

## Default capabilities

| Capability | Claude Code | Cursor | VS Code |
|---|---|---|---|
| `optimize_native` | `true` | `false` | `false` |
| `optimize_mcp` | `true` | `true` | `true` |
| `post_tool_steer` | `true` | `true` | `true` |

`optimize_native` controls whether Confire optimizes native tool
output (Bash, Read, WebFetch) for that agent. It's off by default
for Cursor and VS Code because those agents don't support output
replacement — only Claude Code can receive a rewritten response.

## Overriding defaults

```bash
# Disable MCP optimization for Cursor only
confire config set hosts.cursor.optimize_mcp=false

# Disable steering context injection for VS Code
confire config set hosts.vscode.post_tool_steer=false

# Enable native optimization for Cursor (steer-mode only)
confire config set hosts.cursor.optimize_native=true
```

Agent IDs: `claude-code`, `cursor`, `vscode`.

:::note
Enabling `optimize_native` for Cursor or VS Code turns on
optimization for those tools, but the agent will receive
savings via `additional_context` steering rather than output
replacement — the original output is still visible. This can
still reduce context noise through the steering envelope.
:::

## Checking effective capabilities

```bash
confire status
```

The "Agents" section shows which agents are detected and what
hooks are active.

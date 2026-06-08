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
| `native_output_replaceable` | `true` | `false` | `false` |
| `post_tool_steer` | `true` | `true` | `true` |

**`native_output_replaceable`** — whether Confire can replace the
tool output that the agent receives. Only Claude Code's hook API
supports output replacement. For Cursor and VS Code, processed
results are delivered via the `additional_context` steering envelope
instead.

**`post_tool_steer`** — whether Confire injects a standardized
`[Confire post_tool steer]` block into `additional_context` after
each tool call. Enabled by default on all supported agents.

## Overriding defaults

```bash
# Disable steering context injection for VS Code
confire config set hosts.vscode.post_tool_steer=false

# Disable steering for Cursor
confire config set hosts.cursor.post_tool_steer=false
```

Agent IDs: `claude-code`, `cursor`, `vscode`.

## Checking effective capabilities

```bash
confire status
```

The "Agents" section shows which agents are detected and what
hooks are active.

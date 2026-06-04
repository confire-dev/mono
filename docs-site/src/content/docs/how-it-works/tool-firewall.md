---
title: Tool Firewall
description: Block, allow, or gate individual MCP tool calls with policy rules.
---

The tool firewall intercepts every MCP tool call your agent makes and evaluates it against your policy rules before the call reaches the actual MCP server.

## Why this matters

AI agents are eager. Given the opportunity to call a tool, they will — often more broadly than you intended. The firewall puts you back in control without needing to change your agent's configuration or the MCP servers it talks to.

## How a tool call flows

1. Agent sends a `tools/call` MCP message to the proxy
2. Policy engine checks it against your rules (first match wins)
3. One of three outcomes:
   - **allow** → forwarded to the real MCP server, result returned to agent
   - **block** → error returned immediately, real server never called
   - **require-approval** → user is prompted, then either allowed or blocked

## Matching rules

Rules match on tool name, server name, arguments, or combinations:

```yaml
rules:
  # Match any tool named "bash"
  - when:
      tool: bash
    do: block

  # Match tools from a specific server
  - when:
      server: filesystem
      tool: write_file
    do: require-approval

  # Match on argument content (regex)
  - when:
      tool: bash
      args:
        command: "rm\\s+-rf"
    do: block
    reason: "Destructive rm -rf is not allowed"
```

## Approval flow

When a rule says `require-approval`, the proxy pauses the request and waits for your input:

```
┌──────────────────────────────────────────────┐
│  Tool call pending approval                  │
│                                              │
│  Tool:   write_file                          │
│  Server: filesystem                          │
│  Args:   { "path": "/etc/hosts", ... }       │
│                                              │
│  [Allow once]  [Always allow]  [Block]       │
└──────────────────────────────────────────────┘
```

"Always allow" creates a temporary exception for the duration of the session.

## Default behavior

If no rule matches, the default action is **allow**. To flip to a deny-by-default posture:

```yaml
version: 1
default: block

rules:
  - when:
      tool: read_file
    do: allow
  - when:
      tool: list_directory
    do: allow
```

## Audit log

Every tool call and its outcome (allowed / blocked / approved) is written to `~/.confire/audit.log`:

```
2025-01-15T10:23:01Z  ALLOW    read_file    {"path": "src/main.ts"}
2025-01-15T10:23:04Z  BLOCK    bash         {"command": "curl ..."}
2025-01-15T10:23:09Z  APPROVE  write_file   {"path": "README.md"}
```

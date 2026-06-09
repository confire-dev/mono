---
title: Client support
description: Confire supports Claude Code, Cursor, and VS Code.
---

Confire supports Claude Code, Cursor, and VS Code. Across supported
clients, Confire focuses on the same core behavior:

- review risky tool calls before they run,
- inspect tool results after they return,
- add firewall context back to the agent,
- record local security events.

## Supported clients

| Client | Support |
|---|---|
| Claude Code | Supported |
| Cursor | Supported |
| VS Code | Supported |

## What Confire does

Confire hooks into supported agent/tool events.

Before tools run, Confire can evaluate the tool call and return a
policy decision: `allow`, `warn`, `review`, or `block`.

After tools return, Confire can inspect the result and add firewall
context where supported, such as:

- secret-looking value warnings,
- prompt-injection warnings,
- hidden Unicode warnings,
- MCP risk notes,
- provenance metadata,
- suggested next steps.

## Client differences

Each agent exposes different hook capabilities, so exact behavior can
vary by client and tool type. In general:

- Claude Code supports the deepest hook integration.
- Cursor and VS Code support Confire integrations for supported tool
  events.
- Confire uses the best available mechanism in each client.

You do not need to choose a mode manually. `confire setup` detects
supported clients and installs the right integration.

## Check your setup

```bash
confire status
```

The **Agents** section shows which clients were detected and whether
Confire is connected.

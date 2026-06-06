---
title: Claude Code
description: Confire integrates with Claude Code via PostToolUse hooks.
---

Claude Code is the first fully supported agent. Confire hooks into Claude Code's **PostToolUse** system — every tool output is intercepted, optimized, and returned before the model sees it.

No MCP server, no proxy, no port to configure.

## Setup

```bash
confire setup
```

Installs the hook into `~/.claude/settings.json` (global) or `.claude/settings.json` (project). Restart Claude Code after.

## What gets installed

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "",
        "hooks": [{ "type": "command", "command": "confire hook" }]
      }
    ]
  }
}
```

The empty `matcher` means all tool types are intercepted. Confire routes each one to the appropriate optimizer automatically.

## Verify

```bash
confire status
```

## Remove

```bash
confire reset
```

Removes the hooks and stops the daemon. Binary stays installed.

---
title: Claude Code
description: Confire integrates with Claude Code via PostToolUse hooks.
---

## How it works

Confire uses Claude Code's **PostToolUse hook** — a built-in feature that lets external commands process tool outputs before they're returned to the model. When a tool call finishes, Claude Code pipes the raw output to `confire hook`, which sends it to the daemon for optimization and returns the stripped result.

No MCP server. No proxy. No changes to your existing MCP setup.

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

## Verify

```bash
confire status
```

## Uninstall

```bash
confire reset
```

Removes the hooks and stops the daemon. The binary stays installed.

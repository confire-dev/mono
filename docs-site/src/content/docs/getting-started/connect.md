---
title: Connect your agents
description: Install Confire hooks into Claude Code.
---

Confire uses Claude Code's **PostToolUse hook** to intercept tool outputs. The installer runs `confire setup` automatically, but you can re-run it any time.

## Setup

```bash
confire setup
```

This opens an interactive prompt to choose scope and select agents. Pick **Global** to apply to all projects, or **Local** to apply only to the current repo's `.claude/settings.json`.

After setup, restart Claude Code to activate the hook.

## What setup installs

`confire setup` adds a `PostToolUse` entry to your Claude Code settings:

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

Every time Claude Code finishes a tool call, it pipes the output through `confire hook`, which sends it to the daemon for optimization.

## Start the daemon

```bash
confire start
```

The daemon runs in the background and handles all optimization. Setup starts it automatically, but you can also start it manually.

## Verify everything is running

```bash
confire status
```

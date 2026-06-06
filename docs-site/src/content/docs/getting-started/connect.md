---
title: Connect your agents
description: Install Confire hooks into your AI coding agent.
---

Confire uses your agent's **PostToolUse hook** to intercept tool outputs. The installer runs `confire setup` automatically, but you can re-run it any time.

## Setup

```bash
confire setup
```

Opens an interactive prompt to choose scope and select agents. Pick **Global** to apply to all projects, or **Local** to apply only to the current repo's settings file.

After setup, restart your agent to activate the hook.

## What setup installs (Claude Code)

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

Every time a tool call finishes, Claude Code pipes the output through `confire hook`, which sends it to the daemon for optimization and returns the stripped result.

## Other agents

Hook-based integration for Cursor, VS Code, and other agents is on the roadmap. The daemon is agent-agnostic — once the hook fires, optimization works the same regardless of which agent triggered it.

## Start the daemon

```bash
confire start
```

Runs in the background and handles all optimization. Setup starts it automatically.

## Verify

```bash
confire status
```

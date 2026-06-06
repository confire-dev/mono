---
title: Connect your agent
description: Install Confire hooks into Claude Code, Cursor, or VS Code.
---

The installer runs `confire setup` automatically, but you can re-run
it any time — for example, to add a second agent or switch from
global to local scope.

## Run setup

```bash
confire setup
```

This opens an interactive prompt. You'll pick a scope and select
which agents to hook.

**Global** installs hooks into your user-level settings file and
applies to every project. **Local** writes to `.claude/settings.json`
(or the equivalent) in the nearest git root and applies only to that
repo.

After setup, restart your agent for the hook to take effect.

## What setup installs

For Claude Code, `confire setup` adds a `PostToolUse` hook entry to
your settings file:

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "",
        "hooks": [{ "type": "command", "command": "confire hook" }]
      }
    ],
    "PreToolUse": [
      {
        "matcher": "",
        "hooks": [{ "type": "command", "command": "confire hook" }]
      }
    ]
  }
}
```

Every tool call — before it runs and after it finishes — passes
through `confire hook`, which forwards the event to the daemon.

For Cursor and VS Code, setup writes equivalent hook entries to
their respective settings files.

## Start the daemon

```bash
confire start
```

The daemon runs in the background and handles all firewall evaluation
and optimization. Setup starts it automatically, but run this if you
stopped it with `confire stop`.

## Log in for remote optimizers

Local optimizers (Bash, Read, WebFetch) work without an account.
Remote optimizers for Figma, GitHub, and all other MCP-connected
tools require a Confire account:

```bash
confire login
```

This opens a browser to authenticate. Once logged in, remote
optimizers activate automatically when the relevant tools are called.

## Next step

[Verify your setup →](../getting-started/verify)

---
title: Architecture
description: How Confire fits into the Claude Code stack.
---

## Data flow

```
Claude Code finishes a tool call
  → PostToolUse hook fires
  → confire hook (stdin: raw tool output JSON)
  → daemon (Unix socket at ~/.confire/daemon.sock)
  → optimizer (local or remote)
  → optimized output (stdout → Claude Code)
```

Confire intercepts **tool call responses**, not tool calls themselves. It runs after the tool executes and before the result reaches the model.

## Components

### CLI (`confire`)

The user-facing binary. Handles setup, login, start/stop, and status. Also acts as the hook entrypoint — Claude Code calls `confire hook` on every tool output.

### Daemon

A background process that listens on a Unix socket. The hook sends raw tool output to the daemon, which runs the optimizer and returns the stripped result. The daemon handles:
- Routing to the right optimizer per tool type
- Forwarding to the remote Worker if an API key is present
- Caching and connection management

### Worker (remote)

A Cloudflare Worker at `api.confire.dev`. Runs source-specific remote optimizers — Figma, GitHub PR, Jira, Slack — that require more context or cloud processing. Only used when you're logged in with a paid plan.

## Local vs. remote

| Optimizer        | Mode   | Plan     |
|------------------|--------|----------|
| Bash             | Local  | Free     |
| Read (file)      | Local  | Free     |
| WebFetch         | Local  | Free     |
| Generic JSON     | Local  | Free     |
| Figma            | Remote | Paid     |
| GitHub PR        | Remote | Paid     |
| MCP tools        | Remote | Paid     |

Local optimizers run entirely on your machine. No data leaves unless you're using remote optimizers.

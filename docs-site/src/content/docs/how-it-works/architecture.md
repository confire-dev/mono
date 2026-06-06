---
title: Architecture
description: How Confire fits into the AI agent stack.
---

## Data flow

```
Agent finishes a tool call
  → PostToolUse hook fires
  → confire hook (stdin: raw tool output JSON)
  → daemon (Unix socket at ~/.confire/daemon.sock)
  → optimizer selected by tool type
      local  → runs on your machine (free)
      remote → forwarded to api.confire.dev (paid)
  → optimized output (stdout → agent)
```

Confire intercepts **tool call responses** — after the tool runs, before the result reaches the model.

## Components

### CLI (`confire`)

The user-facing binary. Handles setup, login, start/stop, and status. Also acts as the hook entrypoint — your agent calls `confire hook` for every tool output.

### Daemon

A background process listening on a Unix socket (`~/.confire/daemon.sock`). The hook sends raw tool output to the daemon, which selects and runs the optimizer, then returns the stripped result. Handles:

- Routing to the right optimizer per tool type
- Forwarding to the remote Worker when appropriate
- Connection management and local fallback

### Worker (remote)

A cloud service at `api.confire.dev`. Runs source-aware remote optimizers for tool types that need domain-specific processing — design tools, code review, project management, communication platforms, and more. Only used when logged in with a paid plan.

## Local vs. remote

| Optimizer type         | Mode   | Plan     |
|------------------------|--------|----------|
| Bash                   | Local  | Free     |
| Read (file)            | Local  | Free     |
| WebFetch               | Local  | Free     |
| Generic JSON fallback  | Local  | Free     |
| Figma                  | Remote | Paid     |
| GitHub                 | Remote | Paid     |
| Any MCP-connected tool | Remote | Paid     |

Local optimizers run entirely on your machine. No data leaves unless you're using remote optimizers.

If the remote connection is unavailable, local optimizers continue running and Generic handles the fallback.

## Agent support

Confire currently integrates via **PostToolUse hooks** (Claude Code). Hook-based integration for other agents is on the roadmap — the architecture is agent-agnostic and the daemon works the same regardless of which agent triggers it.

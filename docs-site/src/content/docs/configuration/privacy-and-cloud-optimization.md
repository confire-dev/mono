---
title: Privacy
description: >-
  What data stays on your machine and what is sent to
  Confire's cloud for telemetry and policy sync.
---

Confire is local-first. All firewall decisions and security passes run
in the daemon on your machine. Understanding what does and doesn't leave
your machine is important if you work with sensitive codebases.

## What stays local

The following never leave your machine, regardless of plan:

- Source files and project contents
- Conversation history and agent transcripts
- Tool inputs — the commands and parameters your agent sends
  to tools (file paths, shell commands, search queries)
- Tool output content — raw tool responses are never forwarded
  to the cloud
- Policy rule evaluation — all firewall decisions happen
  locally in the daemon
- Secret redaction and injection scanning — these run
  locally before any content is transmitted

## What is sent to the cloud

When the daemon is connected to `api.confire.dev`, it sends
**structured telemetry events** — not tool output content.

Events sent per session:

- **Security events** — tool name, event type, risk level, action taken,
  pattern category (e.g. `SECRET_REDACTED`). No matched values or raw content.
- **Provenance events** — trust label, MCP server origin, redaction count.
  No tool output content.
- **Session metadata** — integration type, CLI version, total tool call counts.

All telemetry is sent after local security passes run. Matched secret values
are never included — only the fact that a pattern was detected.

## Disabling cloud telemetry

To run in fully local mode with no data leaving your machine:

```bash
confire config set worker_url=
```

Setting `worker_url` to empty disables cloud telemetry and policy sync.
The local firewall — secret redaction, injection guard, built-in rules,
local firewall and context passes — continue running uninterrupted.

Alternatively, don't log in. Without an API key, the daemon uses
local-only mode automatically.

## Analytics

Confire sends anonymous product analytics (security event counts, session
counts, firewall action counts) to help improve the product. No tool
content, file paths, secret values, or user-identifiable data is included.

To opt out:

```bash
confire config set analytics=false
```

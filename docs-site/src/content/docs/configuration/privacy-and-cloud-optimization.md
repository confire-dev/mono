---
title: Privacy and cloud optimization
description: >-
  What data stays on your machine and what is sent to
  Confire's cloud for analytics and policy sync.
---

Confire is local-first. Most processing happens in the daemon
on your machine. Understanding what does and doesn't leave your
machine is important if you work with sensitive codebases.

## What stays local

The following never leave your machine, regardless of plan:

- Source files and project contents
- Conversation history and agent transcripts
- Tool inputs — the commands and parameters your agent sends
  to tools (file paths, shell commands, search queries)
- Policy rule evaluation — all firewall decisions happen
  locally in the daemon
- Secret redaction and injection scanning — these run
  locally before any content is transmitted

## What is sent to the cloud

When you're logged in with a paid plan and the daemon forwards
a tool call to the remote optimizer, the tool's **output** is
sent to `api.confire.dev`.

Before transmission:
1. Secret redaction runs locally — matched values are replaced
   with markers
2. Hidden-unicode stripping runs locally

What reaches the cloud is the already-sanitized tool output,
not the raw response. Confire doesn't transmit the tool name
or inputs alongside the output.

## Free plan

No data is ever transmitted. All processing — optimization and
security — runs locally in the daemon binary.

## Disabling cloud optimization

To prevent any tool output from leaving your machine while
keeping local optimization active:

```bash
confire config set worker_url=
```

Setting `worker_url` to empty disables the remote optimizer.
Local optimizers (Bash, Read, WebFetch, Generic) continue running.

Alternatively, don't log in. Without an API key, the daemon
uses local-only mode.

## Analytics

Confire sends anonymous usage analytics (token savings, optimizer
types, session counts) to help improve the product. No tool
content, file paths, or user-identifiable data is included.

To opt out:

```bash
confire config set analytics=false
```

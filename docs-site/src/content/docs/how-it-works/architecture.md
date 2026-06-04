---
title: Architecture
description: How Confire fits into the AI agent stack.
---

## High-level view

```
 ┌─────────────────────────────────────────────────┐
 │  AI Coding Agent (Claude Code, Cursor, VS Code) │
 └────────────────────────┬────────────────────────┘
                          │  MCP messages (stdio / HTTP)
                          ▼
 ┌─────────────────────────────────────────────────┐
 │               Confire Proxy                     │
 │                                                 │
 │  ① Policy Engine   → allow / block / approve   │
 │  ② Context Filter  → strip noise, redact        │
 │  ③ Injection Guard → scan tool results          │
 └────────────────────────┬────────────────────────┘
                          │  clean, filtered messages
                          ▼
 ┌──────────────┐  ┌──────────────┐  ┌────────────┐
 │  Filesystem  │  │  Shell / Bash│  │  APIs, DBs │
 │  MCP Server  │  │  MCP Server  │  │  MCP Servers│
 └──────────────┘  └──────────────┘  └────────────┘
```

## Components

### Policy Engine

The first thing every MCP message hits. The engine evaluates your policy rules in order and decides:

- **allow** — pass through unchanged
- **block** — return an error to the agent explaining why
- **require-approval** — pause and prompt the user, then proceed or block

### Context Filter / Optimizer

Runs on the *input* side of each tool call — the context payload the agent sends along with its request. The optimizer:

1. Identifies signal vs. noise (irrelevant file content, giant stack traces, repeated boilerplate)
2. Strips or summarises the noisy parts
3. Passes a leaner payload upstream

This reduces token usage and speeds up responses without changing what the agent can do.

### Secret Redaction

Scans all outbound context (and inbound tool results) for patterns matching common secret formats:

- AWS / GCP / Azure credentials
- GitHub, npm, Stripe tokens
- Private keys (PEM blocks)
- Generic high-entropy strings

Matches are replaced with `[REDACTED:type]` before they can appear in any prompt or log.

### Injection Guard

Scans the *output* of tool calls (files read, shell output, API responses) before handing them back to the agent. It looks for embedded prompt-injection payloads — text designed to make the agent ignore its instructions or exfiltrate data.

Suspicious results are flagged or stripped depending on your policy.

## Local vs. remote

All core features run **locally** — no data leaves your machine. Remote features (optimizer updates, remote policy sync, usage analytics) connect to Confire's cloud service and require a paid plan.

| Feature                    | Local | Remote |
|----------------------------|:-----:|:------:|
| Tool Firewall              | ✓     |        |
| Secret Redaction           | ✓     |        |
| Basic context optimizer    | ✓     |        |
| Source-specific optimizers |       | ✓      |
| Remote policy sync         |       | ✓      |
| Usage analytics            |       | ✓      |

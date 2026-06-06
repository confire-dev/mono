---
title: Plans and limits
description: What's included in the free and paid plans.
---

## Free

No account required. Works offline. All firewall features are
included.

- **Tool Firewall** — all built-in rules, all policy modes
- **Context Firewall security** — secret redaction, injection
  guard, hidden-unicode stripping
- **Local optimizers** — Bash, Read, WebFetch, Generic fallback
- **Unlimited** local optimizations
- **All supported agents** — Claude Code, Cursor, VS Code

## Paid

Requires a Confire account (`confire login`).

- Everything in Free
- **Remote optimizers** — Figma, GitHub, Slack, Notion, Jira,
  and every other MCP-connected tool your agent uses
- **Custom rules** — organization-specific firewall rules managed
  in the dashboard and synced via `confire policy pull`
- **Group overrides** — toggle rule groups on/off per team from
  the dashboard
- **Usage dashboard** — token savings, session stats, rule
  activity
- **Top-up credits** — purchase additional optimization credits
  via `confire topup`

See [confire.dev/pricing](https://confire.dev/pricing) for current
pricing and credit limits.

## Usage credits

Remote optimization on paid plans is metered by usage credits.
Your current balance and limit are shown in `confire status` and
in the dashboard. When credits run low, local optimizers continue
running — only remote optimization pauses.

Purchase more credits:

```bash
confire topup
```

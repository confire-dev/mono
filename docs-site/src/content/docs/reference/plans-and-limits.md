---
title: Plans and limits
description: What's included in Free, what's in Dev, and what's coming in Team.
---

## Free

No account required. All firewall and local protection features are
included at no cost — protection is never paywalled.

- **Tool Firewall** — all built-in rules, all policy modes
- **Rate policy** — runaway loop detection + call rate cap (runs locally, no account)
- **Ask-budget** — auto-acknowledge warn-action calls up to 5 per session
- **Context Firewall** — secret redaction, injection guard, hidden-unicode stripping
- **Local optimizers** — Bash, Read, WebFetch, Generic fallback
- **All supported agents** — Claude Code, Cursor, VS Code
- **500 remote optimizations/month**
- **Basic per-session stats** — call counts, savings, firewall events

Local protection runs entirely in the daemon binary with no network
round-trip and no account required.

## Dev — $10/month · $90/year

For daily AI coding with Confire always on.

- Everything in Free
- **5,000 remote optimizations/month**
- **Configurable rate thresholds** — tighten `runaway_loop_threshold`
  and `call_rate_threshold` via the dashboard; values sync to the
  daemon automatically
- **Custom dashboard guardrails** — organization-specific firewall
  rules managed in the dashboard and synced via `confire policy pull`
- **Policy sync** — keep rules in sync across machines
- **Security event history** — blocked, reviewed, and warned calls
  stored in the dashboard and queryable per session
- **Provenance tracking history** — MCP trust labels and sanitization
  events stored in the dashboard
- **Updated optimizer and risk-rule library** — priority access to
  new adapters and rule updates
- **Larger input payloads**
- **Top-up credits** — purchase additional remote optimization
  credits via `confire topup`
- **Early access to new clients/adapters**

[Start Dev →](https://confire.dev/#pricing)

## Team — planned

Shared policy management, team usage dashboard, custom optimizers,
audit controls, and SSO. Contact us to join the waitlist.

[Join waitlist →](https://confire.dev/#pricing)

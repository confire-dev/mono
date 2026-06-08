---
title: Hook phases
description: The four events Confire intercepts in every agent session.
---

Confire registers hooks for four phases of every agent session.
Understanding what runs at each phase helps you reason about when
firewall decisions are made and when the context pipeline runs.

## SessionStart

When you open a new session in your agent, Confire's daemon receives
a `SessionStart` event. It uses this to reload configuration, pull
the latest policy rules from cache, and — if you're logged in — warm
the connection to `api.confire.dev` for cloud policy sync. You may
see a one-time welcome message in the agent output after first installing.

## PreToolUse

Before a tool executes, the daemon receives the tool name and input.
The Tool Firewall evaluates the call against all active policy rules
and returns one of:

- **allow** — tool proceeds normally, nothing shown to the agent
- **warn** — tool proceeds, but the agent receives an advisory
  note in context
- **review** — tool is paused; the agent sees the reason and
  must confirm before proceeding
- **block** — tool is prevented from running; the agent sees
  the block reason

## PostToolUse

After a tool finishes, the daemon receives the full tool output.
The Context Firewall runs in sequence:

1. Security passes: secret redaction, hidden-unicode stripping,
   injection detection (MCP tools only)
2. Context passes: noise trimming, MCP output normalization

The daemon returns the processed output to the hook, which passes it
back to the agent in place of the raw result. If nothing changed,
the original output passes through unchanged.

## SessionEnd

When the agent session closes, the daemon records session stats
(tools processed, rules fired, events detected). If you're logged in,
these sync to your dashboard.

---

:::note
Confire only intercepts tool calls. It never reads your source files,
modifies project state, or changes agent configuration. Tool output
content is never forwarded to the cloud — only structured telemetry
events (risk level, action taken, session metadata) are sent.
:::

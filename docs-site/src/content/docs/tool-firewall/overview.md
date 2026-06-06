---
title: Tool Firewall
description: >-
  What runs before every tool call — policy evaluation,
  risk scoring, and pre-execution review.
---

The Tool Firewall is Confire's PreToolUse layer. Before any tool
executes, the daemon evaluates the call against active policy rules
and returns a decision to the agent.

## The four outcomes

**Allow** — no rule matched, or all matches are suppressed by the
current mode. The tool runs normally with no visible effect.

**Warn** — a rule matched at warn severity. The tool runs, but the
agent receives an advisory note in context describing the concern.

**Review** — a rule matched at review severity. The tool is paused.
The agent sees the reason and must explicitly confirm before
proceeding. Use `confire bypass-next` if you've reviewed the
situation and want to let the next call through.

**Block** — a rule matched at block severity. The tool does not run.
The agent sees the block reason. No bypass is available — this is a
hard stop.

## What gets evaluated

The Tool Firewall evaluates:

- **Native tools** — Bash commands are checked against pattern-based
  rules (git operations, secret file reads, and so on)
- **MCP tools** — all MCP tool calls are scored by the risk
  classifier and checked against MCP-specific rules

## Policy modes

The active policy mode controls how strictly rules are enforced.
In `observe` mode, the firewall records matches without blocking
or reviewing. In `strict` mode, all rules fire at their stated
severity. See [Policy modes](policy-modes) for the full breakdown.

## Related pages

- [Built-in rules](built-in-rules) — what ships in the binary
- [Policy modes](policy-modes) — observe, balanced, strict, bypass
- [MCP risk classifier](mcp-risk-classifier) — automatic MCP scoring
- [Custom rules](custom-rules) — cloud-managed rule additions
- [Bypass and approvals](bypass-and-approvals) — one-shot overrides

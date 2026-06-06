---
title: Tool Firewall
description: >-
  What runs before every tool call — policy evaluation,
  risk scoring, and pre-execution review.
---

Your agent decides to run `rm -rf ./old-docs`. Before it executes,
Confire intercepts it:

```
CONFIRE REVIEW REQUIRED

Rule:              Review destructive filesystem change
Claude is about to run: Bash — rm -rf docs-site/src/content/docs/clients docs-site/src/content/docs/how-it-works
Risk:              high severity
Why this matters:  Recursive force delete permanently removes files with no trash recovery.

ACTION REQUIRED — ask the user:
"Confire flagged this command. Do you want me to run it anyway?
If yes: run 'confire bypass-next' in your terminal, then tell me to retry."
```

That's a review. The tool is paused. Your agent is waiting for you.

## The four outcomes

**Allow** — no rule matched. The tool runs with no visible effect.

**Warn** — a low-severity rule matched. The tool runs, but the
agent receives an advisory note in context:

```
[Confire warning] Review mutating MCP tool: mcp__stripe__create_charge — This MCP tool appears to mutate external state.
```

**Review** — a high-severity rule matched. The tool is paused
until the user approves. The agent cannot proceed on its own —
it must ask you. Run `confire bypass-next` to allow the next call
through, or `confire off` to disable the firewall temporarily.

**Block** — a rule matched with `block` action. The tool does not
run. No retry path exists from the agent — the user must act
outside the agent session to resolve it:

```
CONFIRE BLOCKED TOOL CALL

Rule:    Delete remote repository
Blocked: Bash — gh repo delete my-org/my-repo --yes
Reason:  Permanent repository deletion cannot be undone.
```

## What gets evaluated

The Tool Firewall evaluates every tool call before it runs:

- **Native tools** — Bash commands are matched against
  pattern-based rules (git operations, secret file reads, and
  so on)
- **MCP tools** — all MCP tool calls are scored by the risk
  classifier and checked against MCP-specific rules

## Policy modes

The active mode controls how strictly rules are enforced. In
`observe` mode the firewall records matches without acting. In
`balanced` mode (default) rules fire at their stated severity.
In `strict` mode additional rules activate. See
[Policy modes](policy-modes).

## Related pages

- [Built-in rules](built-in-rules) — what ships in the binary
- [Policy modes](policy-modes) — observe, balanced, strict, bypass
- [MCP risk classifier](mcp-risk-classifier) — automatic MCP scoring
- [Custom rules](custom-rules) — cloud-managed rule additions
- [Bypass and approvals](bypass-and-approvals) — one-shot overrides

---
title: MCP risk classifier
description: >-
  How Confire scores MCP tool calls based on name and
  input parameters before they execute.
---

The MCP risk classifier runs at PreToolUse for every MCP tool call.
It assigns a risk score based on the tool name and input parameter
names, then decides whether to review, warn, or allow.

## How scoring works

The classifier awards points across three tiers. Scores accumulate —
a tool that matches multiple tiers gets points from each.

| Tier | Pattern | Points | Example names |
|---|---|---|---|
| 1 — shell/exec | Tool name contains shell execution terms | 40 | `bash`, `shell`, `exec`, `run_command`, `eval`, `terminal` |
| 2 — write/destroy | Tool name contains destructive operation terms | 30 | `__write_`, `__delete_`, `__drop_`, `__wipe_`, `__truncate_` |
| 3 — exfiltration | Tool name contains data-sending terms | 20 | `__send_`, `__email_`, `__upload_`, `__export_`, `__webhook_` |

In addition, if the tool's input parameters include names that
suggest credentials (for example, `password`, `token`, `secret`,
`api_key`), the classifier adds a warn-level flag regardless of
the score.

## Score thresholds

- **Score ≥ 50** → review: the tool call is paused; the agent
  sees the score and reason
- **Score 20–49** → warn: the tool runs but the agent receives
  an advisory note
- **Score < 20** → allow: the classifier passes the call through

## What the scoring catches

The classifier is name-based — it scores tool names, not what the
tools actually do. An MCP tool named `execute_bash` scores 40 points
and gets reviewed. An MCP tool named `get_user` scores 0 and passes
through.

This means the classifier is a heuristic, not a guaranteed catch.
A poorly named dangerous tool won't be scored. It also means that
well-named safe tools with terms like `update` or `create` may get
scored — the `mcp-mutation` built-in rule (not the classifier)
handles that category.

## Relationship to built-in rules

The MCP risk classifier and the `mcp-mutation` built-in rule are
separate mechanisms that can both fire on the same tool call. The
classifier scores by name pattern; the built-in rule checks for
specific mutation verbs. The higher-severity result wins.

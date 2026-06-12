---
title: What is Confire?
description: A local firewall daemon that inspects tool activity before tools run and after they return.
---

Confire is a local firewall for AI coding agents. It runs as a
background daemon and inspects supported tool activity before each
tool runs and after it returns.

Confire helps answer two questions:

1. Should this tool call be allowed to run?
2. Did this tool result contain anything the agent or user should know about?

## Two firewalls, one daemon

### Tool Firewall

Runs before supported tool calls. The Tool Firewall evaluates each call
against policy rules and returns one of four decisions:

- `allow`
- `warn`
- `review`
- `block`

Risky actions can be flagged before execution, including:

- force pushes
- destructive Git operations
- secret-file reads
- database resets
- deploys
- package publishing
- mutating MCP actions

### Tool Result Firewall

Runs after supported tool calls. The Tool Result Firewall inspects tool
results and returns additional firewall context to the agent where
supported. This context can include:

- secret-looking value warnings
- hidden Unicode warnings
- prompt-injection-like content warnings
- MCP risk notes
- provenance metadata
- policy decisions
- suggested next steps

Confire does not need to replace the original tool output to be useful.
It gives the agent and user security context around what just happened.

## Event lifecycle

Every supported agent session follows this lifecycle:

```
SessionStart  → daemon loads config and latest local policies
PreToolUse    → Tool Firewall evaluates the tool call
               (tool runs if allowed)
PostToolUse   → Tool Result Firewall inspects the result
SessionEnd    → local stats and metadata sync
```

Confire hooks into `PreToolUse` and `PostToolUse` where the client
supports those events. It does not modify source files, project state,
or your agent's task. It only evaluates tool activity and returns
policy decisions or firewall context.

## Local-first

Firewall decisions run locally in the Confire daemon. Built-in rules
work without an account. When you are logged in, Confire can sync
metadata-only security events to the dashboard.

**Synced metadata can include:**

- rule ID
- action taken
- risk level
- tool category
- client
- session ID
- timestamp

**Synced metadata does not include:**

- raw tool output
- source code
- secret values
- full command output
- `.env` contents

Confire is a guardrail layer, not a perfect security boundary. It helps
agents and users notice risky actions and suspicious tool results before
they become bigger mistakes.

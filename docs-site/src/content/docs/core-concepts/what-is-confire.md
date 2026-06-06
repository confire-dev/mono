---
title: What is Confire?
description: The two-pillar model — Context Firewall and Tool Firewall.
---

Confire controls what AI agents can **do** and what tool output they
can **see**. It runs as a background daemon and intercepts every tool
call your agent makes — before the tool runs and after it finishes.

## Two firewalls, one daemon

**Tool Firewall (PreToolUse)** — evaluates each tool call before it
executes. Matches the call against a set of policy rules and returns
`allow`, `warn`, `review`, or `block`. Risky commands — force pushes,
secret file reads, mutating MCP actions — get flagged before anything
happens.

**Context Firewall (PostToolUse)** — processes tool output before it
enters the model context. Runs security passes (secret redaction,
hidden-unicode stripping, injection detection) and optimization passes
(noise trimming, MCP normalization) on every response. The model sees
cleaner, safer output.

## The event lifecycle

Every agent session follows this sequence:

```
SessionStart  → daemon loads config and latest policies
PreToolUse    → Tool Firewall evaluates the call
               (tool runs if allowed)
PostToolUse   → Context Firewall processes the output
SessionEnd    → stats sync
```

Confire intercepts at `PreToolUse` and `PostToolUse`. It never
modifies source files, project state, or agent configuration — only
tool inputs and outputs.

## Local vs. remote processing

Core optimizers run entirely on your machine — no account required,
no data leaves. Remote optimizers (Figma, GitHub, all MCP tools) run
in Confire's cloud at `api.confire.dev` and require a paid plan. When
you're offline or the connection drops, local processing continues
uninterrupted.

Policy rules are evaluated locally in all modes. Your tool calls are
never forwarded to the cloud for firewall decisions — only tool output
is sent for optimization, and only for remote optimizer types.

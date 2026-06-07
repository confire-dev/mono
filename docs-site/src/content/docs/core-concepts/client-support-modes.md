---
title: Client support modes
description: >-
  What Confire can do varies by AI coding agent. Understand
  the difference before installing.
---

Confire supports Claude Code, Cursor, and VS Code. The features
available in each depend on what the agent's hook API exposes.

## Capability table

| Feature | Claude Code | Cursor | VS Code |
|---|---|---|---|
| Tool Firewall (PreToolUse) | ✓ | ✓ | ✓ |
| Context Firewall (PostToolUse) | ✓ | ✓ | ✓ |
| Replace native tool output | ✓ | — | — |
| Optimize native tools (Bash, Read, WebFetch) | ✓ | — | — |
| Optimize MCP tool output | ✓ | ✓ | ✓ |
| Post-tool steering via `additional_context` | ✓ | ✓ | ✓ |

## What "replace native tool output" means

Claude Code's hook API lets Confire return a modified `toolOutput`
that the agent uses in place of the original. This is how native
tool optimization works: the daemon strips noise from Bash stdout,
caps large file reads, and cleans up WebFetch responses — then
returns the trimmed version.

Cursor and VS Code don't expose output replacement through their
hook APIs. Confire can't rewrite what the agent sees from native
tools like `bash` or `read`. Instead, it uses the
`additional_context` mechanism: the agent receives the original
output plus a steering note in context. For this reason, native
tool optimization (Bash, Read, WebFetch) is only fully effective
in Claude Code.

## MCP tools work everywhere

MCP tool output can be replaced in all supported agents. When your
agent calls a Figma, GitHub, Slack, or any other MCP tool, Confire
can return optimized and sanitized output regardless of which
agent you're using.

## Checking your client mode

```bash
confire status
```

The "Agents" section shows which clients were detected and whether
hooks are installed.

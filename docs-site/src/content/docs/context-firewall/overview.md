---
title: Context Firewall
description: >-
  What runs after every tool call, before output enters the
  model context.
---

The Context Firewall is Confire's PostToolUse layer. After every tool
call, before the result reaches the model, the daemon processes the
output through a pipeline of security and optimization passes.

## What it does

The pipeline runs in this order for every tool call:

1. **Security passes** (MCP tools only): secret redaction,
   hidden-unicode stripping, prompt-injection detection
2. **Optimization passes**: noise trimming, token reduction,
   source-aware MCP normalization

Security runs first. Optimization never runs on plaintext secrets —
by the time the optimizer sees the content, sensitive values are
already redacted.

## What the agent sees

For Claude Code, the agent receives the processed output in place of
the original. For Cursor and VS Code, the agent receives the original
output alongside a steering note in `additional_context` that
describes what was found or trimmed.

If none of the passes changed anything, the original output passes
through unchanged. Confire is a no-op when there's nothing to do.

## Local vs. remote

Security passes (secret redaction, injection detection) run locally
in the daemon — no data leaves your machine for these checks.

Optimization passes are split:

- **Local** (free): Bash, Read, WebFetch, Generic fallback — run in
  the daemon binary, zero network, ~1ms
- **Remote** (paid): Figma, GitHub, all MCP-connected tools — sent
  to `api.confire.dev` for source-aware processing

Only tool output is sent to the remote optimizer. Your source files,
conversation history, and project structure are never transmitted.

## Related pages

- [Optimizers](optimizers) — how noise trimming works per tool type
- [Secret redaction](secret-redaction) — what patterns are detected
- [Injection guard](injection-guard) — hidden unicode and
  instruction injection
- [MCP output sanitization](mcp-output-sanitization) — the full
  MCP security pipeline
- [Context limits](context-limits) — built-in output size caps

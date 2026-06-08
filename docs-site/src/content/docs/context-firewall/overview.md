---
title: Context Firewall
description: >-
  What runs after every tool call, before output enters the
  model context.
---

The Context Firewall is Confire's PostToolUse layer. After every tool
call, before the result reaches the model, the daemon processes the
output through a security pipeline. Noise trimming runs as a secondary
pass on whatever the security layer doesn't modify.

## What it does

The pipeline runs in this order for every tool call:

1. **Security passes** (MCP tools only): secret redaction,
   hidden-unicode stripping, prompt-injection detection
2. **Optimization passes**: noise trimming, source-aware MCP normalization

Security runs first. By the time any optimization runs, sensitive values
are already redacted.

## What the agent sees

For Claude Code, the agent receives the processed output in place of
the original. For Cursor and VS Code, the agent receives the original
output alongside a steering note in `additional_context` that
describes what was found or changed.

If none of the passes changed anything, the original output passes
through unchanged. Confire is a no-op when there's nothing to do.

## Local vs. remote

All security passes run locally in the daemon — no data leaves your
machine for redaction, injection detection, or unicode stripping.

Optimization passes are split:

- **Local** (free): Bash, Read, WebFetch, Generic fallback — run in
  the daemon binary, zero network, ~1ms
- **Remote** (paid): Figma, GitHub, all MCP-connected tools — sent
  to `api.confire.dev` for source-aware processing

Only structured telemetry events are sent to the cloud. Tool output
content is never forwarded.

## Related pages

- [Secret redaction](../secret-redaction) — what patterns are detected
- [Injection guard](../injection-guard) — hidden unicode and
  instruction injection
- [MCP output sanitization](../mcp-output-sanitization) — the full
  MCP security pipeline
- [Optimizers](../optimizers) — how noise trimming works per tool type

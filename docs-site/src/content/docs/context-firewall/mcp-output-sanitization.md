---
title: MCP output sanitization
description: >-
  The combined security pipeline Confire runs on every MCP
  tool response.
---

Every MCP tool response passes through the sanitization pipeline
before optimization runs. The pipeline combines three passes in
a single handler: unicode stripping, secret redaction, and
injection detection.

## Why MCP specifically

Native tool output (Bash, Read, WebFetch) comes from your own
machine and is reasonably trusted. MCP tool output comes from
external servers and third-party services — the content is
untrusted. Sanitization runs on MCP output only for this reason.

## Pipeline order

For each MCP PostToolUse event:

1. **Secret redaction** — scan the full response for credential
   patterns; replace matched values with redaction markers
2. **Hidden-unicode stripping** — walk every string in the
   response recursively; remove tag-block, zero-width, and
   BiDi override characters
3. **Injection detection** — scan for instruction-like patterns
   in the sanitized content; flag if found

If none of the passes find anything, the original output passes
through unchanged and nothing is noted in context.

## What the agent sees

When at least one pass fires, Confire replaces the MCP output with
the sanitized version. In Claude Code, the agent also receives a
summary in `additional_context`:

```
findings:
  secrets_redacted=1
  secret_types=api_key
  injection_sanitized=true
```

The model sees the clean output. It doesn't see the original
values that were redacted.

## Relationship to optimization

Sanitization always runs before optimization. The remote optimizer
(Figma, GitHub, etc.) receives already-sanitized content — secrets
are redacted before any data leaves your machine for cloud
optimization.

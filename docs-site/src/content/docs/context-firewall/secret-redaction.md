---
title: Secret redaction
description: >-
  How Confire detects and redacts common secret patterns in
  MCP tool output.
---

When an MCP tool returns output, Confire scans it for common
credential and secret patterns before the content reaches the model
context. Matched values are replaced with a redaction marker.

## What it covers

Confire looks for high-confidence patterns including:

- API keys and tokens (common service-specific formats)
- Private key blocks (`-----BEGIN * PRIVATE KEY-----`)
- Connection strings with embedded credentials
- Authorization header values
- Common environment variable names paired with secret-like values

The patterns are tuned for precision. Confire prefers missing a
secret over false-positiving on normal content.

:::caution
Secret redaction detects common patterns — it doesn't catch
everything. Don't rely on it as your only defense against secrets
in MCP output. Treat it as a safety net, not a guarantee.
:::

## What happens when a secret is found

The matched value is replaced with a redaction marker in the output.
The model sees the sanitized version, not the original. The daemon
records how many values were redacted and what types were found
(without recording the values themselves).

If your agent uses Claude Code, you'll see a note in the
`additional_context` when redaction ran:

```
findings:
  secrets_redacted=2
  secret_types=api_key,private_key
```

## Scope

Secret redaction runs on MCP tool output only. Native tool output
(Bash, Read, WebFetch) doesn't go through this pass. If a Bash
command echoes a secret, that's outside Confire's redaction scope —
use shell-level controls to prevent that.

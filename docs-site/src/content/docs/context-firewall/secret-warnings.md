---
title: Secret warnings
description: >-
  How Confire inspects tool results for secret-looking patterns and
  adds firewall context back to the agent.
---

Confire can inspect supported tool results for common credential and
secret-looking patterns.

When it finds something suspicious, Confire adds firewall context back
to the agent where supported and records a metadata-only security
event.

Secret detection is a safety net. It is not a guarantee that every
secret will be found.

## What it checks

Confire looks for high-confidence patterns such as:

- API keys and tokens,
- private key blocks,
- connection strings with embedded credentials,
- authorization header values,
- common environment variable names paired with secret-looking values,
- database URLs,
- cloud provider credentials,
- webhook tokens.

The patterns are tuned for precision. Confire should avoid noisy false
positives on normal content.

## What happens when a secret-looking value is found

When Confire detects a secret-looking value, it can add firewall
context such as:

```
CONFIRE SECURITY CONTEXT
flags:   secret_like_value_detected
types:   api_key, private_key
risk:    this tool result appears to contain credential-like data
action:  do not copy, publish, or send this value externally
```

Confire records metadata about the finding, such as:

- rule ID,
- secret type,
- count,
- tool category,
- client,
- session ID,
- timestamp.

Confire does not record the secret value itself.

## Secret-file access

Secret-file reads are usually handled before execution by the Tool
Firewall. If an agent tries to read `.env`, Confire can pause the
action before the file is read:

```
CONFIRE REVIEW REQUIRED
rule:    Review secret-file access
risk:    .env files often contain API keys,
         tokens, database URLs, and credentials
action:  run `confire bypass-next` to allow once, then retry
```

## Scope

Secret warnings can run on supported tool results, including shell
output, file reads, fetched content, and MCP results, depending on
the client integration and tool surface.

Confire's strongest protection is reviewing sensitive access before
it happens. Post-tool secret warnings are a second layer that helps
the agent and user notice when credential-like data appeared in a
result.

:::caution
Secret detection catches common patterns, not everything. Do not rely
on Confire as your only defense against secret exposure. Use normal
secret-management practices: least-privilege credentials, `.gitignore`,
environment isolation, and repository secret scanning.
:::

---
title: Secret Redaction
description: Automatically strip secrets and credentials before they appear in prompts or logs.
---

Secret Redaction scans all context and tool results for credentials before they reach the model or get written to any log.

## What gets redacted

Confire detects and redacts:

| Category               | Examples                                                    |
|------------------------|-------------------------------------------------------------|
| AWS credentials        | `AKIA...`, `aws_secret_access_key`                          |
| GitHub tokens          | `ghp_...`, `github_pat_...`                                 |
| npm tokens             | `npm_...`                                                   |
| Stripe keys            | `sk_live_...`, `pk_live_...`                               |
| Twilio                 | `SK...`, auth tokens                                        |
| OpenAI / Anthropic     | `sk-...`, `sk-ant-...`                                      |
| Generic API keys       | High-entropy strings adjacent to `key`, `token`, `secret`  |
| Private keys           | PEM blocks (`-----BEGIN RSA PRIVATE KEY-----`)             |
| Passwords in URLs      | `https://user:pass@host/db`                                 |
| `.env` file values     | Any `KEY=VALUE` lines with secrets                          |

## How it works

Redaction runs on two sides:

**Outbound (context):** Before the agent's context payload is sent to an MCP server or logged, Confire scans it for secret patterns and replaces matches with `[REDACTED:type]`.

**Inbound (tool results):** After an MCP server returns a result, Confire scans the payload before handing it to the agent. This catches secrets accidentally returned by tools (e.g., a file read that included `.env`, or an API response that echoed credentials).

## Redaction format

Secrets are replaced with typed placeholders so you know what was removed:

```
OPENAI_API_KEY=sk-proj-abc123...  →  OPENAI_API_KEY=[REDACTED:openai-key]
Authorization: Bearer ghp_xyz     →  Authorization: Bearer [REDACTED:github-token]
```

## Configuration

Redaction is on by default. To disable it (not recommended):

```yaml
redaction:
  enabled: false
```

To add custom patterns:

```yaml
redaction:
  enabled: true
  extra-patterns:
    - name: internal-token
      pattern: 'INT-[A-Za-z0-9]{32}'
      replacement: '[REDACTED:internal-token]'
```

To allowlist patterns that trigger false positives:

```yaml
redaction:
  allowlist:
    - 'test_key_placeholder'
    - 'example_secret_here'
```

## Audit log

Each redaction is recorded:

```
2025-01-15T10:23:01Z  REDACTED  openai-key  context.messages[2].content
2025-01-15T10:23:04Z  REDACTED  github-token  tool_result.content
```

The actual secret value is never written to the log — only its location and type.

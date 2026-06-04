---
title: Custom guardrails
description: Add your own detection patterns and response behaviors beyond the built-in rules.
---

Custom guardrails let you extend Confire's built-in detection with patterns specific to your environment — internal tokens, proprietary file types, domain-specific injection phrases.

## Secret redaction patterns

Add your own patterns to the redaction engine:

```yaml
redaction:
  extra-patterns:
    - name: internal-api-token
      pattern: 'MYCO-[A-Za-z0-9]{40}'
      replacement: '[REDACTED:internal-api-token]'

    - name: database-connection-string
      pattern: 'postgresql://[^@]+@[^\s]+'
      replacement: '[REDACTED:db-connection-string]'
```

Patterns are Go-compatible regular expressions. Test them:

```bash
confire redaction test --pattern 'MYCO-[A-Za-z0-9]{40}' --input "token: MYCO-abc123xyz"
```

## Injection guard patterns

Add phrases to the injection detector:

```yaml
injection-guard:
  extra-patterns:
    - name: internal-override-phrase
      pattern: 'INTERNAL_AGENT_COMMAND:'
      action: strip    # override per-pattern: strip | flag | block
```

## Tool argument transforms

You can rewrite tool arguments before they reach the real server. This is useful for sanitizing input or adding required parameters:

```yaml
transforms:
  - when:
      tool: bash
    replace:
      args:
        command:
          # Strip any trailing pipe-to-curl patterns
          pattern: '\|\s*curl\s+.*$'
          replacement: ''
```

Transforms run after the policy engine allows a call and before it's forwarded.

## Response transforms

You can also modify what comes back from a tool before the agent sees it:

```yaml
response-transforms:
  - when:
      tool: read_file
    strip-patterns:
      - 'INTERNAL_NOTE:.*$'   # strip internal annotations
```

## Alerts

Trigger a notification when certain patterns are detected, without blocking:

```yaml
alerts:
  - name: sensitive-file-access
    when:
      tool: read_file
      args:
        path: "(password|secret|credential|private_key)"
    notify: terminal   # terminal | webhook
    message: "Agent accessed a potentially sensitive file: {args.path}"

  - name: webhook-alert
    when:
      category: network
    notify: webhook
    webhook-url: "https://hooks.example.com/confire-alerts"
```

## Importing shared guardrails

Reference guardrail files from a central location (useful for teams):

```yaml
import:
  - ~/.confire/company-guardrails.yaml
  - ./guardrails/security.yaml
```

Imported files are merged in order. Local rules always take precedence over imported ones.

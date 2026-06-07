---
title: Injection guard
description: >-
  How Confire detects hidden-unicode attacks and
  prompt-injection patterns in MCP tool output.
---

MCP tool output is untrusted content from external systems. Confire
scans it for two classes of attacks before it enters the model
context: hidden-unicode manipulation and prompt-injection patterns.

## Hidden-unicode stripping

Attackers can embed invisible characters in text to steer model
behavior without the content being visible to humans. Confire strips:

- **Tag block characters** (U+E0000–U+E007F) — used to hide
  instructions in otherwise-normal text
- **Zero-width characters** — zero-width space, zero-width
  non-joiner, zero-width joiner, and similar
- **Bidirectional override characters** — BiDi overrides that can
  reverse displayed text to hide malicious content

Stripping is recursive — it walks every string in the JSON response,
not just the top-level value.

## Injection detection

Confire scans for text patterns that resemble instruction injection:
content that tells the model to ignore previous instructions,
adopt a new persona, or perform actions on behalf of the tool
response. When a match is found, the content is flagged.

:::caution
Injection detection looks for known patterns. A well-crafted
injection that avoids those patterns won't be caught. Confire
helps reduce the attack surface — it doesn't eliminate it.
:::

## What happens on detection

If hidden unicode or injection patterns are found, the content is
sanitized (the offending characters or text are removed) and the
output is replaced with the cleaned version. The agent receives
a note in `additional_context` when this runs:

```
findings:
  injection_sanitized=true
```

## Scope

These passes run on MCP tool output only, at PostToolUse. They
don't run on native tool output (Bash, Read, WebFetch).

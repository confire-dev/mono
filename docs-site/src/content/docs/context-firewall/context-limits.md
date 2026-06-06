---
title: Context limits
description: >-
  Built-in output size caps that prevent large tool responses
  from flooding the model context.
---

Confire enforces built-in size limits on tool output. These caps
prevent a single large tool response from consuming the entire
context window and degrading agent performance.

## Current limits

The limits are fixed in the current version. They apply automatically
to all sessions — no configuration needed.

| Tool | Line cap | Byte cap |
|---|---|---|
| Bash | 200 lines | 40KB |
| Read | 500 lines | 80KB |

When output exceeds a limit, Confire trims it and inserts a marker
so the model knows content was removed:

```
[confire: 312 lines omitted]
```

or

```
[confire: 24680 bytes omitted]
```

## What's preserved

Confire's trimming logic prioritizes useful content:

- **Bash**: the first 20 lines and last 50 lines of general output
  are always kept. For test output, pass/fail summaries and error
  lines are kept. For build output, error lines are kept.
- **Read**: the first 500 lines are kept; the tail is dropped with
  a hint to use `offset`/`limit` parameters to read further.

## What's never trimmed

Regardless of output size:

- Security notices from Confire (redaction warnings, injection
  alerts)
- Error lines and failure indicators
- File paths and identifiers that anchor the content

## Planned

Configurable per-tool and per-agent context limits are planned for
a future paid-plan release. When available, you'll be able to set
tighter or looser budgets per tool type through the dashboard.

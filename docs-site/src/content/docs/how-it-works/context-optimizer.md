---
title: Context Optimizer
description: Reduce token usage by stripping noise from agent context before it hits the model.
---

AI coding agents attach a lot of context to every request: open files, tool results, conversation history, imported modules. Most of it is irrelevant to the current task. The Context Optimizer removes the noise before it reaches the model.

## What it does

The optimizer sits between the proxy and the outbound MCP call. It receives the full context payload, analyzes it, and returns a trimmed version that preserves the signal.

Typical reductions:

- **30–60%** fewer tokens on file-heavy tasks
- Near-zero reduction on short, focused requests (the optimizer gets out of the way)

## Optimizer types

### Universal fallback optimizer

Available on the Free plan. Applies generic heuristics:

- Truncates repeated boilerplate (imports, license headers)
- Collapses long stack traces to first/last few frames
- Strips binary content and non-UTF-8 sequences
- Removes duplicate blocks

### Source-specific optimizers (Dev plan and above)

Separate optimizers tuned per content type:

| Optimizer          | Specialization                              |
|--------------------|---------------------------------------------|
| `typescript`       | Strips type annotations when types are inferable |
| `python`           | Removes docstrings, collapses imports        |
| `rust`             | Strips lifetimes and doc comments            |
| `json-large`       | Summarizes large JSON payloads               |
| `stacktrace`       | Extracts root cause, hides framework frames  |
| `html`             | Strips scripts, styles, hidden elements      |
| `markdown-long`    | Summarizes large documents                   |
| `git-diff`         | Highlights changed lines, collapses context  |
| `package-lock`     | Replaces with dependency summary             |
| … and more         |                                              |

### PreCompact optimizer (Pro plan)

Runs before Claude Code's built-in context compaction. Cleans up the session context window before it grows too large, preserving decisions and discarding implementation noise. Keeps sessions coherent for much longer.

## Configuration

```yaml
optimizer:
  enabled: true
  strategy: auto          # auto | conservative | aggressive
  max-tokens: 40000       # hard cap on output context size
  preserve:
    - "*.test.ts"         # never strip test files
    - "CLAUDE.md"         # never strip project instructions
```

**Strategies:**

- `auto` — adapts based on how close you are to the model's context limit
- `conservative` — only removes clearly redundant content
- `aggressive` — prioritizes token savings even at the cost of some detail

## Measuring savings

```bash
confire stats
```

```
Last 7 days
  Requests:        142
  Tokens sent:     1,840,000
  Tokens saved:      610,000  (33%)
  Estimated cost:   -$1.83
```

The dashboard shows the same data with charts over time.

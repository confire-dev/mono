---
title: Noise Trimming
description: How the Context Firewall reduces noise from tool output before it enters context.
---

The Context Firewall includes a noise trimming layer that cleans up tool
output before it enters context. Confire routes each tool call's output to
the most specific handler available. If no handler matches, the Generic
fallback runs. Trimming only replaces output when the result is strictly
smaller than the original — Confire never inflates output.

All noise trimming runs locally in the daemon binary with no network
round-trip and no account required.

## Bash

Cleans shell output while preserving all actionable content.
The logic adapts to the type of output:

- **Test output** — keeps pass/fail summary lines, drops individual
  passing test names and progress output
- **Build output** — keeps error lines, drops progress and
  success-only noise
- **General output** — strips ANSI escape codes, progress-bar lines,
  and collapses runs of identical consecutive lines into a single
  line with a `[confire: N identical lines omitted]` marker

Output is never blindly truncated by line count. An emergency
512KB cap applies only to outputs that exceed that threshold after
all structural cleanup, with an explicit marker at the cut point.

Typical reduction: **60–95%**

## Read

Passes file content through unchanged by default. No line cap, no
byte cap for normal files. An emergency cap applies only when
file content exceeds 1MB:

```
[confire: N bytes omitted — file exceeds 1MB; use offset/limit to read further]
```

For files larger than 1MB, Confire preserves the first 800KB and
appends the marker above. This fires rarely — most source files are
well under this threshold.

Typical reduction: **0%** on normal files, emergency cap on very large files only

## WebFetch

Strips HTML structure noise from fetched pages: `<script>`,
`<style>`, `<nav>`, `<header>`, `<footer>`, and repeated layout
boilerplate. Returns the readable text content of the page.

Typical reduction: **80–90%**

## MCP (Generic)

Fallback for any MCP tool type without a specific handler. Walks
the JSON response and strips common noise: deeply nested empty
objects, null fields, arrays of nulls, and redundant metadata
keys.

Runs last, after all specific handlers have declined.

## How routing works

The daemon inspects the tool name and output shape, then selects:

1. The matching local handler (Bash, Read, WebFetch)
2. Generic fallback for MCP tools

---
title: Optimizers
description: How the Context Firewall reduces noise through source-aware optimization.
---

The Context Firewall includes an optimization layer that trims noise from tool
output before it enters context. Confire routes each tool call's output to
the most specific optimizer available. If no optimizer matches, the Generic
fallback runs. Optimization only replaces output when the result is strictly
smaller than the original — Confire never inflates output.

## Local optimizers (free)

These run in the daemon binary with no network round-trip and no
account required.

### Bash

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

### Read

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

### WebFetch

Strips HTML structure noise from fetched pages: `<script>`,
`<style>`, `<nav>`, `<header>`, `<footer>`, and repeated layout
boilerplate. Returns the readable text content of the page.

Typical reduction: **80–90%**

### Generic

Fallback for any tool type without a specific optimizer. Walks
the JSON response and strips common noise: deeply nested empty
objects, null fields, arrays of nulls, and redundant metadata
keys.

Runs last, after all specific optimizers have declined.

## Remote optimizers (paid)

Remote optimizers run in Confire's cloud and handle tool types that
require domain knowledge to reduce meaningfully. They activate
automatically when the daemon forwards a matching MCP tool call —
no configuration needed.

Requires a Confire account (`confire login`) and a paid plan.

### Figma

Converts raw Figma MCP output — verbose component trees, nested
style objects, layout descriptors, and auto-layout metadata — into
a compact section map the model can reason about. Returns the
structure and semantics, not the render data.

Typical reduction: **98%**

### GitHub

Summarizes pull request diffs and repository data. Focuses on
meaningful code changes, drops generated files, lock files, and
whitespace-only diffs. Preserves change intent without raw patch
noise.

Typical reduction: **85–95%**

### Other MCP tools

Confire applies source-aware optimizers for Slack, Notion, Jira,
Atlassian, Google Drive, Playwright, and other MCP-connected
platforms. Each optimizer is built specifically for that tool's
output format and schema.

When your agent connects a new MCP tool, Confire routes its output
through the best matching optimizer or falls back to Generic if no
specific one exists yet.

## How routing works

The daemon inspects the tool name and output shape, then selects:

1. The most specific remote optimizer (if paid and online)
2. The matching local optimizer (Bash, Read, WebFetch)
3. Generic fallback

Remote optimizers only run for MCP tools. Native agent tools
(Bash, Read, WebFetch) always use local optimizers, even on a paid
plan. If the remote connection drops, local optimizers continue
uninterrupted.

---
title: Context Optimizer
description: How Confire strips noise from tool outputs to save tokens.
---

The context optimizer is Confire's core feature. After every tool call, your agent gets back raw output — often thousands of tokens of noise. The optimizer strips it down to signal.

## Local optimizers (free, offline)

These run entirely on your machine with no account required.

### Bash

Trims shell output to keep failures, errors, and meaningful results. Removes long chains of progress lines, repeated output, and irrelevant stdout.

Typical reduction: **60–95%**

### Read (file)

Caps large file reads before they reach the model. Works in two phases:
- **Pre-tool**: if a file exceeds the threshold, signals the agent to read only a slice
- **Post-tool**: strips boilerplate and repeated patterns from the returned content

Typical reduction: **50–80%** on large files

### WebFetch

Strips HTML structure, CSS, navigation, ads, and repeated layout elements from fetched pages. Returns readable text content.

Typical reduction: **80–90%**

### Generic

Fallback for any tool type without a specific optimizer. Strips common JSON noise — deeply nested empty objects, null fields, redundant metadata.

## Remote optimizers (paid)

Remote optimizers run in Confire's cloud and handle tool types that require domain-specific processing. They activate automatically when the agent calls a supported tool type — no configuration needed.

### Figma

Converts raw Figma MCP output (verbose component trees, style objects, layout data) into a compact section map the model can reason about.

Typical reduction: **98%**

### GitHub

Summarizes pull request diffs and repository data — focuses on meaningful changes, strips generated files, lock files, and whitespace-only diffs.

Typical reduction: **85–95%**

### Platform-specific MCP optimizers

Confire supports source-aware optimizers for any MCP-connected platform. As you connect tools to your agent — project managers, communication tools, code review systems, design tools, databases — Confire applies an optimizer built specifically for each tool's output format.

New optimizers are added continuously. Supported tools expand as the platform grows.

## How routing works

The daemon inspects the tool name and output shape and routes to the appropriate optimizer automatically. If no specific optimizer matches, Generic runs as fallback.

Remote optimizers only activate when:
1. You're logged in (`confire login`) with a paid plan
2. The daemon can reach `api.confire.dev`

If the connection drops, local optimizers continue running uninterrupted.

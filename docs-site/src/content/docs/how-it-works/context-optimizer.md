---
title: Context Optimizer
description: How Confire strips noise from tool outputs to save tokens.
---

The context optimizer is Confire's core feature. After every tool call, Claude Code gets back raw output — often thousands of tokens of noise. The optimizer strips it down to what the model actually needs.

## Local optimizers (free)

### Bash

Trims shell output to keep failures, errors, and key results. Removes long chains of progress lines, repeated output, and irrelevant stdout.

Typical reduction: **60–95%**

### Read (file)

Caps large file reads before they're sent to the model. Works in two phases:
- **Pre-tool**: if a file exceeds the threshold, Confire signals Claude Code to read only a slice
- **Post-tool**: strips boilerplate and repeated patterns from the returned content

Typical reduction: **50–80%** on large files

### WebFetch

Strips HTML structure, CSS, navigation, ads, and repeated layout elements from fetched pages. Returns readable text content.

Typical reduction: **80–90%**

### Generic

A fallback for any tool type without a specific optimizer. Strips common JSON noise patterns — deeply nested empty objects, null fields, duplicate keys.

## Remote optimizers (paid)

Remote optimizers run in Confire's cloud and handle tool types that need domain-specific context to optimize well.

### Figma

Converts raw Figma MCP output (verbose JSX-like component trees) into a compact section map the model can reason about.

Typical reduction: **98%**

### GitHub PR

Summarizes pull request diffs, focusing on meaningful changes and stripping generated files, lock files, and whitespace-only diffs.

Typical reduction: **85–95%**

### MCP tools

Platform-specific optimizers for Jira, Slack, Linear, and other MCP-connected tools.

## How routing works

The daemon inspects the tool name and output shape and routes to the appropriate optimizer. If no specific optimizer matches, Generic runs as fallback. Remote optimizers only activate when:
1. You're logged in (`confire login`)
2. An API key is stored locally
3. The daemon can reach `api.confire.dev`

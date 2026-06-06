# Confire — Domain Glossary

## Product

**Confire** — a context + tool firewall for AI coding agents. Controls what agents can _do_ (Tool Firewall) and what tool output they can _see_ (Context Firewall).

Core framing: "Confire controls what AI agents can do and what tool output they can see."

Do **not** frame it as an optimizer-first product. Optimization is one component of the Context Firewall.

---

## The Two Pillars

**Context Firewall** — the PostToolUse layer. Runs after every tool call, before output enters the model context. Responsibilities:
- Secret redaction (high-confidence pattern matching)
- Hidden-unicode stripping (tag block, zero-width, BiDi overrides)
- Prompt-injection detection and sanitization
- MCP output normalization
- Token/context optimization (noise stripping)

**Tool Firewall** — the PreToolUse layer. Runs before every tool call. Responsibilities:
- Policy-rule evaluation (block / review / warn / allow)
- MCP risk classification
- Mutating-MCP review
- Custom and built-in rule enforcement
- Bypass-next approval flow (planned)

---

## Components

**CLI** (`confire`) — the user-facing binary. Handles setup, start/stop, login, status, policy management, and acts as the hook entrypoint.

**Daemon** — background process listening on `~/.confire/daemon.sock`. Receives all hook events, routes them through the firewall and optimizer pipeline, returns results to the hook.

**Worker** — Cloudflare Worker at `api.confire.dev`. Runs remote optimizers (Figma, GitHub, MCP tools) for paid-plan users.

**Platform** — web dashboard at `confire.dev`. Handles account, billing, and custom rule management.

---

## Phases

**PreToolUse** — fires before a tool runs. The Tool Firewall operates here.

**PostToolUse** — fires after a tool runs, before output reaches the model. The Context Firewall operates here.

---

## Optimizers

**Local optimizers** — run in the CLI daemon, free, offline, no account required. Cover: Bash, Read, WebFetch, Generic (fallback).

**Remote optimizers** — run in the Worker, require a paid plan. Cover: Figma, GitHub, Slack, Notion, Jira, and any MCP-connected tool.

---

## Policy

**Policy mode** — controls how strictly the Tool Firewall enforces rules. Values: `observe` (log only), `balanced` (default), `strict` (enforce all), `bypass` (fully passive).

**Built-in rules** — shipped in the binary. Cover destructive git operations, GitHub CLI mutations, MCP risk patterns.

**Custom rules** — cloud-managed, fetched via `confire policy pull`. Require a paid plan.

**Rule actions** — `block`, `review`, `warn`, `allow`.

---

## Hosts

**Host** — an AI coding agent that Confire integrates with. Currently: Claude Code (full support), Cursor (MCP + steer only), VS Code (MCP + steer only).

**Hook strategy** — the integration mechanism. Claude Code uses PostToolUse hooks (native output replacement possible). Cursor and VS Code use MCP proxy mode (output replacement not possible; Confire uses `additional_context` steering instead).

---

## Information Architecture

```
/
Getting Started
  Install
  Connect your agent
  Verify installation

Core Concepts
  What is Confire?
  Hook phases
  Client support modes

Context Firewall
  Overview
  Optimizers
  Secret redaction
  Injection guard
  MCP output sanitization
  Context limits

Tool Firewall
  Overview
  Built-in rules
  Policy modes
  MCP risk classifier
  Custom rules
  Bypass and approvals

Configuration
  Config file
  Per-agent settings
  Privacy and cloud optimization

Reference
  CLI commands
  Plans and limits
```

---

## Reader assumption

Target reader: a developer already running Claude Code, Cursor, or VS Code with AI agent/tool workflows. Knows what tool calls, MCP, and hooks are. Does not know what Confire installs, which phases it intercepts, what happens locally vs. remotely, which features are per-client, or how policy modes behave.

Docs should be **short, direct, product-specific**. Do not explain what an AI agent is. If hook phases are mentioned, frame them as Confire's event model only:
- SessionStart: load config/policies
- PreToolUse: Tool Firewall runs
- PostToolUse: Context Firewall runs
- SessionEnd: stats/sync

Include a small "New to MCP/hooks?" callout or glossary link where relevant — at the bottom, not in the main flow.

---

## Wording constraints

Use:
- "reviews risky tool calls"
- "detects common secret patterns"
- "sanitizes prompt-injection-like content"
- "helps keep context cleaner and safer"
- "local-first policy evaluation"

Avoid:
- "guarantees safety"
- "prevents all prompt injection"
- "catches every secret"
- "makes MCP safe"

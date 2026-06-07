---
title: Trust Labels & Provenance
description: How Confire tracks the origin and trust level of every tool output.
---

Every tool output that passes through the Context Firewall receives a
**provenance label** — a small metadata record that captures where the output
came from and how much it should be trusted. Labels are evaluated locally on
your machine and are never sent to the cloud in raw form.

## What a provenance label contains

```json
{
  "context_id":         "abc123",
  "session_id":         "sess_xyz",
  "client":             "claude_code",
  "source_tool":        "mcp__github__get_file_contents",
  "mcp_server":         "github",
  "origin_domain":      "github.com",
  "trust_level":        "mcp_trusted",
  "flags":              [],
  "redactions_count":   0,
  "sanitization_count": 0,
  "risk_score":         10,
  "created_at":         "2025-06-07T12:00:00Z"
}
```

Labels are written to `~/.confire/sessions/<session_id>.jsonl` — one JSON
line per tool call, in append order. The file is local-only; only aggregate
metadata is synced to the cloud when analytics consent is given.

## Trust levels

| Trust level           | Meaning                                                                 |
|-----------------------|-------------------------------------------------------------------------|
| `trusted_local`       | Native tool (Bash, Read, Write, Edit) running on your machine           |
| `workspace`           | Tool scoped to the current project directory                            |
| `external_untrusted`  | WebFetch or external HTTP call — content from an untrusted domain       |
| `mcp_trusted`         | Output from a known, allow-listed MCP server                            |
| `mcp_unknown`         | Output from an MCP server not on the trusted list                       |
| `sensitive`           | Any output that triggered secret redaction                              |
| `agent_generated`     | Synthetic content produced by the model itself (not a real tool result) |

Trust levels are assigned by the `Classify` function in the `provenance` package
immediately after PostToolUse. The level influences the Context Firewall's
cross-tool flow detection (see below).

## Known (trusted) MCP servers

Confire ships with a built-in allow-list of well-known MCP servers:

`github`, `gitlab`, `linear`, `jira`, `confluence`, `slack`, `notion`,
`figma`, `stripe`, `supabase`, `vercel`, `cloudflare`, `amplitude`,
`sentry`, `datadog`, `pagerduty`

Servers not on this list receive `mcp_unknown` trust level and a +25 risk
score. You can extend the list via your `~/.confire/config.yaml`.

## Flag constants

Flags are additive — a single label can carry multiple flags.

| Flag                        | Set when                                              |
|-----------------------------|-------------------------------------------------------|
| `secret_redacted`           | One or more secrets were removed from the output      |
| `prompt_injection_detected` | Classifier matched an injection pattern               |
| `hidden_unicode_stripped`   | Zero-width or invisible characters were removed       |
| `hidden_text_removed`       | CSS-hidden text (opacity:0, display:none) was removed |
| `credential_lure_detected`  | Phishing or credential-harvest pattern matched        |
| `unknown_mcp`               | Source is an unknown MCP server                       |
| `mutation_result`           | Tool is known to mutate state (write, send, deploy)   |
| `sanitized`                 | Any sanitization pass ran on this output              |

## Cross-tool flow detection

Before each PreToolUse the daemon evaluates five chain rules against the ring
buffer of the last 10 provenance labels in the current session:

| Rule | Trigger | Action |
|------|---------|--------|
| **untrusted → secret read** | `external_untrusted` or `mcp_unknown` output preceded a file read of an SSH key or `.env` file | Warn |
| **secret read → external send** | `sensitive` output preceded an outbound network call | Block (strict) / Warn (balanced) |
| **injection → shell** | `prompt_injection_detected` flag preceded a Bash/shell tool call | Warn |
| **unknown MCP → mutating MCP** | `unknown_mcp` flag preceded a mutating MCP tool call | Warn |
| **credential lure → message send** | `credential_lure_detected` flag preceded a messaging tool (Slack, email, etc.) | Warn |

Flow rule decisions are logged to the session JSONL file alongside individual
provenance labels. They are also surfaced in the session summary printed when
the daemon stops.

## Session summary

When a session ends, Confire prints a trust distribution summary:

```
Session ended  (23 tool calls, 0 blocked, 2 warned)

Trust distribution:
  trusted_local       16
  mcp_trusted          4
  external_untrusted   2
  mcp_unknown          1
```

Unknown MCP servers and external-untrusted calls are highlighted so you can
review which tools introduced external content.

## Roadmap

- **v1.1:** User-defined trusted MCP server list via `~/.confire/config.yaml`
- **v1.1:** Dashboard view of per-session provenance timelines
- **v1.2:** Flow rule tuning — adjust which rules run in balanced vs strict mode

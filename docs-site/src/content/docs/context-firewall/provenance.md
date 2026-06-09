---
title: Trust labels and provenance
description: >-
  How Confire attaches trust labels to supported tool events and uses
  provenance to detect risky activity chains.
---

Confire can attach trust labels to supported tool events.

A trust label is a small metadata record that describes where a tool
result came from, how trusted it is, and whether Confire detected
anything security-relevant. Trust labels help Confire answer questions
like:

- Did this result come from a local tool or an external service?
- Did it come from a known or unknown MCP server?
- Did it contain secret-looking values?
- Did it contain prompt-injection-like text?
- Did a risky action happen after untrusted content appeared?

## Example label

```json
{
  "context_id":   "ctx_abc123",
  "session_id":   "sess_xyz",
  "client":       "claude_code",
  "source_tool":  "mcp__github__get_file_contents",
  "mcp_server":   "github",
  "origin_domain":"github.com",
  "trust_level":  "mcp_trusted",
  "flags":        [],
  "risk_score":   10,
  "created_at":   "2026-06-07T12:00:00Z"
}
```

Labels are stored locally as session metadata. When dashboard sync is
enabled, Confire syncs metadata-only security events. Raw tool output,
source code, secret values, and full command output are not synced by
default.

## Trust levels

| Trust level          | Meaning                                                   |
|----------------------|-----------------------------------------------------------|
| `trusted_local`      | A local tool running on your machine                      |
| `workspace`          | A tool scoped to the current project                      |
| `external_untrusted` | External web or HTTP content                              |
| `mcp_trusted`        | A known or user-trusted MCP server                        |
| `mcp_unknown`        | An MCP server not yet trusted                             |
| `sensitive`          | A result or action involving sensitive data               |
| `agent_generated`    | Content produced by an agent or subagent                  |

Trust levels are not permanent guarantees. They are signals used by
Confire policies and session summaries.

## Flags

Flags are added when Confire detects security-relevant patterns.

| Flag                        | Meaning                                             |
|-----------------------------|-----------------------------------------------------|
| `secret_like_value_detected`| Credential-like content appeared                    |
| `prompt_injection_detected` | Instruction-like content appeared in a tool result  |
| `hidden_unicode_detected`   | Invisible or control characters appeared            |
| `credential_lure_detected`  | Credential-harvest or phishing-like text appeared   |
| `unknown_mcp`               | Source was an unknown MCP server                    |
| `mutation_result`           | Tool appears to change external or project state    |
| `large_output`              | Result was unusually large                          |

## Unknown MCP servers

Known MCP servers can be trusted by built-in rules, user
configuration, or future registry metadata. Unknown MCP servers
receive a more cautious label:

```
trust_level: mcp_unknown
flag:        unknown_mcp
```

Unknown does not always mean dangerous. It means Confire should treat
the result more carefully until the server is trusted.

## Flow detection

Provenance enables Confire to detect risky chains across tool events.

| Flow                                         | Example action        |
|----------------------------------------------|-----------------------|
| Untrusted content → secret-file read         | review                |
| Sensitive data → external send               | review or block       |
| Prompt-injection-like result → shell command | warn or review        |
| Unknown MCP result → mutating MCP action     | warn or review        |
| Credential-lure content → message send       | warn or review        |

These rules are designed to catch suspicious sequences, not just
isolated tool calls.

## Session summary

At the end of a session, Confire can summarize trust labels and
security events:

```
Session ended  (23 tool calls, 0 blocked, 2 warned)

Trust distribution:
  trusted_local        16
  mcp_trusted           4
  external_untrusted    2
  mcp_unknown           1
```

This helps you see which tools introduced local, external, or unknown
context during the session.

## Roadmap

Upcoming provenance features may include:

- user-defined trusted MCP server lists,
- dashboard provenance timelines,
- configurable flow rules,
- signed registry metadata for known MCP servers and advisories.

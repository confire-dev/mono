---
title: Hook phases
description: The four events Confire intercepts in an agent session and what happens at each phase.
---

Confire can receive events at key points in an agent session.
Understanding each phase helps you reason about when firewall decisions
are made and when security context is added.

## SessionStart

When you open a new session in your agent, Confire receives a
`SessionStart` event where supported. Confire uses this phase to:

- load local configuration,
- load cached policy rules,
- check daemon health,
- prepare the local firewall engine,
- optionally check for synced policy updates if you are logged in.

You may see a one-time welcome or status message in the agent output
after first installing Confire.

## PreToolUse

Before a supported tool executes, Confire receives the tool name and
input. The Tool Firewall evaluates the call against active policy rules
and returns one of four decisions:

**allow** — the tool proceeds normally.

**warn** — the tool proceeds, but Confire adds an advisory note where
the client supports it.

**review** — the tool is paused. Confire shows the reason and asks for
user approval before retrying. Typical review examples:

- force pushes
- destructive Git operations
- secret-file reads
- database resets
- production deploys
- package publishing
- mutating MCP actions

**block** — the tool is prevented from running. Confire shows the block
reason. Blocks should be reserved for high-confidence dangerous
behavior.

## PostToolUse

After a supported tool finishes, Confire receives information about the
tool result. The Tool Result Firewall inspects the result and can return
additional firewall context to the agent where supported.

This firewall context can include:

- secret-looking value warnings,
- hidden Unicode warnings,
- prompt-injection-like content warnings,
- MCP risk notes,
- provenance metadata,
- policy decisions,
- suggested next steps.

Confire does not need to replace the original tool output to be useful.
It helps the agent and user understand whether a tool result contained
suspicious or security-relevant content.

Depending on the client and tool surface, Confire may add advisory
context, record metadata, or enforce policy decisions. Some clients may
not allow Confire to replace native tool output.

## SessionEnd

When the agent session closes, Confire records session stats. Session
stats can include:

- tools inspected,
- rules fired,
- warnings shown,
- reviews requested,
- blocks applied,
- suspicious result events detected,
- policy mode used.

If you are logged in, metadata-only security events can sync to your
dashboard.

---

:::note
Confire only inspects supported agent/tool events. It does not modify
your source files, project state, or task instructions. It does not
need to read your repository outside the tool events your agent already
performs.

Raw tool output is not forwarded to Confire Cloud by default. When
dashboard sync is enabled, Confire sends metadata-only security events
— risk level, action taken, rule ID, client, timestamp, and session
metadata.
:::

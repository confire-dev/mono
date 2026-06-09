---
title: Verify your setup
description: Confirm Confire is installed, active, and doing its job.
---

## Check status

```bash
confire status
```

This shows your local Confire setup, including connected agents, account
status, firewall mode, daemon status, and optional dashboard sync.

A healthy setup looks like:

```
[confire status]

  Agents
  ✓  Claude Code         found · hooks installed

  Account
  ✓  Signed in           you@example.com · Free plan

  Firewall
  ✓  Mode                balanced
  ✓  Daemon              running
  ✓  Tool Firewall       active
  ✓  Tool Result Firewall active

  Dashboard Sync
  ✓  Security events     metadata-only sync enabled
```

If hooks are not installed, run:

```bash
confire setup
```

If the daemon is stopped, run:

```bash
confire start
```

## Test the Tool Firewall

Use `confire policy test` to simulate a tool-input evaluation without
running the command:

```bash
confire policy test 'git push --force'
```

Expected output:

```
Action:   review
Rule:     Review destructive Git operation
Severity: high
Reason:   Force push can rewrite remote branch history and affect open PRs.
```

Try a safe command to confirm it passes through:

```bash
confire policy test 'git status'
```

Expected output:

```
Action:   allow
```

## Test the Tool Result Firewall

Start a session in your agent and run a supported tool command that
produces a result, such as a shell command, MCP call, or fetched
external content.

After the tool returns, Confire inspects the result and may add firewall
context back to the agent, including:

- secret-looking value warnings
- prompt-injection warnings
- hidden Unicode warnings
- MCP risk notes
- provenance metadata
- policy decisions

To check recent findings, run:

```bash
confire events
```

Or open the dashboard and go to **Overview → Recent activity**.

If your client does not support replacing native tool output, Confire
still records metadata and returns advisory firewall context where
supported.

## Next step

[Try a review flow →](../connect)

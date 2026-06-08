---
title: Verify your setup
description: Confirm Confire is installed, active, and doing its job.
---

## Check status

```bash
confire status
```

This shows four sections: detected agents and hook state, account
and plan, context firewall status, and telemetry. A healthy setup looks like:

```
[confire status]

  Agents
  ✓  Claude Code         found ✓  hook installed ✓

  Account
  ✓  API key            you@example.com · Pro plan

  Context Firewall
  ✓  Status             active (cloud + local)

  Telemetry
  ✓  Analytics          on
```

If the hook shows as not installed, run `confire setup`. If the
daemon is stopped, run `confire start`.

## Test the Tool Firewall

Use `confire policy test` to simulate a PreToolUse evaluation without
running anything:

```bash
confire policy test 'git push --force'
```

Expected output:

```
Action:   review
Rule:     Review destructive git operation
Severity: high
Reason:   Force push can rewrite remote branch history and affect open PRs.
```

Try a safe command to confirm it passes through:

```bash
confire policy test 'git status'
# Action: allow
```

## See the Context Firewall in action

Start a session in your agent and run an MCP tool or a shell command
that produces noisy output. After the tool finishes, Confire processes
the output through the context pipeline and records any findings.

To see what Confire detected, check the dashboard or run:

```bash
confire status
```

The "Context Firewall" section shows active status. Security events
appear in the dashboard under **Overview → Recent activity**.

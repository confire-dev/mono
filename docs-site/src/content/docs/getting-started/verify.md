---
title: Verify your setup
description: Confirm Confire is installed, active, and doing its job.
---

## Check status

```bash
confire status
```

This shows four sections: detected agents and hook state, account
and plan, optimizer status, and telemetry. A healthy setup looks like:

```
[confire status]

  Agents
  ✓  Claude Code         found ✓  hook installed ✓

  Account
  ✓  API key            you@example.com · Pro plan · 4120/10000 credits

  Optimizer
  ✓  Status             active (cloud + local fallback)

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

Start a session in your agent and run a tool that produces large
output — a shell command, a file read, or a web fetch. After the
tool finishes, you'll see a Confire notification in the agent output:

```
🔥 Confire saved ~3k tokens on this Bash response.
```

Savings appear when output was trimmed by at least the configured
minimum (default: 5,000 tokens). You can adjust this threshold with
`confire config set notifications.min_saved_tokens=1000` to see all
saves.

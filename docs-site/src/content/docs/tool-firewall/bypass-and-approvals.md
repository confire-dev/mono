---
title: Bypass and approvals
description: >-
  How to approve a single reviewed tool call with bypass-next, and
  when to change policy mode instead.
---

When Confire returns a review, the tool is paused and the agent waits
for user approval.

```
CONFIRE REVIEW REQUIRED

Rule:              Review destructive Git operation
Tool:              Bash
Command:           git push origin main --force
Risk:              high severity
Why this matters:  Force push can rewrite remote branch history and affect open PRs.

ACTION REQUIRED
Confire flagged this command. Do you want the agent to run it anyway?
If yes, run:
  confire bypass-next
Then ask the agent to retry.
```

If you checked the command and it is intentional, run:

```bash
confire bypass-next
```

Then tell the agent to retry. That's it.

## What bypass-next does

`confire bypass-next` creates a one-shot approval. The next reviewed
tool call is allowed once, then the approval is consumed automatically.
After that, Confire returns to normal enforcement.

This is not the same as disabling the firewall.

## When to use it

Use `bypass-next` when:

- you understand why Confire reviewed the action,
- the command is intentional,
- you want to allow this specific retry,
- you want the firewall to stay active afterward.

Common examples:

- intentional force push,
- intentional destructive cleanup,
- approved production deploy,
- approved secret-file inspection,
- approved mutating MCP action.

## Review vs. block

`bypass-next` is for review outcomes.

A blocked tool call does not run through normal retry approval. If a
block fires on something your team wants to allow, change the policy
intentionally instead of bypassing it ad hoc:

- edit a custom rule,
- change a rule group,
- switch policy mode,
- update project policy.

## Temporarily changing mode

For short periods where you want Confire to observe without
interrupting, use observe mode:

```bash
confire mode observe
```

Return to balanced mode:

```bash
confire mode balanced
```

Use bypass mode only when you intentionally want enforcement off:

```bash
confire mode bypass
```

Return to normal enforcement:

```bash
confire mode balanced
```

Balanced mode is recommended for daily work.

## Check current mode

```bash
confire status
```

The status output shows the active policy mode and whether the local
firewall is running.

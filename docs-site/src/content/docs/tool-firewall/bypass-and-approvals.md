---
title: Bypass and approvals
description: >-
  How to approve a single reviewed tool call without
  disabling the firewall.
---

The firewall just fired a review. Your agent is paused:

```
CONFIRE REVIEW REQUIRED

Rule:              Review destructive git operation
Claude is about to run: Bash — git push origin main --force
Risk:              high severity
Why this matters:  Force push can rewrite remote branch history and affect open PRs.

ACTION REQUIRED — ask the user:
"Confire flagged this command. Do you want me to run it anyway?
If yes: run 'confire bypass-next' in your terminal, then tell me to retry."
```

You've checked it. It's intentional. Run:

```bash
confire bypass-next
```

Then tell the agent to retry. That's it.

## What bypass-next does

`confire bypass-next` sets a one-shot approval flag. The very next
PreToolUse event skips all firewall evaluation and the tool runs.
After that single use the flag is automatically cleared — the
firewall returns to normal enforcement immediately. This is a
one-shot approval, not disabling the firewall.

## Disabling the firewall for a longer sequence

If you need to run several operations in a row without review
interruptions, switch to bypass mode:

```bash
confire off          # sets mode to bypass
# ... do your work ...
confire on           # re-enables balanced mode
```

Restart the daemon after switching:

```bash
confire stop && confire start
```

## Block vs. review

`bypass-next` works only for review outcomes. A **block** has no
retry path from the agent — the tool won't run regardless. If a
block fires on something you intend to do, the right fix is a
custom rule override or a group disable. See
[Custom rules](custom-rules).

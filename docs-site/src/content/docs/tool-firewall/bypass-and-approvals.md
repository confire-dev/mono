---
title: Bypass and approvals
description: >-
  How to let a single reviewed tool call through without
  changing your policy mode.
---

When the Tool Firewall reviews a tool call and you've confirmed
it's safe, you can let the next call through without disabling
the firewall entirely.

## Bypass-next

```bash
confire bypass-next
```

Sets a one-shot flag that causes the firewall to pass through the
very next PreToolUse event without review, warn, or block. After
that one event, the flag is automatically cleared and the firewall
returns to normal enforcement.

This is the right tool when your agent is paused on a review and
you've inspected the command and want to proceed. It doesn't change
your mode or persist beyond the single call.

## Disabling the firewall temporarily

If you need to run a sequence of operations without review
interruptions, you can switch to bypass mode:

```bash
confire off          # sets mode to bypass
# ... do your work ...
confire on           # re-enables balanced mode
```

`confire off` disables the firewall completely. `confire on`
re-enables it at `balanced`. Remember to restart the daemon after
switching:

```bash
confire stop && confire start
```

## Permanent changes

If a built-in rule fires too often on something your team considers
safe, the right fix is a custom rule override or a group disable —
not leaving the firewall off. See [Custom rules](custom-rules) for
managing rule groups.

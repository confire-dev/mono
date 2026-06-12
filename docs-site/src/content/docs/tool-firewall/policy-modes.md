---
title: Policy modes
description: >-
  How Confire's four modes control firewall enforcement
  across all rules.
---

The active policy mode determines how strictly the Tool Firewall
enforces rules. The mode applies globally. It doesn't change which
rules exist, only how they fire.

## Modes

**`observe`**: the firewall runs and records all matches but never
blocks, reviews, or warns. Use this to audit what would fire in your
workflow before enabling enforcement. Matches appear in daemon logs.

**`balanced`** (default): built-in rules fire at their stated
severity. Review-severity rules pause the tool and ask for
confirmation. Warn-severity rules allow the tool but add context.
Block-severity rules prevent execution.

**`strict`**: all rules fire, including rules gated behind strict
mode (`min_mode: strict`). Higher-risk heuristics that are off by
default in balanced mode become active.

**`bypass`**: the firewall is fully passive. No rules fire, no
blocks, no reviews, no warnings. The daemon continues to run for
context processing; only the firewall is disabled.

## Switching modes

### Quick toggles

```bash
confire on     # enable firewall, set mode to balanced
confire off    # set mode to bypass (firewall disabled)
```

### Set a specific mode

```bash
confire config set mode=observe
confire config set mode=balanced
confire config set mode=strict
confire config set mode=bypass
```

Restart the daemon after changing mode:

```bash
confire stop && confire start
```

### Check the current mode

```bash
confire policy status
```

## Choosing a mode

Start with `observe` if you want to understand what Confire would
do in your existing workflow before it intercepts anything.
`balanced` is the right default for most developers. It catches
the high-risk operations without creating review friction for normal
work. Use `strict` if you want maximum coverage. Use `bypass`
temporarily when you need to run something uninterrupted and
trust what you're doing.

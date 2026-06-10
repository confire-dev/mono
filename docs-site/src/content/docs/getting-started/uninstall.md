---
title: Uninstall
description: Remove Confire from your AI coding agent and your machine.
---

## Remove Confire integrations

`confire reset` removes Confire's hooks from all supported agent settings and
stops the daemon. The Confire binary remains on your machine after this step.

```bash
confire reset
```

After `confire reset`, your agents continue normally. Tool calls pass through
unmodified — Confire no longer intercepts anything.

## Stop the daemon

If the daemon is still running, stop it:

```bash
confire stop
```

## Remove the Confire binary

Remove the binary from `~/.local/bin`:

```bash
rm ~/.local/bin/confire
```

## Remove Confire config and data

```bash
rm -rf ~/.config/confire
```

This removes local event history, policy cache, and configuration.

## Verify removal

```bash
which confire
```

If nothing is returned, the binary has been removed. Your agent settings files
(`.claude/settings.json`, `.cursor/hooks.json`, etc.) can be checked manually
to confirm no Confire entries remain after `confire reset`.

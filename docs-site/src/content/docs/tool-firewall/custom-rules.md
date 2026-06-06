---
title: Custom rules
description: >-
  Add organization-specific firewall rules managed through
  the Confire dashboard.
---

Custom rules extend the built-in rule set with your own
organization-specific policies. They're managed in the Confire
dashboard and synced to local devices on demand.

:::note
Custom rules require a paid plan. The `confire policy pull` command
and dashboard rule management are available on paid plans only.
:::

## How it works

Custom rules are stored in the Confire cloud and cached locally at
`~/.confire/policies/cache.json`. The daemon merges them with the
built-in rules at startup and uses the combined set for all
evaluations.

The same rule schema applies to custom rules as to built-in rules:
phase, action, severity, match conditions, and message. Custom rules
can cover any tool type and any phase.

## Pulling the latest rules

```bash
confire policy pull
```

This fetches the latest custom rules from your account and writes
them to the local cache. Run this after making changes in the
dashboard, or set it up in your team's onboarding flow so everyone
starts with the current rule set.

## Group overrides

The dashboard lets you toggle rule groups on or off. These overrides
sync alongside the rules when you run `confire policy pull`. For
example, if your team has reviewed and approved all MCP tools you
use, you can disable the `mcp.risk_classifier` group in the dashboard
and pull — all devices that pull will suppress that group without
changing the rule definitions.

## Checking rule state

```bash
confire policy status
```

Shows a count of built-in rules, custom rules, and the last time
the cache was fetched.

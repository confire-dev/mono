---
title: Custom guardrails
description: >-
  Extend Confire with project, team, or organization-specific rules,
  managed in the dashboard, synced and evaluated locally.
---

Custom rules let you extend Confire with project, team, or
organization-specific guardrails. They are managed in the Confire
dashboard, synced to your local machine, and evaluated by the Confire
daemon.

Custom rules are useful when your team has policies that Confire cannot
know by default, such as:

- review production deploy commands,
- block access to specific credential files,
- review commands touching customer data,
- require review for internal MCP tools,
- warn when agents use unapproved external services,
- block risky actions in specific repositories.

## Availability

Custom rules are part of Dev early access. The local firewall and
built-in rules continue to work on the Free plan without an account.

## How it works

Custom rules are stored in Confire Cloud and pulled to your local
machine. After syncing, the daemon merges:

1. built-in rules,
2. synced custom rules,
3. local configuration.

Evaluation still runs locally. Raw tool inputs and outputs do not need
to be sent to Confire Cloud for custom rules to work.

## Pull the latest rules

After changing rules in the dashboard, run:

```bash
confire policy pull
```

This fetches the latest custom rules for your account and updates the
local policy cache. You can use this in onboarding scripts so new
devices start with the current rule set.

## Check rule state

```bash
confire policy status
```

This shows:

- active policy mode,
- built-in rule count,
- custom rule count,
- last policy sync,
- enabled rule groups.

## Rule groups

The dashboard can group related rules together. Examples:

- Git safety
- Secret-file access
- MCP mutations
- Deploy protection
- Database protection
- External sends
- Organization rules

Rule groups make it easier to tune policy for a project or team. For
example, a team may choose to:

- review all production deploys,
- warn on unknown MCP tools,
- block repository deletion,
- require review before sending content to Slack or email.

## Example custom rule

```yaml
id: review-production-deploy
name: Review production deploys
phase: PreToolUse
action: review
severity: high
match:
  tool: Bash
  command_contains:
    - "vercel --prod"
    - "wrangler deploy"
    - "terraform apply"
message: Production deploys should be reviewed before execution.
```

If a matching command appears, Confire pauses the tool call and asks
for approval.

## Local-first behavior

Custom rules are evaluated locally by the Confire daemon. When
dashboard sync is enabled, Confire can sync metadata-only events such
as:

- rule ID,
- action taken,
- severity,
- client,
- timestamp,
- session ID.

Raw command output, source code, secrets, and full tool results are
not synced.

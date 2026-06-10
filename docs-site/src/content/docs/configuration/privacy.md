---
title: Privacy
description: >-
  What stays local by default and what can sync to Confire Cloud
  when cloud features are enabled.
---

Confire is local-first. Firewall decisions run in the Confire daemon
on your machine. Built-in rules work without an account, and raw tool
content is not required for Confire Cloud features.

This page explains what stays local and what can sync when cloud
features are enabled.

## What stays local by default

The following stay on your machine by default:

- source files and project contents,
- conversation history and agent transcripts,
- raw tool output,
- secret values,
- full command output,
- full file contents,
- policy evaluation inputs,
- firewall rule evaluation.

Confire evaluates built-in rules locally. Custom rules are pulled to
your machine and evaluated locally.

## What can sync to Confire Cloud

When you are logged in and dashboard sync is enabled, Confire can sync
metadata-only security events.

Examples of synced metadata:

- rule ID,
- action taken,
- risk level,
- tool category,
- client,
- session ID,
- timestamp,
- policy mode,
- MCP server name,
- MCP tool name,
- trust label,
- finding type.

Examples of finding types:

- `secret_like_value_detected`,
- `prompt_injection_detected`,
- `hidden_unicode_detected`,
- `unknown_mcp`,
- `mutating_mcp_action`,
- `destructive_git_operation`.

Synced metadata does not include matched secret values or raw tool
output.

## Dashboard sync

Dashboard sync powers features such as:

- security event history,
- session summaries,
- connected clients,
- policy sync status,
- custom rule visibility,
- provenance timelines.

Disable dashboard sync:

```bash
confire config set dashboard_sync.enabled=false
```

The local firewall continues to work with built-in rules.

## Policy sync

If you use custom rules, Confire pulls policy metadata from Confire
Cloud and caches it locally. Policy sync can include:

- rule IDs,
- rule names,
- match conditions,
- actions,
- severity,
- rule groups.

Policy sync does not require sending raw tool output or source code to
Confire Cloud.

Disable automatic policy sync:

```bash
confire config set policy_sync.auto_pull=false
```

Pull policies manually:

```bash
confire policy pull
```

## Local-only mode

You can use Confire without logging in. Without an account, Confire
runs in local-only mode:

- built-in rules work,
- Tool Firewall works,
- Tool Result Firewall works,
- local events can be recorded,
- cloud dashboard sync is disabled,
- custom cloud rules are unavailable.

You can also turn off dashboard sync while staying logged in:

```bash
confire config set dashboard_sync.enabled=false
```

## Product analytics

Confire may collect product analytics when enabled, such as:

- install success or failure,
- CLI version,
- client type,
- feature usage counts,
- error codes,
- aggregate firewall action counts.

Product analytics does not include source code, raw tool output,
secret values, full shell commands, full file paths, or conversation
transcripts.

Opt out:

```bash
confire config set analytics.enabled=false
```

## Security Registry

When enabled, Confire may fetch signed Security Registry updates.
Registry updates can include:

- known risky MCP server metadata,
- public advisories,
- recommended rule metadata,
- suspicious tool patterns.

Registry updates are downloaded to your machine and evaluated locally.
They do not require sending your tool output to Confire Cloud.

## Support reports

If you report a bug or suspicious MCP server, only include information
you are comfortable sharing. Avoid including raw secrets, private
source code, or full tool output in support reports.

---

Confire's default model:

- rules and registry metadata can come down,
- metadata-only security events can go up,
- raw tool content stays local by default.

:::caution
Confire is a guardrail layer, not a guarantee. Use it together with
normal security practices: least-privilege credentials, secret
scanning, environment isolation, and careful MCP server review.
:::

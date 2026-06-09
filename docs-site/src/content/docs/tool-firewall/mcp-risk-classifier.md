---
title: MCP risk classifier
description: >-
  How Confire classifies MCP tool calls before they run — using tool
  names and input signals to identify likely risky behavior.
---

The MCP risk classifier runs before supported MCP tool calls. It helps
Confire identify MCP tools that may execute commands, write data,
delete resources, send information externally, or handle credentials.

The classifier is a heuristic. It looks at tool names and input
parameter names. It does not prove what a tool actually does.

## What it looks for

Confire scores MCP tools using signals such as:

| Signal | Examples | Typical action |
|---|---|---|
| Shell or execution terms | `shell`, `exec`, `bash`, `terminal`, `run_command`, `eval` | review |
| Destructive or write terms | `write`, `delete`, `drop`, `wipe`, `truncate`, `remove` | review |
| External send terms | `send`, `email`, `upload`, `export`, `webhook`, `post` | warn or review |
| Deployment terms | `deploy`, `publish`, `release`, `promote` | review |
| Payment or account terms | `refund`, `charge`, `transfer`, `invite`, `approve` | review |
| Credential parameters | `password`, `token`, `secret`, `api_key`, `credential` | warn or review |

A tool can match multiple signals. Higher-risk combinations may be
escalated.

## Example: shell-like MCP tool

```
tool: mcp__unknown_server__execute_bash

CONFIRE REVIEW REQUIRED
rule:    MCP risk classifier
risk:    tool name suggests shell execution
action:  run `confire bypass-next` to allow once, then retry
```

## Example: external send tool

```
tool: mcp__slack__post_message

CONFIRE REVIEW REQUIRED
rule:    Review mutating MCP action
risk:    this MCP tool can send content to an external service
action:  run `confire bypass-next` to allow once, then retry
```

## What the classifier catches

The classifier is useful for broad MCP coverage, especially with
unknown or custom MCP servers. It can catch tools like:

- `mcp__server__execute_shell`
- `mcp__server__delete_record`
- `mcp__server__send_email`
- `mcp__server__upload_file`
- `mcp__server__trigger_deploy`
- `mcp__server__create_refund`

## What it cannot guarantee

The classifier is name-based. A dangerous tool with a harmless name
may not be scored correctly. For example:

```
mcp__server__get_user
```

could be safe, or it could hide risky behavior behind a vague name.

A safe tool with a risky name may also be reviewed. For example:

```
mcp__server__update_local_cache
```

may be harmless, but Confire may still treat it cautiously because it
contains a mutation verb.

## Relationship to built-in MCP rules

The MCP risk classifier works alongside built-in MCP rules:

- The classifier scores broad risk signals.
- Built-in rules catch known mutation verbs and sensitive action
  categories.
- Custom guardrails can override or extend behavior.
- Future Security Registry metadata can improve MCP-specific decisions.

When multiple rules match, Confire uses the highest-severity result.

## Testing MCP tools

Test how Confire would classify an MCP tool name without running it:

```bash
confire policy test 'mcp__github__merge_pull_request'
```

Example output:

```
Action:   review
Rule:     Review mutating MCP action
Severity: medium
Reason:   MCP tool name suggests it can mutate external state.
```

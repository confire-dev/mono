---
title: Policy rules
description: Complete reference for the Confire policy file syntax.
---

Policy files are YAML. Confire looks for them in two places:

- `~/.confire/policy.yaml` — global, applies to all projects
- `<project-root>/.confire.yaml` — project-specific, merges with global

Project rules are evaluated first. If no rule matches at the project level, global rules are checked.

## File structure

```yaml
version: 1          # required, must be 1
default: allow      # action when no rule matches: allow | block

rules:
  - name: my-rule   # optional, used in logs
    when: ...       # matching criteria
    do: allow       # action: allow | block | require-approval
    reason: "..."   # optional, shown in block/approval messages
```

## The `when` block

All fields in `when` are optional. A rule with an empty `when` matches everything.

### `tool`

Match on tool name (string or glob):

```yaml
when:
  tool: bash
```

```yaml
when:
  tool: "write_*"   # glob: write_file, write_document, etc.
```

### `server`

Match on MCP server name:

```yaml
when:
  server: filesystem
  tool: write_file
```

### `args`

Match on argument values. Values are treated as regular expressions:

```yaml
when:
  tool: bash
  args:
    command: "curl|wget|nc"   # regex: any of these words
```

```yaml
when:
  tool: write_file
  args:
    path: "^\\/etc"   # starts with /etc
```

### `category`

Match by built-in category:

```yaml
when:
  category: destructive
```

Built-in categories: `network`, `destructive`, `read`, `write`, `execute`, `auth`

## Actions

### `allow`

Pass the tool call through unchanged.

```yaml
- when:
    tool: read_file
  do: allow
```

### `block`

Return an error to the agent. The real tool server is never called.

```yaml
- when:
    tool: delete_file
  do: block
  reason: "File deletion is disabled"
```

### `require-approval`

Pause and prompt the user. The user can allow, block, or allow-always.

```yaml
- when:
    tool: bash
  do: require-approval
```

## Rule evaluation

Rules are evaluated **top-to-bottom**. The **first match wins**. If no rule matches, the `default` action is used (defaults to `allow`).

```yaml
rules:
  - when:
      tool: read_file
      args:
        path: "\\.env"     # .env files → always block
    do: block

  - when:
      tool: read_file      # other reads → allow
    do: allow

  - when:
      category: write      # all writes → approve
    do: require-approval
```

## Full example

```yaml
version: 1
default: allow

rules:
  # Hard blocks
  - name: no-env-reads
    when:
      tool: read_file
      args:
        path: "(\\.env|\\.env\\..*|credentials\\.json)"
    do: block
    reason: "Reading credential files is not allowed"

  - name: no-destructive-shell
    when:
      tool: bash
      args:
        command: "rm\\s+(-rf?|--recursive)"
    do: block
    reason: "Destructive rm is disabled"

  # Approvals
  - name: approve-shell
    when:
      tool: bash
    do: require-approval

  - name: approve-writes-outside-src
    when:
      tool: write_file
      args:
        path: "^(?!src/)"   # not starting with src/
    do: require-approval
```

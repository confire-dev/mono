---
title: Test a policy
description: Write a simple policy rule and see Confire enforce it.
---

Policies are YAML files that tell Confire what to do with tool calls and context. This guide walks you through writing and testing your first rule.

## Default policy location

```
~/.confire/policy.yaml        # global (all projects)
<project-root>/.confire.yaml  # project-specific (takes precedence)
```

## A minimal policy

Create `.confire.yaml` in your project root:

```yaml
version: 1

rules:
  - name: block-shell
    when:
      tool: bash
    do: block
    reason: "Shell execution requires manual approval"
```

This blocks any `bash` tool call your agent attempts.

## Test it

With `confire start` running in the background, trigger a tool call from your agent or use the CLI:

```bash
confire test --tool bash --args '{"command": "ls"}'
```

You should see:

```
✗ BLOCKED  bash  →  "Shell execution requires manual approval"
```

## Approve instead of block

Change `do: block` to `do: require-approval` to get an interactive prompt:

```yaml
rules:
  - name: approve-shell
    when:
      tool: bash
    do: require-approval
```

Next time the agent calls `bash`, Confire will pause and ask:

```
Agent wants to run: bash {"command": "ls -la"}
Allow? [y/N/always]
```

## Allow specific tools

You can combine rules. For example, allow `read_file` freely, require approval for `write_file`, and block `delete_file`:

```yaml
version: 1

rules:
  - name: allow-reads
    when:
      tool: read_file
    do: allow

  - name: approve-writes
    when:
      tool: write_file
    do: require-approval

  - name: block-deletes
    when:
      tool: delete_file
    do: block
```

Rules are evaluated top-to-bottom. The first match wins.

## Next steps

- [Full policy syntax reference →](/configuration/policy-rules)
- [Custom guardrails →](/configuration/custom-guardrails)
- [How the firewall works →](/how-it-works/tool-firewall)

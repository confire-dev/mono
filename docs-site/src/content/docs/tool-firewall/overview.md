---
title: Tool Firewall
description: >-
  Confire's PreToolUse layer. Evaluates supported tool calls before
  they run and decides whether to allow, warn, review, or block.
---

The Tool Firewall is Confire's PreToolUse layer.

Before a supported tool runs, Confire evaluates the action against
policy rules and decides whether it should be allowed, warned,
reviewed, or blocked.

## Example: destructive filesystem change

Your agent decides to run:

```bash
rm -rf docs-site/src/content/docs/clients docs-site/src/content/docs/how-it-works
```

Before it executes, Confire intercepts it:

```
CONFIRE REVIEW REQUIRED

Rule:              Review destructive filesystem change
Tool:              Bash
Command:           rm -rf docs-site/src/content/docs/clients docs-site/src/content/docs/how-it-works
Risk:              high severity
Why this matters:  Recursive force delete permanently removes files with no trash recovery.

ACTION REQUIRED
Confire flagged this command. Do you want the agent to run it anyway?
If yes, run:
  confire bypass-next
Then ask the agent to retry.
```

That is a review. The tool is paused. Your agent is waiting for user
approval.

## The four outcomes

### Allow

No rule matched, or the action is considered safe. The tool runs
normally.

```
Action: allow
```

### Warn

A lower-severity rule matched. The tool can proceed, but Confire adds
an advisory note where supported.

```
CONFIRE WARNING
Rule:    Review mutating MCP tool
Tool:    mcp__stripe__create_charge
Risk:    this MCP tool appears to mutate external state
Action:  allowed with warning
```

### Review

A higher-severity rule matched. The tool is paused until the user
approves.

```
CONFIRE REVIEW REQUIRED
Rule:    Review force pushes
Tool:    Bash
Command: git push --force-with-lease origin main
Risk:    force pushes can rewrite remote branch history
Action:  run `confire bypass-next` to allow once, then retry
```

`confire bypass-next` is a one-shot approval. It allows the next
matching call through and is then consumed.

### Block

A block rule matched. The tool does not run.

```
CONFIRE BLOCKED TOOL CALL
Rule:    Delete remote repository
Tool:    Bash
Command: gh repo delete my-org/my-repo --yes
Reason:  Permanent repository deletion cannot be undone.
```

Blocks are reserved for high-confidence dangerous behavior.

## What gets evaluated

The Tool Firewall evaluates supported tool calls before they run.

### Shell commands

Confire can review risky shell commands such as:

- `rm -rf`
- `git push --force`
- `git reset --hard`
- `gh repo delete`
- `terraform destroy`
- package publishing
- database reset commands

### Secret-file access

Confire can review attempts to read sensitive files such as:

- `.env`
- `.env.production`
- private keys
- SSH keys
- credential files

### MCP tools

Confire can evaluate MCP tool calls by looking at:

- MCP server name,
- MCP tool name,
- whether the tool appears to mutate state,
- whether the tool touches external systems,
- whether the server is known or unknown,
- policy rules and registry metadata when available.

Examples: `mcp__stripe__create_refund`,
`mcp__slack__post_message`, `mcp__github__merge_pull_request`,
`mcp__database__delete_record`.

## Policy modes

The active policy mode controls how strictly Confire responds.

| Mode      | Behavior                                                      |
|-----------|---------------------------------------------------------------|
| `observe` | Record matches without interrupting the agent                 |
| `balanced`| Default. Warns and reviews common risky actions               |
| `strict`  | More sensitive. Reviews or blocks more aggressively           |
| `bypass`  | Temporarily disables enforcement                              |

Use `balanced` for normal daily work. Use `observe` when first rolling
Confire out to see what it would catch without interrupting users.

## Related pages

- [Built-in rules](../built-in-rules)
- [Policy modes](../policy-modes)
- [MCP security](../../context-firewall/mcp-security)
- [Custom guardrails](../custom-rules)
- [Bypass and approvals](../bypass-and-approvals)

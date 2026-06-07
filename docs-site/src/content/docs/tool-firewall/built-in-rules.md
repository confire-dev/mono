---
title: Built-in rules
description: The policy rules that ship in the Confire binary.
---

Confire ships a set of built-in rules embedded in the binary. They
cover common high-risk operations and activate in `balanced` mode
by default.

Use `confire policy test <command>` to simulate what any rule would
do for a given input without running anything.

## Destructive git operations

**Rule ID**: `git-destructive` | **Action**: review | **Severity**: high

Fires on Bash commands that match destructive git patterns:

| Pattern | Reason |
|---|---|
| `git push ... --force` | Force push rewrites remote branch history and affects open PRs. |
| `git reset --hard` | Hard reset permanently discards local commits and staged changes. |
| `git clean -f` | Deletes untracked files; not recoverable. |
| `git rebase main` (or master) | Rebasing onto main rewrites commit history. |
| `git branch -D` | Hard-deletes a branch with no remote recovery. |

## GitHub CLI mutations

**Rule ID**: `github-cli-review` | **Action**: review | **Severity**: medium

Fires on `gh` commands that mutate shared state:

- `gh pr close` — closes a PR, discarding review comments
- `gh pr merge` — merges a PR, triggers CI/CD pipelines

## Secret file reads

**Rule ID**: `secret-file-read` | **Action**: review | **Severity**: medium

Fires when Bash or Read accesses files commonly containing
credentials:

- `.env`, `.env.local`, `.env.prod`, `.env.staging`
- `id_rsa`, `id_ed25519`
- `.aws/credentials`
- `.kube/config`
- `.npmrc`, `.pypirc`

## Mutating MCP tool calls

**Rule ID**: `mcp-mutation` | **Action**: review | **Severity**: medium

Fires on MCP tool calls whose name contains a mutation verb:
`create`, `update`, `delete`, `remove`, `send`, `post`, `publish`,
`merge`, `approve`, `refund`, `charge`, `transfer`, `invite`,
`execute`, `run`, `write`, `deploy`, `close`, `archive`.

This rule catches broad MCP write operations across all connected
platforms — a Notion page creation, a Stripe charge, a GitHub PR
merge via MCP are all covered by the same pattern.

## Output security rules (PostToolUse)

Two built-in rules operate at PostToolUse and drive the Context
Firewall's security passes:

| Rule ID | Action | What it does |
|---|---|---|
| `secret-scan-output` | redact | Replaces credential-like values in tool output |
| `injection-scan-output` | sanitize | Removes instruction-like text from tool output |

These fire automatically and don't interact with policy modes.

## Testing rules

```bash
# Test a specific command
confire policy test 'git push origin main --force'

# Test an MCP tool name
confire policy test 'mcp__github__merge_pull_request'

# Test something safe
confire policy test 'git status'
```

See all active rules and counts:

```bash
confire policy status
```

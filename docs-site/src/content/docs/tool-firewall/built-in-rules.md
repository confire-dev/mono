---
title: Built-in rules
description: The policy rules that ship in the Confire binary.
---

Confire ships a set of built-in rules embedded in the binary. They
cover common high-risk operations and activate in `balanced` mode
by default.

Use `confire policy test <command>` to simulate what any rule would
do for a given input without running anything.

## Rate policy rules

Rate policy rules fire on call patterns across a session, not on
individual command content. They run before all other rules — if a
rate rule fires, no further rule evaluation happens for that call.

### Runaway loop detection

**Rule ID**: `rate.runaway_loop` | **Action**: block | **Severity**: high

Fires when the same tool call (same tool name + same input prefix)
is repeated **20 or more times within a 5-minute window**. This
catches agents stuck in a retry loop before they cause wide damage.

```
[Confire rate.runaway_loop] Bash has been called 20 times with the
same input in the past 5 minutes — this looks like a runaway retry loop.

To continue immediately: run `confire off` to disable the firewall
for this batch, then `confire on` when done.
To raise the threshold permanently: add "runaway_loop_threshold": 50
to ~/.confire/config.json, then run `confire stop && confire start`.
```

The default threshold is 20 identical calls per 5-minute window.
Raise it with `runaway_loop_threshold` in `~/.confire/config.json`
if legitimate batch work triggers false positives.

### Call rate cap

**Rule ID**: `rate.call_cap` | **Action**: block | **Severity**: high

Fires when more than **100 tool calls of any kind** occur within a
1-minute window. This prevents a misconfigured or runaway agent from
flooding tool calls at machine speed.

```
[Confire rate.call_cap] 101 tool calls in the past 1 minute exceeds
the rate limit (>100). Session blocked to prevent runaway agent behavior.

To continue immediately: run `confire off` to disable the firewall
for this batch, then `confire on` when done.
To raise the threshold permanently: add "call_rate_threshold": 200
to ~/.confire/config.json, then run `confire stop && confire start`.
```

The default threshold is 100 calls per minute. Raise it with
`call_rate_threshold` in `~/.confire/config.json` for high-volume
automation workflows.

Both rate rules are programmatic — they don't appear in
`confire policy status` or `confire policy test` output in v1.

## Destructive git operations

**Rule ID**: `git-destructive` | **Action**: review | **Severity**: high | **Irreversible**

Fires on Bash commands that match destructive git patterns:

| Pattern | Reason |
|---|---|
| `git push ... --force` | Force push rewrites remote branch history and affects open PRs. |
| `git reset --hard` | Hard reset permanently discards local commits and staged changes. |
| `git clean -f` | Deletes untracked files; not recoverable. |
| `git rebase main` (or master) | Rebasing onto main rewrites commit history. |
| `git branch -D` | Hard-deletes a branch with no remote recovery. |

This rule is marked **irreversible** — ask-budget skips never apply
to it. It always surfaces as a review regardless of session budget.

## GitHub CLI mutations

**Rule ID**: `github-cli-review` | **Action**: review | **Severity**: medium

Fires on `gh` commands that mutate shared state:

- `gh pr close` — closes a PR, discarding review comments
- `gh pr merge` — merges a PR, triggers CI/CD pipelines

## GitHub CLI block

**Rule ID**: `github-cli-block` | **Action**: block | **Severity**: high | **Irreversible**

Fires on `gh repo delete`. Repository deletion is permanent — no
retry path exists from the agent.

## Secret file reads

**Rule ID**: `secret-file-read` | **Action**: review | **Severity**: medium

Fires when Bash or Read accesses files commonly containing
credentials:

- `.env`, `.env.local`, `.env.prod`, `.env.staging`
- `id_rsa`, `id_ed25519`
- `.aws/credentials`
- `.kube/config`
- `.npmrc`, `.pypirc`

## Destructive filesystem operations

**Rule ID**: `filesystem-destructive` | **Action**: review | **Severity**: high | **Irreversible**

Fires on `rm -rf`, `rm -fr`, and `chmod -R 777`. These rules are
marked **irreversible** — ask-budget skips never apply.

## Destructive database operations

**Rule ID**: `database-destructive` | **Action**: review | **Severity**: high | **Irreversible**

Fires on `DROP DATABASE`, `DROP TABLE`, `TRUNCATE TABLE`, unguarded
`DELETE FROM`, `supabase db reset`, `supabase db push`, and
`prisma migrate reset`. Marked **irreversible**.

## Production deployments

**Rule ID**: `deploy-prod` | **Action**: review | **Severity**: high | **Irreversible**

Fires on `vercel --prod`, `netlify deploy --prod`, `fly deploy`, and
`railway up`. Marked **irreversible** — deploying to production
always surfaces as a review.

## Mutating MCP tool calls

**Rule ID**: `mcp-mutation` | **Action**: review | **Severity**: medium

Fires on MCP tool calls whose name contains a mutation verb:
`create`, `update`, `delete`, `remove`, `send`, `post`, `publish`,
`merge`, `approve`, `refund`, `charge`, `transfer`, `invite`,
`execute`, `run`, `write`, `deploy`, `close`, `archive`.

This rule catches broad MCP write operations across all connected
platforms — a Notion page creation, a Stripe charge, a GitHub PR
merge via MCP are all covered by the same pattern.

## The irreversible flag

Rules marked **Irreversible** bypass the ask-budget mechanism
entirely. In a normal session, warn-action calls can be
auto-acknowledged up to the session budget (default 5) and passed
through silently. Irreversible rules are exempt — they always
surface as a review, regardless of how much budget remains.

The five irreversible rules: `git-destructive`, `github-cli-block`,
`filesystem-destructive`, `database-destructive`, `deploy-prod`.

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

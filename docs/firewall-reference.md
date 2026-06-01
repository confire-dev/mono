# Confire Firewall — Complete Behavior Reference

What Confire does in each hook phase: when it triggers, what action it takes, and how the outcome is surfaced to Claude and the user.

---

## Overview

Confire runs a **two-pillar firewall** across every Claude Code tool call:

| Pillar | Phase | Does what |
|--------|-------|-----------|
| **Tool Firewall** | PreToolUse | Evaluates the tool call _before_ execution — allows, warns, reviews, or blocks |
| **Context Firewall** | PostToolUse | Sanitizes and optimizes raw tool output _before_ it enters Claude's working context |

Both pillars are controlled by a single policy mode (`balanced` by default). If anything fails or times out, Confire passes through silently — it never breaks Claude Code.

---

## Hook phases

### SessionStart

**Trigger:** Claude Code begins a new session (once per window/process launch).

**What Confire does:**
1. Records session metadata locally: `session_id`, `cwd`, `started_at`, `mode`.
2. Loads the merged rule set: built-in rules (always) + cached custom rules from `~/.confire/policies/cache.json` if present.
3. If an API key is present, spawns a background goroutine to fetch fresh custom rules from the backend — on success, updates the cache. On network failure, uses cached rules silently.
4. Writes status to stderr. Injects `additionalContext` into Claude only when something requires Claude's awareness (not logged in, strict/bypass mode active, policy sync failure). Silent on a normal healthy start.

**How you notice it:**

_In terminal (stderr), always:_
```
[confire] v0.x.y active — balanced mode, 22 rules loaded
```

_In Claude's context, only when relevant:_
```
⚠️ Confire: not logged in — cloud optimization disabled. Run `confire login`.
```
or, if strict mode is active:
```
[Confire] Strict mode active. Dangerous tool calls will be blocked, not just reviewed.
```

A healthy session with balanced mode and an account injects nothing into Claude's context. Saving context is the product; Confire should not spend context announcing itself.

**Failure behavior:** If the daemon is not running or the hook times out, Claude starts normally. No block, no error shown to user.

---

### PreToolUse — Tool Firewall

**Trigger:** Claude Code is about to execute any tool call (Bash, Read, Edit, Write, MCP tools, etc.).

**What Confire does:**
1. Checks for the one-shot bypass flag at `~/.confire/.bypass-next`. If present, clears the flag and passes through immediately.
2. Checks `config.IsFirewallEnabled()`. If mode is `bypass` or firewall is disabled, passes through.
3. Evaluates all enabled policy rules (built-in + cached custom) against the tool name and input.
4. Picks the highest-priority matching rule action (`block > review > warn > allow`).
5. In `observe` mode, downgrades `block` and `review` to `warn` — never blocks in observe.
6. Returns a decision to the hook process.

**Hook process behavior after daemon responds:**
- `ResultPassthrough` → exit 0, no stdout — tool proceeds normally.
- `ResultWarn` → exit 0, writes `additionalContext` JSON to stdout — tool proceeds but Claude sees the advisory.
- `ResultReview` or `ResultBlock` → writes block JSON to stdout, exits with code 2 — Claude Code cancels tool execution and surfaces the message to Claude.

**Decision outcomes:**

#### allow (passthrough)
No output. Tool runs immediately. Applies to safe commands: `git status`, `git diff`, `gh pr view`, `ls`, `cat`, read-only MCP tools (`get_*`, `list_*`, `search_*`).

#### warn
Tool runs. Claude receives one advisory line as context — visible in the conversation but not blocking:
```
[Confire warning] Review force push: Bash — Force push can rewrite remote branch history and affect open PRs.
```
Used in `observe` mode for all matches, and when `action: warn` is set explicitly on a rule.

#### review

Tool is **blocked**. Claude Code surfaces the reason to Claude and the user. Because the message is returned as the block reason, **Claude reads it** and is expected to ask the user for a yes or no before proceeding.

The review message is written to instruct Claude to ask the user for explicit confirmation:

```
CONFIRE REVIEW REQUIRED

Rule:              Review force push
Claude is about to run: Bash — git push --force-with-lease origin main
Risk:              high severity
Why this matters:  Force push can rewrite remote branch history and affect open PRs.

ACTION REQUIRED — ask the user:
"Confire flagged this command. Do you want me to run it anyway?
If yes, run: confire bypass-next — then tell me to retry."
```

**The correct flow for a review:**
1. Confire blocks the tool and Claude sees the message above.
2. Claude asks the user: "Do you want me to run this anyway?"
3. If yes: user runs `confire bypass-next` in a terminal, then tells Claude to retry.
4. On retry: bypass-next flag is consumed, tool runs once without review.
5. Flag is deleted automatically — the next identical command gets reviewed again.

This requires user deliberate action. Claude cannot bypass Confire on its own.

#### block
Tool is **blocked permanently** for this invocation. No retry path. Shown to Claude and user:

```
CONFIRE BLOCKED TOOL CALL

Rule:    Block repo deletion
Blocked: Bash — gh repo delete myorg/myrepo
Reason:  Repository deletion is irreversible. Use the GitHub web UI with explicit confirmation.
```

Unlike review, block does not invite retry. The user must take deliberate action outside Claude Code. Claude should communicate this clearly and not attempt workarounds.

**Built-in rules (balanced mode):**

| Trigger | Action | Reason |
|---------|--------|--------|
| `git push --force` / `--force-with-lease` | review | Rewrites remote history, affects open PRs |
| `git reset --hard` | review | Discards local commits permanently |
| `git clean -f[d]` | review | Deletes untracked files, unrecoverable |
| `git rebase main/master/origin/main` | review | Rewrites commit history |
| `git branch -D <branch>` | review | Unrecoverable if no remote |
| `gh pr close` | review | Discards review comments |
| `gh pr merge` | review | Modifies shared branch, triggers CI |
| `gh repo delete` | **block** | Irreversible |
| `rm -rf` / `rm -fr` | review | Permanent file deletion |
| `chmod -R 777` | review | World-writable, security risk |
| `DROP DATABASE/TABLE`, `TRUNCATE TABLE` | review | Destructive database operation |
| `DELETE FROM <table>;` (no WHERE) | review | Deletes all rows |
| `supabase db reset` | review | Wipes local database |
| `supabase db push` | review | Applies migrations to remote |
| `prisma migrate reset` | review | Drops and recreates database |
| `npm/pnpm/yarn/bun publish` | review | Public, irreversible |
| `vercel --prod`, `netlify deploy --prod` | review | Affects live users |
| `fly deploy`, `railway up` | review | Affects live users |
| `docker push` | review | May overwrite a registry tag |
| `kubectl apply/delete` | review | Modifies live Kubernetes workloads |
| `terraform apply/destroy` | review | Modifies real cloud infrastructure |
| Reading `.env`, `id_rsa`, `.aws/credentials`, `.kube/config` | review | May expose secrets |
| `mcp__*__create_*` / `delete_*` / `send_*` / `merge_*` / `deploy_*` / `approve_*` / `refund_*` / `charge_*` / `run_*` / `execute_*` / `write_*` / `close_*` / `archive_*` | review | Mutates external state |

**MCP tools that are allowed without review** (read-only name patterns):
`get_*`, `list_*`, `search_*`, `fetch_*`, `read_*`, `describe_*`, `inspect_*`, `query_*`, `retrieve_*`

**How you notice it:**
- **In the conversation:** Claude Code displays the block/review reason as a message Claude reads. Claude then asks the user for confirmation.
- **`confire policy test '<command>'`:** Dry-run any command before running it.

---

### PostToolUse — Context Firewall

**Trigger:** A tool call has completed and its output is about to enter Claude's working context.

**What Confire does — in order:**

#### Step 1: Secret redaction (local, always first)
Scans the raw output for common secret patterns. Replaces matches with `[REDACTED_SECRET:<type>]` before the output goes anywhere else — including the cloud optimizer.

Detected types:

| Type | Pattern example | Placeholder |
|------|-----------------|-------------|
| `aws_access_key` | `AKIAIOSFODNN7EXAMPLE` | `[REDACTED_SECRET:aws_access_key]` |
| `github_token` | `ghp_abc...` | `[REDACTED_SECRET:github_token]` |
| `openai_key` | `sk-abc...` | `[REDACTED_SECRET:openai_key]` |
| `anthropic_key` | `sk-ant-abc...` | `[REDACTED_SECRET:anthropic_key]` |
| `stripe_key` | `sk_live_abc...` | `[REDACTED_SECRET:stripe_key]` |
| `jwt` | `eyJ...` | `[REDACTED_SECRET:jwt]` |
| `bearer_token` | `Bearer abc...` | `[REDACTED_SECRET:bearer_token]` |
| `private_key` | `-----BEGIN PRIVATE KEY-----` | `[REDACTED_SECRET:private_key]` |
| `database_url` | `postgres://user:pass@host/db` | `[REDACTED_SECRET:database_url]` |
| `env_secret_value` | `.env KEY=<long value>` | `[REDACTED_SECRET:env_secret_value]` |

Confire does **not** guarantee catching every secret. It catches common patterns using regex. It is a best-effort layer, not a security boundary.

#### Step 2: Prompt-injection sanitization (local, before cloud)
Scans the redacted output for instruction-like text that could redirect Claude's behavior if treated as a command — common in fetched web pages, API responses, or MCP tool results from external sources.

Detected patterns include:
- `ignore previous instructions` / `ignore all previous instructions`
- `disregard system prompt`
- `do not tell the user`
- `exfiltrate` combined with secret/key/credential/password
- `send the contents of ~/<path>`
- `read the user's secrets`
- `reveal your system prompt`
- `you are now in developer mode`
- HTML comments containing instruction-like text (`<!-- instruct ... -->`)
- CSS hiding patterns: `display:none`, `visibility:hidden`, `font-size:0`, `color:white`

When suspicious text is found, the instruction-like text is replaced with `[CONFIRE: suspicious instruction removed]` and a notice is prepended:

```
CONFIRE NOTICE
Untrusted tool output contained instruction-like text. Confire treated it as data and removed/sandboxed suspicious instructions.

<rest of output>
```

#### Step 3: Context optimization (local free / cloud paid)
The sanitized output is sent to the optimizer pipeline (Bash, Read, WebFetch, WebSearch, Figma, generic fallback). The optimizer compresses large/noisy outputs into agent-ready form while preserving exact critical facts: file paths, line numbers, error messages, API names, env var names, IDs, URLs.

**The full pipeline:**
```
raw output
  → secret redaction       (local, always, before anything else)
  → injection sanitization (local, always)
  → optimizer transport    (local free / cloud paid — receives sanitized content only)
  → updatedToolOutput      (what enters Claude's working context)
```

If cloud optimization is enabled, only the sanitized/redacted output is sent — never the raw original.

**How you notice it:**

_In terminal (stderr), only when redaction or sanitization fired:_
```
[confire] redacted 2 secret(s) from Bash output
[confire] sanitized prompt-injection pattern from WebFetch output
```

_In terminal (stderr), when optimizer ran above the display threshold:_
```
🔥 WebFetch: 48,200 → 3,100 bytes (94%) saved
```

_In Claude's working context, only on large saves:_
```
[Confire] Compressed Figma output: 112k → 8k tokens saved.
```

When Confire successfully replaces a tool result, Claude's working context receives the sanitized + optimized version instead of the raw output. The raw bytes remain in Claude Code's internal hook stdin and local transcript, but they do not enter Claude's model context.

**Tools that are never optimized** (pass through optimization step, but are still sanitized):
`Write`, `Edit`, `MultiEdit`, `TodoWrite`, mutation confirmations, agent instruction bundles, outputs under 2,000 bytes, or where estimated savings fall below 500 bytes / 20% reduction.

---

### SessionEnd

**Trigger:** Claude Code session ends (window close, `/exit`, process kill).

**What Confire does:**
1. Finalizes session stats: total calls, optimized calls, raw bytes, optimized bytes, firewall stats (blocked, reviewed, warned, sanitized calls, secrets redacted).
2. Prints a session summary to stderr if savings were meaningful (above display threshold).
3. Syncs aggregate metadata to the backend asynchronously — does not block shutdown. On network failure, queues locally for retry.
4. Returns `passthrough` immediately.

**What is never sent to the backend:** raw tool output, redacted secret values, prompt-injection content, full tool inputs/outputs.

**What is sent:** aggregate counts only — total calls, bytes before/after, tool type categories, session duration, secrets redacted count (number, not values), blocked/reviewed counts.

**How you notice it:**

_In terminal (stderr):_
```
🔥 Confire session summary
   Calls optimized: 12/18
   Context saved:   ~140,000 tokens
   Biggest win:     figma (92%)
   Firewall:        2 reviewed, 1 secret redacted
```

Nothing is printed if savings and firewall events were below the display threshold.

---

### PreCompact _(reserved, not active)_

**Trigger:** Would fire before Claude Code compacts the conversation transcript.

**Current behavior:** Disabled. All PreCompact events pass through immediately. No output, no side effects.

**Architecture:** The phase is mapped (`context.pre-compact`) and the daemon is wired to receive it, but no handler is registered and no action is taken. Enabling it requires an explicit opt-in config flag — it is not planned for launch.

---

## Modes

Set with `confire on`, `confire off`, or `mode` in `~/.confire/config.json`. Restart the daemon to apply.

| Mode | PreToolUse | PostToolUse |
|------|------------|-------------|
| `observe` | Rules match but block/review downgrades to warn — never blocks | Sanitization + optimization run |
| `balanced` _(default)_ | Review dangerous; block only the most destructive | Sanitization + optimization run |
| `strict` | Block dangerous; review more broadly | Stricter sanitization; tighter optimization budgets |
| `bypass` | All events pass through — no review, no block | No sanitization, no optimization |

`observe` is useful for trying Confire without risk. `bypass` is a full off-switch.

---

## The review UX flow

When a review fires, the intended user experience is:

```
1. Claude tries to run: git push --force-with-lease
2. Confire blocks it and returns the review message.
3. Claude reads the message and asks the user:
   "Confire flagged this as a force push. Do you want me to run it anyway?"
4. User decides:
   a. "No, use a safer command." → Claude finds an alternative.
   b. "Yes, proceed." → User runs `confire bypass-next` in terminal, tells Claude to retry.
5. Claude retries. Bypass flag is consumed. Tool runs once.
6. Next identical command gets reviewed again.
```

Claude cannot bypass Confire autonomously. The bypass-next flag requires a terminal action by the user. This is intentional — it keeps a human in the loop for flagged operations.

---

## One-shot overrides

### `confire bypass-next`
Writes a flag at `~/.confire/.bypass-next`. The next PreToolUse event passes through without review or block, regardless of what rule it matches. The flag is deleted automatically after one use.

Run this in your terminal after Claude presents a review message and you have decided to allow it.

### `confire off`
Sets `mode: bypass` in `~/.confire/config.json`. Applies after daemon restart (`confire stop && confire start`). Disables the entire firewall until `confire on` is run.

---

## CLI quick reference

```
confire on                      Enable firewall (balanced mode)
confire off                     Disable firewall (bypass mode)
confire bypass-next             Allow next tool call to skip review (one-shot)

confire policy test '<cmd>'     Dry-run: show what Confire would do
confire policy status           Show active mode, rule counts, cache age
confire policy pull             Fetch custom rules from dashboard (paid plan)
```

**`confire policy test` examples:**
```
$ confire policy test 'git push --force'
Action:   review
Rule:     Review force push
Severity: high
Reason:   Force push can rewrite remote branch history and affect open PRs.
Source:   builtin

$ confire policy test 'git status'
Action:   allow
Reason:   No matching rule — tool would proceed normally.

$ confire policy test 'mcp__stripe__create_refund'
Action:   review
Rule:     Review mutating MCP tool
Severity: medium
Reason:   This MCP tool appears to mutate external state.
Source:   builtin
```

---

## Rule tiers

| Tier | Where | Who edits | When available |
|------|-------|-----------|----------------|
| **Built-in** | CLI binary (`policy/builtin.go`) | Confire team via CLI releases | Always — no account, no network |
| **Custom** | Confire dashboard | You, via the dashboard UI | Paid plan — fetched and cached locally; all evaluation runs locally |

Custom rules use the same schema as built-in rules. They can override built-ins by ID or add new rules. All rule evaluation runs locally — tool inputs and outputs never leave the machine for policy decisions.

---

## Privacy model

| Data | What happens |
|------|-------------|
| Raw tool output | Never sent to Confire Cloud by default. Local redaction/sanitization runs first. If cloud optimization is enabled, only the sanitized/redacted content is sent — never the original. |
| Redacted secret values | Replaced before any further processing. Never stored locally or sent anywhere. |
| Prompt-injection content | Removed before optimization. Never stored or sent. |
| Tool inputs (PreToolUse) | Evaluated locally against rules. Never sent to the backend. |
| Session metadata | Stored locally (`~/.confire/sessions/`). Aggregate counts (not raw content) synced to backend if logged in. |
| Custom rule cache | Pulled from backend on sync, stored at `~/.confire/policies/cache.json`. Never sent back. |
| Secrets-redacted count | Sent as a number in session metadata. The secret values are never sent. |

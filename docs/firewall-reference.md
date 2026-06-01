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
2. Loads the merged rule set: built-in rules + any cached custom rules from `~/.confire/policies/cache.json`.
3. If an API key is present, spawns a background goroutine to fetch fresh custom rules from the backend — on success, updates the cache. On network failure, uses cached rules silently.
4. Returns a brief one-line notification that Claude sees as `additionalContext`.

**How you notice it:**

_In Claude's context (visible in conversation):_
```
✓ Confire v0.x.y active (cloud optimizer)
```
or, if not logged in:
```
⚠️ Confire: not logged in — run `confire login` to enable optimization.
```

_Nothing is written to stderr during SessionStart unless there is a hard error._

**Failure behavior:** If the daemon is not running or the hook times out, Claude starts normally. No block, no error shown to user.

---

### PreToolUse — Tool Firewall

**Trigger:** Claude Code is about to execute any tool call (Bash, Read, Edit, Write, MCP tools, etc.).

**What Confire does:**
1. Checks for the one-shot bypass flag at `~/.confire/.bypass-next`. If present, clears the flag and passes through immediately.
2. Checks `config.IsFirewallEnabled()`. If mode is `bypass` or firewall is disabled, passes through.
3. Evaluates all enabled policy rules (built-in + cached custom) against the tool name and input.
4. Picks the highest-priority matching rule action (`block > review > warn > allow`).
5. In `observe` mode, downgrades `block` and `review` to `warn` (never blocks).
6. Returns a decision to the hook process.

**Hook process behavior after daemon responds:**
- `ResultPassthrough` → exit 0, no stdout — tool proceeds normally.
- `ResultWarn` → exit 0, writes `additionalContext` JSON to stdout — tool proceeds but Claude sees the warning.
- `ResultReview` or `ResultBlock` → writes block JSON to stdout, exits with code 2 — Claude Code cancels tool execution.

**Decision outcomes:**

#### allow (passthrough)
No output. Tool runs immediately. Applies to safe commands: `git status`, `git diff`, `gh pr view`, `ls`, `cat`, read-only MCP tools (`get_*`, `list_*`, `search_*`).

#### warn
Tool runs. Claude receives one advisory line in context:
```
[Confire warning] Review force pushes: Bash — Force push can rewrite remote branch history and affect open PRs.
```
Used in `observe` mode for all matches, and when `action: warn` is explicitly set.

#### review
Tool is blocked. Claude Code shows the reason to Claude and to the user in the terminal:

```
CONFIRE REVIEW REQUIRED

Rule:              Review force push
Claude is about to run: Bash — git push --force-with-lease origin main
Risk:              high severity
Why this matters:  Force push can rewrite remote branch history and affect open PRs.

How to proceed:
- If this is intentional, tell Claude to proceed and it will retry.
- Or use a safer alternative command/tool.
```

Claude can be told to allow and retry, at which point the user can run `confire bypass-next` to clear one invocation.

#### block
Tool is blocked permanently (for this invocation). Shown to Claude and user:

```
CONFIRE BLOCKED TOOL CALL

Rule:    Block repo deletion
Blocked: Bash — gh repo delete myorg/myrepo
Reason:  Repository deletion is irreversible. Use the GitHub web UI with explicit confirmation.
```

Unlike review, block does not invite retry. The user must take deliberate action outside Claude Code.

**Built-in rules that trigger (balanced mode):**

| Trigger | Action | Reason shown |
|---------|--------|--------------|
| `git push --force` / `--force-with-lease` | review | Rewrites remote history, affects open PRs |
| `git reset --hard` | review | Discards local commits permanently |
| `git clean -f[d]` | review | Deletes untracked files, unrecoverable |
| `git rebase main/master/origin/main` | review | Rewrites commit history |
| `git branch -D <branch>` | review | Unrecoverable if no remote |
| `gh pr close` | review | Discards all review comments |
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
| `docker push` | review | Overwrites a registry tag |
| `kubectl apply/delete` | review | Modifies live Kubernetes workloads |
| `terraform apply/destroy` | review | Modifies real cloud infrastructure |
| Reading `.env`, `id_rsa`, `.aws/credentials`, `.kube/config` | review | May expose secrets |
| Any `mcp__*__create_*` / `delete_*` / `send_*` / `merge_*` / `deploy_*` / `approve_*` / `refund_*` / `charge_*` / etc. | review | Mutates external state |

**MCP tools that are allowed without review** (read-only name pattern):
`get_*`, `list_*`, `search_*`, `fetch_*`, `read_*`, `describe_*`, `inspect_*`, `query_*`, `retrieve_*`

**How you notice it (user-facing):**
- **Terminal:** Claude Code displays the block/review reason inline in the conversation.
- **stderr:** None (daemon logs to its own stderr, not the hook process's).
- **`confire policy test '<command>'`:** Dry-run any command to preview the decision before running it.

---

### PostToolUse — Context Firewall

**Trigger:** A tool call has completed and its raw output is about to enter Claude's working context.

**What Confire does — in order:**

#### Step 1: Secret redaction
Scans the raw output string for common secret patterns. Replaces matches with `[REDACTED_SECRET:<type>]` before the output goes anywhere else.

Detected types:

| Secret type | Pattern example | Placeholder |
|-------------|-----------------|-------------|
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

Confire does **not** guarantee catching every secret. It catches common patterns. It is a best-effort layer, not a security boundary.

#### Step 2: Prompt-injection sanitization
Scans the (now-redacted) output for instruction-like text that could redirect Claude's behavior if treated as a command.

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

When suspicious text is found, it is replaced with `[CONFIRE: suspicious instruction removed]` and the output is prefixed with:

```
CONFIRE NOTICE
Untrusted tool output contained instruction-like text. Confire treated it as data and removed/sandboxed suspicious instructions.

<rest of output>
```

#### Step 3: Context optimization
The sanitized output is sent to the optimizer pipeline (Bash, Read, WebFetch, WebSearch, Figma, generic fallback). The optimizer compresses large/noisy outputs into agent-ready form while preserving exact critical facts (file paths, line numbers, error messages, API names, env var names, IDs).

The full pipeline:
```
raw output
  → secret redaction      (local, always)
  → injection sanitization (local, always)
  → optimizer transport   (local free / cloud paid)
  → updatedToolOutput     (what Claude sees)
```

**How you notice it:**

_In terminal (stderr):_ Only if redaction or sanitization fired:
```
[confire] redacted 2 secret(s) from Bash output
[confire] sanitized prompt-injection pattern from WebFetch output
```

_In terminal (stderr):_ Optimization savings (when above threshold):
```
🔥 WebFetch: 48,200 → 3,100 bytes (94%) saved
```

_In Claude's context:_ On large saves, one line added as `additionalContext`:
```
[Confire] Compressed figma output: 112k → 8k tokens saved.
```

_Claude never sees the raw output_ when Confire replaces it. It only sees the sanitized + optimized version.

**Tools that are NOT optimized** (always pass through optimization, though still sanitized):
`Write`, `Edit`, `MultiEdit`, `TodoWrite`, mutation confirmations, agent instruction bundles, outputs under 2,000 bytes, outputs where estimated savings are below 500 bytes or 20%.

---

### SessionEnd

**Trigger:** Claude Code session ends (window close, `/exit`, process kill).

**What Confire does:**
1. Finalizes session stats: total calls, optimized calls, raw bytes, optimized bytes, firewall stats (blocked, reviewed, warned, sanitized, secrets redacted).
2. Prints a session summary to stderr if savings were meaningful.
3. Syncs metadata to the backend asynchronously. If the network is unavailable, queues the event locally for retry.
4. Returns `passthrough` immediately so Claude Code shutdown is not delayed.

**Raw content is never sent to the backend.** Only metadata and aggregate counts.

**How you notice it:**

_In terminal (stderr):_
```
🔥 Confire session summary
   Calls optimized: 12/18
   Context saved:   ~140,000 tokens
   Biggest win:     figma (92%)
   Firewall:        2 reviewed, 1 secret redacted
```

If savings were below the display threshold, nothing is printed.

---

### PreCompact _(reserved, not active)_

**Trigger:** Would fire before Claude Code compacts the conversation transcript.

**Current behavior:** Disabled. All PreCompact events pass through immediately. No output, no side effects.

**Architecture:** The phase is mapped (`context.pre-compact`) and the engine is wired to receive it, but no handler is registered. Enabling it requires an explicit config flag in a future release.

---

## Modes

The mode controls how aggressively the firewall enforces rules. Set with `confire on`, `confire off`, or by editing `~/.confire/config.json`.

| Mode | PreToolUse behavior | PostToolUse behavior |
|------|---------------------|----------------------|
| `observe` | Rules fire but all block/review downgrades to warn — never blocks | Sanitization runs; optimization runs |
| `balanced` _(default)_ | Review dangerous actions; block only the most destructive | Sanitization + optimization runs |
| `strict` | Block dangerous actions; review more broadly | Stronger sanitization; tighter optimization budgets |
| `bypass` | All PreToolUse events pass through immediately | All PostToolUse events pass through (no sanitization, no optimization) |

---

## One-shot overrides

### `confire bypass-next`
Writes a flag file at `~/.confire/.bypass-next`. The next PreToolUse event — regardless of what rule it would match — passes through without review or block. The flag is deleted automatically after one use.

Use this when Confire has reviewed a command and you have confirmed it is intentional.

### `confire off`
Sets mode to `bypass` in `~/.confire/config.json`. Applies after daemon restart. Disables the firewall for all future sessions until `confire on` is run.

---

## CLI commands quick reference

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

| Tier | Where defined | Who edits | When available |
|------|--------------|-----------|----------------|
| **Built-in** | CLI binary (`policy/builtin.go`) | Confire team, via CLI releases | Always — no account, no network |
| **Custom** | Confire dashboard | You, via dashboard | Paid plan; fetched and cached locally; evaluated locally |

Custom rules use the same schema as built-in rules and can override built-ins by ID or add new ones. All evaluation runs locally — raw tool inputs and outputs never leave the machine.

---

## Privacy model

| Data | What happens |
|------|-------------|
| Raw tool output | Never sent anywhere. Sanitized locally, optimized locally or in the cloud worker. |
| Redacted secrets | Replaced before any further processing. Never stored or sent. |
| Prompt-injection content | Removed before optimization. Never stored or sent. |
| Session metadata | Stored locally (`~/.confire/sessions/`). Aggregate counts synced to backend if logged in. |
| Custom rule cache | Fetched from backend, stored locally (`~/.confire/policies/cache.json`). Never sent back. |

---
title: Troubleshooting
description: Common issues and how to fix them.
---

## Daemon won't start

**Symptom:** `confire status` shows `daemon: stopped` or `confire start` exits immediately.

**Fix:**
1. Check for a stale socket: `rm -f ~/.confire/confire.sock`
2. Check for a stale PID file: `rm -f ~/.confire/daemon.pid`
3. Restart: `confire start`
4. View daemon logs: `confire daemon --foreground` to see errors in real time.

---

## Hook not intercepting tool calls

**Symptom:** Tools fire normally but no optimization or filtering occurs.

**Fix:**
1. Run `confire status` — confirm `hook: installed` and `daemon: running`.
2. If the hook is missing, re-run `confire setup` for your client.
3. Confirm the hook path is registered in your client's settings. See the [Claude Code](/clients/claude-code), [Cursor](/clients/cursor), or [VS Code](/clients/vscode) setup guides.
4. Restart your editor after hook installation.

---

## Legitimate tool calls are being blocked

**Symptom:** A policy rule is blocking a tool call you want to allow.

**Fix:**
1. Run `confire stats` to see recent blocked calls and the matching rule.
2. Edit your policy file (default: `~/.confire/policy.yaml`) to add an `allow` rule above the blocking `deny` rule. Rules are evaluated top-to-bottom.
3. Run `confire policy validate` to check for syntax errors.
4. See the [Policy Rules](/configuration/policy-rules) reference for rule syntax.

---

## Secret is being redacted unexpectedly

**Symptom:** A value in a tool response is replaced with `[REDACTED:type]` but it isn't actually a secret.

**Fix:**
This can happen with high-entropy strings adjacent to words like `key`, `token`, or `secret`. You can disable redaction for specific patterns by adjusting your policy, or file a report at **hi@confire.dev** with a sanitized example so we can tune the heuristic.

---

## `confire login` fails or hangs

**Symptom:** The browser opens but the callback never completes, or `login` times out.

**Fix:**
1. Make sure nothing is blocking port `9999` on localhost (the OAuth callback server).
2. Try: `lsof -i :9999` to check for conflicts, then `confire login` again.
3. If your browser doesn't open automatically, copy the URL from the terminal and open it manually.
4. If the issue persists, email **hi@confire.dev** with the output of `confire login --debug`.

---

## `confire status` shows `auth: invalid key`

**Symptom:** The daemon reports an invalid or expired API key.

**Fix:**
1. Re-authenticate: `confire login`
2. If that doesn't help, reset credentials and log in again: `confire reset && confire login`
3. Check that your system keychain is accessible (some headless/CI environments don't have one).

---

## High token usage — optimizer doesn't seem to be working

**Symptom:** Token counts in your AI client haven't changed after installing Confire.

**Fix:**
1. Confirm `confire status` shows `optimizer: active`.
2. Check your plan — local-only optimizers cover Bash, Read, WebFetch, and Generic tools. Cloud optimizers (GitHub, Jira, Slack, etc.) require an authenticated session.
3. Run `confire stats` to see per-tool optimization stats. If a tool shows `0 bytes saved`, it may not have a dedicated optimizer yet.

---

## Rate limit errors

**Symptom:** You see `429 Too Many Requests` in daemon logs or tool calls are slowed down.

**Fix:**
Confire enforces per-minute and monthly limits by plan. See [Plans](/reference/plans) for your tier's limits.

- **Per-minute throttling** is temporary — wait a moment and calls will resume.
- **Monthly quota exhausted** — purchase additional credits from the dashboard or upgrade your plan.

Run `confire stats` to see your current usage.

---

## Getting more help

If none of the above resolves your issue, email **hi@confire.dev** with:

- Output of `confire version` and `confire status`
- Your OS and editor/client
- What you expected vs. what happened
- Any relevant logs from `confire daemon --foreground`

# Privacy

This document describes what data Confire collects, what stays on your machine, and how we handle the data that does leave it.

## What stays on your machine

The following never leaves your device:

- **Source code and file contents** — tool inputs and outputs from your editor stay local unless you are authenticated and explicitly using cloud optimizers.
- **Secrets and credentials** — AWS keys, GitHub tokens, API keys, private keys, and other high-entropy strings are detected and replaced with `[REDACTED:type]` by the local secret redaction step *before* any data is transmitted.
- **Conversation history and prompts** — session transcripts stored by Claude Code, Cursor, or VS Code are not read or transmitted by Confire.
- **Your OAuth credentials** — authentication with GitHub/Google is handled by Supabase Auth. Confire never sees your OAuth tokens directly.

## What is sent to the cloud

When you are authenticated and the cloud optimizer is enabled, Confire sends tool call payloads to our Cloudflare Worker (`api.confire.dev`) for optimization. A payload looks like this:

```json
{
  "host": "claude-code",
  "phase": "tool.post",
  "session": { "id": "...", "cwd": "...", "transcriptPath": "..." },
  "tool": { "name": "...", "input": {...}, "output": {...}, "durationMs": 123 }
}
```

The tool `output` field contains the raw MCP tool response (e.g. a GitHub API response, a Jira ticket, Figma metadata). This is the content being optimized to reduce token usage. It is processed in-memory and is not stored permanently.

If you do not authenticate or the remote optimizer is unavailable, Confire falls back to local-only optimization — no network call is made.

## What we store

| Data | Where | Retention | Purpose |
|------|-------|-----------|---------|
| Account email & auth state | Supabase | Until account deletion | Authentication |
| API key (hashed) | Supabase | Until revoked | Request authorization |
| Tool call summaries (tool type, raw bytes, optimized bytes, reduction %) | Supabase | Rolling 12 months | Usage stats, billing |
| Monthly usage counters | Supabase | Rolling 12 months | Quota enforcement |
| Credit ledger entries | Supabase | Permanent | Billing audit trail |
| Stripe subscription & payment events | Supabase + Stripe | Per Stripe policy | Billing |
| Activation events (login, hook install, first optimization) | Amplitude | 24 months | Product analytics |
| Performance metrics (latency, error rates) | Cloudflare Analytics Engine | 90 days | Infrastructure monitoring |

Tool call summaries contain **aggregate statistics only** (byte counts, reduction ratios, timestamps). They do not contain the content of your tool calls.

## Telemetry

Confire collects anonymized product telemetry via Amplitude. This includes events like `cli_session_started`, `hook_installed`, and `tool_call_optimized`. No personally identifiable content is attached to these events — they are used to understand feature adoption and improve the product.

There is currently no opt-out mechanism for telemetry, but one is planned.

## API keys

Your Confire API key is generated during `confire login` and stored in your **OS keychain** (macOS Keychain, Linux Secret Service, or Windows Credential Manager). It is never written to a config file in plaintext. You can revoke it at any time from the dashboard or by running `confire reset`.

## Data deletion

To delete your account and all associated data, email **hi@confire.dev** with the subject `Account deletion request`. We will process it within 30 days.

## Contact

For privacy questions or concerns: **hi@confire.dev**

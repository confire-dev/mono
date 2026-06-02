# Policy & Firewall System

Internal reference for Confire's rule-based firewall that runs on every MCP tool call. Covers architecture, rule authoring, group overrides, and the handler pipeline.

---

## Overview

The policy system is a two-phase firewall that wraps every tool call the CLI intercepts:

- **PreToolUse** — evaluate before the tool executes; can block, request review, or warn.
- **PostToolUse** — evaluate after the tool returns; can sanitize, redact, or normalize the output.

All rules fire only on MCP tools by default (`mcp_only: true`). Native Claude Code tools (Bash, Read, WebFetch, etc.) are never touched by MCP-specific rules.

---

## Architecture

```
Claude Code hook
       │
       ▼
  cli/cmd/daemon.go (handleConn)
       │
       ├─ PhaseToolPre ──► guardrail.Run(event)      ← policy Engine + MCPRiskHandler
       │
       └─ PhaseToolPost ─► runMCPPostPipeline(event)  ← MCPSanitizeHandler
                                                          MCPNormalizeHandler
                                                          MCPStatsWriter
```

### Packages

| Package | Role |
|---------|------|
| `cli/policy` | Rule types, engine, bundle loader, cache |
| `cli/intercept` | Handler interface + MCP handler implementations |
| `cli/guardrail` | Thin wrapper that calls `policy.Engine.EvaluatePreTool` and translates to `InterceptResult` |
| `worker/src/handlers/policy.ts` | API endpoints for reading/writing per-user group overrides |

---

## Rule Types

Rules are defined in JSON bundles embedded in the binary (`cli/policy/rules/`).

### Rule struct

```go
type Rule struct {
    ID       string     // unique, e.g. "mcp.secrets.output"
    Name     string     // human-readable
    Enabled  bool
    Phase    RulePhase  // "PreToolUse" | "PostToolUse" | "Any"
    Action   RuleAction // see Actions below
    Severity Severity   // "low" | "medium" | "high" | "critical"
    MinMode  Mode       // only fires in this mode or stricter (optional)
    Group    string     // e.g. "mcp.secrets" — used for dashboard toggles
    Match    RuleMatch
    Message  string     // shown to the user
    Source   string     // "builtin" | "cloud"
}
```

### RuleMatch fields

| Field | Type | Description |
|-------|------|-------------|
| `mcp_only` | bool | Only fire on MCP tools (all MCP rules set this) |
| `tool_name` | string | Exact tool name match |
| `tool_names` | []string | Any of these tool names |
| `tool_prefix` | string | Tool name starts with this |
| `tool_name_regex` | string | Regex matched against tool name (case-insensitive) |
| `command_contains` | []string | Bash command contains any of these strings |
| `command_regex` | string | Regex matched against Bash command |
| `input_param_scan` | bool | Only fire if input has credential-like param names |
| `output_min_bytes` | int | Only fire if output JSON size ≥ N bytes |
| `output_secret_scan` | bool | Delegate to secret scanner (PostToolUse) |
| `output_injection_scan` | bool | Delegate to injection scanner (PostToolUse) |

### Actions

| Action | Phase | Effect |
|--------|-------|--------|
| `block` | Pre | Prevent tool from running; show reason to user |
| `review` | Pre | Block + invite user to allow once |
| `warn` | Pre/Post | Allow but inject advisory context |
| `redact` | Post | Replace output with redacted version |
| `sanitize` | Post | Replace output with sanitized version |
| `optimize` | Post | Replace output with normalized/packed version |
| `allow` | Pre | Explicit allow (overrides lower-priority rules) |
| `passthrough` | Any | No-op |

Priority order (highest wins): `block > review > warn = redact = sanitize > optimize > allow = passthrough`.

### Modes

| Mode | Behaviour |
|------|-----------|
| `observe` | Collect signals only — `block` and `review` downgrade to `warn` |
| `balanced` | Default — all actions fire |
| `strict` | All actions fire; `MinMode: strict` rules also activate |
| `bypass` | Firewall disabled entirely for this session |

---

## Built-in Rule Bundles

### `cli/policy/rules/builtin.json`

Native tool rules (Bash, git, gh, etc.). Not covered here — see file directly.

### `cli/policy/rules/mcp.json`

All MCP-specific rules. 11 rules across 7 groups:

| Rule ID | Group | Phase | Action | Trigger |
|---------|-------|-------|--------|---------|
| `mcp.risk.shell_name` | `mcp.risk_classifier` | Pre | review | Tool name matches shell/exec patterns |
| `mcp.risk.write_name` | `mcp.risk_classifier` | Pre | review | Tool name matches write/delete/destroy |
| `mcp.risk.exfil_name` | `mcp.risk_classifier` | Pre | review | Tool name matches send/upload/webhook |
| `mcp.risk.credential_params` | `mcp.risk_classifier` | Pre | warn | Input params contain credential-like names |
| `mcp.cross_tool.output_ref` | `mcp.cross_tool` | Post | warn | Output contains cross-tool steering phrases |
| `mcp.secrets.output` | `mcp.secrets` | Post | redact | Output contains secret-like strings |
| `mcp.injection.output` | `mcp.injection` | Post | sanitize | Output contains instruction-like text |
| `mcp.unicode.output` | `mcp.unicode` | Post | sanitize | Always — strip hidden Unicode |
| `mcp.normalizer.output` | `mcp.normalizer` | Post | optimize | Always — prune nulls + truncate arrays |
| `mcp.budget.large_output` | `mcp.budget` | Post | optimize | Output ≥ 1 KB |
| `mcp.budget.output` | `mcp.budget` | Post | optimize | Output ≥ 4 KB |

---

## Handler Pipeline

### PreToolUse

1. **`guardrail.Handler`** — runs `policy.Engine.EvaluatePreTool`; fires builtin + cloud rules.
2. **`MCPRiskHandler`** — runs if guardrail returned passthrough; scores tool name + input params (0–100); returns warn (≥20) or review (≥50).

The first non-passthrough result wins.

### PostToolUse — MCP tools (`runMCPPostPipeline`)

Executed in order; each pass mutates `event.Tool.Output` in place:

1. **`MCPSanitizeHandler`**
   - Unicode strip: removes U+E0000–U+E007F (tag block), U+200B/C/D (zero-width), U+FEFF, U+00AD (soft hyphen), U+202A–U+202E (BiDi overrides), U+2066–U+2069, ALM/LRM/RLM, and all other `unicode.Cf` format chars.
   - Secrets redaction: 20 curated high-confidence patterns (AWS key/secret, GitHub tokens, Stripe, Slack bot/user/webhook, OpenAI, Anthropic, JWT, PEM private key header, Postgres/MySQL URIs, SendGrid, Twilio, npm, PyPI, Heroku, generic `secret=`/`api_key=` KV). Matches replaced with `[REDACTED:type]`.
   - Injection detection: 9 regex patterns covering "ignore previous instructions", "you are now", "act as", `<system>` tags, "system prompt", and cross-tool steering (`use the \w+ tool`, `call the \w+ function`, `invoke mcp__`, `mcp__\w+__\w+`).

2. **`MCPNormalizeHandler`**
   - Null/empty pruner: removes `null` values and empty `[]`/`{}` from the JSON tree.
   - Array truncation: arrays > 20 items → keep first 10 + last 3, insert `{"_confire_truncated": true, "_omitted_count": N, "_total": M}` marker.
   - String budget: strings > 8 KB → keep first 80% + `…[confire: truncated N bytes]…` + last 5%.

3. **`MCPStatsWriter`** (async, non-blocking)
   - Writes one row to `mcp_tool_events` and upserts `mcp_server_stats` in SQLite (`~/.confire/stats.db`).
   - Never delays a tool call — runs in a detached goroutine.

### PostToolUse — non-MCP (native) tools

Uses the existing `sanitizeOutput()` path in `daemon.go` backed by the `sanitize` package (not the intercept handlers above).

---

## Rule Group Overrides

Rules are grouped by the `Group` string field. Groups enable dashboard-level toggles — paid users can disable a group (e.g. turn off `mcp.unicode`) without editing individual rules.

### How it works

1. **Supabase** stores per-user overrides in `user_policy_overrides (user_id, group_id, enabled)`.
2. **Worker** serves `GET /v1/policy` → `{ group_overrides: { "mcp.unicode": false, … } }`.
3. **CLI** fetches on every `confire sync`, stores in `~/.confire/policy_cache.json` alongside cloud rules.
4. `policy.LoadRules()` calls `applyGroupOverrides()` which sets `Rule.Enabled = false` for any rule whose group is disabled.

### API endpoints

```
GET  /v1/policy
Authorization: Bearer <api_key>
→ { "group_overrides": { "mcp.secrets": true, "mcp.unicode": false } }

PATCH /v1/policy/groups
Authorization: Bearer <api_key>
Content-Type: application/json
{ "group_overrides": { "mcp.unicode": false } }
→ 200 OK  (paid plan required — firewallGroupToggles feature flag)
→ 403     (free plan)
```

### Feature gate

`firewallGroupToggles` is a boolean in the `Plan.features` object. It is stored in the Supabase `plans` table and served via Cloudflare KV — **not** hardcoded in TypeScript. Free plan: `false`. Dev/Pro/Enterprise: `true`.

Default behaviour when no overrides exist: **all groups enabled**.

---

## Adding a New Rule

### Option A — data only (recommended for new patterns)

Add an entry to `cli/policy/rules/mcp.json` or `builtin.json`:

```json
{
  "id": "mcp.risk.my_new_rule",
  "name": "My new check",
  "group": "mcp.risk_classifier",
  "enabled": true,
  "phase": "PreToolUse",
  "action": "warn",
  "severity": "medium",
  "match": {
    "mcp_only": true,
    "tool_name_regex": "my_pattern"
  },
  "message": "Shown to the user."
}
```

Run `go test ./policy/...` — the bundle loader validates all rules at startup (panics on invalid regex).

### Option B — new match field

1. Add the field to `RuleMatch` in `cli/policy/types.go`.
2. Add evaluation logic to `matchesPreTool` or `matchesPostTool` in `cli/policy/engine.go`.
3. Add a test case in `cli/policy/engine_test.go`.

### Option C — new handler (complex logic)

Implement the `intercept.Handler` interface:

```go
type Handler interface {
    ID() string
    Phases() []Phase
    Matches(e InterceptEvent) bool
    Run(e InterceptEvent) (InterceptResult, error)
}
```

Wire it into `daemon.go` (`runMCPPostPipeline` for PostToolUse, or the `PhaseToolPre` block for PreToolUse).

---

## Stats Schema

SQLite at `~/.confire/stats.db`. MCP tables are created by `intercept.MigrateMCPSchema(db)` on daemon start.

```sql
CREATE TABLE mcp_server_stats (
    server_id          TEXT PRIMARY KEY,
    total_calls        INTEGER DEFAULT 0,
    total_bytes_in     INTEGER DEFAULT 0,
    total_bytes_out    INTEGER DEFAULT 0,
    secrets_redacted   INTEGER DEFAULT 0,
    injections_flagged INTEGER DEFAULT 0,
    blocks             INTEGER DEFAULT 0,
    reviews            INTEGER DEFAULT 0,
    noisiness_score    REAL    DEFAULT 0,
    riskiness_score    REAL    DEFAULT 0,
    last_seen          INTEGER
);

CREATE TABLE mcp_tool_events (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    ts              INTEGER NOT NULL,
    server_id       TEXT    NOT NULL,
    tool_name       TEXT    NOT NULL,
    phase           TEXT    NOT NULL,
    outcome         TEXT    NOT NULL,   -- passthrough|warn|review|block|sanitize|redact
    risk_score      INTEGER,
    bytes_in        INTEGER,
    bytes_out       INTEGER,
    tokens_est_in   INTEGER,            -- bytes_in / 4
    tokens_est_out  INTEGER,            -- bytes_out / 4
    secrets_found   INTEGER DEFAULT 0,
    injection_flags INTEGER DEFAULT 0,
    unicode_hits    INTEGER DEFAULT 0,
    redacted        INTEGER DEFAULT 0,
    truncated       INTEGER DEFAULT 0
);
```

---

## File Index

| File | Purpose |
|------|---------|
| `cli/policy/types.go` | Rule, RuleMatch, RuleAction, Mode, Severity types |
| `cli/policy/engine.go` | EvaluatePreTool / EvaluatePostTool matching logic |
| `cli/policy/bundle.go` | Embedded JSON bundle loader + pattern expansion |
| `cli/policy/cache.go` | Local policy cache (builtin + cloud rules + group overrides) |
| `cli/policy/rules/builtin.json` | Built-in native tool rules |
| `cli/policy/rules/mcp.json` | Built-in MCP firewall rules |
| `cli/guardrail/guardrail.go` | Thin handler wrapping the policy engine |
| `cli/intercept/types.go` | InterceptEvent, Tool, InterceptResult, Phase constants |
| `cli/intercept/engine.go` | Handler dispatch loop |
| `cli/intercept/mcp_risk.go` | PreToolUse risk classifier handler |
| `cli/intercept/mcp_sanitize.go` | PostToolUse secrets + unicode + injection handler |
| `cli/intercept/mcp_normalize.go` | PostToolUse null pruner + array truncation handler |
| `cli/intercept/mcp_stats.go` | Async SQLite MCP stats writer |
| `cli/intercept/steer.go` | PostToolSteer context injector |
| `cli/cmd/daemon.go` | Main daemon — wires all handlers into the connection loop |
| `worker/src/handlers/policy.ts` | GET /v1/policy, PATCH /v1/policy/groups |
| `worker/src/lib/plans.ts` | Plan type with firewallGroupToggles feature flag |
| `docs/supabase-schema.sql` | user_policy_overrides table + RLS policies |

# Confire Product Roadmap

Internal document. Not published. Last updated: 2026-06-07.

---

## Product vision

Confire is a context firewall for AI coding agents.

Every tool call an agent makes — reading files, fetching URLs, running shell commands, calling MCP servers — passes through Confire before the output enters the model's context window. Confire decides: block it, redact it, flag it, or let it through. The agent never sees the original if Confire rewrote it.

The key insight: the context window is the attack surface. Prompt injection, secret exfiltration, and credential lures all flow through tool output. Confire is the last gate before that output becomes agent behavior.

---

## Plans

### Free — protect my local agent
- Input firewall: block / review / warn on dangerous tool calls
- Output sanitization: secret redaction + prompt injection removal
- Session summary on exit
- Works with Claude Code, Cursor, and other hook-compatible agents
- Local only — no sync, no dashboard

### Dev — customize my own firewall ($12/mo)
- Everything in Free
- Custom guardrail rules (YAML)
- Provenance labels: per-tool-call trust level and flags in local JSONL
- Cross-tool flow detection: 5 dangerous chain rules
- Cloud sync: security events → Supabase (dashboard at confire.dev)
- `confire review` — recent security events in terminal
- `confire stats` — trust distribution and session activity

### Team — org-wide agent security ($49/mo per seat)
Team is not "more credits." Team is organization-wide agent security.

- Everything in Dev
- Shared guardrail policies pushed to all team agents
- Agent/device inventory: see every agent running Confire in your org
- Central dashboard: security events across all agents, all sessions
- Per-agent permissions: different rule sets for different agents or CI roles
- Audit logs: full history of blocked/reviewed/bypassed actions
- Fleet controls: kill switch, force-upgrade, emergency lockdown
- SAML SSO + Slack/PagerDuty alerts

Teams need this because they have many users, cloud agents, CI agents, production tools, and shared secrets. The threat model changes when you have 20 engineers each running autonomous agents with access to production MCP servers.

---

## Current: v1.0 (shipped)

### What shipped
- **Phase 0** — Local context firewall
  - Unix socket daemon intercepts all tool calls (pre + post)
  - PreToolUse: block / review / warn on 60+ policy rules
  - PostToolUse: secret redaction + prompt injection removal
  - `confire bypass-next` single-use bypass token
  - Session summary on exit
  - ANSI color-coded terminal output (orange=review, red=block, yellow=warn, green=allowed)
  - Plain-text model-facing messages (no ANSI in JSON fields)

- **Phase 2** — Provenance labels
  - Per-tool-call `ProvenanceLabel` with trust level and flags
  - Trust levels: `trusted_local`, `external_untrusted`, `mcp_trusted`, `mcp_unknown`, `sensitive`, `agent_generated`
  - Local append-only JSONL at `~/.confire/sessions/<session_id>.jsonl`
  - Cloud sync via `provenance_event` telemetry to worker
  - Session summary: trust distribution line when external/unknown calls present
  - Single source of truth: `provenance.TrustedMCPServers` map (16 known servers)

- **Phase 3** — Cross-tool flow detection
  - 5 dangerous chain rules evaluated before each PreToolUse
  - Ring buffer of last 10 provenance labels per session
  - Rules: untrusted→secret-read, secret-read→external-send (block in strict), injection→shell, unknown-mcp→mutating-mcp, credential-lure→message
  - `firewall.CheckFlowRules()` runs before per-call guardrails

### What is NOT in v1.0
- Custom guardrail rules (v1.1)
- `confire review` terminal command (v1.1)
- `confire stats` trust distribution command (v1.1)
- Team fleet controls (v1.3)
- Dashboard UI (v1.1)

---

## Upcoming: v1.1 — Dev tier complete

Target: Q3 2026.

- Custom guardrail rules via `~/.confire/rules.yaml`
  - YAML format: id, name, phase, action, severity, match
  - Hot-reload: daemon watches for changes
  - Scoped rules: per-project `.confire/rules.yaml` overrides global
- `confire review` — last 20 security events, reverse chronological
- `confire stats` — full trust distribution table (internal / external trusted / external untrusted / mcp_unknown)
- Dashboard v1 at confire.dev:
  - Security events timeline
  - Trust distribution chart
  - Top blocked/reviewed rules
- `confire report <pattern>` — submit a new attack pattern for review

---

## Upcoming: v1.2 — Platform + agent identity

Target: Q4 2026.

- Agent identity: each agent gets a stable device ID
- Per-agent rule profiles: different guardrail configs for interactive vs CI agents
- `confire agents` — list all registered agents and their last activity
- Webhook notifications: Slack/email on block or critical review
- Rule bundles: one-click rule packs for common stacks (GitHub + Slack, AWS + Terraform)
- `confire update` — pull latest rule bundle from worker

---

## Upcoming: v1.3 — Team tier

Target: Q1 2027.

- Organization policies: shared rules synced to all team agents
- Team dashboard: cross-agent security events, trust distribution fleet-wide
- Per-agent permissions matrix: read-only vs read-write vs exec per agent role
- Audit logs: full tamper-evident event history (append-only, signed)
- Fleet controls:
  - Force-upgrade all agents to a minimum CLI version
  - Remote rule push without requiring agent restart
  - Emergency lockdown: block all tool calls organization-wide
- SAML SSO
- PagerDuty / OpsGenie integration for critical blocks

---

## Implementation phases (technical breakdown)

| Phase | Description | Status |
|---|---|---|
| Phase 0 | Local context firewall (block/review/warn/sanitize) | Done |
| Phase 1 | Custom guardrails + basic dashboard (v1.1) | Planned |
| Phase 2 | Provenance labels + local JSONL + cloud sync | Done |
| Phase 3 | Cross-tool flow detection (5 chain rules) | Done |
| Phase 4 | Worker update: provenance_event ingestion | Done |
| Phase 5 | Supabase schema (provenance_events table) | Done (migration 007) |
| Phase 6 | Docs + platform update | Planned |
| Phase 7 | Team fleet controls | Planned |

---

## Product principles

1. **The context window is the attack surface.** Every tool output that enters the model is an attack vector. Confire operates at this boundary.

2. **Trust is earned, not assumed.** Every tool call gets a provenance label. Unknown MCP servers are `mcp_unknown` until explicitly trusted. External content is `external_untrusted` by default.

3. **Never touch the raw content.** Only metadata (counts, flags, trust level, domain, tool name) leaves the user's machine. No raw tool output, no secret values, no `.env` content syncs to the cloud. Ever.

4. **Plain text for the model, color for the human.** ANSI codes go only to stderr for the operator. The `Reason` and `Context` fields sent to the agent are always plain text. The agent explains its reasoning; the human approves.

5. **One bypass token, consumed on use.** `confire bypass-next` generates a single-use token. The next tool call consumes it and is logged as bypassed. No persistent disable, no `--force` escape hatch.

6. **Free for individuals, paid for teams.** Free tier is feature-complete for solo developers. Team tier is not "more credits" — it is organizational security infrastructure.

7. **Upgrade paths are additive.** Each plan tier adds capabilities; nothing is removed. A Team user still gets all Dev features.

---

## Open questions

- Should `confire review` show flow rule triggers separately from per-call blocks? (leaning yes — flow context is important)
- For the Team dashboard, should cross-agent aggregation be opt-in per agent? (yes — some orgs will have compliance requirements)
- Rule bundle versioning: semver or hash-pinned? (semver for human readability, hash for integrity)
- Should bypass events block the session summary line from showing "clean session"? (yes — bypasses are security activity)

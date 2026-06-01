# Content Strategy — Confire

**Skill:** content-strategy  
**Primary goal:** Organic search → free signups  
**Secondary:** Authority in AI agent tooling + security

See also: existing pillar draft in `docs/copy/content-strategy.md` — this doc updates positioning for security-on-free and v1 hosts.

---

## Business context

| Field | Value |
|-------|-------|
| Product | CLI + cloud optimizer + local firewall for AI coding agents |
| ICP | Daily Claude Code / Cursor / VS Code users feeling cost or safety pain |
| Plans | Free (500 req + built-in security), Developer $10/mo |
| Tagline | Less noise in context = cheaper, faster, sharper AI agents |
| Site | confire.dev |

---

## Content pillars (5)

### Pillar 1: AI Agent Token Cost & Context Efficiency
**Why:** Direct problem Confire solves. High search intent.

Subtopics:
- Context window limits in agentic workflows
- Tool output verbosity (Bash, GitHub, MCP)
- Measuring savings with `confire stats`
- Free tier limits and when to upgrade

### Pillar 2: Claude Code, Cursor & VS Code Optimization
**Why:** Primary distribution channels for v1.

Subtopics:
- Hook setup (`confire setup`)
- PostToolUse vs PreToolUse
- Cross-agent comparison (cost, hooks, security)
- VS Code Copilot hooks parity

### Pillar 3: AI Agent Security & Guardrails
**Why:** Differentiator — security included free; competitors often ignore.

Subtopics:
- Pre-tool firewall patterns (block/warn/review)
- Dangerous Bash patterns agents run
- Secret redaction (best-effort)
- Custom rules on Developer tier

### Pillar 4: Bash & MCP Output Compression
**Why:** Technical, low-competition long-tail.

Subtopics:
- Bash output cleaning for LLMs
- GitHub API response compression
- Jira/Figma/Slack MCP noise
- Local vs cloud optimization

### Pillar 5: Business Case for Agent Efficiency
**Why:** Shareable; founder/engineering-lead audience.

Subtopics:
- Monthly Claude Code cost breakdown
- ROI of context optimization
- Team rollout (roadmap tease)

---

## Searchable vs shareable

**Priority: searchable first** (developers search pain before solutions).

| Type | Formats | Channels |
|------|---------|----------|
| Searchable | How-tos, comparisons, hook guides | SEO, docs |
| Shareable | Benchmark posts, cost transparency | HN, Reddit, X |

---

## Priority topics (Tier 1 — build first)

| # | Title | Type | Keyword | Stage | Score |
|---|-------|------|---------|-------|-------|
| 1 | Why your Claude Code sessions cost so much | Both | claude code cost | Awareness | 9.2 |
| 2 | How Claude Code hooks work (with Confire) | Searchable | claude code hooks | Implementation | 9.1 |
| 3 | AI agent security: pre-tool guardrails | Searchable | ai coding agent security | Consideration | 9.0 |
| 4 | How to reduce token usage in Claude Code | Searchable | reduce claude code tokens | Consideration | 8.9 |
| 5 | Confire compression benchmarks (our numbers) | Shareable | — | Awareness | 8.7 |
| 6 | Claude Code vs Cursor: context and cost | Searchable | claude code vs cursor | Consideration | 8.5 |
| 7 | What fills your context window in agents | Searchable | context window agents | Awareness | 8.4 |
| 8 | VS Code + Copilot hooks with Confire | Searchable | vscode copilot hooks | Implementation | 8.2 |

### Tier 2 (weeks 9–16)

- Best AI agent observability tools 2026
- GitHub API compression for LLM agents
- MCP tool output optimization guide
- Real cost of Claude Code for a month (transparent numbers)
- Custom firewall rules on Confire Developer

---

## Topic cluster map

```
PILLAR 1: Token Cost
├── Hub: Why Claude Code sessions cost so much
│   ├── How to measure token usage
│   ├── Tool output as the hidden cost
│   └── Free vs Developer limits

PILLAR 2: Agent Optimization
├── Hub: Claude Code hooks explained
│   ├── PostToolUse setup
│   ├── Cursor hook config
│   └── VS Code Copilot integration

PILLAR 3: Security
├── Hub: AI agent security firewall
│   ├── PreToolUse block/warn/review
│   ├── Built-in rules (free)
│   └── Custom rules (Developer)

PILLAR 4: Compression
├── Hub: Bash output for LLMs
│   ├── GitHub API noise
│   └── MCP optimizer deep dives

PILLAR 5: Business case
├── Hub: ROI of context optimization
    └── Team rollout (waitlist)
```

---

## Editorial calendar (first 90 days)

| Week | Piece | Pillar |
|------|-------|--------|
| 1 | Claude Code hooks guide | 2 |
| 2 | Why sessions cost so much | 1 |
| 3 | AI agent security firewall | 3 |
| 4 | Reduce token usage how-to | 1 |
| 5 | Claude Code vs Cursor cost | 2 |
| 6 | Context window explainer | 1 |
| 7 | Compression benchmark post | 1 + shareable |
| 8 | VS Code setup guide | 2 |
| 9 | GitHub API compression | 4 |
| 10 | MCP optimization overview | 4 |
| 11 | Monthly cost transparency post | 5 |
| 12 | Token benchmarks across agents | 5 |

---

## Content ideation sources

**Forum research:**
- `site:reddit.com "claude code" expensive OR tokens`
- `site:news.ycombinator.com claude code context`
- Claude / Cursor Discord #help channels

**Sales/support signals (when live):**
- Where setup fails (agent detection, daemon not running)
- Surprise at free security vs paid custom rules
- Which optimizer unlock drives Developer upgrades

---

## Resource requirements

| Asset | Owner | Notes |
|-------|-------|-------|
| Benchmark reproducibility | Eng | Script + sample outputs |
| Screenshots | Eng/design | `confire stats`, hook settings |
| Docs sync | Eng | Firewall reference → public docs |
| Blog CMS | Marketing | Astro blog on platform/ |

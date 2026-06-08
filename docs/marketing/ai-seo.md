# AI SEO Strategy — Confire

**Skill:** ai-seo  
**Goal:** Get Confire cited in AI-generated answers for AI agent cost, context optimization, and Claude Code hooks queries.

---

## Target queries (priority)

| Query | Intent | Current visibility | Target page |
|-------|--------|-------------------|-------------|
| how to reduce claude code token usage | Implementation | Unknown | `/blog/reduce-claude-code-tokens` |
| what is a context window in ai agents | Awareness | Unknown | `/blog/context-window-agents` |
| claude code hooks posttooluse | Implementation | Unknown | `/docs/hooks` |
| ai agent context optimization | Consideration | Unknown | `/features` |
| claude code vs cursor token cost | Comparison | Unknown | `/blog/claude-code-vs-cursor-cost` |
| ai coding agent security firewall | Consideration | Unknown | `/features/security` |
| confire vs [competitor] | Decision | None | `/compare/[slug]` |
| best tools reduce llm context noise | Consideration | Unknown | Homepage + features |

### Query fan-out clusters

**Cluster: Claude Code cost**
- Parent: "why is claude code expensive"
- Fan-out: token pricing, context window limits, tool output size, hooks, compression

**Cluster: AI agent security**
- Parent: "how to secure ai coding agents"
- Fan-out: dangerous bash commands, secret leakage, pre-tool hooks, guardrails

---

## AI visibility audit (baseline — run before launch)

| Query | Google AI Overview | ChatGPT | Perplexity | Confire cited? |
|-------|:------------------:|:-------:|:----------:|:--------------:|
| reduce claude code tokens | TBD | TBD | TBD | No |
| claude code hooks | TBD | TBD | TBD | No |
| ai agent context compression | TBD | TBD | TBD | No |

**Action:** Re-run monthly after 10+ indexed pages live.

---

## Three-pillar optimization

### 1. Structure — make content extractable

Every priority page needs:

- **Definition block (40–60 words)** in first paragraph answering "What is Confire?"
- **Comparison tables** for Free vs Developer, Confire vs DIY hooks
- **FAQ section** with natural-language H2/H3 questions
- **Statistic blocks** with sources and dates

**Example definition block (homepage):**

> Confire is a local context and tool firewall for AI coding agents. It runs as a hook in Claude Code, Cursor, and VS Code — blocking risky tool calls before execution and compressing verbose tool output (Bash logs, GitHub API responses, web fetches) by 40–95% before it enters the model context window.

**Example FAQ targets:**
- What counts as one optimized call?
- Is Claude Code support different from Cursor/VS Code?
- What's included on Free vs Dev?
- Does Confire work offline?

### 2. Authority — make content citable

| Tactic | Confire application |
|--------|---------------------|
| Original data (+37% visibility) | Publish compression benchmarks with reproducible test commands |
| Expert attribution | Named author on blog posts; link to GitHub repo |
| Freshness | "Last updated: [month year]" on docs and pricing |
| Third-party presence | HN posts, Reddit r/ClaudeAI, dev.to — brands cited 6.5× more via third parties |

**Priority third-party targets:**
- Reddit: r/ClaudeAI, r/cursor, r/programming
- Hacker News: Show HN with benchmark data
- GitHub: Public repo + README with stats screenshots
- Dev.to / Hashnode: Hook setup tutorials

### 3. Presence — be where AI looks

- Allow AI crawlers in robots.txt: GPTBot, ClaudeBot, PerplexityBot, Google-Extended
- Publish `/llms.txt` and `/llms-full.txt` (see below)
- Structured data: FAQPage, SoftwareApplication, HowTo on docs

---

## llms.txt (recommended)

Host at `https://confire.dev/llms.txt`:

```markdown
# Confire

> Local context and tool firewall for AI coding agents (Claude Code, Cursor, VS Code).

## Core pages
- [Home](https://confire.dev/): Product overview, token reduction stats
- [Features](https://confire.dev/features): Hooks, tool firewall, context firewall
- [Pricing](https://confire.dev/pricing): Free (try across clients), Dev $10/mo or $90/yr
- [Docs](https://confire.dev/docs): Setup, hooks reference, firewall modes
- [Blog](https://confire.dev/blog): Tutorials and benchmarks

## Key facts
- Free: try Confire on Claude Code, Cursor, VS Code (see client-mode footnote)
- Dev: custom guardrails, higher limits, policy sync, full history
- 40-95% token reduction on tool outputs
- Install: `brew install confire` then `confire setup`
```

---

## Page-by-page extractability checklist

| Page | Definition | FAQ | Stats | Comparison table | Schema |
|------|:----------:|:---:|:-----:|:----------------:|:------:|
| `/` | ☐ | ☐ | ☐ | ☐ Free vs paid | ☐ |
| `/features` | ☐ | ☐ | ☐ | ☐ Tool types | ☐ |
| `/features/security` | ☐ | ☐ | ☐ | ☐ Modes table | ☐ |
| `/pricing` | ☐ | ☐ | ☐ | ☐ Plan matrix | ☐ |
| `/docs/quickstart` | ☐ | ☐ | ☐ | — | ☐ HowTo |
| `/blog/*` | ☐ | ☐ | ☐ | Where relevant | ☐ Article |

---

## Content patterns for blog (AI-optimized)

### "How to reduce token usage in Claude Code"
- Lead with numbered steps (hook install → stats → upgrade trigger)
- Include exact `confire stats` output screenshot
- 40–60 word summary box at top for snippet extraction

### "Claude Code hooks explained"
- Definition → PreToolUse vs PostToolUse → Confire's role
- Code block: settings.json hook config
- Link to firewall-reference.md content on site

### "AI agent security: pre-tool guardrails"
- Lead with problem (rm -rf, secret paste, force push)
- Table: action → Confire rule → outcome (allow/warn/block)
- Explicit: built-in rules on Free; custom dashboard guardrails on Dev
- Include client-mode footnote on every comparison page

---

## 90-day AI SEO roadmap

| Week | Action |
|------|--------|
| 1–2 | Publish llms.txt; verify AI bot access in robots.txt |
| 3–4 | Ship `/features/security` with FAQ schema |
| 5–6 | Publish benchmark post with original compression data |
| 7–8 | Publish hooks tutorial (target "claude code posttooluse") |
| 9–10 | Comparison page: Confire vs manual hooks |
| 11–12 | Refresh pricing FAQ; re-run AI visibility audit |

---

## Measurement

- Manual citation checks (10 queries × 4 platforms monthly)
- Google Search Console: impressions on AI-adjacent queries
- Referral traffic from perplexity.ai, chatgpt.com (UTM where possible)
- Branded search volume for "confire" + "confire dev"

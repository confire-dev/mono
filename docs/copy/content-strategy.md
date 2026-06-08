# Content Strategy: Confire

## Business Context

**Product:** Confire — a CLI tool (Go) that acts as a context and tool firewall for AI coding agents. Reviews risky tool calls before execution and sanitizes noisy tool output before it enters context.  
**Core value:** Reduces context noise by 40–95% on Bash logs, web fetches, and other verbose tool outputs. Blocks or reviews risky tool calls.  
**Integration:** Works as a hook for Claude Code, Cursor, and VS Code.  
**Infrastructure:** All-local daemon. No cloud round-trip for processing.  
**Plans:** Free, Dev ($10/mo or $90/yr), Team (coming soon).  
**Tagline direction:** "Less noise in context = cheaper, faster, sharper AI agents."  
**Website:** confire.dev  

**Ideal Customer Profile:**
- AI developers building or running autonomous agents
- Power users of Claude Code, Cursor, or Cline who feel token costs
- AI infrastructure engineers caring about latency + cost at scale
- Developer-tool founders instrumenting agents for production

**Primary Content Goals:** Organic search traffic → signups, developer awareness, authority in the AI agent tooling space.

---

## Content Pillars (3–5 Core Topics Confire Will Own)

### Pillar 1: AI Agent Token Cost & Context Efficiency
**Rationale:** This is the direct problem Confire solves. High search intent from developers experiencing runaway token bills. Product connection is 1:1.

Subtopic clusters:
- How LLM token costs work in agentic workflows
- Context window limits and how agents hit them
- Tool call output verbosity and why it matters
- Measuring token usage in Claude Code / Cursor sessions
- Reducing agent API costs without losing quality

### Pillar 2: Claude Code & AI Coding Agent Optimization
**Rationale:** Claude Code and Cursor are the primary distribution channels. Developers searching for optimization, hooks, and setup guides are Confire's exact buyer.

Subtopic clusters:
- Claude Code hooks: what they are and how they work
- Extending and customizing Claude Code behavior
- Cursor rules and extensions for power users
- Comparing AI coding agents (Claude Code vs Cursor vs Cline)
- Running Claude Code in production / CI pipelines

### Pillar 3: Developer Tooling for AI Agents
**Rationale:** Positions Confire in the growing "AI DevOps" / AI infrastructure space. Broad enough to own long-tail; aligned with ICP's professional identity.

Subtopic clusters:
- Agent observability and output logging
- Tool call optimization patterns
- Local vs cloud processing for agent hooks
- AI agent architecture best practices
- Prompt engineering for agentic systems

### Pillar 4: Bash & Tool Output Compression Techniques
**Rationale:** Highly searchable, technical, low-competition. Developers hitting this content have immediate practical need and are a natural fit for Confire's free tier.

Subtopic clusters:
- Filtering and cleaning Bash command output
- Compressing GitHub API responses for LLM consumption
- Summarizing web fetch results for AI context
- Structured logging patterns for agent tools
- Reducing noise in Figma / file outputs sent to AI

### Pillar 5: The Business Case for AI Agent Efficiency
**Rationale:** Addresses decision-makers and founders who manage costs. Shareable content with viral potential in founder/engineering communities.

Subtopic clusters:
- Total cost of ownership for AI-assisted development
- ROI calculations for AI coding tools
- Scaling agentic workflows without scaling spend
- When to self-host vs use cloud inference
- Token cost benchmarks across AI coding agents

---

## Searchable vs Shareable Framework Applied to Confire

**Priority: searchable content first.** Confire is a technical tool with a highly searchable problem space. Developers search for their pain before they look for solutions.

### Searchable Content (Foundation)
Target developers who are actively:
- Experiencing high token costs in Claude Code or Cursor
- Hitting context window limits on complex tasks
- Searching for how to use Claude Code hooks
- Looking for ways to reduce AI API costs

Every searchable piece must match exact search intent, answer the full question, and include a natural product mention where honest.

### Shareable Content (Demand creation)
Target the developer/founder community on:
- Hacker News (technical depth + data required)
- r/LocalLLaMA, r/ClaudeAI, r/cursor, r/programming
- X/Twitter AI developer community
- Indie Hackers, Product Hunt

Best formats for Confire: data-driven benchmarks, transparent cost breakdowns, counterintuitive takes on AI agent architecture.

---

## Priority Topics (First 20 Pieces)

Scored on: Customer Impact (40%), Content-Market Fit (30%), Search Potential (20%), Resources (10%).

### TIER 1: High Priority (Build First)

| # | Topic | Type | Keyword Target | Buyer Stage | Score |
|---|-------|------|----------------|-------------|-------|
| 1 | How Claude Code hooks work (and how to write your own) | Searchable | "claude code hooks" | Implementation | 9.1 |
| 2 | Why your Claude Code sessions cost so much (and how to fix it) | Both | "claude code cost" / "reduce claude code tokens" | Awareness | 9.0 |
| 3 | How to reduce token usage in Claude Code | Searchable | "reduce token usage claude code" | Consideration | 8.8 |
| 4 | Claude Code vs Cursor vs Cline: a developer's comparison | Searchable | "claude code vs cursor" | Consideration | 8.5 |
| 5 | What is a context window and why does it fill up so fast in agents? | Searchable | "context window limit agents" | Awareness | 8.4 |
| 6 | Confire: how we compress tool call outputs by up to 95% [data post] | Shareable | — | — | 8.3 |
| 7 | Claude Code setup guide for power users | Searchable | "claude code setup" | Implementation | 8.2 |
| 8 | How to control what goes into Claude's context window | Searchable | "control claude context window" | Awareness | 8.1 |

### TIER 2: Medium Priority (Build Next)

| # | Topic | Type | Keyword Target | Buyer Stage | Score |
|---|-------|------|----------------|-------------|-------|
| 9 | Best tools for AI agent observability in 2025 | Searchable | "ai agent observability tools" | Consideration | 7.8 |
| 10 | How to set up Cursor rules that actually work | Searchable | "cursor rules setup" | Implementation | 7.6 |
| 11 | The real cost of running Claude Code for a month (our numbers) | Shareable | — | Awareness | 7.5 |
| 12 | Bash output compression techniques for AI context | Searchable | "bash output for LLM" | Implementation | 7.4 |
| 13 | How to use the Claude Code PostToolUse hook | Searchable | "claude code posttooluse" | Implementation | 7.3 |
| 14 | Token cost benchmarks: Claude Code vs Cursor vs raw API | Both | "claude code token cost benchmark" | Awareness | 7.2 |
| 15 | What is MCP and how does it change AI agent architecture? | Searchable | "what is MCP claude" | Awareness | 7.1 |

### TIER 3: Build as Capacity Allows

| # | Topic | Type | Keyword Target | Buyer Stage | Score |
|---|-------|------|----------------|-------------|-------|
| 16 | How Confire's context firewall works under the hood | Shareable | "ai agent tool firewall" | Awareness | 6.9 |
| 17 | GitHub API response compression for LLM agents | Searchable | "github api llm compression" | Implementation | 6.8 |
| 18 | Alternatives to Claude Code for AI-assisted development | Searchable | "claude code alternatives" | Consideration | 6.7 |
| 19 | How to run Claude Code in CI/CD pipelines | Searchable | "claude code CI pipeline" | Implementation | 6.6 |
| 20 | AI agent cost calculator: estimate your monthly spend | Searchable (tool) | "ai agent cost calculator" | Decision | 6.5 |

---

## Topic Cluster Map

```
PILLAR 1: AI Agent Token Cost & Context Efficiency
├── What is a context window? (Awareness hub)
│   ├── Why context windows fill up in agents
│   ├── How to measure token usage in Claude Code
│   └── Context window limits by model (comparison)
├── Reducing token costs (Consideration hub)
│   ├── How to reduce token usage in Claude Code
│   ├── Token cost benchmarks across AI coding tools
│   └── Free vs Dev: when to upgrade Confire
└── Tool call output optimization (Implementation hub)
    ├── Why tool outputs are the biggest context hog
    ├── How Confire compresses tool call outputs
    └── Bash output cleaning techniques

PILLAR 2: Claude Code & AI Coding Agent Optimization
├── Claude Code guides (Implementation hub)
│   ├── Claude Code setup for power users
│   ├── How Claude Code hooks work
│   ├── Using PostToolUse and PreToolUse hooks
│   └── Writing custom Claude Code hooks
├── Agent comparison (Consideration hub)
│   ├── Claude Code vs Cursor vs Cline
│   ├── Claude Code alternatives
│   └── Best AI coding agents for [use case]
└── Customization & automation (Implementation hub)
    ├── Cursor rules that actually work
    ├── Automating Claude Code with shell scripts
    └── Running Claude Code in CI pipelines

PILLAR 3: Developer Tooling for AI Agents
├── AI agent architecture (Awareness hub)
│   ├── What is MCP and how does it work?
│   ├── AI agent tool design patterns
│   └── Local-first processing for agent hooks
└── Observability & monitoring (Consideration hub)
    ├── Best AI agent observability tools
    ├── Logging tool call inputs and outputs
    └── Debugging agentic workflows

PILLAR 4: Bash & Tool Output Compression
├── Bash output for LLM (Implementation hub)
│   ├── Filtering Bash output before sending to AI
│   ├── Compressing GitHub API responses
│   └── Cleaning web fetch results for agents
└── Structured output patterns (Implementation hub)
    ├── JSON compression techniques for LLM context
    └── Figma file output summarization

PILLAR 5: Business Case for AI Efficiency
├── Cost & ROI (Awareness/Decision hub)
│   ├── Real cost of Claude Code for a month
│   ├── AI agent cost calculator
│   └── ROI of optimizing agent context
└── Scaling agents (Consideration hub)
    ├── Scaling agentic workflows without scaling spend
    └── When to self-host AI vs use cloud inference
```

---

## Content Ideation Sources

**Forum research to conduct:**
- `site:reddit.com "claude code" cost OR token OR expensive`
- `site:reddit.com cursor "context window" OR "token limit"`
- `site:news.ycombinator.com claude code tokens`
- `site:reddit.com/r/LocalLLaMA agent context`
- Discord: Claude, Cursor, Cline servers — #help and #general

**Competitor content to analyze:**
- Cursor.com/blog
- Cline.bot docs
- Anthropic docs on Claude Code
- Any tools in the "context management" or "agent optimization" space

**Sales/support signals to extract:**
- First questions new users ask in setup
- Most common reasons people hit the free tier limit
- What surprises developers most about token costs
- What "95% reduction" claim people are skeptical of

---

## Editorial Calendar (First 90 Days)

**Month 1 (Foundation):**
- Week 1: Claude Code hooks guide (Pillar 2)
- Week 2: Why Claude Code sessions cost so much (Pillar 1)
- Week 3: How to reduce token usage in Claude Code (Pillar 1)
- Week 4: Claude Code setup for power users (Pillar 2)

**Month 2 (Authority):**
- Week 5: Claude Code vs Cursor vs Cline comparison (Pillar 2)
- Week 6: What is a context window? (Pillar 1)
- Week 7: Confire compression data post (shareable, Pillar 1)
- Week 8: How to control Claude's context window (Pillar 1)

**Month 3 (Expansion):**
- Week 9: Best AI agent observability tools (Pillar 3)
- Week 10: Bash output compression techniques (Pillar 4)
- Week 11: Real cost of Claude Code (shareable, Pillar 5)
- Week 12: Token cost benchmarks (Pillar 1 + 5)

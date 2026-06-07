# Technical SEO: Confire (confire.dev)

## Target Audience Search Behavior

Confire's buyers are technical. They search with specificity:
- They know what Claude Code is, they just want it cheaper/faster
- They compare tools before committing (Claude Code vs Cursor)
- They search for implementation help (how to write hooks, set up tools)
- They don't search for "AI agent optimizer" — they search for their symptom ("claude code expensive", "context window full")

---

## Primary Keyword Targets

### Tier 1: High-Intent, High-Fit (Build pages for these first)

| Keyword | Est. Monthly Volume | Difficulty | Intent | Page Type |
|---------|--------------------|-----------|---------|----|
| claude code hooks | 800–2k | Low | Implementation | Guide |
| reduce token usage claude code | 400–800 | Low | Solution | Guide |
| claude code cost | 600–1.2k | Low–Med | Awareness | Guide + product |
| claude code setup | 1k–3k | Med | Implementation | Guide |
| claude code posttooluse | 200–400 | Very low | Implementation | Guide |
| context window limit claude | 600–1.2k | Low | Awareness | Guide |
| claude code vs cursor | 1k–3k | Med | Comparison | Comparison page |
| cursor rules setup | 800–2k | Low–Med | Implementation | Guide |

### Tier 2: Volume Plays (Strong searchable content)

| Keyword | Est. Monthly Volume | Difficulty | Intent | Page Type |
|---------|--------------------|-----------|---------|----|
| ai agent token cost | 300–600 | Low | Awareness | Guide |
| what is mcp claude | 400–1k | Low | Awareness | Explainer |
| claude code alternatives | 500–1k | Med | Consideration | Comparison |
| ai agent observability | 200–500 | Low | Consideration | Guide + tools list |
| bash output llm | 100–300 | Very low | Implementation | Guide |
| claude code ci pipeline | 200–400 | Low | Implementation | Guide |
| context window optimization | 300–700 | Low–Med | Solution | Guide |
| cursor ai cost | 400–800 | Low–Med | Awareness | Guide |

### Tier 3: Long-Tail, High Conversion

| Keyword | Intent | Page Type |
|---------|--------|-----------|
| how to reduce claude code api costs | Solution | Guide |
| claude code hook examples | Implementation | Guide with code |
| tool call output too long | Symptom | Guide |
| claude code token usage stats | Implementation | Guide |
| how to write a claude code hook | Implementation | Tutorial |
| mcp tool output compression | Implementation | Guide |
| github api response too large for context | Symptom | Guide |
| confire ai optimizer | Branded | Homepage / Product page |

---

## Site Architecture

```
confire.dev/
├── / (Homepage)
├── /pricing
├── /docs/
│   ├── /docs/quickstart
│   ├── /docs/setup
│   ├── /docs/claude-code
│   ├── /docs/cursor
│   ├── /docs/hooks
│   └── /docs/api
├── /blog/
│   ├── /blog/claude-code-hooks-guide
│   ├── /blog/reduce-token-usage-claude-code
│   ├── /blog/claude-code-cost
│   ├── /blog/claude-code-vs-cursor-vs-cline
│   └── ... (all content under /blog)
└── /changelog
```

**URL conventions:**
- Lowercase, hyphen-separated slugs only
- No trailing slashes
- Blog post URLs: `/blog/[target-keyword-phrase]`
- No dates in URLs (content stays evergreen)

---

## Page-by-Page SEO Specs

### Homepage (/)

**Target keywords:** confire, ai agent firewall, ai agent security, mcp firewall, prompt injection protection, claude code security  
**Search intent:** Branded + solution-aware

**Meta title:** `Confire — Context Firewall for AI Coding Agents`  
**Meta description:** `Confire is a context firewall for Claude Code, Cursor, and Cline. Reviews risky tool calls and cleans noisy outputs before they reach your agent.`

**H1:** `Keep AI coding agents cleaner and safer.`  
**H2 structure:**
- What Confire does (two-pillar value prop block)
- How it works (tool firewall + context firewall)
- Capabilities (bento grid)
- Integrations (Claude Code, Cursor, Cline badges)
- Pricing

**Schema:** SoftwareApplication, Organization, FAQPage

---

### /pricing

**Target keywords:** confire pricing, ai agent optimizer pricing, claude code cost reducer  
**Meta title:** `Confire Pricing — Free, Dev, and Pro Plans`  
**Meta description:** `Start free with 500 optimizations/month. Upgrade to Dev or Pro as your agent usage grows. No surprise bills.`

**H1:** `Simple pricing. Built for how developers actually work.`

---

### /blog/claude-code-hooks-guide

**Target keywords:** claude code hooks, posttooluse hook claude code, claude code hook examples  
**Meta title:** `Claude Code Hooks: Complete Guide (with Examples)`  
**Meta description:** `Learn how Claude Code hooks work, how to use PreToolUse and PostToolUse, and how to write your own hook to control agent behavior.`

**H1:** `Claude Code Hooks: Complete Guide`  
**H2 structure:**
- What are Claude Code hooks?
- PreToolUse vs PostToolUse: when to use each
- How to write a Claude Code hook (step-by-step)
- Hook examples: filtering output, logging, compression
- Advanced: chaining hooks and error handling
- FAQ

**Internal links to:** /docs/hooks, /blog/reduce-token-usage-claude-code  
**Schema:** Article, HowTo, FAQPage

---

### /blog/reduce-token-usage-claude-code

**Target keywords:** reduce token usage claude code, lower claude code costs, claude code token limit  
**Meta title:** `How to Reduce Token Usage in Claude Code (Practical Guide)`  
**Meta description:** `Practical techniques to cut Claude Code token consumption: hook-based output compression, context pruning, and tool call filtering.`

**H1:** `How to Reduce Token Usage in Claude Code`  
**H2 structure:**
- Why token usage spikes in Claude Code sessions
- The biggest culprit: tool call output verbosity
- Technique 1: Use PostToolUse hooks to compress output
- Technique 2: Prune context between turns
- Technique 3: Filter Bash output before it enters context
- Technique 4: Use Confire for automated optimization
- Benchmarks: before vs after
- FAQ

**Internal links to:** /blog/claude-code-hooks-guide, /pricing  
**Schema:** Article, HowTo, FAQPage

---

### /blog/claude-code-cost

**Target keywords:** claude code cost, how much does claude code cost, claude code expensive  
**Meta title:** `Claude Code Cost: What You Actually Pay and How to Cut It`  
**Meta description:** `A clear breakdown of Claude Code costs — API pricing, token consumption patterns, and the fastest ways to reduce your monthly AI bill.`

**H1:** `Claude Code Cost: What You Actually Pay and How to Reduce It`  
**H2 structure:**
- How Claude Code pricing works (API tokens, not seats)
- What drives token costs up in real sessions
- Average monthly costs by developer type
- 5 ways to reduce your Claude Code bill
- When it's worth paying more vs optimizing
- FAQ

**Schema:** Article, FAQPage

---

### /blog/claude-code-vs-cursor-vs-cline

**Target keywords:** claude code vs cursor, claude code vs cursor vs cline, best ai coding agent  
**Meta title:** `Claude Code vs Cursor vs Cline: Which AI Coding Agent Should You Use?`  
**Meta description:** `An honest comparison of Claude Code, Cursor, and Cline. We cover cost, customization, context handling, and which is best for different workflows.`

**H1:** `Claude Code vs Cursor vs Cline: A Developer's Honest Comparison`  
**H2 structure:**
- Quick verdict (summary table)
- Claude Code: strengths, weaknesses, best for
- Cursor: strengths, weaknesses, best for
- Cline: strengths, weaknesses, best for
- Cost comparison (token cost in real usage)
- Context handling: how each manages tool output
- Which should you use? (decision matrix)
- FAQ

**Schema:** Article, FAQPage, Table markup

---

### /blog/context-window-guide

**Target keywords:** context window limit, what is context window, context window claude  
**Meta title:** `What Is a Context Window? Why It Fills Up and What to Do`  
**Meta description:** `Understand context windows in LLMs and AI agents — how they fill up, what happens when they do, and how to keep them clean.`

**H1:** `What Is a Context Window? (And Why It Matters for AI Agents)`  
**H2 structure:**
- What a context window is
- How context fills up in agentic sessions
- Why tool call outputs are the worst offender
- What happens when you hit the limit
- How to extend effective context capacity
- FAQ

**Schema:** Article, FAQPage

---

## Technical SEO Checklist

### Core Infrastructure
- [ ] Canonical URLs on all pages (self-referencing)
- [ ] robots.txt: allow all, disallow /docs/internal, /admin
- [ ] XML sitemap at /sitemap.xml, submitted to GSC
- [ ] HTTPS enforced, no mixed content
- [ ] Hreflang: not needed initially (English-only)

### Performance (Core Web Vitals)
- [ ] LCP < 2.5s: optimize hero image, use next-gen formats (WebP/AVIF)
- [ ] CLS < 0.1: reserve space for images, avoid layout shifts
- [ ] INP < 200ms: minimize blocking JS
- [ ] Static pages where possible (Astro, Next static export, or similar)
- [ ] CDN for assets (Cloudflare Pages is a natural fit given Cloudflare Worker backend)

### On-Page Basics
- [ ] Every page has unique title (50–60 chars) and description (150–160 chars)
- [ ] H1: exactly one per page, matches primary keyword intent
- [ ] H2–H4: descriptive, keyword-present where natural
- [ ] Images: descriptive alt text, compressed, lazy-loaded below fold
- [ ] Internal links: 2–4 per post, anchor text descriptive (not "click here")
- [ ] External links: open in new tab, rel="noopener"

### Schema Markup (Structured Data)
- Homepage: `SoftwareApplication`, `Organization`
- Blog posts: `Article` with author, datePublished, dateModified
- How-to posts: `HowTo` schema
- FAQ sections: `FAQPage` schema
- Comparison posts: `Table` markup + `Article`

### Developer SEO (LLM/AI Search Discovery)
- [ ] `llms.txt` at confire.dev/llms.txt — machine-readable product summary for LLM crawlers
- [ ] Clear, structured content so AI search (Perplexity, ChatGPT browse, Claude) can cite correctly
- [ ] Product facts in consistent language across all pages (name, core metric "up to 95%", integrations)
- [ ] FAQ sections on key pages to capture conversational queries

---

## Meta Title & Description Templates

**Blog post title formula:**
`[Target Keyword]: [Value/Outcome/Clarity hook] | Confire`

**Blog post description formula:**
`[What the post answers]. [One concrete takeaway or proof point]. [Optional: what readers will be able to do].`

**Keep titles under 60 characters. Descriptions 150–160 characters.**

---

## Internal Linking Strategy

**Hub pages link to:** All spokes in their cluster + /pricing + /docs/quickstart  
**Spoke pages link to:** Their hub + 1–2 related spokes + /pricing (where honest)  
**Blog posts must always have:** At least one link to docs or a related guide  

**Anchor text rules:**
- Use exact or near-exact keyword match for pillar hub links
- Use descriptive phrases for supporting pages
- Never use "click here," "read more," or generic anchors

---

## Tracking & Measurement

**Set up in Google Search Console:**
- Submit sitemap
- Monitor impressions/clicks for target keywords weekly
- Watch for crawl errors on new posts

**Key metrics to track (monthly):**
- Organic sessions by page
- Keyword rank movement for Tier 1 targets
- Click-through rate from SERPs (benchmark: 3–5% for informational, 5–10% for branded)
- Signups attributed to organic (UTM or first-touch)
- Pages indexed vs submitted

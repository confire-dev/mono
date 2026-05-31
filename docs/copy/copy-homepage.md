# Homepage Copy: Confire (confire.dev)

---

## Above the Fold / Hero

**H1:**
Less noise in context.  
Cheaper, faster, sharper AI agents.

**Subheadline (20–30 words):**
Confire sits between your AI agent and the model context. It compresses tool call outputs — Bash logs, GitHub responses, web fetches — before they eat your tokens.

**Hero CTA (primary):**
`Get started free →`

**Secondary CTA:**
`See how it works`

**Social proof line (below CTAs):**
Works with Claude Code, Cursor, and Cline. Free up to 500 calls/month.

**Hero stat callout (visual block below CTA):**
```
40–95%
token reduction on tool call outputs
```

---

## How It Works (3-Step Section)

**Section label:** How Confire works

**Step 1**  
**Heading:** Install in 30 seconds  
**Body:** One command adds Confire as a hook to your AI agent. No code changes, no configuration required to start.  
```bash
confire setup
```

**Step 2**  
**Heading:** Confire intercepts tool outputs  
**Body:** Every time your agent runs a Bash command, fetches a URL, or calls a tool, Confire processes the response before it reaches the model context.

**Step 3**  
**Heading:** Your agent gets clean, compressed context  
**Body:** The model sees exactly what it needs — no verbose stack traces, no 10,000-line GitHub API blobs, no noise. Fewer tokens. Same results.

---

## Results / Proof Section

**Section heading:** Real token reduction. Not rounding.

**Subhead:** We compress the verbose parts of agentic workflows that eat context without adding value.

**Stat cards:**

| Tool Output Type | Typical Reduction |
|-----------------|-------------------|
| Bash command logs | 60–90% |
| GitHub API responses | 70–95% |
| Web page fetches | 50–85% |
| Figma file outputs | 40–80% |
| General tool output | 40–95% |

**Supporting copy:**  
Confire uses a cloud optimizer (Cloudflare Worker) with a local fallback. Latency is under 50ms on optimized calls. Your agent doesn't wait — it just gets better input.

---

## Features Section

**Section heading:** Built for how developers actually use AI agents

**Feature 1**  
**Icon:** hook/connector  
**Heading:** Drop-in hook for Claude Code & Cursor  
**Body:** Confire works as a PostToolUse hook. No changes to your workflow. Works with Claude Code, Cursor, and Cline out of the box.

**Feature 2**  
**Icon:** cloud + server  
**Heading:** Cloud optimizer with local fallback  
**Body:** Processing happens on Cloudflare's edge network. If the cloud is unreachable, the local optimizer kicks in automatically. Always on.

**Feature 3**  
**Icon:** chart/graph  
**Heading:** Usage stats and savings tracking  
**Body:** Run `confire stats` to see how many tokens you've saved, which tools produce the most noise, and your monthly usage across calls.

**Feature 4**  
**Icon:** lock/shield  
**Heading:** Your data stays yours  
**Body:** Confire processes tool output content to compress it, but does not store, log, or share the content of your tool calls.

**Feature 5**  
**Icon:** bolt/speed  
**Heading:** Go from 100,000-token sessions to 20,000  
**Body:** Context bloat is the number one reason agents get confused, slow, and expensive. Confire attacks it at the source.

**Feature 6**  
**Icon:** CLI  
**Heading:** Simple CLI, no dashboard required  
**Body:** `confire setup` · `confire login` · `confire stats` · `confire start` · `confire stop`. That's the whole surface area.

---

## Integration Section

**Section heading:** Plugs into your existing stack

**Subhead:** If you use Claude Code, Cursor, or Cline today, Confire starts working in under a minute.

**Integration badges/logos:**
- Claude Code (Anthropic)
- Cursor
- Cline
- Cloudflare (infrastructure)

**Small-print copy:**  
Confire uses the standard hook protocol for each agent. No custom forks, no private APIs.

---

## Pricing Preview (Homepage Teaser)

**Section heading:** Start free. Upgrade when you need to.

**Plan cards:**

**Free**
500 optimized calls/month  
Cloud + local optimizer  
Claude Code, Cursor, Cline support  
`Get started free →`

**Dev**  
[X] calls/month  
Priority cloud optimizer  
Usage analytics  
Email support  
`Start Dev plan →`

**Pro**  
Unlimited calls  
SLA-backed uptime  
Team usage dashboard  
Priority support  
`Talk to us →`

**Subtext below pricing:**  
500 free calls goes a long way when every call saves 40–95% of token output. Most developers start free and never need more.

---

## Final CTA Section

**Heading:** Stop paying for context you don't need.

**Body:**  
Every Bash log your agent dumps into context, every bloated GitHub response, every 200KB web fetch — you're paying to process all of it. Confire removes the noise before the model ever sees it.

**CTA:** `Get started free — no card required`

**Under CTA:**  
Takes 30 seconds to install. Works immediately.

---

## Footer

**Tagline:** Confire — Less noise in context = cheaper, faster, sharper AI agents.

**Links:**
- Docs
- Pricing
- Blog
- GitHub
- Changelog

**Legal:** Privacy Policy · Terms of Service

---

## SEO Notes for Homepage

- **Title tag:** `Confire — Reduce AI Agent Token Costs by Up to 95%`
- **Meta description:** `Confire is a context optimizer for Claude Code, Cursor, and Cline. Cut token costs on Bash logs, GitHub responses, and web fetches without changing how you work.`
- **H1:** matches hero headline (or close variant)
- **Schema:** SoftwareApplication + FAQPage

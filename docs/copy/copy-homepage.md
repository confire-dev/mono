# Homepage Copy: Confire (confire.dev)

---

## Above the Fold / Hero

**H1:**
Keep AI coding agents cleaner and safer.

**Subheadline (20–30 words):**
Confire reviews risky tool calls before they run and sanitizes noisy tool output before it enters context.

**Hero CTA (primary):**
`Start free →`

**Secondary CTA:**
`See how it works`

**Social proof line (below CTAs):**
Works with Claude Code, Cursor, and VS Code.

**Client-mode footnote (below social proof):**
Claude Code supports full hook-based firewall mode. Cursor and VS Code support MCP gateway mode for tools routed through Confire.

**Hero stat callout (visual block below CTA):**
```
5 layers
of output protection per tool call
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
**Body:** Confire works as a PreToolUse and PostToolUse hook. No changes to your workflow. Works with Claude Code, Cursor, and VS Code out of the box.

**Feature 2**  
**Icon:** loop/stop  
**Heading:** Stops runaway agent loops automatically  
**Body:** When your agent gets stuck repeating the same command, Confire blocks it after 20 identical calls in 5 minutes — before it can cause damage at scale. No config required.

**Feature 3**  
**Icon:** shield/check  
**Heading:** Irreversible actions always get a second look  
**Body:** Commands that can't be undone — `rm -rf`, force push, database resets, production deploys — are always surfaced for review. The ask-budget mechanism never skips them.

**Feature 4**  
**Icon:** cloud + server  
**Heading:** Cloud optimizer with local fallback  
**Body:** Processing happens on Cloudflare's edge network. If the cloud is unreachable, the local optimizer kicks in automatically. Always on.

**Feature 5**  
**Icon:** chart/graph  
**Heading:** Usage stats and savings tracking  
**Body:** Run `confire stats` to see how many tokens you've saved, which tools produce the most noise, and your monthly usage across calls.

**Feature 6**  
**Icon:** lock/shield  
**Heading:** Your data stays yours  
**Body:** Local protection runs entirely on your machine — block, loop detection, and budget decisions never leave the daemon. Remote optimizations receive only sanitized content, never raw tool output.

---

## Integration Section

**Section heading:** Plugs into your existing stack

**Subhead:** If you use Claude Code, Cursor, or Cline today, Confire starts working in under a minute.

**Integration badges/logos:**
- Claude Code (Anthropic)
- Cursor
- VS Code
- Cloudflare (infrastructure)

**Small-print copy:**  
Claude Code uses full PreToolUse/PostToolUse hooks. Cursor and VS Code use MCP gateway mode for tools routed through Confire.

---

## Pricing Preview (Homepage Teaser)

**Section heading:** Start free. Upgrade when Confire becomes part of your daily agent workflow.

**Plan cards:**

**Free — $0/mo**  
For trying Confire with your AI coding agent.

Try Confire’s local firewall for Claude Code, Cursor, and VS Code. Runaway loop detection, risky action review, and output sanitization — all local, all free.

- Claude Code full firewall
- Cursor + VS Code MCP gateway
- Runaway loop + call rate protection
- 500 remote optimizations/month
- Basic savings stats

`Start free →`

**Dev — $10/mo · $90/yr**  
For daily AI coding with Confire always on.

Use Confire daily with higher limits, custom firewall rules, updated source-specific optimizers, and full history.

- Everything in Free
- 5,000 remote optimizations/month
- Custom dashboard guardrails
- Configurable rate thresholds
- Security event + provenance history
- Optimization history

`Start Dev →`

**Team — Coming soon**  
Shared policies, audit logs, team dashboard.

`Notify me →`

**Footnote below pricing:**  
Claude Code supports full hook-based firewall mode. Cursor and VS Code support MCP gateway mode for tools routed through Confire. Remote optimizations receive sanitized/redacted content only.

---

## Final CTA Section

**Heading:** Keep AI coding agents cleaner and safer.

**Body:**  
Risky tool calls get reviewed before they run. Noisy outputs get sanitized before they hit your context window. Start free on Claude Code, Cursor, or VS Code.

**CTA:** `Start free — no card required`

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

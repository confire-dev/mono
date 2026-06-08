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
**Heading:** Your agent gets clean, filtered context  
**Body:** The model sees exactly what it needs — no verbose stack traces, no bloated API responses, no noise. Less context pollution. Same results.

---

## Results / Proof Section

**Section heading:** Real noise reduction. Not rounding.

**Subhead:** Confire trims the verbose parts of agentic workflows that eat context without adding value.

**Stat cards:**

| Tool Output Type | Typical Reduction |
|-----------------|-------------------|
| Bash command logs | 60–90% |
| GitHub API responses | 70–95% |
| Web page fetches | 50–85% |
| Figma file outputs | 40–80% |
| General tool output | 40–95% |

**Supporting copy:**  
All processing runs locally in the Confire daemon. No proxy, no cloud round-trip.

---

## Features Section

**Section heading:** Built for how developers actually use AI agents

**Feature 1**  
**Icon:** hook/connector  
**Heading:** Drop-in hook for Claude Code & Cursor  
**Body:** Confire works as a PostToolUse hook. No changes to your workflow. Works with Claude Code, Cursor, and VS Code out of the box.

**Feature 2**  
**Icon:** shield  
**Heading:** Local-first by default  
**Body:** All firewall decisions and context passes run in the daemon on your machine. No tool output leaves your machine.

**Feature 3**  
**Icon:** chart/graph  
**Heading:** Usage stats and savings tracking  
**Body:** Run `confire stats` to see how many tool calls were processed, which tools produce the most noise, and your session activity.

**Feature 4**  
**Icon:** lock/shield  
**Heading:** Your data stays yours  
**Body:** Tool output content is processed locally and never forwarded to the cloud. Only structured telemetry events (risk level, action taken) are sent when you're logged in.

**Feature 5**  
**Icon:** bolt/speed  
**Heading:** Go from 100,000-token sessions to 20,000  
**Body:** Context bloat is the number one reason agents get confused, slow, and expensive. Confire attacks it at the source.

**Feature 6**  
**Icon:** CLI  
**Heading:** Simple CLI, no dashboard required  
**Body:** `confire setup` · `confire login` · `confire status` · `confire start` · `confire stop`. That's the whole surface area.

---

## Integration Section

**Section heading:** Plugs into your existing stack

**Subhead:** If you use Claude Code, Cursor, or VS Code today, Confire starts working in under a minute.

**Integration badges/logos:**
- Claude Code (Anthropic)
- Cursor
- VS Code

**Small-print copy:**  
Claude Code uses full PreToolUse/PostToolUse hooks. Cursor and VS Code use MCP gateway mode for tools routed through Confire.

---

## Pricing Preview (Homepage Teaser)

**Section heading:** Start free. Upgrade when Confire becomes part of your daily agent workflow.

**Plan cards:**

**Free — $0/mo**  
For trying Confire with your AI coding agent.

Try Confire's local firewall for Claude Code, Cursor, and VS Code. Clean noisy tool output and review risky agent actions with built-in rules.

- Claude Code full firewall
- Cursor + VS Code MCP gateway
- Built-in noise trimming
- Basic security event stats

`Start free →`

**Dev — $10/mo · $90/yr**  
For daily AI coding with Confire always on.

Use Confire daily with higher limits, custom firewall rules, and local policy sync.

- Everything in Free
- Custom dashboard guardrails
- Policy sync
- Full security event history

`Start Dev →`

**Team — Coming soon**  
Shared policies, audit logs, team dashboard.

`Notify me →`

**Footnote below pricing:**  
Claude Code supports full hook-based firewall mode. Cursor and VS Code support MCP gateway mode for tools routed through Confire.

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

- **Title tag:** `Confire — Context and Tool Firewall for AI Coding Agents`
- **Meta description:** `Confire is a context and tool firewall for Claude Code, Cursor, and VS Code. Review risky tool calls and clean noisy outputs before they reach your agent.`
- **H1:** matches hero headline (or close variant)
- **Schema:** SoftwareApplication + FAQPage

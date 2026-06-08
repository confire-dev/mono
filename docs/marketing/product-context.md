# Confire — Product Marketing Context

Shared foundation for all marketing docs. Update this first when positioning changes.

**Last updated:** June 2026  
**Site:** confire.dev  
**Public tiers:** Free + Dev ($10/mo or $90/yr). Team — coming soon.

---

## Headlines (pricing + homepage)

**Above plans:**  
Start free. Upgrade when Confire becomes part of your daily agent workflow.

**Or stronger hero:**  
Keep AI coding agents cleaner and safer.

**Subline:**  
Confire reviews risky tool calls before they run and sanitizes noisy tool output before it enters context.

**Footnote (required on pricing + features):**  
Claude Code supports full hook-based firewall mode. Cursor and VS Code support MCP gateway mode for tools routed through Confire.

---

## One-line

Confire reviews risky agent tool calls before they run and sanitizes noisy tool output before it enters context.

## Category claim

**AI agent context + tool firewall** — safety and efficiency for supported coding clients, not a single-host plugin.

## The clean split

| | Free | Dev |
|---|------|-----|
| **Best for** | Trying Confire | Daily AI coding |
| **Price** | $0 | $10/mo or $90/yr |
| **Core promise** | Full context + tool firewall across supported clients | Higher limits, custom guardrails, policy sync, full history |

**Do not frame as:** "Security free, optimizers paid."  
**Do frame as:** "Free lets you try Confire end-to-end on supported clients. Dev is for always-on daily use with custom rules and a continuously updated library."

---

## Client support (honest modes)

| Client | Mode | What Confire does |
|--------|------|-------------------|
| **Claude Code** | Full hooks | PreToolUse risky-action review + PostToolUse sanitization + noise trimming |
| **Cursor** | MCP gateway | Protects/sanitizes tools routed through Confire |
| **VS Code** | MCP gateway | Protects/sanitizes tools routed through Confire |

Always include the footnote when listing clients. Never imply Cursor/VS Code get the same hook depth as Claude Code.

**Not v1:** OpenCode, OpenClaw, Windsurf. Do not lead with Cline.

---

## Free — $0

**For:** Trying Confire with your AI coding agent.

**Card copy:**  
Try Confire's local firewall for Claude Code, Cursor, and VS Code. Clean noisy tool output and review risky agent actions with built-in rules.

**CTA:** Start free

### Includes

**Claude Code**
- Full hook-based firewall
- PreToolUse risky action review
- PostToolUse output sanitization + noise trimming

**Cursor**
- MCP gateway mode
- Protects/sanitizes Confire-routed MCP/tool calls

**VS Code**
- MCP gateway mode
- Protects/sanitizes Confire-routed MCP/tool calls

**Platform**
- Local CLI
- Built-in firewall rules (risky Git/GitHub, destructive shell/DB, publish/deploy warnings, mutating MCP review)
- Secret redaction
- Prompt-injection pattern sanitization
- Built-in noise trimming (Bash, WebFetch, MCP, Generic)
- Basic token/context savings stats
- Local policy evaluation
- Local-only mode

---

## Dev — $10/mo or $90/yr

**For:** Daily AI coding with Confire always on.

**Card copy:**  
Use Confire daily with higher limits, custom firewall rules, and local policy sync across supported clients.

**CTA:** Start Dev

### Includes everything in Free, plus

- Custom guardrail rules from dashboard
- Remote policy sync to local CLI
- Rule cache for offline/local evaluation
- Firewall history (reviewed, blocked, sanitized, secrets redacted)
- Larger input payload limits
- More detailed savings stats (bytes, estimated tokens, estimated cost, top noisy tools)
- Tool-use guidance ("do not refetch", safer/narrower tool, MCP mutates external state)
- Cursor/VS Code MCP gateway advanced rules
- Early access to new clients/adapters
- Priority fixes for client/tool compatibility

**Top-ups (Dev only):** $5 = 5,000 additional requests, never expire.

---

## Team — coming soon

Small third card on pricing: **Team — Coming soon / Contact us**

Future: shared policies, audit logs, team dashboard, self-hosting, admin-enforced rules. Do not sell details until shipped.

---

## Shorter pricing-card bullets

### Free card
- ✓ Claude Code full firewall
- ✓ Cursor + VS Code MCP gateway
- ✓ Local CLI
- ✓ Built-in risky action review
- ✓ Secret + prompt-injection sanitization
- ✓ Built-in noise trimming
- ✓ Basic savings stats

### Dev card
- ✓ Everything in Free
- ✓ Custom dashboard guardrails
- ✓ Remote policy sync
- ✓ Firewall history
- ✓ Larger payloads
- ✓ Detailed savings stats
- ✓ Tool-use guidance
- ✓ Early access to new adapters

---

## Ideal customer profile

**Primary:** Developers trying or daily-driving Claude Code, Cursor, or VS Code with AI agents.

**Secondary:** Eng leads who want guardrails + context efficiency without re-platforming per editor.

## Jobs to be done

1. "I want to try a firewall without committing" → Free, all supported clients
2. "Confire is part of every session now" → Dev (limits, custom rules, history)
3. "My agent runs dangerous commands" → built-in rules (Free); custom rules (Dev)
4. "Tool output bloats my context" → built-in noise trimming + stats (Free); detailed stats + larger payloads (Dev)

## Brand voice

- Direct, technical, honest about client modes
- Numbers when provable; no overpromise on Cursor/VS Code vs Claude Code
- Free is real try-out, not a crippled trial
- Silent SessionStart on healthy sessions — don't spam context

## Primary CTAs

- Free: `Start free` — no card
- Dev: `Start Dev`
- At limit: point to pricing; frame as daily-workflow upgrade, not "unlock security"

## Open decisions

- [ ] Team waitlist fields (team size?)
- [ ] Case studies / logos
- [ ] Product Hunt date

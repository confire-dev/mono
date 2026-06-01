# Confire — Product Marketing Context

Shared foundation for all marketing docs. Update this first when positioning changes.

**Last updated:** June 2026  
**Site:** confire.dev  
**Stage:** Pre-GA launch — Free + Developer tiers active; Team/Enterprise on roadmap

---

## One-line

Confire is a local context and tool firewall for AI coding agents — it blocks risky tool calls and compresses noisy outputs before they reach the model.

## Category claim

**AI agent context firewall** — not just compression, not just security. Confire sits between your agent and the model: PreToolUse for safety, PostToolUse for context efficiency.

## Core value proposition

Less noise in context = cheaper, faster, sharper AI agents.

- **40–95% token reduction** on Bash logs, GitHub responses, web fetches, MCP tool output
- **Built-in security rules** on every plan — dangerous commands reviewed or blocked before they run
- **Drop-in hooks** for Claude Code, Cursor, and VS Code — `confire setup` in 30 seconds

## Positioning: free vs paid

| | Free | Developer ($10/mo) |
|---|------|-------------------|
| **Security / firewall** | ✅ Built-in rules (balanced/strict/observe) | ✅ Same + custom policy rules from dashboard |
| **Context optimization** | ✅ Local optimizers + 500 cloud req/mo | ✅ All 13 MCP optimizers + 5,000 req/mo |
| **Hosts** | Claude Code, Cursor, VS Code | Same |
| **Stats** | `confire stats` CLI | + 30-day history, web dashboard |
| **API** | — | 1M tokens/mo |
| **Top-ups** | — | $5 = 5,000 requests |

**Key message:** Security is not a paywall. Every developer gets the context firewall. Paid unlocks configurable rules and deep MCP optimization.

## Ideal customer profile (ICP)

**Primary:** Developers using Claude Code, Cursor, or VS Code daily who feel token costs or context bloat.

**Secondary:**
- Engineering leads managing AI-assisted workflows across a team
- Security-conscious orgs wanting guardrails before tool execution
- AI infrastructure engineers optimizing agent cost at scale

**Not for v1:** OpenCode, OpenClaw, Windsurf (no post-tool steer API). Cline deprioritized — marketing should say Claude Code, Cursor, VS Code.

## Jobs to be done

1. "My Claude Code sessions cost too much" → show savings via `confire stats`
2. "My agent keeps reading 10k-line test output" → Bash/GitHub compression
3. "I'm worried the agent will rm -rf or leak secrets" → built-in firewall (free)
4. "Jira/Figma MCP output is unreadable garbage" → Developer-tier optimizers

## Brand voice

- **Direct, technical, no hype** — write for skeptical senior engineers
- **Show numbers** — 97% on test output, 87% on GitHub PRs, not "massive savings"
- **Honest about limits** — free tier is real, not a trial; firewall is best-effort not a security boundary
- **Context is the product** — Confire should not spam the agent UI; silent when healthy

## Competitors / alternatives

| Alternative | Confire angle |
|-------------|---------------|
| Manual `.claudeignore` / rules | Reactive; Confire is automatic per tool call |
| Cursor built-in context | Host-specific; Confire works cross-agent |
| Raw hooks / DIY compression | Confire ships rules + cloud optimizers + stats |
| Nothing | Pay full token cost + no pre-tool guardrails |

## Proof points

- Bash test output: up to 97% reduction
- GitHub PR responses: up to 87% reduction
- WebFetch extraction: 80–90%
- 24+ built-in firewall rules loaded at session start
- Cloud optimizer on Cloudflare Workers, <50ms typical latency
- Local fallback when offline or over quota

## Primary CTAs

- Homepage / docs: `Get started free →` (no card)
- CLI at limit: `confire.dev/pricing`
- Upgrade nudge: Developer for MCP optimizers + custom rules

## Open decisions (marketing)

- [ ] Finalize homepage hero for security + optimization dual pillar
- [ ] Team tier waitlist copy on pricing page
- [ ] Case studies / logos (none shipped yet)
- [ ] Product Hunt date TBD

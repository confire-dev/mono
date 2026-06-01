# Copy Editing Audit — Confire

**Skill:** copy-editing  
**Scope:** `docs/copy/` + marketing alignment with [product-context.md](./product-context.md)  
**Goal:** Honest Free/Dev split and client-mode wording before GA

---

## Executive summary

**Wrong direction (fixed):** Framing Free as “security only” and Dev as “MCP optimizers unlock” — undersells Free and misstates the product.

**Correct direction:** Free = try Confire across supported clients (full Claude Code firewall + Cursor/VS Code gateway + built-in rules + basic optimizers). Dev = daily use (limits, custom guardrails, updated library, history). Always footnote client modes.

**Priority fixes:** Pricing page (done in `copy-pricing.md`), homepage hero, remove Cline/Pro references, add support footnote everywhere clients are listed.

---

## Sweep 1: Accuracy & freshness

| File | Finding | Fix | Priority |
|------|---------|-----|----------|
| `copy-pricing.md` | Old Dev/Pro, $95/yr | ✅ Rewritten — Free + Dev + Team coming soon | Done |
| `copy-homepage.md` | Cline, old pricing table | Claude Code/Cursor/VS Code; Free + Dev + Team soon | P0 |
| `copy-features.md` | Cline, Pro-only features | v1 clients + client-mode section | P0 |
| `copy-ctas.md` | Cline, Dev/Pro | Free/Dev CTAs | P0 |
| `copy-onboarding.md` | Cline in setup, Pro upgrade | Three clients + Dev upgrade | P0 |
| `docs/marketing/*` | “Developer”, security-as-paywall | ✅ product-context.md is source of truth | Done |

---

## Sweep 2: Required messaging

Every page listing clients must include:

> Claude Code supports full hook-based firewall mode. Cursor and VS Code support MCP gateway mode for tools routed through Confire.

Pricing must use:

- **Headline:** Start free. Upgrade when Confire becomes part of your daily agent workflow.
- **Plans:** Free ($0), Dev ($10/mo or $90/yr), Team (coming soon)
- **Never:** “Security free, optimizers paid” as primary pitch

---

## Sweep 3: Homepage hero (recommended)

**H1:** Keep AI coding agents cleaner and safer.

**Sub:** Confire reviews risky tool calls before they run and sanitizes noisy tool output before it enters context.

**Social proof:** Works with Claude Code, Cursor, and VS Code.

---

## Sweep 4: Pricing card copy (canonical)

See `product-context.md` and `copy-pricing.md` for full lists and short bullet cards.

---

## Final QA

- [ ] Free described as “try across supported clients”
- [ ] Dev described as “daily workflow” not “unlock firewall”
- [ ] Client-mode footnote on pricing, features, homepage
- [ ] No Pro tier on public pricing
- [ ] Team = coming soon only
- [ ] $90/year (not $95) for Dev

# Copy Editing Audit — Confire

**Skill:** copy-editing  
**Scope:** Existing draft copy in `docs/copy/` vs current product (security free, v1 hosts, Developer pricing)  
**Goal:** Refresh before GA — not full rewrite

---

## Executive summary

Existing copy is strong on **context optimization** but underplays **security/firewall** (now a free-tier pillar) and still references **Cline** in several places (v1 ships Claude Code, Cursor, VS Code only). Pricing copy uses placeholder `[price]` and old plan names (Dev/Pro vs Free/Developer).

**Priority fixes:** Homepage hero dual pillar, host list, pricing accuracy, security messaging, remove Cline from v1 CTAs.

---

## Sweep 1: Accuracy & freshness

| File | Finding | Impact | Recommendation | Priority |
|------|---------|--------|----------------|----------|
| `copy-homepage.md` | Social proof says "Cline" | Wrong audience signal | Change to "Claude Code, Cursor, and VS Code" | P0 |
| `copy-homepage.md` | No mention of free security/firewall | Misses key differentiator | Add sub-bullet under hero or new feature block | P0 |
| `copy-features.md` | Supported agents lists Cline | Out of scope v1 | Remove Cline; add VS Code | P0 |
| `copy-pricing.md` | Uses Dev/Pro plan names | Mismatch with product | Align to Free / Developer / Team (soon) | P0 |
| `copy-pricing.md` | `[price]` placeholders | Can't ship | Use $10/mo Developer, $95/yr | P0 |
| `copy-pricing.md` | No firewall on free called out | Under-sells free tier | Add row: "Built-in tool firewall" on Free | P0 |
| `copy-ctas.md` | Multiple Cline references | Same as above | Global replace for v1 surfaces | P1 |
| `copy-onboarding.md` | Setup lists Cline option | Confusing setup flow | Claude Code, Cursor, VS Code only | P0 |
| `content-strategy.md` | Cline in ICP/distribution | Strategy drift | Update to match v1 hosts | P1 |

---

## Sweep 2: Clarity & structure

| File | Finding | Recommendation |
|------|---------|----------------|
| `copy-homepage.md` | Hero only mentions compression | Split value: "Secure tool calls. Compress context." |
| `copy-features.md` | "Wall of features" in compression table | Keep table — it's proof; add intro sentence on *why* each matters |
| `copy-pricing.md` | FAQ doesn't explain firewall tiers | Add: "Built-in rules free; custom rules on Developer" |

**Suggested homepage hero refresh:**

> **H1:** Secure your agent. Compress your context.  
> **Sub:** Confire hooks into Claude Code, Cursor, and VS Code — blocking risky tool calls before they run and cutting noisy output by up to 95% before it hits your context window.

---

## Sweep 3: Voice & tone

| Issue | Location | Fix |
|-------|----------|-----|
| "Customers love us" absent — good | — | Keep using numbers not adjectives |
| "No per-seat nonsense" in pricing meta | copy-pricing.md | Fine for DevTools voice; keep |
| Firewall described as "security boundary" nowhere | copy-features | Add honest line from firewall-reference: best-effort, not guarantee |

---

## Sweep 4: Conversion & CTAs

| File | Finding | Recommendation |
|------|---------|----------------|
| `copy-ctas.md` | Strong free CTAs | Add security-angle variant: "Free firewall + 500 optimizations/mo" |
| `copy-pricing.md` | Upgrade trigger weak on security | "Need custom rules for your team? → Developer" |
| `copy-onboarding.md` | Welcome could mention firewall silently active | One line after setup: "Built-in guardrails are on — run `confire help`" |

---

## Sweep 5: SEO & meta

| Page | Current meta issue | Fix |
|------|-------------------|-----|
| Features | No "security" in title | `Confire Features — Context Optimization & Agent Firewall` |
| Pricing | "Dev and Pro" in title | `Confire Pricing — Free & Developer Plans` |
| Homepage | OK direction | Add "AI agent firewall" to meta description |

---

## Sweep 6: Specific line edits (high-impact)

### copy-homepage.md — social proof line

**Before:** `Works with Claude Code, Cursor, and Cline. Free up to 500 calls/month.`  
**After:** `Works with Claude Code, Cursor, and VS Code. Free firewall + 500 optimizations/month.`

### copy-pricing.md — Free tier bullets

**Add:**
- Built-in tool firewall (balanced mode)
- 24+ security rules: block, warn, or review risky commands
- Local optimizers (Bash, Read, WebFetch)

**Add to Developer:**
- Custom firewall rules from dashboard
- 13 MCP-specific optimizers (GitHub, Figma, Jira, …)

### copy-features.md — new section (insert after Hook-Based Integration)

**Heading:** Tool firewall before execution  
**Body:** Before Claude runs a Bash command or MCP call, Confire evaluates it against built-in rules — force pushes, recursive deletes, secret patterns, and more. Balanced mode warns; strict mode blocks. Included on every plan.

---

## Sweep 7: Final QA checklist

- [ ] All v1 host references consistent (Claude Code, Cursor, VS Code)
- [ ] Free tier explicitly includes security
- [ ] Developer tier explicitly includes custom rules + MCP optimizers
- [ ] No placeholder pricing
- [ ] Cline removed from user-facing v1 copy (can stay in "roadmap" if needed)
- [ ] Honest firewall disclaimer on security page
- [ ] CTAs say "no card required" on free paths

---

## Refresh vs rewrite matrix

| File | Action |
|------|--------|
| copy-homepage.md | Refresh (hero + social proof + security block) |
| copy-features.md | Refresh (add firewall section, fix agents) |
| copy-pricing.md | Rewrite plan table + FAQ |
| copy-ctas.md | Refresh (host names + security CTA variants) |
| copy-onboarding.md | Refresh (agent picker, post-setup message) |
| copy-blog-posts.md | Refresh per post when published |
| content-strategy.md | Superseded by `docs/marketing/content-strategy.md` for strategy |

---

## Next step

Apply P0 edits to `docs/copy/` and platform Astro pages, then re-run this audit before launch.

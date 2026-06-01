# Marketing Plan — Confire (12-Month)

**Skill:** marketing-plan  
**Stage:** Bootstrapped / pre-seed — $0–2K/mo marketing spend  
**Growth phase:** $0–10K ARR (grueling) → path to $10K–100K  
**Deliverable:** Notion-paste-ready AARRR plan

---

## 1. Executive summary

**Three big bets (12 months):**
1. **Own Claude Code + Cursor + VS Code SEO** — become the default answer for "reduce agent context" and "AI coding agent firewall"
2. **Product-led proof** — every user sees `confire stats` savings; free security removes signup friction
3. **Community-led WOM** — GitHub Discussions + stats screenshots as referral engine

**90-day priorities:**
1. Ship refreshed site + docs + 4 blog posts
2. Launch on Product Hunt + HN
3. Hit 500 activated installs (completed `confire setup` + 1 optimized session)

**12-month outcome target:**
- 2,000+ free accounts
- 150+ Developer subscribers (~$1,500 MRR)
- Top-3 organic presence for 3 core keywords

---

## 2. Strategic frame

| Element | Definition |
|---------|------------|
| **Category** | AI agent context firewall |
| **ICP** | Daily Claude Code / Cursor / VS Code users |
| **Business model** | Freemium → Developer $10/mo → Team (future) |
| **Voice** | Direct, numeric, skeptical-engineer-friendly |
| **Non-negotiables** | Security on free; no context spam; honest limits |

---

## 3. Current state

**Scored 0–5 vs 17-section rubric (from materials):**

| Section | Score | Notes |
|---------|:-----:|-------|
| Brand & positioning | 3 | Strong draft copy; security underplayed |
| Website | 3 | Platform exists; pages need refresh |
| Content / SEO | 2 | Strategy written; few live posts |
| Paid acquisition | 0 | Not started |
| Email / lifecycle | 1 | Templates in docs/copy; not wired |
| Product-led growth | 4 | `confire stats` is strong PLG hook |
| Community | 1 | Not launched |
| Analytics | 2 | Worker telemetry; funnel gaps |
| Pricing & packaging | 4 | plan-new.md clear |
| Launch motion | 2 | Launch doc ready; not executed |

**Binding constraint:** Single founder/small team bandwidth — sequence organic before paid.

**Open decisions:** Product Hunt date; first case study; Team tier GA timing.

---

## 4. Acquisition

**Goal:** Strangers → confire.dev visit → signup

| Channel | Now | 90-day | 12-month | Skill |
|---------|-----|--------|----------|-------|
| SEO / blog | Draft | 8 posts live | 24 posts + programmatic integration pages | content-strategy, ai-seo |
| HN / Reddit | — | Launch spike | Monthly shareable posts | launch, community-marketing |
| Product Hunt | — | Launch day | Re-launch on Team tier | launch |
| DevRel / GitHub | Repo private/partial | Public docs + Discussions | Stars → install funnel | marketing-ideas #133 |
| Paid search | Skip | $300 test | Scale if CAC < $30 | ads (future) |
| Podcast / borrowed | Skip | 1 guest pitch | 3–4 appearances | launch |

**Skipped (with rationale):**
- LinkedIn ads — ICP is individual devs first, not enterprise (yet)
- International SEO — English-only v1

---

## 5. Activation

**Goal:** Signup → `confire setup` → first optimized session → sees savings

**Activation event:** User runs agent session with Confire hook; `confire stats` shows >0 optimized calls.

| Move | Owner | Skill |
|------|-------|-------|
| Reduce setup steps to one command | Eng | onboarding |
| One-time welcome (not spam) | Eng | onboarding |
| Post-setup email at +1h with troubleshooting | Marketing | emails |
| Docs quickstart < 5 min TTV | Eng | onboarding |
| Dashboard shows savings (Developer) | Eng | — |

**Target:** 60% signup → setup within 7 days; 40% → activation event within 14 days.

See [onboarding.md](./onboarding.md) for full flow design.

---

## 6. Retention

**Goal:** Weekly agent usage with Confire enabled

| Move | When |
|------|------|
| Monthly stats email ("You saved X tokens") | Month 2 |
| Changelog + new optimizer announcements | Biweekly |
| Firewall rule updates (silent) | Ongoing |
| Re-engage stalled: no stats in 14 days | Month 3 |
| Team waitlist nurture | Pre-Team launch |

**Churn signals:** `confire off` or uninstall; no API calls 30 days post-setup.

---

## 7. Referral

**Goal:** Users bring other devs

| Move | Timeline |
|------|----------|
| Stats screenshot culture in Discord | Launch |
| README "Optimized with Confire" badge | Month 2 |
| Referral credit on Team waitlist | Month 3 |
| Advocate program (10+ referrals) | Month 6 |
| Affiliate for DevTools creators | Month 9 |

**Skip until 500 activated users:** Formal affiliate payouts.

---

## 8. Revenue

**Packaging (current):**

| Plan | Price | Upgrade trigger |
|------|-------|-----------------|
| Free | $0 | 500 req exhausted; wants MCP optimizers |
| Developer | $10/mo | Daily use; GitHub/Figma pain |
| Team | $20/seat (soon) | Long sessions; multi-engineer |

**Monetization moves:**
- In-product upgrade at 80% quota (CLI stderr — already exists)
- Pricing page comparison: free security vs paid custom rules
- Top-ups $5/5K req for heavy Developer users
- Team waitlist with founding pricing ($15/seat grandfather)

**Psychology:** Anchor Developer against Claude API spend saved — see marketing-psychology.md.

---

## 9. 90-day roadmap

| Weeks | Theme | Actions |
|-------|-------|---------|
| 1–2 | Unblock | P0 copy fixes; docs public; analytics |
| 3–4 | Foundation | 2 blog posts; security features page |
| 5–8 | Velocity | Launch PH/HN; Discord; 2 more posts |
| 9–12 | Compound | Comparison pages; onboarding emails; first paid test |

---

## 10. 12-month outlook

| Quarter | Milestone |
|---------|-----------|
| Q2 2026 | GA launch; 500 activated users |
| Q3 2026 | 100 Developer subs; Team waitlist 500+ |
| Q4 2026 | Team beta; programmatic SEO pages |
| Q1 2027 | Team GA; case studies; optional seed raise |

**Funding unlock:** If seed closes ($5–15K/mo marketing), add paid search + part-time content contractor.

---

## 11. Marketing operations stack

| AARRR | Skills | Tools |
|-------|--------|-------|
| Acquisition | content-strategy, ai-seo, launch, marketing-ideas | Astro blog, GSC, Plausible/GA |
| Activation | onboarding, copy-editing | Resend/email, Supabase auth |
| Retention | emails, community-marketing | Customer.io or Resend sequences |
| Referral | community-marketing, referrals | Discord, referral links |
| Revenue | marketing-psychology, pricing | Stripe, in-CLI upgrade prompts |

---

## 12. Tactical idea bank (sample)

| # | Idea | AARRR | Status |
|---|------|-------|--------|
| 1 | Easy keyword SEO | Acquisition | **Now** |
| 11 | Comparison pages | Acquisition | Q2 |
| 15 | Stats as free tool | Acquisition | **Now** |
| 38 | Reddit | Acquisition | **Now** |
| 47 | Founder welcome email | Activation | **Now** |
| 78 | Product Hunt | Acquisition | Launch |
| 79 | Team waitlist referrals | Referral | Q3 |
| 93 | Stats viral loop | Referral | Q2 |

Full 139: see [marketing-ideas.md](./marketing-ideas.md).

---

## 13. Measurement, RACI, open decisions

**North star:** Weekly activated users (setup + ≥1 optimization/session)

**Leading indicators:**

| Stage | Metric |
|-------|--------|
| Acquisition | Site visits, signup rate |
| Activation | Setup completion %, time-to-first-stats |
| Retention | WAU with hook enabled |
| Referral | Invites / badge clicks |
| Revenue | Free→Developer conversion; MRR |

**RACI (solo founder default):**

| Area | R | A | C | I |
|------|---|---|---|---|
| Content | Founder | Founder | Community | — |
| Product | Eng | Founder | — | Users |
| Launch | Founder | Founder | Friends | Community |

**Open decisions:**
- [ ] CAC unknown — run $300 ad test in Q2
- [ ] Product Hunt date
- [ ] Email provider (Resend vs Customer.io)
- [ ] Public GitHub repo scope

---

*Plan version 1.0 — June 2026. Revise as v2 after launch retrospective.*

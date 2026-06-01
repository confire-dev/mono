# Launch Strategy — Confire v1 GA

**Skill:** launch  
**Launch target:** General availability — Free + Dev tiers  
**v1 scope:** Claude Code (full hooks), Cursor + VS Code (MCP gateway); honest client-mode footnote on all surfaces

---

## What we're launching

| Component | Status |
|-----------|--------|
| CLI (Homebrew + direct install) | ✅ |
| Cloud optimizer + local fallback | ✅ |
| Built-in firewall (free) | ✅ |
| Dev tier ($10/mo or $90/yr) | ✅ |
| Custom guardrails + policy sync | ✅ Dev only |
| Team tier | 🚧 Coming soon card only |

**Launch type:** Full GA for Free + Dev — not a limited beta.

---

## ORB channel strategy

### Owned (invest here first)

| Channel | Action | Goal |
|---------|--------|------|
| **Docs** | Quickstart + firewall reference public | Activation |
| **Blog** | 2 posts live at launch (hooks + cost) | SEO + credibility |
| **Email list** | Capture on pricing waitlist + blog | Owned audience |
| **GitHub** | Public repo, issues, Discussions | Community + trust |
| **Product** | `confire stats` as shareable proof | Retention + WOM |

### Rented (drive to owned)

| Channel | Tactic | CTA |
|---------|--------|-----|
| **X/Twitter** | Benchmark thread + setup GIF | confire.dev/docs |
| **LinkedIn** | "AI agent infra" angle for eng leads | Blog post |
| **Reddit** | r/ClaudeAI, r/cursor — value-first posts | Docs, not hard sell |
| **HN** | Show HN: "Confire – context firewall for Claude Code" | Homepage |

### Borrowed

| Channel | Target | Pitch |
|---------|--------|-------|
| AI/dev podcasts | Latent Space, dev tooling shows | "Cut agent context 40–95%" |
| Newsletter swaps | DevTools newsletters | Guest post on hooks |
| YouTube | Claude Code power users | Demo + affiliate later |

---

## Launch timeline (4 weeks)

### Week −2: Pre-launch

- [ ] Landing page copy refresh (see copy-editing.md)
- [ ] `/features/security` page live
- [ ] Analytics on signup + `confire login` completion
- [ ] Launch assets: 3 screenshots, 1 demo GIF, 1 benchmark chart
- [ ] Product Hunt draft (tagline, gallery, maker comment)
- [ ] HN Show HN post drafted

### Week −1: Soft launch

- [ ] Email 50 beta/friend users — "GA next week"
- [ ] GitHub Discussions + Discord open
- [ ] Fix P0 bugs from soft traffic
- [ ] `confire setup` tested clean on all 3 hosts

### Week 0: Launch week

**Day 1 (Tuesday recommended for PH):**
- [ ] Product Hunt live (if using)
- [ ] Show HN posted
- [ ] Blog announcement published
- [ ] X thread + LinkedIn post
- [ ] Email list announcement

**Day 2–7:**
- [ ] Reply to every PH/HN comment
- [ ] Monitor setup failures in support channels
- [ ] Ship hotfix if hook regression
- [ ] Collect stats screenshots from early users

### Week +1: Post-launch

- [ ] Onboarding email sequence active (see onboarding.md)
- [ ] Roundup: "What we learned from launch week"
- [ ] Comparison page draft: vs DIY hooks
- [ ] Plan next feature launch (Team waitlist push)

---

## Product Hunt checklist

**Tagline (≤60 chars):** Context firewall for Claude Code, Cursor & VS Code

**Description lead:**
> Confire reviews risky tool calls and sanitizes noisy output before it hits your context window. Claude Code: full hooks. Cursor & VS Code: MCP gateway. Start free on all supported clients.

**Gallery:**
1. Hero — stats screenshot
2. Hook diagram (PreToolUse + PostToolUse)
3. Pricing table
4. Firewall rules in action
5. Demo GIF — `confire setup` → session → stats

**Maker comment:** Personal story — why context noise + agent safety, benchmark numbers, free tier includes firewall.

---

## Launch messaging matrix

| Audience | Lead message | Proof |
|----------|--------------|-------|
| Individual dev | "Try the firewall across your editor" | Free card copy |
| Daily driver | "Custom rules + 5,000 remote/mo" | Dev card copy |
| Team lead | "Team — coming soon" | Waitlist |

---

## Post-launch momentum (ongoing launches)

| Update size | Channels |
|-------------|----------|
| **Major** (Team tier, new host) | Blog + email + PH + social |
| **Medium** (new optimizer, VS Code parity) | Changelog + X + in-app |
| **Minor** (bugfix, rule tweak) | Changelog only |

**Rule:** Ship visible improvements every 2–3 weeks minimum to sustain narrative.

---

## Launch checklist (condensed)

### Pre-launch
- [x] CLI install paths (Homebrew, direct)
- [ ] Landing + pricing + features + docs live
- [ ] Email capture
- [ ] Launch assets
- [ ] PH listing prepared
- [ ] Onboarding flow tested end-to-end

### Launch day
- [ ] PH + HN + blog + social
- [ ] Team on support duty 8+ hours
- [ ] Status page clean

### Post-launch
- [ ] Welcome email sequence
- [ ] Feedback → GitHub issues
- [ ] Week-1 retrospective doc

---

## Risks

| Risk | Mitigation |
|------|------------|
| Hook breaks on agent update | Host capability tests; fast patch release |
| "Security" overpromise | Honest copy: best-effort guardrails |
| Free tier abuse | Rate limits already in worker |
| Low PH ranking | HN + Reddit parallel; don't bet everything on PH |

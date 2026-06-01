# Community Marketing — Confire

**Skill:** community-marketing  
**Stage:** Pre-launch → early community (0–500 members)  
**Primary goal:** Word-of-mouth + product feedback + support deflection

---

## Community identity

**Who members are:** Developers who treat AI coding agents as daily infrastructure — not toys. They care about token economics, agent safety, and shipping faster without burning context.

**Identity statement:** *"Engineers who refuse to pay for noise in their context window."*

**Not:** A generic "AI enthusiasts" group. Specificity drives retention.

---

## Platform recommendation

| Platform | Verdict | Rationale |
|----------|---------|-----------|
| **GitHub Discussions** | ✅ Primary | Matches ICP; ties to open issues; SEO benefit |
| **Discord** | ✅ Secondary | Real-time help; Claude/Cursor dev culture |
| **Slack** | ⏸ Later | When Team tier ships for org invites |
| **Reddit** | ✅ Outpost | r/ConfireDev or participate in r/ClaudeAI — don't own yet |
| **Circle / Forum** | ❌ Skip | Too much overhead pre-PMF |

---

## 90-day launch plan

### Phase 1: Founding members (Weeks 1–4)

1. **Recruit 30–50 manually** — DM beta users, GitHub stargazers, friends using Claude Code daily
2. **Open GitHub Discussions** with categories: `Show & Tell`, `Help`, `Feature Ideas`, `Optimizer Requests`
3. **Seed 10 posts** before announcing publicly (stats screenshots, setup tips, firewall rule explanations)
4. **Discord:** Single server, channels: `#welcome`, `#setup-help`, `#stats-wins`, `#feature-requests`
5. **Weekly ritual:** "What did Confire save you this week?" thread (stats screenshots encouraged)

### Phase 2: Public growth (Weeks 5–8)

1. Pin **New Member Journey** in Discord + Discussions README
2. Link community from docs footer, post-setup CLI message (one-time welcome only)
3. **Surface wins:** Retweet/quote user stats posts; add to `/customers` when available
4. **Monthly AMA** with founder — firewall roadmap, new optimizers, Team tier preview

### Phase 3: Advocate program (Weeks 9–12)

1. Identify users who share unprompted (GitHub stars + social mentions)
2. Launch **Confire Advocates** — referral link, early Team pricing, swag for top referrers
3. **Community Expert badge** in Discord for members who answer setup questions

---

## Channel architecture (Discord)

| Channel | Purpose | Posting guidelines |
|---------|---------|-------------------|
| `#welcome` | Rules + start here | Read-only; bot posts intros |
| `#introduce-yourself` | New member intros | Agent used, biggest context pain |
| `#setup-help` | Install/debug | Search before asking; paste `confire status` |
| `#stats-wins` | Savings screenshots | Require `confire stats` output |
| `#firewall` | Security rules discussion | Built-in vs custom (paid) clarity |
| `#feature-requests` | Product input | Upvote via reactions |
| `#announcements` | Ship notes | Team-only post |

---

## New member journey

**Pinned welcome (Discord + Discussions):**

> Welcome to Confire. Three things to do today:
> 1. Run `confire setup` and pick your agent (Claude Code, Cursor, or VS Code)
> 2. Run a session, then `confire stats` — post your savings in #stats-wins
> 3. Read the firewall docs if you care about pre-tool safety (free on all plans)

**DM template (founding members, manual):**

> Hey [name] — saw you're using Confire. We're building a small community of devs optimizing agent context. Would love your stats screenshot or setup feedback in [link]. No pitch — just trying to learn what breaks.

---

## Community flywheel

```
Dev hits token pain → finds Confire → posts stats win
        ↑                                      ↓
   new devs discover via social/search ← shared screenshot
```

**Accelerators:**
- Make stats output shareable (pretty terminal or `confire stats --json` → badge generator — future)
- Credit community feedback in changelog ("Thanks @user for the Jira optimizer idea")

---

## Health metrics (track weekly)

| Metric | Target (Month 3) |
|--------|------------------|
| New member post rate (7 days) | >40% |
| Thread reply rate | >60% |
| Non-staff posts | >70% of volume |
| DAU/MAU (Discord) | >15% |

**Warning signs:**
- Only team posting
- Setup questions unanswered >24h
- Same 5 people = 80% of activity

---

## Ambassador program brief

**Criteria:** 3+ referrals OR public write-up OR 10+ community helps

**Benefits:**
- Founding Team pricing when shipped ($15/seat grandfathered)
- Early access to new optimizers
- `@Advocate` role + referral link with tracking

**Tools provided:**
- Referral URL via dashboard (when built)
- Shareable one-liner: "Confire cut my Claude Code context by 73% — free firewall included"
- Logo badge for README: "Optimized with Confire"

---

## Support deflection

1. Turn top 20 Discord/Discussions questions into docs FAQ
2. Bot auto-reply: "Have you run `confire status`?" with link to troubleshooting
3. When community answer becomes a doc fix, announce and credit the member

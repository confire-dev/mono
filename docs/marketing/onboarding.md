# Onboarding Strategy — Confire

**Skill:** onboarding  
**Activation definition:** User completes `confire setup`, runs ≥1 agent session with hooks active, and sees non-zero savings in `confire stats`.

---

## Funnel (target)

```
Signup → confire login/setup → first hooked session → confire stats → (Developer upgrade)
100%      70%                   50%                    40%              5–10%
```

**Focus:** Biggest drop expected at signup → setup (activation energy). Minimize steps.

---

## Aha moment

**The moment:** User runs `confire stats` and sees concrete byte/token reduction after a normal coding session.

**Supporting aha:** First PreToolUse warn/block — "Confire just stopped something risky" (security aha, free tier).

**Correlate retention with:**
- Setup within 24h of signup
- ≥10 optimized calls in first week
- `confire stats` run at least once

---

## Flow design

### Step 0: Signup (web)

| Element | Recommendation |
|---------|----------------|
| Approach | Value-first — show what free includes |
| Fields | GitHub OAuth preferred; email fallback |
| Copy | "Free firewall + 500 optimizations/mo. No card." |
| CTA | `Create free account` |

### Step 1: Post-signup (immediate)

**Product-first** — don't ask preferences before value.

```
Account created.

Run this in your terminal:
  brew install confire   # if needed
  confire login
  confire setup

Pick Claude Code, Cursor, or VS Code — takes ~30 seconds.
```

**Single next action.** No dashboard tour yet.

### Step 2: `confire setup` (CLI)

| Step | UX |
|------|-----|
| Detect agent | Auto-select if settings found |
| Install hook | Show exactly what file changed |
| Start daemon | `confire start` if not running |
| Confirm | "You're set. Start a coding session." |

**Avoid:** Multi-page wizard, account verification delays.

### Step 3: First session (agent)

- **One-time welcome** in Claude Code only when `welcome_pending` — not every session
- Healthy sessions: **silent** (no context spam)
- User runs normal work — tests, reads, fetches

### Step 4: Activation nudge

After first session (or +30 min), stderr hint:

```
Run `confire stats` to see what Confire saved this session.
```

Optional email +1h with same CTA if no stats run yet.

### Step 5: Developer upgrade (optional)

Trigger at 80% quota OR when stats show high MCP noise with "available on Developer" message.

---

## Onboarding checklist (in-app / email)

**5 items — order by value:**

1. ☐ Create account
2. ☐ Run `confire setup`
3. ☐ Complete one agent session
4. ☐ Run `confire stats`
5. ☐ Star on GitHub / join Discord (optional)

Celebrate completion: "You're saving context. Here's what to try next: strict mode, `confire help`."

---

## Empty states

| Surface | Empty state copy |
|---------|------------------|
| Dashboard (pre-setup) | "No data yet. Run `confire setup` to connect your agent." |
| Stats (zero calls) | "No optimizations yet. Use your agent — Confire runs automatically." |
| Custom rules (free) | "Built-in rules active. Custom rules available on Developer." |

---

## Email sequence (trigger-based)

| Trigger | Timing | Subject | CTA |
|---------|--------|---------|-----|
| Welcome | Immediate | Welcome to Confire | `confire setup` docs link |
| No setup | +24h | Quick setup for Claude Code / Cursor | Troubleshooting |
| No setup | +72h | Need help connecting? | Reply / Discord |
| Activation | First stats | You saved X bytes | Share stats / upgrade |
| Stalled | +14d inactive | Your firewall is still ready | `confire status` |
| Limit warning | 80% quota | Running low on optimizations | Pricing |

**Rule:** Email reinforces in-app actions — doesn't duplicate docs.

---

## Stalled user recovery

**Stalled =** signed up, no API activity 7 days.

1. Email: address top blockers (daemon not running, wrong agent, hook not saved)
2. Include: `confire status` output to paste in Discord #setup-help
3. High-value accounts: manual DM offer to pair setup (founding users)

---

## Multi-host specifics

| Host | Setup friction | Mitigation |
|------|----------------|------------|
| Claude Code | Low — settings.json hook | Default path in docs |
| Cursor | Medium — hook format | Host-specific doc section |
| VS Code | Medium — Copilot hooks | Parity tested before marketing |

---

## Metrics plan

| Metric | Target (90d) |
|--------|--------------|
| Signup → setup (7d) | 60% |
| Setup → activation (14d) | 40% |
| Time to activation | <24h median |
| Day 7 retention | 30% |
| Free → Developer (30d) | 5% |

**Instrument:** Supabase events: `signup`, `api_key_created`, `first_optimization`, `stats_viewed`, `upgrade_clicked`.

---

## Audit findings → fixes

| Finding | Impact | Recommendation | Priority |
|---------|--------|----------------|----------|
| Cline in setup picker | Confusion | Remove for v1 | P0 |
| SessionStart spam (fixed) | Trust erosion | Silent healthy sessions | Done |
| No post-signup email | Drop at setup | Wire welcome sequence | P1 |
| Upgrade before value | Resentment | Only nudge after stats | P1 |

---

## Experiments

1. **Setup video vs text** — 60s Loom in welcome email
2. **GitHub-only signup** — measure completion vs email
3. **Stats in setup completion** — run sample optimization demo
4. **Security-first vs cost-first** onboarding copy A/B

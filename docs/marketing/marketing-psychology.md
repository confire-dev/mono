# Marketing Psychology — Confire

**Skill:** marketing-psychology  
**Application:** Homepage, pricing, CLI upgrade prompts, onboarding

---

## Core behavior we want

| Stage | Desired behavior |
|-------|------------------|
| Awareness | "I have a context cost / safety problem" |
| Consideration | "Confire might solve this better than DIY" |
| Signup | Create free account (no card) |
| Activation | Run `confire setup` and use agent |
| Upgrade | Dev when Confire is daily workflow (quota, custom rules, history) |

---

## Principles applied to Confire

### Loss aversion (pricing + CLI)

**Insight:** People feel losses ~2× more than equivalent gains.

**Application:**
- CLI at limit: *"Confire paused — you're receiving full-size tool output until reset"* (loss frame) vs *"Upgrade for more"* (gain frame). Lead with loss.
- Stats view: *"You would have spent ~$X extra on tokens this month without Confire"* before upgrade CTA.

### Anchoring (pricing page)

**Insight:** First number seen sets reference point.

**Application:**
- Anchor against **Claude API spend**, not Confire price: "Save $40–200/mo in tokens → Confire Dev $10/mo"
- Show **Team** (coming soon) to make Dev feel accessible
- Order: Free → Dev → Team (ascending)

### Social proof (homepage)

**Insight:** Uncertainty reduces action; others' behavior guides decisions.

**Application:**
- "Works with Claude Code, Cursor, VS Code" — host logos
- Stats screenshots from real sessions (with permission)
- GitHub star count when public
- **Avoid:** fake testimonials

### Reciprocity (free tier)

**Insight:** Give first → obligation to try / upgrade later.

**Application:**
- **Free is a real try-out** — full Claude Code firewall + Cursor/VS Code gateway + built-in rules + noise trimming
- 500 remote optimizations/mo is usable, not a 7-day trial
- `confire stats` shows savings before asking for Dev

### Endowment effect (post-setup)

**Insight:** People value what they already have.

**Application:**
- After setup, user "owns" their optimized workflow — `confire off` feels like loss
- Custom guardrails on Dev + policy sync increase switching cost (ethical)

### Activation energy (onboarding)

**Insight:** Starting friction kills conversion even if product is easy overall.

**Application:**
- Single command: `confire setup`
- Pre-select detected agent (Claude Code if `.claude` exists)
- GitHub OAuth signup — no password friction

### Goal-gradient (onboarding checklist)

**Insight:** Motivation increases near completion.

**Application:**
- Setup checklist: Account ✓ → Setup ✓ → First session → Run stats (4 steps)
- Progress bar in dashboard (future)

### Paradox of choice (pricing)

**Insight:** Too many options paralyze.

**Application:**
- v1: Free + Dev + Team (coming soon) — only two purchasable tiers
- Don't expose Enterprise on self-serve pricing

### Authority (content)

**Insight:** Expertise signals increase trust.

**Application:**
- Publish reproducible benchmarks with commands
- Link to firewall-reference docs
- Named author on technical posts

### Scarcity (ethical use only)

**Insight:** Limited availability increases urgency.

**Application:**
- Team founding pricing: "First 100 teams lock $15/seat" — real limit only if enforced
- **Avoid:** fake countdown timers

### Framing (security messaging)

**Insight:** Same facts, different frames → different reactions.

| Frame | Copy |
|-------|------|
| Gain | "Keep your agent safe" |
| Loss | "Stop accidental force pushes and secret leaks before they run" |
| **Use loss frame for security** — devs respond to disaster prevention |

### Status-quo bias (upgrade)

**Insight:** Default is to do nothing.

**Application:**
- Confire enabled by default after setup — opt-out (`confire off`) not opt-in
- Balanced mode default — strict available but not forced

### IKEA effect (custom rules)

**Insight:** People value what they helped create.

**Application:**
- Developer custom rules from dashboard — user-configured policies feel "mine"
- Community-submitted rule ideas credited in changelog

---

## Page-specific recommendations

### Homepage

| Element | Psychology | Implementation |
|---------|------------|----------------|
| Hero | Loss + clarity | Lead with cost OR safety pain, not feature list |
| Stats block | Authority + anchoring | Specific % not "up to" alone |
| CTA | Low activation energy | "Get started free — no card" |
| Social proof | Bandwagon | Host logos + one stat screenshot |

### Pricing

| Element | Psychology |
|---------|------------|
| Free includes real firewall + gateway | Reciprocity — try before daily commit |
| Dev custom rules + continuous updates | Contrast — daily workflow, not crippled free |
| FAQ on limits | Loss aversion at quota |

### CLI upgrade prompt

```
⚠️ 412/500 optimizations used this month.
    Custom guardrail rules on Dev.
    Upgrade: confire.dev/pricing
```

Combines: loss aversion (quota), specific anchor (87%), clear action.

---

## Experiments to run

| Hypothesis | Test | Metric |
|------------|------|--------|
| Loss-framed CLI limit beats gain-framed | A/B stderr copy | Upgrade click rate |
| Security in hero increases signup | Hero A/B | Signup rate |
| Founder email increases setup | Welcome A/B | Setup completion |
| Benchmark post increases Developer conv | Before/after publish | Free→Dev rate |

---

## Models quick reference for Confire

| Challenge | Model |
|-----------|-------|
| Low signup | Activation energy, friction |
| Won't upgrade | Loss aversion, anchoring vs API spend |
| Don't trust security claims | Authority, pratfall ("best-effort, not guarantee") |
| Churn after setup | Endowment, goal-gradient emails |
| Pricing confusion | Paradox of choice — simplify tiers |

# Confire — Pricing & Feature Specification

Internal reference document. 
Defines all plans (active + future), feature sets, 
and upgrade triggers.

Last updated: June 2026
Status: Free + Developer launching now. Team + Enterprise: roadmap.

---

## Pricing Overview

| Plan        | Monthly  | Annual   | Status      |
|-------------|----------|----------|-------------|
| Free        | $0       | —        | ✅ Active   |
| Developer   | $10      | $95      | ✅ Active   |
| Team        | $20/seat | $195/seat| 🚧 Soon     |
| Enterprise  | Custom   | Custom   | 🚧 Roadmap  |

Annual savings: ~20% (presented as "2 months free")

Top-up packs: $5 = 5,000 additional requests
(available on Developer, never expire)

---

## Free — $0/month
**Status: Active**

For developers trying Confire on personal projects
or learning AI-assisted workflows. Real value, real limits.

### What's included

**Platform support**
- Claude Code hook
- Cursor support
- VS Code support
- Local CLI via Homebrew
- Cross-platform: macOS, Linux, Windows (WSL)

**Local optimizers (always offline)**
- Bash output compression (97% on test output)
- File read deduplication (up to 97% on repeats)
- WebFetch content extraction (80-90%)
- Generic universal fallback (10-30% on any tool)

**Remote optimization**
- 500 requests per month
- Generic remote optimizer only
- Local fallback when limit reached

**Stats and visibility**
- `confire stats` command
- Session-level savings
- Today / month / all-time views
- Tool breakdown

**Privacy**
- Anonymous opt-in telemetry
- Never stores response content
- Local stats own forever

**Support**
- Community support (GitHub issues)
- Public documentation

### What's not included
- Tool-specific MCP optimizers (GitHub, Figma, Jira, etc.)
- Confire API access
- Dashboard
- Pre-compact intelligence
- Session memory features
- Top-up purchases (Developer required)

### The experience
Sessions feel snappier. Tests output cleanly.
Stats show real dollar savings. The agent stays
coherent longer before compaction.

### Upgrade trigger
500 requests runs out in 3-7 days of active use.
Stats display shows the gap:
> "GitHub MCP: 87% reduction available on Developer"
> "Figma: 98% reduction available on Developer"

---

## Developer — $10/month or $95/year
**Status: Active. Launch tier.**

For individual developers using Confire daily across
their full workflow. The workhorse tier.

### Everything in Free, plus:

### Remote optimization (5,000 req/month)

**All 13 tool-specific optimizers:**

| Tool                  | Reduction | Noise stripped              |
|-----------------------|-----------|------------------------------|
| Figma sparse          | 98%       | Raw SVG nodes → section map  |
| Confluence (ADF)      | 96%       | ADF wire format → markdown   |
| Google Drive          | 91%       | Content snippets auto-embed  |
| GitHub PR             | 87%       | Bot comments                 |
| Jira (ADF)            | 85%       | ADF JSON + avatar URLs       |
| Amplitude cohort      | 76%       | Duplicate clause JSON        |
| ClickUp               | 64-76%    | type_config dropdowns        |
| Zapier                | 54%       | Verbose descriptions         |
| Amplitude context     | 46%       | 17 null user fields          |
| Slack                 | 37%       | User IDs + timestamps        |
| Figma JSX             | 28-35%    | data-node-id + CSS vars      |
| Fireflies             | routing   | Smart tool selection         |
| Notion                | 4-76%     | Scales with noise ratio      |

**Plus:**
- Larger input payloads accepted (up to 1MB)
- Auto-updating rules when APIs change
- New tool optimizers added regularly

### Confire API access (1M tokens/month)

Use Confire optimizers in your own applications,
not just AI coding agents.

**Endpoint:** `POST https://api.confire.dev/v1/optimize`

**Use cases:**
- CV/resume builders
- Document processing apps
- Customer support classifiers
- RAG pipeline optimization
- Email parsing
- Legal document analysis

**Compatibility:** Works before any LLM provider
(OpenAI, Anthropic, Groq, Cohere, etc.)

**Overage:** $0.50/1M tokens

### Top-up packs

When you need more than 5,000 requests/month:

- $5 = 5,000 additional requests
- Top-ups never expire
- Stack with monthly allowance
- One command: `confire top-up`

### Stats and dashboard

- 30-day optimization history
- Per-tool breakdown
- Web dashboard at confire.dev/dashboard
- Cost projection: "you saved $X this month"
- Export data (CSV, JSON)

### Support

- Email support (48h response)
- Priority on bug reports
- Optimizer requests considered for roadmap

### The experience
Jira tickets arrive readable, not as ADF JSON.
Figma returns clean section maps with node IDs.
GitHub PRs surface human comments, not bot noise.
The agent reads signal instead of noise — and
suddenly seems much smarter on the same model.

### Upgrade triggers to Team
- Hitting compaction in long sessions
- Wanting session continuity across days
- Need PreCompact intelligence
- Multi-engineer team adoption
- Want session memory features

---

## Team — $20/seat/month or $195/seat/year
**Status: Coming soon. Email signup on pricing page.**

For engineering teams using Confire across multiple
developers. Adds session intelligence and team
management features.

### Planned features (not yet shipped)

**Everything in Developer per seat, plus:**

### Unlimited fair-use optimization
- No monthly request cap
- Soft cap: 100k requests/month per seat
- Per-minute rate limit: 1,000 req/min
- Top-ups still available for outliers

### Session Memory Guard
- Maintains session continuity beyond compaction
- Structured signal storage across sessions
- "Resume yesterday's session" works correctly
- Cross-tool memory (Jira + Figma + GitHub context unified)

### PreCompact Context Optimizer
- Confire takes over compaction from Claude Code
- Structured compaction preserves decisions, paths, errors
- Sessions run 2-3x longer before quality degradation
- No more mid-task "lobotomization"

### Local memory packs
- Download tool-specific memory packs for offline work
- Optimizers + recent rule updates bundled
- Work without network — no quality drop
- Sync when back online

### Team management
- Centralized billing
- Per-engineer usage tracking
- Team dashboard with cross-engineer stats
- Tool allowlist/blocklist enforcement
- Centralized configuration management

### Extended API
- 10M tokens/month per seat
- Reduced overage: $0.40/1M tokens
- Webhook callbacks for async optimization
- Custom optimizer rules

### Stats and history
- 90-day optimization history
- Team-wide aggregated views
- Per-engineer drill-down
- Cost attribution by team/project

### Support
- Priority email support (24h response)
- Shared Slack channel access
- Quarterly check-ins

### Launch plan
Email capture on pricing page:
"Notify me when Team launches"
+ optional team size field

Existing Developer users get founding pricing
when Team ships ($15/seat/month grandfathered).

---

## Enterprise — Custom
**Status: Roadmap. Contact sales.**

For engineering organizations running Confire
across teams with security, compliance, or scale needs.

### Planned features

**Everything in Team, plus:**

### Self-hosted deployment
- Deploy on your own infrastructure
- Supported targets:
  - Cloudflare Workers (your account)
  - AWS Lambda
  - Google Cloud Run
  - Docker containers (Kubernetes, Akamai, bare metal)
- Data never leaves your network
- Source code provided for security review

### Local-only optimizer mode
- All processing on-premise
- No external network calls required
- Optimizer rules bundled at deploy time
- Updates pushed via your CI/CD

### Team policy controls
- Per-team request limits
- Tool allowlist/blocklist enforcement
- Centralized configuration management
- Role-based access control

### Custom optimizers
- Build optimizers for internal MCP tools
- Submit company-specific optimizers
- Confire team helps build and maintain
- Optimizers stay private to your org

### Compliance and audit
- Audit logging for all optimization activity
- Configurable log retention
- Data residency controls
- SOC2 Type II compliance documentation
- DPA (Data Processing Agreement) available

### Identity and access
- SSO/SAML support (roadmap, Q3 2026)
- SCIM provisioning (roadmap)
- API key rotation policies
- IP allowlisting

### Support and SLA
- Dedicated Slack channel
- 99.9% uptime SLA (if using hosted)
- Named customer success contact
- Quarterly business reviews
- Custom training for engineering teams

### Pricing model
- Annual contracts (1-year minimum)
- Self-host: starts at $499/year base + per-seat
- Hosted: starts at $2,500/year base + per-seat
- Custom optimizer development billed separately

### The experience
Confire runs entirely on internal infrastructure.
Engineers use it daily without thinking.
Security signs off in one review cycle.
Token costs across the org drop 30-50%.
AI-assisted development becomes infrastructure
the team doesn't have to think about.

---

## Upgrade Path Logic

### Free → Developer ($10/mo)
**Triggered by:** 500 request limit hit  
**Felt improvement:** All 13 specific optimizers
transform Jira/Figma/GitHub experience. Same agent
suddenly feels much smarter because it reads signal
not noise. API access unlocks new use cases.

### Developer → Top-up ($5 packs)
**Triggered by:** 5,000 request limit hit  
**Felt improvement:** Continue working without
interruption. Never expires. Cheaper per-request
than base subscription.

### Developer → Team ($20/seat)
**Triggered by:** Long sessions, team adoption,
3+ top-ups per month  
**Felt improvement:** Sessions run 2-3x longer.
PreCompact replaces lossy default. Memory Guard
makes "tomorrow's session" actually continue
today's work. Team-wide visibility and control.

### Team → Enterprise (Custom)
**Triggered by:** Security review, self-host needs,
50+ engineers, custom internal tools  
**Felt improvement:** Self-hosted = SOC2 compliant.
Custom optimizers for internal MCPs. Org-wide
deployment with policy controls.

---

## Top-up Economics

**Why top-ups beat metered overage:**

| Metered overage          | Top-up packs              |
|--------------------------|---------------------------|
| Surprise bill anxiety    | Conscious purchase        |
| Auto-charges credit card | One-click checkout        |
| Refund disputes          | Clean transactions        |
| Hard to budget           | Predictable spend         |
| Same revenue             | Same revenue              |
| Lower retention          | Higher retention          |

**Top-up margin:**
- Cost per request: ~$0.000001 (Cloudflare Worker)
- Top-up price per request: $0.001
- Margin: 99.9%

**Conversion path:**

0 top-ups/mo  → happy Developer user
1-2 top-ups   → power user, still cheaper than Team
3+ top-ups    → "Team plan = unlimited" prompt fires
→ email captures Team waitlist

---

## API Overage Pricing

Developer tier includes 1M optimized tokens/month.

**Overage:** $0.50 per 1M tokens optimized

**Example economics:**

CV Builder app processing 1,000 CVs/month:
- Average CV input: 8,000 tokens
- Total: 8M tokens
- Confire cost: 1M free + 7M × $0.50/1M = $3.50
- OpenAI savings: ~$80-100/month (from optimized input)
- Net saving to developer: ~$77-96/month
- Net revenue to Confire: $3.50/month

**At scale:**
SaaS processing 100k documents/month:
- 800M tokens through Confire
- $0.50 × 800 = $400/month to Confire
- OpenAI savings: ~$8,000-10,000/month
- Customer wins 20x return on Confire spend

---

## Annual Pricing Strategy

Annual pricing for Developer: $95/year
- Monthly equivalent: $7.92
- Effective savings: 20.8%
- Presented as: "2 months free"

**Why annual matters:**
- Procurement-friendly for small teams
- Cash flow advantage (year of revenue upfront)
- Higher retention (annual subscribers churn ~50% less)
- Some companies require annual contracts

**Launch decision:**
Start with monthly only on Day 1.
Add annual toggle when 50+ active subscribers exist
and 3+ request annual billing.

---

## What This Pricing Communicates

### Free is real, not a trial
500 requests forever. Generic optimizer works on any
tool. Local install with no account friction. Not a
crippled demo.

### Developer is the workhorse
$10/month is impulse-buy territory. All 13 specific
optimizers = transformative jump from Free. API access
included = unlocks new use cases beyond coding agents.
This is the primary conversion target.

### Top-ups respect users
Heavy users don't get cut off. Top-ups never expire.
$5 feels like an add-on, not "buying another month."
Margins are high. Users feel in control.

### Team is for organizations (when ready)
$20/seat doubles per-user price but adds session
memory, unlimited fair-use, and team management.
Don't ship this until org features exist.

### Enterprise is for serious orgs
Self-hosted is the SOC2 path. Custom optimizers
create lock-in. Per-seat economics scale. SSO
positions for IT review. Don't undersell — higher
prices = serious vendor.

---

## Launch Configuration

**Day 1 pricing page shows:**

┌─────────────┬─────────────────┐
│    Free     │   Developer     │
│     $0      │      $10        │
│  /month     │     /month      │
│             │                 │
│  500 req/mo │  5,000 req/mo   │
│  Local      │  All optimizers │
│  Generic    │  API access     │
│             │  Dashboard      │
│             │  $5 top-ups     │
└─────────────┴─────────────────┘
──────────────────────────────────
Team plans coming soon
For organizations needing
multi-seat management, SSO,
and self-hosted deployment.
→ Get notified
──────────────────────────────────

**Email capture for Team:**
- Email field
- Optional: team size (1-10 / 10-50 / 50-500 / 500+)
- Optional: company name

**The email signup becomes your enterprise
pipeline before you write a single sales page.**

---

## Stripe Product Configuration

### Products to create

1. **Developer Monthly**
   - Recurring: $10.00/month
   - Trial: none
   - Includes: 5,000 requests/month

2. **Developer Annual**
   - Recurring: $95.00/year
   - Trial: none
   - Includes: 5,000 requests/month
   - (Add when launching annual)

3. **Top-up Pack**
   - One-time: $5.00
   - Includes: 5,000 additional requests
   - Stackable, never expires

### Future products

4. **Team Monthly** ($20/seat)
   - Recurring per quantity
   - Includes: unlimited fair-use

5. **Team Annual** ($195/seat)
   - Recurring per quantity

6. **Enterprise** - manual invoicing, no Stripe product needed

---

## Open Questions

1. **When to launch annual?**
   Recommendation: monthly only Day 1. Add annual
   when first 50 customers exist.

2. **When to launch Team?**
   Recommendation: ship when org-level management
   (session memory, team dashboard, centralized billing)
   exists. Estimate: 6-8 weeks post-launch.

3. **When to launch Enterprise sales?**
   Recommendation: reactive only at first. Let
   inbound from Team waitlist drive conversations.
   Build sales page when 5+ enterprise leads in pipeline.

4. **Should API have separate tier?**
   Recommendation: keep bundled. Most early API users
   are individual developers building side projects.
   They already fit Developer tier. Pure API customers
   at scale → Team tier with custom limits.

5. **Free tier abuse prevention?**
   Recommendation: rate limit per IP + email verification.
   500 req/mo limits ROI of abuse anyway. Don't
   over-engineer this on Day 1.

---

## Pricing Evolution Roadmap

**Month 1 (launch):**
- Free + Developer only
- Monthly billing only
- Top-ups available
- Team email capture

**Month 2-3:**
- Annual billing toggle
- Email Team waitlist with progress updates
- First Team beta with select customers

**Month 4-6:**
- Team plan launches
- Existing Developer users grandfathered to Team
- Enterprise sales page (if 5+ inbound leads)

**Month 6-12:**
- Enterprise self-hosted option
- SSO/SAML
- Custom optimizer development service
- Volume discounts for >100 seats



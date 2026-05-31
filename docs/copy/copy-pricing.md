# Pricing Page Copy: Confire

**URL:** confire.dev/pricing  
**Meta title:** `Confire Pricing — Free, Dev, and Pro Plans`  
**Meta description:** `Start free with 500 optimized calls/month. Upgrade to Dev or Pro as your AI agent usage grows. No surprise bills, no per-seat nonsense.`

---

## Hero

**H1:** Simple pricing. Built for how developers actually work.

**Subhead:**  
Confire pricing is based on optimized calls — one per tool output your agent processes. Most developers never leave the free tier. When you're ready to scale, it's a simple upgrade.

---

## Pricing Table

### Free
**$0/month**

- 500 optimized calls/month
- Cloud optimizer (Cloudflare Worker)
- Local fallback optimizer
- Claude Code, Cursor, Cline support
- Usage stats via `confire stats`

**CTA:** `Get started free`  
**Sub-CTA:** No credit card required.

---

### Dev
**$[price]/month**

Everything in Free, plus:
- [X,000] optimized calls/month
- Priority cloud optimizer (lower latency)
- Extended usage analytics
- Email support

**CTA:** `Start Dev plan`  
**Sub-CTA:** Cancel anytime.

---

### Pro
**$[price]/month**

Everything in Dev, plus:
- Unlimited optimized calls
- Per-tool-type configuration
- Team usage pooling (unlimited seats)
- Team dashboard
- SLA-backed uptime
- Priority support

**CTA:** `Get Pro`  
**Sub-CTA:** [Link to contact] — want a custom quote?

---

## FAQ Section

**Q: What counts as one "optimized call"?**  
Each tool output that Confire processes counts as one call. If your agent runs a Bash command, that's one call. If it fetches a URL, that's another. If it calls the GitHub API, that's another. Tool calls that result in empty output or errors may not count against your limit.

**Q: What happens when I hit my free tier limit?**  
Confire pauses optimization for the rest of the month. Your agent keeps working — it just receives unoptimized (full-size) tool outputs until your limit resets. You can upgrade at any time to resume optimization immediately.

**Q: Does the local optimizer count against my limit?**  
Yes. All optimized calls — whether processed by the cloud or local optimizer — count toward your monthly limit. The local optimizer is a fallback for availability, not a way to bypass limits.

**Q: Is there a free trial of Dev or Pro?**  
The Free plan is effectively a permanent free trial. You get real optimization on 500 calls/month with no time limit. Upgrade when the cap matters to you.

**Q: Can I change plans anytime?**  
Yes. Upgrade, downgrade, or cancel from your account settings or by emailing us. No runaround.

**Q: Do you offer annual billing?**  
[Yes — 2 months free / Not yet — coming soon.]

**Q: Is there a student or open source discount?**  
Email us. We'll work something out.

**Q: How do I know how many calls I'm using?**  
Run `confire stats` in your terminal at any time. You'll see your current month's call count, estimated tokens saved, and reset date.

**Q: What is a "team seat" on Pro?**  
Pro plans don't have per-seat limits. Your team shares a pooled call limit. Everyone logs in with their own account but contributes to and draws from the same pool. Usage by team member is visible in the team dashboard.

---

## Usage Estimator (Inline Tool)

**Heading:** Not sure which plan you need?

Estimate your monthly call volume:

```
Average Claude Code or Cursor sessions per week: [  ]
Average tool calls per session: [  ]

→ Estimated monthly calls: ~[X]

Recommendation: [Free / Dev / Pro]
```

**Typical profiles:**

| Developer type | Typical sessions/week | Est. calls/month | Plan |
|----------------|----------------------|------------------|------|
| Casual (side projects, occasional) | 3–5 | 150–400 | Free |
| Active (daily dev, full-time coding) | 10–20 | 800–2,000 | Dev |
| Heavy (agents in CI, team workflows) | 30+ | 3,000+ | Pro |

---

## Guarantee / Risk Reversal

**Heading:** If Confire doesn't save you tokens, you owe us nothing.

**Body:**  
The Free plan has no time limit and no credit card requirement. If your first session doesn't show measurable token savings in `confire stats`, uninstall is one command: `confire uninstall`. We'd genuinely rather you leave than stay with a tool that doesn't work for you.

---

## Bottom CTA

**Heading:** Start with free. Upgrade when it matters.

**CTA button:** `Create free account`

**Subtext:**  
500 calls/month · No card required · Works in 2 minutes

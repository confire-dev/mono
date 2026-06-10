# Confire Launch Strategy

**Situation:** v0.13.0 shipped, infrastructure complete, research papers validate the threat model. Ready for a **public early access launch** with a clear path to full launch.

---

## Where We Are in the Five Phases

- Phase 1 Internal ✅ done
- Phase 2 Alpha ✅ done
- Phase 3 Beta → finish this week
- Phase 4 Early Access → **launch target**
- Phase 5 Full Launch → 3–6 weeks after early access validates

The research papers (MCP client prompt injection / ClawGuard) hand us the perfect narrative to kick Phase 4 into gear publicly.

---

## The Core Launch Narrative

**Lead with the research, not the product.**

> AI coding agents are gaining tool access faster than their security model is maturing.
>
> Recent research evaluated seven MCP clients — Claude Code, Cursor, Cline, Continue, Gemini CLI — and found significant differences in how they handle prompt injection, tool poisoning, hidden parameters, and unauthorized tool invocation.
>
> Confire adds a local runtime firewall around tool activity: risky actions are reviewed before execution, and suspicious tool results are flagged after they return.

This framing does four things:
1. Names the exact tools our users already use (SEO + resonance)
2. Cites external authority so we don't sound like a vendor claiming there's a problem
3. Positions us as the practical fix, not the academic solution
4. The word "firewall" is immediately understood — no category education needed

---

## Owned Channels

**Email list** — highest priority. Every install should be an invitation to join. Gate the "custom rules" and "security event history" tiers behind an account to capture emails from the most engaged users.

**Blog at confire.dev/blog** — SEO engine. Three posts to write before launch:

1. **"MCP prompt injection is not theoretical"** — summarize the research papers, cite them properly, show what Confire does about it. Target: `mcp prompt injection`, `claude code security`, `cursor mcp security`. This is the moat post.

2. **"How Claude Code, Cursor, and Cline handle tool poisoning differently"** — a comparison post based on the first paper. Ranks for all the tool names the paper mentions. Drives traffic from users of each tool.

3. **"What is a tool firewall?"** — define the category we're creating. This is the post that gets linked from HN and Reddit threads.

**Changelog** — publish publicly at confire.dev/changelog. Developers check changelogs. Every release is a touchpoint.

---

## Rented Channels (Pick Two)

**Twitter/X — primary rented channel.** Developer security Twitter is active and engaged. The research paper angle is extremely tweetable.

Thread structure:
1. Hook: "Researchers evaluated 7 MCP clients for prompt injection. The results are interesting."
2. What they found (summarize neutrally — don't overclaim)
3. What this means for developers using Claude Code / Cursor / Cline
4. What Confire does about it (one sentence, practical)
5. CTA: try it free, no account needed

Post on launch day. Quote-tweet the actual papers if they have public links.

**Hacker News — secondary rented channel.** Two submissions planned:

1. **Show HN: Confire – local tool firewall for Claude Code and MCP clients** — on early access launch day. Lead with what it does, not what it prevents. Include the research angle in the first comment.

2. Submit the "MCP prompt injection is not theoretical" blog post — one week after the Show HN.

---

## Borrowed Channels (Start Outreach This Week)

**Security researchers and paper authors** — reach out to the authors of both papers. Not asking for endorsement — asking if they'd be willing to try it and share feedback. If one tweets about it, that's enormous earned signal.

**Developer security podcasters** — Security Now, The Changelog, Darknet Diaries, SANS ISC. Pitch: "MCP tool security is underexplored and we built a local firewall for it."

**AI coding YouTubers** — send a free Pro account to 5–10 YouTubers who cover Claude Code, Cursor, or AI-assisted development. No script, no ask — just hope they find it interesting enough to mention. (The TRMNL model: Snazzy Labs, $500K in sales from one unsolicited review.)

Target creators who have covered: Claude Code setup, Cursor deep-dives, MCP configuration, AI agent workflows.

**Developer newsletters** — TLDR, Pointer, Bytes. The "MCP security research" angle is genuinely newsworthy.

---

## Launch Week Sequence

**Day -7 (prep):**
- Publish "MCP prompt injection is not theoretical" blog post (don't announce yet, let it index)
- Finalize confire.dev landing page — headline should be "Local tool firewall for AI coding agents"
- Make sure free tier install flow is frictionless: `curl ... | sh` → working in < 5 minutes

**Day -3 (soft signal):**
- Post a single tweet: "Something that should exist for AI coding agents: a local firewall that reviews tool calls before they run. We built it. confire.dev" — no thread, no hype, just the statement. Gauge response.

**Day 0 (launch):**
- Show HN: **Confire – local runtime firewall for Claude Code, Cursor, and MCP clients**
- Twitter/X thread (full research-backed thread)
- Email to existing users/waitlist: "Early access is open — what we built and why"
- Link confire.dev/blog post from the HN thread

**Day +1:**
- Respond to every HN comment
- Follow up with anyone who engaged on Twitter
- DM/email the research paper authors

**Day +7:**
- Publish "How Claude Code, Cursor, and Cline handle tool poisoning differently"
- Submit as a second HN link post
- Start personal outreach to YouTubers with free Pro accounts

---

## Product Hunt (Plan, Don't Rush)

Do Product Hunt 3–4 weeks after early access, when we have:
- Testimonials from real users
- A polished demo GIF or video showing a blocked destructive tool call
- 50+ people ready to upvote on launch day

**Hook:** "The firewall your AI coding agent is missing" — before/after showing a risky `git push --force` getting blocked vs. executing silently.

---

## Messaging Constraints

Carry these forward from CONTEXT.md. Hard rules during launch.

| Say | Don't say |
|-----|-----------|
| "reviews risky tool calls" | "prevents all prompt injection" |
| "sanitizes prompt-injection-like content" | "guarantees safety" |
| "local runtime firewall" | "ClawGuard for production" |
| "helps keep context cleaner and safer" | "makes MCP safe" |
| "detects common secret patterns" | "catches every secret" |

The research papers establish that the problem is real — not that Confire is the complete solution.

---

## The Free Tier as Distribution Wedge

Local, offline, no account required, unlimited. That means:
- Zero friction for the skeptical developer who just wants to try it
- Every `confire setup` completed is a potential evangelist
- Every blocked destructive tool call becomes a story they'll tell

Optimize this path before launch:
```
install → confire setup → confire daemon → first blocked action → "sign up to customize rules"
```

---

## Roadmap as Launch Content

The tier structure maps to a launch narrative that compounds over time:

| Tier | Hook |
|------|------|
| Free | Local tool firewall — works offline, no account |
| Dev | Custom guardrails + history — see what your agent tried to do |
| Registry | Known MCP/tool risk intelligence — community-sourced threat patterns |
| Team | Org-wide runtime agent security policy |

Announce Registry and Team as coming soon on launch day. Two more launch moments in the next 6–12 months.

---

## Launch Checklist

**This week:**
- [ ] Write and publish "MCP prompt injection is not theoretical" post
- [ ] Verify `curl | sh` install flow is < 5 minutes to first working firewall
- [ ] Lock final homepage headline: "Local tool firewall for AI coding agents"
- [ ] Add public changelog at confire.dev/changelog
- [ ] Draft Show HN post and Twitter thread

**Launch day:**
- [ ] Show HN: Confire — local runtime firewall for Claude Code, Cursor, and MCP clients
- [ ] Twitter/X research-backed thread
- [ ] Email to existing users
- [ ] Team on standby to respond to comments all day

**Week after:**
- [ ] "How Claude Code, Cursor, Cline handle tool poisoning differently" post
- [ ] Second HN submission (link post)
- [ ] Outreach to 5–10 AI coding YouTubers (free Pro)
- [ ] Outreach to paper authors

**3–4 weeks after:**
- [ ] Product Hunt listing — after testimonials accumulate
- [ ] Developer newsletter pitches
- [ ] Podcast outreach

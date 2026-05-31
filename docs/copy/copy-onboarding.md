# Onboarding Copy: Confire

Covers: signup flow, welcome email, first-run CLI experience, upgrade prompts, and key in-product messages.

---

## Signup Flow

### Step 1: Landing on the signup page

**Page heading:** Create your free Confire account

**Subhead:** 500 optimizations per month, free forever. No credit card required.

**Form fields:**
- Email address
- Password (or "Continue with GitHub" button — preferred for developer audience)

**CTA button:** `Create account`

**Below form:**  
By creating an account, you agree to the Terms of Service and Privacy Policy.

---

### Step 2: Post-signup / Account created

**Heading:** Account created. Now connect your agent.

**Body:**  
Run this command to connect Confire to Claude Code, Cursor, or Cline:

```bash
confire setup
```

It'll walk you through linking your account and configuring the hook automatically.

**CTA button:** `Open the docs →`

**Or copy your API key:**  
`[key displayed with copy button]`

---

### Step 3: CLI First Run (`confire setup`)

**CLI output — what the user sees:**

```
Confire setup

Which AI agent are you using?
  > Claude Code
    Cursor
    Cline
    Other (manual setup)

Detected Claude Code settings at ~/.claude/settings.json
Adding Confire as a PostToolUse hook...

Done. Confire is now active.

To verify it's working, run a Claude Code session and then:
  confire stats

Your free tier: 500 calls/month remaining.
Account: you@example.com

Need help? confire help | docs: confire.dev/docs
```

---

## Welcome Email

**Subject:** You're set up. Here's what happens next.

**Preheader:** Your first Confire optimization is probably already running.

---

Hey,

Your Confire account is live and (if you've run `confire setup`) your agent is already optimizing.

Here's what to expect:

**Your first session**  
Run your normal Claude Code or Cursor workflow. Confire works in the background as a hook — you won't notice anything different except your sessions will go further before hitting context limits.

**Check your savings**  
After your first session, run:
```
confire stats
```
You'll see how many calls were optimized and an estimate of tokens saved.

**Your free tier: 500 calls/month**  
Each tool call your agent makes (Bash command, URL fetch, GitHub response) counts as one optimization. 500/month is more than most developers use in a week of heavy sessions.

**If you hit the limit or want more**  
Upgrade to Dev or Pro at confire.dev/pricing. Takes 30 seconds.

Questions? Reply to this email. I read every one.

— [Founder name], Confire

P.S. If you run into anything weird during setup, `confire help` is a good first stop. The docs are at confire.dev/docs.

---

## In-Product / CLI Messages

### Usage threshold warnings

**At 80% of free tier (400/500 calls used):**
```
Heads up: You've used 400 of your 500 free calls this month.
Upgrade to Dev for [X] calls/month: confire.dev/pricing
```

**At 100% (limit hit):**
```
You've reached your free tier limit for this month (500/500 calls).
Confire is paused until [reset date], or upgrade now to keep going:
confire.dev/pricing
```

**Reset notification (next month):**
```
Your free tier has reset. You have 500 calls available this month.
```

---

### `confire stats` output template

```
Confire — Usage Summary

This month:
  Calls optimized:    247 / 500
  Estimated tokens saved: ~128,000
  Average reduction:  73%

Top tool types by call volume:
  Bash           → 141 calls (avg 81% reduction)
  Web fetch      → 68 calls  (avg 65% reduction)
  GitHub API     → 38 calls  (avg 89% reduction)

Account: free tier
Resets: June 1, 2026

Upgrade: confire.dev/pricing
```

---

### `confire start` / `confire stop`

**Start:**
```
Confire active. Tool outputs will be optimized before reaching your agent.
```

**Stop:**
```
Confire paused. Your agent will receive unoptimized tool outputs.
Run 'confire start' to resume.
```

---

## Upgrade Flow

### Upgrade prompt (in CLI, hitting limit)

```
You've hit your monthly limit.

Your options:
  1. Wait until [reset date] (free tier resets)
  2. Upgrade to Dev — [X] calls/month for $[price]
  3. Upgrade to Pro — unlimited calls for $[price]

Open pricing: confire.dev/pricing
Or: confire upgrade
```

### Upgrade confirmation email

**Subject:** You're on Dev. Here's what changed.

**Body:**

Hey,

You're now on the Dev plan. Here's what you have:

- [X] optimized calls/month
- Priority cloud optimizer (faster processing)
- Usage analytics at `confire stats`

Your next billing date is [date]. Manage your subscription at confire.dev/account.

If you ever want to cancel or change plans, just reply here or go to your account settings. No runaround.

— [Founder name], Confire

---

## Docs Quickstart Page Copy

**URL:** confire.dev/docs/quickstart  
**H1:** Get started with Confire in 2 minutes

### Step 1: Install

```bash
curl -fsSL https://install.confire.dev | sh
```

Or with Homebrew:
```bash
brew install confire
```

Verify:
```bash
confire --version
```

### Step 2: Log in

```bash
confire login
```

Opens a browser window to authenticate with your Confire account.  
No browser? Use `confire login --token YOUR_API_TOKEN`.

### Step 3: Connect your agent

**For Claude Code:**
```bash
confire setup --agent claude-code
```

Confire will automatically add itself as a PostToolUse hook to your Claude Code settings.

**For Cursor:**
```bash
confire setup --agent cursor
```

**For manual setup:** See the [manual hook configuration guide](/docs/hooks).

### Step 4: Verify it's working

Start a Claude Code or Cursor session with a Bash-heavy task. Then:

```bash
confire stats
```

You should see at least a few calls logged and tokens saved.

That's it. Confire runs automatically from here. You don't need to think about it.

---

**What happens next?**  
- [How Confire compresses tool outputs →](/docs/how-it-works)  
- [View your usage stats →](/docs/stats)  
- [Upgrade your plan →](/pricing)

---

## Error Messages (Friendly, Not Cryptic)

**No internet / cloud unreachable:**
```
Cloud optimizer unreachable. Using local optimizer instead.
Performance is identical — latency may be slightly higher.
```

**Hook not installed / misconfigured:**
```
Confire doesn't seem to be connected to your agent.
Run 'confire setup' to fix this automatically.
```

**Invalid API key:**
```
Your API key isn't working. Try 'confire login' to re-authenticate.
If the problem continues, check confire.dev/account.
```

**Unknown tool type (uncovered output format):**
```
Tool output type not recognized. Passing through unmodified.
Let us know what you're working with: confire.dev/feedback
```

---

## Referral / Word-of-Mouth Prompt (Post-Value Moment)

Triggered after first session where >50% token reduction is achieved:

```
Nice — Confire saved ~34,000 tokens in that session.
That's roughly [$ amount] off your Claude API bill.

If you know other Claude Code or Cursor users, tell them:
confire.dev

(We don't do ads. Word of mouth is how we grow.)
```

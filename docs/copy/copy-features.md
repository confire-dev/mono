# Features Page Copy: Confire

**URL:** confire.dev/features  
**Meta title:** `Confire Features — How AI Agent Context Optimization Works`  
**Meta description:** `See how Confire compresses Bash logs, GitHub responses, and web fetches before they reach your AI agent's context window. Less noise, lower costs.`

---

## Hero

**H1:** Everything Confire does, explained.

**Subhead:**  
Confire is small by design. It does one thing — remove noise from tool call outputs before they enter your AI agent's context. Here's exactly how.

---

## Feature: Hook-Based Integration

**Heading:** Works as a hook. Zero workflow changes.

**Body:**  
Confire integrates through the standard hook protocol for Claude Code, Cursor, and Cline. When you run `confire setup`, it adds itself as a PostToolUse hook in your agent's settings file. From that point, every tool call response passes through Confire before the model sees it.

No code changes. No custom forks. No middleware to maintain.

**Technical detail:**  
For Claude Code, Confire adds a hook to `~/.claude/settings.json`:
```json
{
  "hooks": {
    "PostToolUse": ["confire process"]
  }
}
```

The agent calls tools normally. Confire processes the output. The model gets compressed input.

**Supported agents:**
- Claude Code (Anthropic)
- Cursor
- Cline
- Any agent supporting the PostToolUse hook pattern

---

## Feature: Cloud Optimizer with Local Fallback

**Heading:** Always on. Cloud-fast or locally resilient.

**Body:**  
Confire's primary optimizer runs on Cloudflare's edge network. Processing happens close to where you are — typical latency is under 50ms. If the cloud is unreachable (offline work, network issues, rate limits), a local optimizer kicks in automatically.

You never lose optimization coverage. The local fallback is deterministic and offline-capable. The cloud optimizer handles more complex compression patterns.

**The stack:**
- **Cloud:** Cloudflare Worker — edge-deployed, globally distributed
- **Local:** Go binary included in the CLI install — no dependencies, no runtime

**Failover:** Automatic, silent. Your agent doesn't know or care which optimizer ran.

---

## Feature: Tool Output Compression

**Heading:** Compression that targets the right content.

**Body:**  
Confire doesn't blindly truncate output. It analyzes the structure of each tool's response and removes what's provably noise: redundant stack traces, verbose metadata, repeated headers, binary-encoded content, and other patterns the model doesn't need to act on.

The compressed output retains:
- All actionable information
- Error messages and relevant status codes
- File paths, function names, and structural markers
- Content the agent explicitly needs to respond

**By tool type:**

| Tool Output | What Confire Removes | What It Keeps |
|-------------|---------------------|---------------|
| Bash command output | Repeated lines, progress bars, verbose flags, debug traces | Return codes, errors, key output lines |
| GitHub API responses | Pagination metadata, redundant object fields, blob data | PR titles, status, file diffs, commit messages |
| Web page fetches | Navigation HTML, ads, boilerplate, repeated elements | Body text, headings, code blocks, structured data |
| Figma file outputs | Asset binary metadata, style inheritance repetition | Layer names, component IDs, text content |
| Generic JSON | Deeply nested empty fields, null values, schema boilerplate | Populated fields, IDs, meaningful values |

---

## Feature: Usage Stats and Savings Tracking

**Heading:** See exactly what you're saving.

**Body:**  
`confire stats` gives you a clear view of your optimization activity: how many calls ran, which tool types produced the most noise, and how much token reduction you're getting.

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
```

Token savings are estimated based on tokenizer output for the before/after sizes of each compressed call.

---

## Feature: No Data Storage

**Heading:** Your tool outputs don't live on our servers.

**Body:**  
Confire processes tool output content in-flight to compress it. We do not:
- Store the content of tool calls
- Log what commands you run
- Retain file contents, API responses, or web fetches beyond the processing window
- Share your output data with third parties

What we do retain: call counts, byte-size statistics (for stats/billing), and error logs. No content.

If you work in regulated environments or want full on-device processing, the local optimizer handles everything without any cloud call.

---

## Feature: Simple CLI Surface

**Heading:** Five commands. That's the whole interface.

**Body:**  
Confire is a CLI tool. There's no dashboard to learn, no UI to configure, no settings buried in a web panel. Everything you need is in the terminal.

```
confire setup      # Connect to your AI agent
confire login      # Authenticate your account
confire stats      # View usage and savings
confire start      # Enable optimization (default: on)
confire stop       # Pause optimization
confire help       # Full command reference
```

Advanced configuration is available via a local config file (`~/.confire/config.toml`) for teams or scripts that need to customize behavior.

---

## Feature: Per-Tool-Type Tuning (Pro)

**Heading:** Control which tools Confire optimizes.

**Body (Pro plan):**  
On the Pro plan, you can configure Confire's behavior per tool type. Aggressive compression for Bash output, lighter touch on web fetches, passthrough for specific tools entirely. Useful when you need fine-grained control in production agent pipelines.

Example config:
```toml
[optimizer]
  bash = "aggressive"
  web_fetch = "standard"
  github = "aggressive"
  figma = "standard"
  custom_tools = "passthrough"
```

---

## Feature: Team Usage (Pro)

**Heading:** One plan, your whole team.

**Body (Pro):**  
Pro plans cover your full team under a single account. Usage is pooled (not per-seat). Monitor team-wide optimization stats, token savings, and call volumes from a shared dashboard or via `confire stats --team`.

---

## Comparison Table: Confire vs Doing Nothing

| | Without Confire | With Confire |
|---|---|---|
| Bash log in context | 8,000 tokens | 600–800 tokens |
| GitHub PR response | 12,000 tokens | 400–600 tokens |
| Web page fetch | 15,000 tokens | 2,000–4,000 tokens |
| Context fills up at | Task 3–4 in a session | Task 8–12 in a session |
| Monthly token bill | Baseline | 40–75% lower |
| Agent behavior | Confused by noise | Acting on signal |

---

## CTA

**Heading:** Try it free. See the difference in your first session.

**CTA button:** `Get started — free up to 500 calls/month`

**Subtext:**  
No credit card. No config. One command: `confire setup`.

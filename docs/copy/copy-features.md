# Features Page Copy: Confire

**URL:** confire.dev/features  
**Meta title:** `Confire Features — Context and Tool Firewall for AI Coding Agents`  
**Meta description:** `See how Confire reviews risky tool calls and sanitizes noisy output before it reaches your AI agent's context. Less noise, fewer unsafe actions.`

---

## Hero

**H1:** Everything Confire does, explained.

**Subhead:**  
Confire is small by design. It does two things — review risky tool calls before they run, and remove noise from tool call outputs before they enter your AI agent's context. Here's exactly how.

---

## Feature: Hook-Based Integration

**Heading:** Works as a hook. Zero workflow changes.

**Body:**  
Confire integrates through the standard hook protocol for Claude Code, Cursor, and VS Code. When you run `confire setup`, it adds itself as a PreToolUse and PostToolUse hook in your agent's settings file. From that point, every tool call passes through Confire — before execution and before the output reaches context.

No code changes. No custom forks. No middleware to maintain.

**Technical detail:**  
For Claude Code, Confire adds hooks to `~/.claude/settings.json`:
```json
{
  "hooks": {
    "PreToolUse": ["confire hook"],
    "PostToolUse": ["confire hook"]
  }
}
```

The agent calls tools normally. Confire reviews the call, processes the output. The model gets clean input.

**Supported agents:**
- Claude Code (Anthropic)
- Cursor
- VS Code
- Any agent supporting the PreToolUse/PostToolUse hook pattern

---

## Feature: Local-First by Default

**Heading:** All processing on your machine.

**Body:**  
All firewall decisions and context passes — security, noise trimming, MCP normalization — run in the daemon on your machine. No tool output leaves your machine.

Structured telemetry events (risk level, action taken, session metadata) are sent to the cloud when you're logged in. Tool call content is never included.

**What runs locally:**
- Tool Firewall evaluation (allow/warn/review/block)
- Secret redaction
- Prompt-injection sanitization
- Noise trimming (Bash, WebFetch, MCP, Generic)
- Policy rule evaluation

---

## Feature: Tool Output Noise Trimming

**Heading:** Trimming that targets the right content.

**Body:**  
Confire doesn't blindly truncate output. It analyzes the structure of each tool's response and removes what's provably noise: redundant stack traces, verbose metadata, repeated headers, binary-encoded content, and other patterns the model doesn't need to act on.

The trimmed output retains:
- All actionable information
- Error messages and relevant status codes
- File paths, function names, and structural markers
- Content the agent explicitly needs to respond

**By tool type:**

| Tool Output | What Confire Removes | What It Keeps |
|-------------|---------------------|---------------|
| Bash command output | Repeated lines, progress bars, verbose flags, debug traces | Return codes, errors, key output lines |
| Web page fetches | Navigation HTML, ads, boilerplate, repeated elements | Body text, headings, code blocks, structured data |
| MCP JSON responses | Deeply nested empty fields, null values, schema boilerplate | Populated fields, IDs, meaningful values |
| Generic tool output | Redundant metadata, null arrays, empty objects | Populated, actionable fields |

---

## Feature: Tool Firewall

**Heading:** Review risky commands before they run.

**Body:**  
The Tool Firewall evaluates every tool call against built-in policy rules before execution. It can allow, warn, review, or block based on the risk profile of the call.

Built-in rules cover:
- Destructive git operations (force push, reset --hard)
- Mutating MCP actions (delete, destroy, publish)
- Secret file reads
- Database resets and destructive shell commands

**Modes:** allow, warn, review (pauses for confirmation), block.

---

## Feature: Usage Stats and Savings Tracking

**Heading:** See exactly what you're saving.

**Body:**  
`confire status` and `confire stats` give you a view of your session activity: how many calls processed, which tool types produced the most noise, and how much context reduction you're getting.

Token savings are estimated based on before/after sizes of each processed call.

---

## Feature: No Data Storage

**Heading:** Your tool outputs don't live on our servers.

**Body:**  
Confire processes tool output content locally — it never leaves your machine for processing. We do not:
- Store the content of tool calls
- Log what commands you run
- Retain file contents, API responses, or web fetches beyond the local processing window
- Share your output data with third parties

What we retain (when logged in): structured event metadata (risk level, action taken, session counts). No content.

---

## Feature: Simple CLI Surface

**Heading:** Five commands. That's the whole interface.

**Body:**  
Confire is a CLI tool. There's no dashboard to learn, no UI to configure, no settings buried in a web panel. Everything you need is in the terminal.

```
confire setup      # Connect to your AI agent
confire login      # Authenticate your account
confire status     # View daemon state and hook status
confire start      # Enable firewall (default: on)
confire stop       # Pause firewall
confire help       # Full command reference
```

Advanced configuration is available via a local config file (`~/.confire/config.json`) for teams or scripts that need to customize behavior.

---

## Feature: Custom Guardrail Rules (Dev)

**Heading:** Control which actions Confire reviews.

**Body (Dev plan):**  
On Dev, you can define custom rules from the dashboard and sync them locally. Rules cover any tool type and any phase — allow specific force pushes for your team, block specific MCP mutations, or add review gates on custom tools.

Rules are evaluated locally in the daemon — no cloud call at evaluation time.

---

## Comparison Table: Confire vs Doing Nothing

| | Without Confire | With Confire |
|---|---|---|
| Bash log in context | 8,000 tokens | 600–800 tokens |
| Web page fetch | 15,000 tokens | 2,000–4,000 tokens |
| Risky tool call review | None | Review before execution |
| Context fills up at | Task 3–4 in a session | Task 8–12 in a session |
| Secret leakage risk | Unmitigated | Redacted before context |
| Agent behavior | Confused by noise | Acting on signal |

---

## CTA

**Heading:** Try it free. See the difference in your first session.

**CTA button:** `Get started — free`

**Subtext:**  
No credit card. No config. One command: `confire setup`.

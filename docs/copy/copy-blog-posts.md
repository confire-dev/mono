# Blog Post Copy: Confire — Full Drafts (Priority Posts)

---

## Post 1: Claude Code Hooks — Complete Guide

**URL:** confire.dev/blog/claude-code-hooks-guide  
**Target keyword:** claude code hooks  
**Meta title:** `Claude Code Hooks: Complete Guide (with Examples)`  
**Meta description:** `Learn how Claude Code hooks work, how to use PreToolUse and PostToolUse, and how to write your own hook to control agent behavior.`  
**Buyer stage:** Implementation  
**Type:** Searchable — use-case guide with code examples

---

### Claude Code Hooks: Complete Guide

If you use Claude Code for serious work, hooks are the feature that separates power users from everyone else. They let you intercept what goes into and comes out of tool calls — before the model ever sees it.

This guide covers everything: what hooks are, how they work, how to write one, and practical examples you can use today.

#### What Are Claude Code Hooks?

Claude Code hooks are scripts or commands that run automatically at specific points in the tool-call lifecycle. When your agent executes a tool (a Bash command, a file read, a web fetch), hooks can intercept the input before it runs or process the output before it reaches the model context.

Two hook types exist today:

**PreToolUse** — runs before a tool executes. Useful for:
- Validating or modifying tool inputs
- Blocking certain commands
- Logging what's about to run

**PostToolUse** — runs after a tool executes, before the output enters context. Useful for:
- Compressing verbose outputs
- Filtering noise from logs
- Reformatting data structures

#### Where Hooks Are Configured

Hooks live in `~/.claude/settings.json` (global) or `.claude/settings.json` (project-level):

```json
{
  "hooks": {
    "PreToolUse": ["your-pretooluse-script"],
    "PostToolUse": ["your-posttooluse-script"]
  }
}
```

The value is a command string that Claude Code will execute. You can use any CLI tool, script, or binary.

#### How PostToolUse Hooks Work

When Claude Code runs a tool and gets output, it passes that output to your PostToolUse hook via stdin. Your hook processes it and returns modified output via stdout. Claude Code then puts your hook's stdout into the model context instead of the raw output.

Flow:
```
Tool runs → raw output → PostToolUse hook → processed output → model context
```

Your hook receives:
```json
{
  "tool": "bash",
  "input": {"command": "git log --oneline -100"},
  "output": "... raw tool output ..."
}
```

Your hook should return modified output as plain text or JSON.

#### Example 1: Basic Output Truncation

The simplest possible hook — truncate Bash output to 2000 characters:

```bash
#!/bin/bash
# ~/.claude/hooks/truncate-output.sh

INPUT=$(cat)
OUTPUT=$(echo "$INPUT" | jq -r '.output // .')
TRUNCATED="${OUTPUT:0:2000}"

if [ ${#OUTPUT} -gt 2000 ]; then
  echo "${TRUNCATED}... [truncated: ${#OUTPUT} chars total]"
else
  echo "$OUTPUT"
fi
```

Make it executable: `chmod +x ~/.claude/hooks/truncate-output.sh`

Register it:
```json
{
  "hooks": {
    "PostToolUse": ["~/.claude/hooks/truncate-output.sh"]
  }
}
```

#### Example 2: Strip Node Modules from File Listings

When Claude Code lists a directory, node_modules is useless noise. Filter it:

```bash
#!/bin/bash
INPUT=$(cat)
echo "$INPUT" | jq -r '.output' | grep -v node_modules | grep -v ".git/" | grep -v "__pycache__"
```

#### Example 3: Compress GitHub API Output

GitHub API responses are extremely verbose. This hook extracts just what matters:

```python
#!/usr/bin/env python3
import json, sys

data = json.load(sys.stdin)
output = data.get('output', '')

try:
    parsed = json.loads(output)
    # Keep only essential fields for PR responses
    if isinstance(parsed, dict):
        cleaned = {
            k: v for k, v in parsed.items()
            if k in ['title', 'state', 'body', 'number', 'url', 'head', 'base', 'files']
        }
        print(json.dumps(cleaned, indent=2))
    else:
        print(output[:3000])
except:
    print(output[:3000])
```

#### Example 4: Automated Hook with Confire

If you don't want to write and maintain hooks yourself, Confire handles this automatically. One command sets it up:

```bash
confire setup
```

Confire adds PreToolUse and PostToolUse hooks that review risky calls and trim noisy tool output locally before it reaches context. It handles Bash, web fetches, MCP responses, and more — averaging 40–95% reduction. See your savings with `confire stats`.

#### PreToolUse Hook: Blocking Dangerous Commands

PreToolUse hooks can prevent Claude Code from running commands you want to block:

```bash
#!/bin/bash
INPUT=$(cat)
COMMAND=$(echo "$INPUT" | jq -r '.input.command // ""')

# Block rm -rf on production paths
if echo "$COMMAND" | grep -q "rm -rf /prod"; then
  echo '{"block": true, "reason": "rm -rf on /prod is blocked"}' 
  exit 0
fi

# Allow everything else
echo '{"block": false}'
```

#### Debugging Hooks

If your hook isn't running or is behaving unexpectedly:

1. Test it directly: `echo '{"tool":"bash","output":"test"}' | your-hook-script`
2. Check permissions: `ls -la your-hook-script` (must be executable)
3. Check the path in settings.json (absolute paths are safer than relative)
4. Run `claude --debug` to see hook execution in the agent logs

#### Summary

| Hook type | When it runs | Common uses |
|-----------|-------------|-------------|
| PreToolUse | Before tool executes | Input validation, command blocking, logging |
| PostToolUse | After tool, before context | Output compression, noise filtering, reformatting |

Hooks are the cleanest way to control what your agent sees. Start simple, and build from there.

**See also:**
- [How to reduce token usage in Claude Code →](/blog/reduce-token-usage-claude-code)
- [Confire: automated PostToolUse optimization →](/docs/quickstart)

---

## Post 2: Why Your Claude Code Sessions Cost So Much

**URL:** confire.dev/blog/claude-code-cost  
**Target keywords:** claude code cost, why is claude code expensive  
**Meta title:** `Claude Code Cost: What You Actually Pay and How to Cut It`  
**Meta description:** `A clear breakdown of Claude Code costs — API pricing, token consumption patterns, and the fastest ways to reduce your monthly AI bill.`  
**Buyer stage:** Awareness → Consideration  
**Type:** Searchable + Shareable

---

### Claude Code Cost: What You Actually Pay and How to Cut It

Claude Code doesn't have a flat monthly fee. It runs on Anthropic's API — you pay per token, and the amount you spend depends entirely on how your sessions run.

For light users, it's cheap. For power users working on large codebases with complex multi-step tasks, costs can spike to $50–200+ per month, sometimes more. Here's why, and what to do about it.

#### How Claude Code Pricing Actually Works

Claude Code bills through your Anthropic API account. Every session consumes:

- **Input tokens:** Everything in the context window (your messages, file contents, tool outputs, conversation history)
- **Output tokens:** Claude's responses, code it writes, thoughts it expresses

Input and output tokens are priced differently, and input tokens are typically where costs accumulate in agentic use.

**Current Claude 3.5/Claude 3 pricing (check anthropic.com for current rates):**
- Input: ~$3–15 per million tokens depending on model
- Output: ~$15–75 per million tokens depending on model

For a context-heavy agentic session, you might consume 50,000–200,000 input tokens without blinking.

#### What Makes Costs Spike: Tool Call Outputs

The single biggest driver of Claude Code costs isn't how much Claude writes back — it's how much raw tool output gets stuffed into the context.

When your agent runs a Bash command, the entire output goes into context:

```bash
$ git log --oneline
# → 200 lines of commit history → ~4,000 tokens
```

When it fetches a GitHub PR:
```
# → Full JSON response with all fields → 15,000–40,000 tokens
```

When it reads a web page:
```
# → Full HTML including nav, ads, scripts → 20,000–80,000 tokens
```

These aren't edge cases. This is what normal Claude Code sessions look like.

**A realistic cost breakdown for one session:**

| Tool call | Raw output size | Token estimate |
|-----------|----------------|----------------|
| `find . -type f` | 300 lines | ~3,000 |
| `npm install` log | 500 lines | ~5,000 |
| GitHub PR fetch | Full JSON | ~18,000 |
| Web page fetch × 2 | Full HTML | ~35,000 |
| Bash command × 5 | Mixed output | ~8,000 |
| **Total context** | | **~69,000 tokens** |

At $3/million input tokens, that's about $0.21 per session. If you run 10 complex sessions a day, you're at $2/day or $60/month — just in input tokens, just for tool output.

#### Five Ways to Reduce Your Claude Code Bill

**1. Compress tool output with a PostToolUse hook**  
Write a hook that filters verbose output before it enters context. Even crude truncation (first 2000 chars only) can cut Bash output costs by 70%. [Here's how to write a hook →](/blog/claude-code-hooks-guide)

**2. Use `confire setup` for automated compression**  
Confire adds a PostToolUse hook that compresses tool outputs intelligently — targeting the verbose parts (metadata, redundant fields, repeated patterns) while keeping what the agent needs. Free up to 500 calls/month.

```bash
confire setup
```

**3. Be specific with Bash commands**  
`git log --oneline -10` vs `git log` — the former produces 10 lines. The latter produces everything. Claude Code doesn't need full history to work. Prompt it to use specific, scoped commands.

**4. Use a smaller model for tool-heavy tasks**  
For tasks where Claude is mostly reading file contents and executing commands, the Haiku model is often sufficient and is significantly cheaper. Reserve Sonnet or Opus for tasks requiring more reasoning.

**5. Clear context between distinct tasks**  
Use `/clear` in Claude Code between unrelated tasks. Starting fresh prevents the previous session's tool outputs from inflating the context of the next task.

#### What You Can Realistically Expect to Save

With output compression via hooks or Confire:
- Typical reduction in per-session token cost: **40–75%**
- For a developer spending $60/month: reduced to **$15–36/month**
- For a developer spending $200/month: reduced to **$50–120/month**

Results vary by task type. Bash-heavy and GitHub-heavy workflows see the biggest savings. File reading and reasoning tasks see less benefit (the model still needs to read file content).

#### The Bottom Line

Claude Code costs are driven by context size, and context size is driven by tool output verbosity. The fix is to compress what goes into context before it gets there — either by writing your own hooks or using a tool like Confire.

The underlying AI work doesn't get cheaper, but the noise that doesn't need to be there absolutely can.

**Next:** [How to reduce token usage in Claude Code →](/blog/reduce-token-usage-claude-code)

---

## Post 3: How to Reduce Token Usage in Claude Code

**URL:** confire.dev/blog/reduce-token-usage-claude-code  
**Target keyword:** reduce token usage claude code  
**Meta title:** `How to Reduce Token Usage in Claude Code (Practical Guide)`  
**Meta description:** `Practical techniques to cut Claude Code token consumption: hook-based output compression, context pruning, and tool call filtering.`  
**Buyer stage:** Consideration → Decision  
**Type:** Searchable — practical guide

---

### How to Reduce Token Usage in Claude Code

If you've looked at your Anthropic API usage and felt a jolt — this is for you. Claude Code sessions consume tokens fast, mostly because tool call outputs are verbose by default. Here's a practical, technical guide to cutting consumption without losing quality.

#### Why Token Usage Spikes in Agentic Sessions

Claude Code uses tools constantly. Every tool call (Bash command, file read, URL fetch, API call) generates output, and that output gets appended to the model's context window. The context window is cumulative within a session — older tool outputs don't disappear, they stay in context.

This means a session that runs 20 tool calls might have 100,000+ tokens of tool output in context by the end — most of it noise that Claude no longer needs to act on.

#### Technique 1: PostToolUse Hook for Output Compression

The most impactful technique is intercepting tool output with a PostToolUse hook and compressing it before it enters context.

Minimal implementation:

```bash
#!/bin/bash
# Compress Bash output to first 100 lines + last 10 lines
INPUT=$(cat)
OUTPUT=$(echo "$INPUT" | jq -r '.output // .')
LINES=$(echo "$OUTPUT" | wc -l)

if [ "$LINES" -gt 110 ]; then
  HEAD=$(echo "$OUTPUT" | head -100)
  TAIL=$(echo "$OUTPUT" | tail -10)
  echo "$HEAD"
  echo "... [$(($LINES - 110)) lines omitted] ..."
  echo "$TAIL"
else
  echo "$OUTPUT"
fi
```

Register in `~/.claude/settings.json`:
```json
{"hooks": {"PostToolUse": ["/path/to/compress.sh"]}}
```

For automated, multi-tool compression: `confire setup` handles this for Bash, GitHub, web fetches, and Figma in one step.

#### Technique 2: Scope Your Bash Commands

The way you prompt Claude Code affects how expensive its tool use is. Encourage scoped commands:

| Verbose (expensive) | Scoped (cheap) |
|--------------------|----------------|
| `git log` | `git log --oneline -20` |
| `find . -type f` | `find . -name "*.go" -not -path "*/vendor/*"` |
| `npm test` | `npm test -- --grep "specific test"` |
| `cat large-file.json` | `jq '.key.path' large-file.json` |

You can create a project-level `CLAUDE.md` instruction: "When running Bash commands, always scope output to what's needed. Prefer `-n 20` limits on log commands. Use `jq` to extract specific fields from JSON."

#### Technique 3: Use /clear Between Tasks

Claude Code's `/clear` command resets the context. All previous tool outputs, messages, and conversation history is wiped. The model starts fresh.

Use it between:
- Switching from one feature to another
- Moving from investigation to implementation
- Starting a new file or module

Don't use it in the middle of a task — you'll lose the working context. Use it at natural breakpoints.

#### Technique 4: Filter Specific Tool Types

If you know certain tool types generate the most noise, configure hooks selectively:

```bash
#!/bin/bash
INPUT=$(cat)
TOOL=$(echo "$INPUT" | jq -r '.tool')

case "$TOOL" in
  "bash")
    # Heavy compression for Bash
    echo "$INPUT" | jq -r '.output' | head -200
    ;;
  "web_fetch")
    # Strip HTML, keep text
    echo "$INPUT" | jq -r '.output' | sed 's/<[^>]*>//g' | head -300
    ;;
  *)
    # Pass other tools through unchanged
    echo "$INPUT" | jq -r '.output'
    ;;
esac
```

#### Technique 5: Model Selection

Claude 3 Haiku is significantly cheaper than Sonnet or Opus. For tool-heavy, execution-focused tasks where Claude is primarily running commands and reading files (not reasoning through complex problems), Haiku often performs as well at a fraction of the cost.

Switch via Claude Code's model selector or `--model` flag.

#### Measuring Your Progress

Before optimizing: run `confire stats` or check your Anthropic API dashboard for baseline token usage per session.

After: compare per-session costs over the next week. Look for:
- Average tokens per session (before vs after)
- Cost per day (Anthropic API usage page)
- Most expensive tool types

With PostToolUse compression, most developers see 40–75% reduction in session token costs. The biggest wins come from GitHub API calls and web fetches — both are extremely verbose by default.

**Related:**
- [Claude Code hooks guide →](/blog/claude-code-hooks-guide)
- [Claude Code cost breakdown →](/blog/claude-code-cost)
- [Get started with Confire →](/docs/quickstart)

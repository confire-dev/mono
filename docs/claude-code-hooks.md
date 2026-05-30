# Claude Code Hook Contract

**Source:** Extracted from Claude Code 2.1.158 binary via `strings` — not from public docs (which lag the release).  
**Why it matters:** The entire product depends on `updatedToolOutput` working for ALL tools. Verified in binary before building.

## PostToolUse stdin payload

```json
{
  "session_id":       "uuid",
  "transcript_path":  "/path/to/transcript.jsonl",
  "cwd":              "/current/working/dir",
  "permission_mode":  "default|plan|acceptEdits|auto|dontAsk|bypassPermissions",
  "hook_event_name":  "PostToolUse",
  "tool_name":        "Bash",
  "tool_input":       { "command": "npm test" },
  "tool_response":    "...output string or object...",
  "tool_use_id":      "uuid",
  "duration_ms":      142
}
```

**Critical:** The field is `tool_response`, not `tool_output`. Public docs say `tool_output`; binary says `tool_response`.

## PostToolUse stdout (exit 0) — to replace tool output

```json
{
  "hookSpecificOutput": {
    "hookEventName":     "PostToolUse",
    "updatedToolOutput": "<shrunk output — same shape as tool_response>",
    "additionalContext": "optional string Claude sees as context"
  }
}
```

## The shape rule

`updatedToolOutput` is `safeParse`'d against each tool's own `outputSchema`. If it doesn't match, Claude Code logs:
> `PostToolUse hook returned updatedToolOutput that does not match ${toolName}'s output shape; using original output.`
and discards the optimization. Tools with no schema accept any value.

**Rule: preserve the tool's output shape, shrink only the content inside.**  
Since we receive the exact shape in `tool_response`, this is trivial — return the same type with smaller strings.

## PreToolUse — rewrite input before the tool runs

stdin: same shape but `hook_event_name: "PreToolUse"`, no `tool_response`.  
stdout to rewrite the tool call:
```json
{
  "hookSpecificOutput": {
    "hookEventName": "PreToolUse",
    "updatedInput":  { "file_path": "/foo", "limit": 500 }
  }
}
```

## Exit codes

| Code | Meaning |
|------|---------|
| 0    | Success; hookSpecificOutput (if any) is processed |
| 2    | Blocking error; stderr shown to Claude; for PostToolUse tool already ran |
| other | Non-blocking error; execution continues |

## Matchers

| Pattern | Meaning |
|---------|---------|
| `".*"` or omitted | All tools |
| `"Bash"` | Exact tool name |
| `"Edit\|Write"` | Either |
| `"mcp__.*__.*"` | All MCP tools |
| `"mcp__figma__.*"` | Specific MCP server |

## Complete list of hook events (29)

From binary `hook_event_name` literals:

```
Lifecycle:    SessionStart, SessionEnd, Setup, InstructionsLoaded, ConfigChange, CwdChanged
Prompt:       UserPromptSubmit, UserPromptExpansion
Tools:        PreToolUse, PostToolUse, PostToolUseFailure, PostToolBatch
Turn:         Stop, StopFailure, SubagentStart, SubagentStop
Compaction:   PreCompact, PostCompact
Tasks:        TaskCreated, TaskCompleted
Permissions:  PermissionRequest, PermissionDenied
Files:        FileChanged, WorktreeCreate, WorktreeRemove
Other:        Notification, MessageDisplay, Elicitation, ElicitationResult, TeammateIdle
```

## Settings locations

```
~/.claude/settings.json        user-global (all projects)
.claude/settings.json          project-specific (shareable)
.claude/settings.local.json    project-specific (gitignored)
~/.claude/mcp.json             global MCP servers
.mcp.json                      project MCP servers
```

Hook default timeout: **600 seconds**.

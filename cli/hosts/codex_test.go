package hosts

import (
	"encoding/json"
	"testing"

	"github.com/confire-dev/confire/intercept"
)

func TestCodexRealisticFormats(t *testing.T) {
	t.Run("PreToolUse deny", func(t *testing.T) {
		raw := map[string]any{
			"session_id":      "sess-abc",
			"cwd":             "/project",
			"hook_event_name": "PreToolUse",
			"model":           "codex-mini-latest",
			"tool_name":       "bash",
			"tool_input":      map[string]any{"command": "rm -rf /"},
		}
		var in CodexHookInput
		mustRemarshal(t, raw, &in)
		event := DecodeCodexHookInput(in)
		if event.Host != "codex" || event.Phase != intercept.PhaseToolPre {
			t.Fatalf("unexpected event: %+v", event)
		}
		out, shouldWrite, shouldBlock := EncodeCodexPreToolResult(intercept.InterceptResult{
			Kind:   intercept.ResultBlock,
			Reason: "Destructive command blocked by policy",
		}, "PreToolUse")
		if !shouldWrite || !shouldBlock {
			t.Fatalf("expected write+block, got shouldWrite=%v shouldBlock=%v", shouldWrite, shouldBlock)
		}
		if out.HookSpecificOutput == nil || out.HookSpecificOutput.PermissionDecision != "deny" {
			t.Fatalf("expected permissionDecision=deny, got %+v", out.HookSpecificOutput)
		}
		assertJSONHasPath(t, out, "hookSpecificOutput", "permissionDecision")
		assertJSONHasPath(t, out, "hookSpecificOutput", "permissionDecisionReason")
	})

	t.Run("PreToolUse warn allow", func(t *testing.T) {
		out, shouldWrite, shouldBlock := EncodeCodexPreToolResult(intercept.InterceptResult{
			Kind:    intercept.ResultWarn,
			Context: "advisory context",
		}, "PreToolUse")
		if !shouldWrite || shouldBlock {
			t.Fatalf("warn should write but not block")
		}
		if out.HookSpecificOutput == nil || out.HookSpecificOutput.PermissionDecision != "allow" {
			t.Fatalf("expected allow, got %+v", out.HookSpecificOutput)
		}
	})

	t.Run("PostToolUse additionalContext", func(t *testing.T) {
		raw := map[string]any{
			"session_id":      "sess-abc",
			"hook_event_name": "PostToolUse",
			"tool_name":       "bash",
			"tool_response":   "some output",
		}
		var in CodexHookInput
		mustRemarshal(t, raw, &in)
		event := DecodeCodexHookInput(in)
		if event.Phase != intercept.PhaseToolPost {
			t.Fatalf("expected PhaseToolPost, got %v", event.Phase)
		}
		out, ok := EncodeCodexResult(intercept.InterceptResult{
			Kind:    intercept.ResultAddContext,
			Context: "[Confire post_tool steer]\nflags=secret_like_value_detected",
		}, "PostToolUse")
		if !ok || out.HookSpecificOutput == nil || out.HookSpecificOutput.AdditionalContext == "" {
			t.Fatalf("unexpected post output: %+v ok=%v", out, ok)
		}
		assertJSONHasPath(t, out, "hookSpecificOutput", "additionalContext")
		assertJSONHasPath(t, out, "hookSpecificOutput", "hookEventName")
	})

	t.Run("SessionStart context", func(t *testing.T) {
		raw := map[string]any{
			"session_id":      "sess-abc",
			"hook_event_name": "SessionStart",
			"model":           "codex-mini-latest",
		}
		var in CodexHookInput
		mustRemarshal(t, raw, &in)
		event := DecodeCodexHookInput(in)
		if event.Phase != intercept.PhaseSessionStart {
			t.Fatalf("expected PhaseSessionStart, got %v", event.Phase)
		}
	})

	t.Run("passthrough produces no output", func(t *testing.T) {
		_, ok := EncodeCodexResult(intercept.InterceptResult{Kind: intercept.ResultPassthrough}, "PostToolUse")
		if ok {
			t.Fatal("passthrough should not write")
		}
	})

	t.Run("JSON output round-trip", func(t *testing.T) {
		cont := true
		out := CodexHookOutput{
			Continue: &cont,
			HookSpecificOutput: &codexSpecificOut{
				HookEventName:     "PostToolUse",
				AdditionalContext: "test context",
			},
		}
		data, err := json.Marshal(out)
		if err != nil || !json.Valid(data) {
			t.Fatalf("invalid JSON: %v", err)
		}
	})
}

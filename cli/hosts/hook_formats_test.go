package hosts

import (
	"encoding/json"
	"testing"

	"github.com/confire-dev/confire/intercept"
)

// Realistic stdin payloads shaped like Cursor and VS Code hook docs.

func TestCursorRealisticFormats(t *testing.T) {
	t.Run("preToolUse deny", func(t *testing.T) {
		raw := map[string]any{
			"conversation_id": "conv-abc",
			"cursor_version":    "1.7.0",
			"hook_event_name":   "preToolUse",
			"tool_name":         "Shell",
			"tool_input":        map[string]any{"command": "rm -rf /"},
			"tool_use_id":       "tu-1",
			"cwd":               "/project",
		}
		if !IsCursorHook(raw) || IsVSCodeHook(raw) {
			t.Fatal("host detection failed")
		}
		var in CursorHookInput
		mustRemarshal(t, raw, &in)
		event := DecodeCursorHookInput(in)
		if event.Host != "cursor" || event.Phase != intercept.PhaseToolPre {
			t.Fatalf("unexpected event: %+v", event)
		}
		out, ok, block := EncodeCursorPreToolResult(intercept.InterceptResult{
			Kind:   intercept.ResultReview,
			Reason: "CONFIRE REVIEW REQUIRED",
		})
		if !ok || !block || out.Permission != "deny" {
			t.Fatalf("unexpected preTool output: %+v ok=%v block=%v", out, ok, block)
		}
		assertJSONKeys(t, out, "permission", "user_message", "agent_message")
	})

	t.Run("postToolUse MCP replace", func(t *testing.T) {
		raw := map[string]any{
			"conversation_id": "conv-abc",
			"cursor_version":  "1.7.0",
			"hook_event_name": "postToolUse",
			"tool_name":       "MCP: figma/get_file",
			"tool_output":     `{"document":{"id":"0:1"}}`,
			"tool_use_id":     "tu-2",
		}
		var in CursorHookInput
		mustRemarshal(t, raw, &in)
		event := DecodeCursorHookInput(in)
		if !event.Tool.IsMCP {
			t.Fatal("expected MCP tool")
		}
		out, ok := EncodeCursorResult(intercept.InterceptResult{
			Kind:       intercept.ResultReplaceOutput,
			ToolOutput: map[string]any{"document": map[string]any{"id": "0:1", "trimmed": true}},
			Context:    "[Confire post_tool steer]\nmode=mcp_replaced",
		}, in.ToolName)
		if !ok || out.UpdatedMCPToolOutput == nil || out.AdditionalContext == "" {
			t.Fatalf("unexpected MCP post output: %+v ok=%v", out, ok)
		}
		assertJSONKeys(t, out, "updated_mcp_tool_output", "additional_context")
	})

	t.Run("postToolUse native steer only", func(t *testing.T) {
		out, ok := EncodeCursorResult(intercept.InterceptResult{
			Kind:    intercept.ResultAddContext,
			Context: "[Confire post_tool steer]\ntool=Shell\nmode=native_unreplaceable",
		}, "Shell")
		if !ok || out.UpdatedMCPToolOutput != nil {
			t.Fatalf("native post should steer only: %+v ok=%v", out, ok)
		}
		assertJSONKeys(t, out, "additional_context")
	})
}

func TestVSCodeRealisticFormats(t *testing.T) {
	t.Run("PreToolUse deny", func(t *testing.T) {
		raw := map[string]any{
			"timestamp":      "2026-02-09T10:30:00.000Z",
			"cwd":            "/project",
			"sessionId":      "session-identifier",
			"hookEventName":  "PreToolUse",
			"transcript_path": "/tmp/transcript.json",
			"tool_name":      "runTerminalCommand",
			"tool_input":     map[string]any{"command": "DROP TABLE users;"},
			"tool_use_id":    "tool-123",
		}
		if IsCursorHook(raw) || !IsVSCodeHook(raw) {
			t.Fatal("host detection failed")
		}
		var in VSCodeHookInput
		mustRemarshal(t, raw, &in)
		event := DecodeVSCodeHookInput(in)
		if event.Host != "vscode" || event.Tool.Name != "Bash" {
			t.Fatalf("unexpected event: %+v", event)
		}
		out, ok := EncodeVSCodePreToolResult(intercept.InterceptResult{
			Kind:   intercept.ResultBlock,
			Reason: "Destructive command blocked by policy",
		}, "PreToolUse")
		if !ok || out.HookSpecificOutput == nil {
			t.Fatalf("unexpected output: %+v ok=%v", out, ok)
		}
		if out.HookSpecificOutput.PermissionDecision != "deny" {
			t.Fatalf("expected deny, got %+v", out.HookSpecificOutput)
		}
		assertJSONHasPath(t, out, "hookSpecificOutput", "permissionDecision")
		assertJSONHasPath(t, out, "hookSpecificOutput", "permissionDecisionReason")
	})

	t.Run("PostToolUse additionalContext", func(t *testing.T) {
		raw := map[string]any{
			"sessionId":      "session-identifier",
			"hookEventName":  "PostToolUse",
			"tool_name":      "editFiles",
			"tool_input":     map[string]any{"files": []any{"src/main.ts"}},
			"tool_response": "File edited successfully",
			"tool_use_id":    "tool-456",
		}
		var in VSCodeHookInput
		mustRemarshal(t, raw, &in)
		_ = DecodeVSCodeHookInput(in)
		out, ok := EncodeVSCodeResult(intercept.InterceptResult{
			Kind:    intercept.ResultAddContext,
			Context: "[Confire post_tool steer]\ntool=editFiles\nmode=native_unreplaceable",
		}, "PostToolUse")
		if !ok || out.Continue != true {
			t.Fatalf("unexpected post output: %+v ok=%v", out, ok)
		}
		assertJSONHasPath(t, out, "hookSpecificOutput", "additionalContext")
		assertJSONHasPath(t, out, "hookSpecificOutput", "hookEventName")
	})

	t.Run("SessionStart welcome", func(t *testing.T) {
		out := EncodeVSCodeContext("🔥 Confire context firewall is active.", "SessionStart")
		assertJSONHasPath(t, out, "hookSpecificOutput", "additionalContext")
	})
}

func mustRemarshal(t *testing.T, raw map[string]any, dst any) {
	t.Helper()
	b, err := json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, dst); err != nil {
		t.Fatal(err)
	}
}

func assertJSONKeys(t *testing.T, v any, keys ...string) {
	t.Helper()
	var m map[string]any
	b, _ := json.Marshal(v)
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	for _, k := range keys {
		if _, ok := m[k]; !ok {
			t.Fatalf("expected top-level key %q in %s", k, string(b))
		}
	}
}

func assertJSONHasPath(t *testing.T, v any, path ...string) {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	var cur any
	if err := json.Unmarshal(b, &cur); err != nil {
		t.Fatal(err)
	}
	for _, p := range path {
		m, ok := cur.(map[string]any)
		if !ok {
			t.Fatalf("path %v not found in %s", path, string(b))
		}
		cur, ok = m[p]
		if !ok {
			t.Fatalf("missing key %q in %s", p, string(b))
		}
	}
}

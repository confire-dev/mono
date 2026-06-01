package hosts

import (
	"encoding/json"
	"testing"

	"github.com/confire-dev/confire/intercept"
)

func TestIsCursorHook(t *testing.T) {
	if IsCursorHook(map[string]any{"session_id": "abc"}) {
		t.Fatal("Claude-like payload should not match Cursor")
	}
	if !IsCursorHook(map[string]any{"cursor_version": "1.0", "session_id": "abc"}) {
		t.Fatal("expected Cursor payload to match")
	}
}

func TestDecodeCursorHookInput_ShellAlias(t *testing.T) {
	event := DecodeCursorHookInput(CursorHookInput{
		ConversationID: "conv-1",
		HookEventName:  "preToolUse",
		ToolName:       "Shell",
		ToolInput:      map[string]any{"command": "git status"},
	})
	if event.Tool == nil || event.Tool.Name != "Bash" {
		t.Fatalf("expected Shell to map to Bash, got %+v", event.Tool)
	}
}

func TestDecodeCursorHookInput_MCPDetection(t *testing.T) {
	event := DecodeCursorHookInput(CursorHookInput{
		ConversationID: "conv-1",
		HookEventName:  "postToolUse",
		ToolName:       "MCP: figma/get_file",
		ToolOutput:     `{"nodes":[]}`,
	})
	if event.Tool == nil || !event.Tool.IsMCP {
		t.Fatal("expected MCP tool")
	}

	native := DecodeCursorHookInput(CursorHookInput{
		ConversationID: "conv-1",
		HookEventName:  "postToolUse",
		ToolName:       "Shell",
		ToolOutput:     `{"stdout":"ok"}`,
	})
	if native.Tool == nil || native.Tool.IsMCP {
		t.Fatal("Shell should not be treated as MCP")
	}
}

func TestEncodeCursorPreToolResult_Block(t *testing.T) {
	out, ok, block := EncodeCursorPreToolResult(intercept.InterceptResult{
		Kind:   intercept.ResultBlock,
		Reason: "blocked",
	})
	if !ok || !block || out.Permission != "deny" {
		t.Fatalf("unexpected block output: %+v ok=%v block=%v", out, ok, block)
	}
}

func TestEncodeCursorResult_NativeSteerContext(t *testing.T) {
	steer := `[Confire post_tool steer]
tool=Shell
mode=native_unreplaceable
findings:
  secrets_redacted=1`
	out, ok := EncodeCursorResult(intercept.InterceptResult{
		Kind:    intercept.ResultAddContext,
		Context: steer,
	}, "Shell")
	if !ok || out.UpdatedMCPToolOutput != nil || out.AdditionalContext != steer {
		t.Fatalf("expected steer context, got %+v ok=%v", out, ok)
	}
}

func TestEncodeCursorResult_MCPReplace(t *testing.T) {
	out, ok := EncodeCursorResult(intercept.InterceptResult{
		Kind:       intercept.ResultReplaceOutput,
		ToolOutput: map[string]any{"trimmed": true},
	}, "MCP: figma/get_file")
	if !ok || out.UpdatedMCPToolOutput == nil {
		t.Fatalf("expected MCP output replacement, got %+v ok=%v", out, ok)
	}
	data, _ := json.Marshal(out)
	if !json.Valid(data) {
		t.Fatalf("invalid json: %s", data)
	}
}

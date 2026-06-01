package hosts

import (
	"encoding/json"
	"testing"

	"github.com/confire-dev/confire/intercept"
)

func TestIsVSCodeHook(t *testing.T) {
	if IsVSCodeHook(map[string]any{"session_id": "abc"}) {
		t.Fatal("Claude payload should not match VS Code")
	}
	if !IsVSCodeHook(map[string]any{"sessionId": "abc", "hookEventName": "PreToolUse"}) {
		t.Fatal("expected VS Code payload to match")
	}
	if IsVSCodeHook(map[string]any{"sessionId": "abc", "conversation_id": "x"}) {
		t.Fatal("Cursor payload should not match VS Code")
	}
}

func TestDecodeVSCodeHookInput_TerminalAlias(t *testing.T) {
	event := DecodeVSCodeHookInput(VSCodeHookInput{
		SessionID:     "s1",
		HookEventName: "PreToolUse",
		ToolName:      "runTerminalCommand",
		ToolInput:     map[string]any{"command": "git push --force"},
	})
	if event.Tool == nil || event.Tool.Name != "Bash" {
		t.Fatalf("expected runTerminalCommand to map to Bash, got %+v", event.Tool)
	}
}

func TestEncodeVSCodePreToolResult_Deny(t *testing.T) {
	out, ok := EncodeVSCodePreToolResult(intercept.InterceptResult{
		Kind:   intercept.ResultBlock,
		Reason: "blocked",
	}, "PreToolUse")
	if !ok || out.HookSpecificOutput == nil {
		t.Fatalf("unexpected output: %+v ok=%v", out, ok)
	}
	if out.HookSpecificOutput.PermissionDecision != "deny" {
		t.Fatalf("expected deny, got %+v", out.HookSpecificOutput)
	}
	data, _ := json.Marshal(out)
	if !json.Valid(data) {
		t.Fatalf("invalid json: %s", data)
	}
}

func TestEncodeVSCodeResult_SteerContext(t *testing.T) {
	steer := "[Confire post_tool steer]\ntool=runTerminalCommand"
	out, ok := EncodeVSCodeResult(intercept.InterceptResult{
		Kind:    intercept.ResultAddContext,
		Context: steer,
	}, "PostToolUse")
	if !ok || out.HookSpecificOutput == nil || out.HookSpecificOutput.AdditionalContext != steer {
		t.Fatalf("expected steer context, got %+v ok=%v", out, ok)
	}
}

func TestDefaultCapabilities_VSCodeMatchesCursor(t *testing.T) {
	vs := DefaultCapabilities("vscode")
	cu := DefaultCapabilities("cursor")
	if vs != cu {
		t.Fatalf("vscode and cursor caps should match: vscode=%+v cursor=%+v", vs, cu)
	}
}

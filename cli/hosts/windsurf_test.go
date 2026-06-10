package hosts

import (
	"encoding/json"
	"testing"

	"github.com/confire-dev/confire/intercept"
)

func TestWindsurfRealisticFormats(t *testing.T) {
	t.Run("PreToolUse block", func(t *testing.T) {
		raw := map[string]any{
			"hook_event_name": "PreToolUse",
			"tool_name":       "exec",
			"tool_input":      map[string]any{"command": "git push --force"},
			"session_id":      "sess-1",
			"cwd":             "/project",
		}
		var in WindsurfHookInput
		mustRemarshal(t, raw, &in)
		event := DecodeWindsurfHookInput(in)
		if event.Host != "windsurf" || event.Phase != intercept.PhaseToolPre {
			t.Fatalf("unexpected event: %+v", event)
		}
		out, shouldWrite, shouldBlock := EncodeWindsurfPreToolResult(intercept.InterceptResult{
			Kind:   intercept.ResultReview,
			Reason: "CONFIRE REVIEW REQUIRED",
		})
		if !shouldWrite || !shouldBlock {
			t.Fatalf("expected write+block, got shouldWrite=%v shouldBlock=%v", shouldWrite, shouldBlock)
		}
		if out.Decision != "block" {
			t.Fatalf("expected decision=block, got %q", out.Decision)
		}
		assertJSONKeys(t, out, "decision", "reason")
	})

	t.Run("PreToolUse warn allow", func(t *testing.T) {
		out, shouldWrite, shouldBlock := EncodeWindsurfPreToolResult(intercept.InterceptResult{
			Kind:    intercept.ResultWarn,
			Context: "advisory context",
		})
		if !shouldWrite || shouldBlock {
			t.Fatalf("warn should write but not block")
		}
		if out.Decision != "approve" {
			t.Fatalf("expected decision=approve, got %q", out.Decision)
		}
	})

	t.Run("PostToolUse with context", func(t *testing.T) {
		raw := map[string]any{
			"hook_event_name": "PostToolUse",
			"tool_name":       "exec",
			"tool_output":     "some output",
			"session_id":      "sess-1",
		}
		var in WindsurfHookInput
		mustRemarshal(t, raw, &in)
		event := DecodeWindsurfHookInput(in)
		if event.Phase != intercept.PhaseToolPost {
			t.Fatalf("expected PhaseToolPost, got %v", event.Phase)
		}
		out, ok := EncodeWindsurfResult(intercept.InterceptResult{
			Kind:    intercept.ResultAddContext,
			Context: "secret_like_value_detected",
		})
		if !ok || out.Decision != "approve" {
			t.Fatalf("unexpected post result: %+v ok=%v", out, ok)
		}
	})

	t.Run("PostToolUse passthrough", func(t *testing.T) {
		_, ok := EncodeWindsurfResult(intercept.InterceptResult{Kind: intercept.ResultPassthrough})
		if ok {
			t.Fatal("passthrough should not write")
		}
	})

	t.Run("JSON output round-trip", func(t *testing.T) {
		out := WindsurfHookOutput{Decision: "block", Reason: "test reason"}
		data, err := json.Marshal(out)
		if err != nil || !json.Valid(data) {
			t.Fatalf("invalid JSON: %v", err)
		}
	})
}

// fixture_test.go verifies that real client hook payloads (stored in testdata/hooks/)
// decode to the correct canonical InterceptEvent and produce valid JSON output.
// These fixtures are the ground truth for each client's hook schema.
// When a client ships a new hook format, update the fixture and this test together.
package hosts

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/confire-dev/confire/intercept"
)

func loadFixture(t *testing.T, name string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "hooks", name))
	if err != nil {
		t.Fatalf("fixture %s not found: %v", name, err)
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("fixture %s: invalid JSON: %v", name, err)
	}
	return raw
}

// ── Claude Code ───────────────────────────────────────────────────────────────

func TestFixture_Claude_PreToolUse(t *testing.T) {
	raw := loadFixture(t, "claude-pretooluse.json")
	if IsCursorHook(raw) || IsVSCodeHook(raw) {
		t.Fatal("claude fixture misdetected as another host")
	}
	var in HookInput
	mustRemarshal(t, raw, &in)
	event := DecodeHookInput(in)
	if event.Host != "claude-code" { t.Errorf("host: got %q", event.Host) }
	if event.Phase != intercept.PhaseToolPre { t.Errorf("phase: got %v", event.Phase) }
	if event.Tool == nil { t.Fatal("tool must not be nil") }
	if event.Tool.Name != "Bash" { t.Errorf("tool name: got %q", event.Tool.Name) }
	if event.Session.CWD != "/project" { t.Errorf("cwd: got %q", event.Session.CWD) }
}

func TestFixture_Claude_PostToolUse(t *testing.T) {
	raw := loadFixture(t, "claude-posttooluse.json")
	var in HookInput
	mustRemarshal(t, raw, &in)
	event := DecodeHookInput(in)
	if event.Phase != intercept.PhaseToolPost { t.Errorf("phase: got %v", event.Phase) }
	if event.Tool.Output == nil { t.Error("tool output must be present") }
}

// ── Cursor ────────────────────────────────────────────────────────────────────

func TestFixture_Cursor_PreToolUse(t *testing.T) {
	raw := loadFixture(t, "cursor-pretooluse.json")
	if !IsCursorHook(raw) { t.Fatal("cursor fixture not detected as cursor") }
	var in CursorHookInput
	mustRemarshal(t, raw, &in)
	event := DecodeCursorHookInput(in)
	if event.Host != "cursor" { t.Errorf("host: got %q", event.Host) }
	if event.Phase != intercept.PhaseToolPre { t.Errorf("phase: got %v", event.Phase) }
	if event.Tool.Name != "Bash" { t.Errorf("Shell should map to Bash, got %q", event.Tool.Name) }
	if event.Session.ID != "conv-abc123" { t.Errorf("session ID should use conversation_id") }
}

func TestFixture_Cursor_PostToolUse_MCP(t *testing.T) {
	raw := loadFixture(t, "cursor-posttooluse-mcp.json")
	if !IsCursorHook(raw) { t.Fatal("cursor fixture not detected as cursor") }
	var in CursorHookInput
	mustRemarshal(t, raw, &in)
	event := DecodeCursorHookInput(in)
	if event.Phase != intercept.PhaseToolPost { t.Errorf("phase: got %v", event.Phase) }
	if !event.Tool.IsMCP { t.Error("figma tool should be detected as MCP") }
	if event.Tool.MCPServer != "figma" { t.Errorf("mcp server: got %q", event.Tool.MCPServer) }
	if event.Tool.Output == nil { t.Error("MCP tool output should be parsed from JSON string") }
}

// ── VS Code ───────────────────────────────────────────────────────────────────

func TestFixture_VSCode_PreToolUse(t *testing.T) {
	raw := loadFixture(t, "vscode-pretooluse.json")
	if IsCursorHook(raw) { t.Fatal("vscode fixture misdetected as cursor") }
	if !IsVSCodeHook(raw) { t.Fatal("vscode fixture not detected as vscode") }
	var in VSCodeHookInput
	mustRemarshal(t, raw, &in)
	event := DecodeVSCodeHookInput(in)
	if event.Host != "vscode" { t.Errorf("host: got %q", event.Host) }
	if event.Phase != intercept.PhaseToolPre { t.Errorf("phase: got %v", event.Phase) }
	if event.Tool.Name != "Bash" { t.Errorf("runTerminalCommand should map to Bash, got %q", event.Tool.Name) }
	if event.Session.ID != "vscode-session-identifier" { t.Errorf("session ID: got %q", event.Session.ID) }
}

func TestFixture_VSCode_PostToolUse(t *testing.T) {
	raw := loadFixture(t, "vscode-posttooluse.json")
	if !IsVSCodeHook(raw) { t.Fatal("vscode fixture not detected") }
	var in VSCodeHookInput
	mustRemarshal(t, raw, &in)
	event := DecodeVSCodeHookInput(in)
	if event.Phase != intercept.PhaseToolPost { t.Errorf("phase: got %v", event.Phase) }
	if event.Tool.Output == nil { t.Error("tool output must be present") }
}

// ── Windsurf ──────────────────────────────────────────────────────────────────

func TestFixture_Windsurf_PreToolUse(t *testing.T) {
	raw := loadFixture(t, "windsurf-pretooluse.json")
	if IsCursorHook(raw) || IsVSCodeHook(raw) {
		t.Fatal("windsurf fixture misdetected as cursor or vscode")
	}
	var in WindsurfHookInput
	mustRemarshal(t, raw, &in)
	event := DecodeWindsurfHookInput(in)
	if event.Host != "windsurf" { t.Errorf("host: got %q", event.Host) }
	if event.Phase != intercept.PhaseToolPre { t.Errorf("phase: got %v", event.Phase) }
	if event.Tool.Name != "exec" { t.Errorf("tool name: got %q", event.Tool.Name) }
	if event.Session.ID != "windsurf-sess-1" { t.Errorf("session ID: got %q", event.Session.ID) }
}

func TestFixture_Windsurf_PostToolUse(t *testing.T) {
	raw := loadFixture(t, "windsurf-posttooluse.json")
	var in WindsurfHookInput
	mustRemarshal(t, raw, &in)
	event := DecodeWindsurfHookInput(in)
	if event.Phase != intercept.PhaseToolPost { t.Errorf("phase: got %v", event.Phase) }
}

// ── Codex ─────────────────────────────────────────────────────────────────────

func TestFixture_Codex_PreToolUse(t *testing.T) {
	raw := loadFixture(t, "codex-pretooluse.json")
	if IsCursorHook(raw) || IsVSCodeHook(raw) {
		t.Fatal("codex fixture misdetected")
	}
	var in CodexHookInput
	mustRemarshal(t, raw, &in)
	event := DecodeCodexHookInput(in)
	if event.Host != "codex" { t.Errorf("host: got %q", event.Host) }
	if event.Phase != intercept.PhaseToolPre { t.Errorf("phase: got %v", event.Phase) }
	if event.Tool.Name != "bash" { t.Errorf("tool name: got %q", event.Tool.Name) }
	if in.Model != "codex-mini-latest" { t.Errorf("model: got %q", in.Model) }
}

func TestFixture_Codex_PostToolUse(t *testing.T) {
	raw := loadFixture(t, "codex-posttooluse.json")
	var in CodexHookInput
	mustRemarshal(t, raw, &in)
	event := DecodeCodexHookInput(in)
	if event.Phase != intercept.PhaseToolPost { t.Errorf("phase: got %v", event.Phase) }
}

// ── Schema uniqueness: each fixture is detected by exactly one host ───────────

func TestFixture_HostDetectionExclusive(t *testing.T) {
	cases := []struct {
		file       string
		wantCursor bool
		wantVSCode bool
	}{
		{"claude-pretooluse.json", false, false},
		{"cursor-pretooluse.json", true, false},
		{"vscode-pretooluse.json", false, true},
		{"windsurf-pretooluse.json", false, false},
		{"codex-pretooluse.json", false, false},
	}
	for _, tc := range cases {
		raw := loadFixture(t, tc.file)
		gotCursor := IsCursorHook(raw)
		gotVSCode := IsVSCodeHook(raw)
		if gotCursor != tc.wantCursor {
			t.Errorf("%s: IsCursorHook = %v, want %v", tc.file, gotCursor, tc.wantCursor)
		}
		if gotVSCode != tc.wantVSCode {
			t.Errorf("%s: IsVSCodeHook = %v, want %v", tc.file, gotVSCode, tc.wantVSCode)
		}
	}
}

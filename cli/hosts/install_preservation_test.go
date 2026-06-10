// install_preservation_test.go — verifies that install and uninstall
// never drop unknown top-level keys, never reorder keys, never strip
// unknown fields from existing entries, and never remove non-confire hooks.
//
// These tests apply to every client that modifies a JSON config file.
package hosts

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeJSON writes pretty-printed JSON to path, creating parent dirs.
func writeJSON(t *testing.T, path string, v any) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(b, '\n'), 0644); err != nil {
		t.Fatal(err)
	}
}

func readString(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(b)
}

// ── Claude Code ───────────────────────────────────────────────────────────────

func TestClaudeCode_PreservesUnknownTopLevelKeys(t *testing.T) {
	_, cleanup := settingsFixture(t, map[string]any{
		"theme":          "dark",
		"enabledPlugins": map[string]any{"my-plugin": true},
		"customArray":    []any{"a", "b"},
	})
	defer cleanup()

	if err := installClaudeCodeHooks("/usr/bin/confire"); err != nil {
		t.Fatalf("install: %v", err)
	}

	content := readString(t, claudeSettingsPath())
	for _, want := range []string{`"theme"`, `"dark"`, `"enabledPlugins"`, `"my-plugin"`, `"customArray"`} {
		if !strings.Contains(content, want) {
			t.Errorf("after install: missing %q", want)
		}
	}

	if err := uninstallClaudeCodeHooks(); err != nil {
		t.Fatalf("uninstall: %v", err)
	}
	content = readString(t, claudeSettingsPath())
	for _, want := range []string{`"theme"`, `"dark"`, `"enabledPlugins"`, `"customArray"`} {
		if !strings.Contains(content, want) {
			t.Errorf("after uninstall: missing %q", want)
		}
	}
}

func TestClaudeCode_PreservesUnknownFieldsInExistingEntries(t *testing.T) {
	_, cleanup := settingsFixture(t, map[string]any{
		"hooks": map[string]any{
			"PostToolUse": []any{
				map[string]any{
					"matcher":       "Bash",
					"hooks":         []any{map[string]any{"type": "command", "command": "/usr/bin/audit"}},
					"customField":   "should-survive",
					"anotherField":  42,
				},
			},
		},
	})
	defer cleanup()

	installClaudeCodeHooks("/usr/bin/confire")

	content := readString(t, claudeSettingsPath())
	for _, want := range []string{`"customField"`, `"should-survive"`, `"anotherField"`, `/usr/bin/audit`} {
		if !strings.Contains(content, want) {
			t.Errorf("unknown field dropped: missing %q", want)
		}
	}
}

func TestClaudeCode_PreservesKeyOrderOnInstallAndUninstall(t *testing.T) {
	dir := t.TempDir()
	orig := os.Getenv("HOME")
	os.Setenv("HOME", dir)
	defer os.Setenv("HOME", orig)

	p := filepath.Join(dir, ".claude", "settings.json")
	os.MkdirAll(filepath.Dir(p), 0700)
	// "zLast" intentionally after "theme" — non-alphabetical order must survive.
	os.WriteFile(p, []byte(`{"zLast":"z","theme":"dark","alpha":"a"}`), 0600)

	installClaudeCodeHooks("/usr/bin/confire")
	content := readString(t, p)

	zPos := strings.Index(content, `"zLast"`)
	themePos := strings.Index(content, `"theme"`)
	alphaPos := strings.Index(content, `"alpha"`)
	hooksPos := strings.Index(content, `"hooks"`)

	if !(zPos < themePos && themePos < alphaPos && alphaPos < hooksPos) {
		t.Errorf("key order not preserved: zLast=%d theme=%d alpha=%d hooks=%d", zPos, themePos, alphaPos, hooksPos)
	}
}

// ── Cursor ────────────────────────────────────────────────────────────────────

func TestCursor_PreservesUnknownTopLevelKey(t *testing.T) {
	home := redirectHome(t)
	p := filepath.Join(home, ".cursor", "hooks.json")
	writeJSON(t, p, map[string]any{
		"version":      1,
		"customSetting": "keep-me",
		"hooks": map[string]any{
			"preToolUse": []any{
				map[string]any{"command": "/usr/bin/other", "timeout": 30, "metadata": "keep-field"},
			},
		},
	})

	if err := installCursorHooksAt(p, "/usr/bin/confire"); err != nil {
		t.Fatalf("install: %v", err)
	}
	content := readString(t, p)
	for _, want := range []string{`"customSetting"`, `"keep-me"`, `/usr/bin/other`, `"timeout"`, `"metadata"`, `"keep-field"`} {
		if !strings.Contains(content, want) {
			t.Errorf("after install: missing %q", want)
		}
	}

	if err := uninstallCursorHooksAt(p); err != nil {
		t.Fatalf("uninstall: %v", err)
	}
	content = readString(t, p)
	for _, want := range []string{`"customSetting"`, `"keep-me"`, `/usr/bin/other`, `"timeout"`, `"metadata"`} {
		if !strings.Contains(content, want) {
			t.Errorf("after uninstall: missing %q", want)
		}
	}
	if strings.Contains(content, "confire") {
		t.Error("confire entry not removed by uninstall")
	}
}

func TestCursor_PreservesHooksKeyOrder(t *testing.T) {
	home := redirectHome(t)
	p := filepath.Join(home, ".cursor", "hooks.json")
	// sessionEnd comes before preToolUse — non-alphabetical order must survive.
	os.MkdirAll(filepath.Dir(p), 0700)
	os.WriteFile(p, []byte(`{"version":1,"hooks":{"sessionEnd":[],"preToolUse":[]}}`), 0644)

	installCursorHooksAt(p, "/usr/bin/confire")
	content := readString(t, p)

	sessionEndPos := strings.Index(content, `"sessionEnd"`)
	preToolPos := strings.Index(content, `"preToolUse"`)
	if sessionEndPos < 0 || preToolPos < 0 || sessionEndPos > preToolPos {
		t.Errorf("hooks key order not preserved: sessionEnd=%d preToolUse=%d\n%s", sessionEndPos, preToolPos, content)
	}
}

func TestCursor_NonConfireEntriesPreserved(t *testing.T) {
	home := redirectHome(t)
	p := filepath.Join(home, ".cursor", "hooks.json")
	writeJSON(t, p, map[string]any{
		"version": 1,
		"hooks": map[string]any{
			"preToolUse": []any{
				map[string]any{"command": "/usr/bin/other-tool", "timeout": 30},
			},
		},
	})

	installCursorHooksAt(p, "/usr/bin/confire")
	uninstallCursorHooksAt(p)

	content := readString(t, p)
	if !strings.Contains(content, `/usr/bin/other-tool`) {
		t.Error("non-confire entry removed by uninstall")
	}
	if strings.Contains(content, "confire") {
		t.Error("confire entry not removed")
	}
}

// ── VS Code ───────────────────────────────────────────────────────────────────

func TestVSCode_PreservesUnknownTopLevelKey(t *testing.T) {
	home := redirectHome(t)
	p := filepath.Join(home, ".copilot", "hooks", "confire.json")
	writeJSON(t, p, map[string]any{
		"metadata": map[string]any{"version": "2.0", "author": "me"},
		"hooks": map[string]any{
			"PreToolUse": []any{
				map[string]any{"type": "command", "command": "/usr/bin/other", "customField": true},
			},
		},
	})

	if err := installVSCodeHooks("/usr/bin/confire", false); err != nil {
		t.Fatalf("install: %v", err)
	}
	content := readString(t, p)
	for _, want := range []string{`"metadata"`, `"version"`, `"2.0"`, `"author"`, `/usr/bin/other`, `"customField"`} {
		if !strings.Contains(content, want) {
			t.Errorf("after install: missing %q", want)
		}
	}

	uninstallVSCodeHooks(false)
	content = readString(t, p)
	if !strings.Contains(content, `"metadata"`) {
		t.Error("metadata key dropped by uninstall")
	}
	if !strings.Contains(content, `/usr/bin/other`) {
		t.Error("non-confire entry removed by uninstall")
	}
	if strings.Contains(content, "confire") {
		t.Error("confire entry not removed")
	}
}

func TestVSCode_PreservesHooksKeyOrder(t *testing.T) {
	home := redirectHome(t)
	p := filepath.Join(home, ".copilot", "hooks", "confire.json")
	os.MkdirAll(filepath.Dir(p), 0755)
	// Stop before PreToolUse — must stay that way.
	os.WriteFile(p, []byte(`{"hooks":{"Stop":[],"PreToolUse":[]}}`), 0644)

	installVSCodeHooks("/usr/bin/confire", false)
	content := readString(t, p)

	stopPos := strings.Index(content, `"Stop"`)
	prePos := strings.Index(content, `"PreToolUse"`)
	if stopPos < 0 || prePos < 0 || stopPos > prePos {
		t.Errorf("hooks key order not preserved: Stop=%d PreToolUse=%d\n%s", stopPos, prePos, content)
	}
}

// ── Windsurf ──────────────────────────────────────────────────────────────────

func TestWindsurf_PreservesUnknownTopLevelKey(t *testing.T) {
	home := redirectHome(t)
	p := filepath.Join(home, ".windsurf", "hooks.json")
	// Windsurf uses a flat format: top-level keys ARE events + any unknown fields.
	writeJSON(t, p, map[string]any{
		"metadata":   "global-setting",
		"PreToolUse": []any{map[string]any{"command": "/usr/bin/other", "onError": "retry"}},
	})

	if err := installWindsurfHooks("/usr/bin/confire", false); err != nil {
		t.Fatalf("install: %v", err)
	}
	content := readString(t, p)
	for _, want := range []string{`"metadata"`, `"global-setting"`, `/usr/bin/other`, `"onError"`, `"retry"`} {
		if !strings.Contains(content, want) {
			t.Errorf("after install: missing %q", want)
		}
	}

	uninstallWindsurfHooks(false)
	content = readString(t, p)
	if !strings.Contains(content, `"metadata"`) {
		t.Error("unknown top-level key dropped by uninstall")
	}
	if !strings.Contains(content, `/usr/bin/other`) {
		t.Error("non-confire entry removed by uninstall")
	}
	if strings.Contains(content, "confire") {
		t.Error("confire entry not removed")
	}
}

func TestWindsurf_PreservesEventKeyOrder(t *testing.T) {
	home := redirectHome(t)
	p := filepath.Join(home, ".windsurf", "hooks.json")
	os.MkdirAll(filepath.Dir(p), 0700)
	// SessionEnd before PreToolUse — non-alphabetical, must stay that way.
	os.WriteFile(p, []byte(`{"SessionEnd":[],"PreToolUse":[]}`), 0644)

	installWindsurfHooks("/usr/bin/confire", false)
	content := readString(t, p)

	sessionPos := strings.Index(content, `"SessionEnd"`)
	prePos := strings.Index(content, `"PreToolUse"`)
	if sessionPos < 0 || prePos < 0 || sessionPos > prePos {
		t.Errorf("event key order not preserved: SessionEnd=%d PreToolUse=%d\n%s", sessionPos, prePos, content)
	}
}

// ── Codex ─────────────────────────────────────────────────────────────────────

func TestCodex_PreservesUnknownTopLevelKey(t *testing.T) {
	home := redirectHome(t)
	p := filepath.Join(home, ".codex", "hooks.json")
	writeJSON(t, p, map[string]any{
		"version":  1,
		"settings": map[string]any{"debug": true},
		"hooks": map[string]any{
			"PreToolUse": []any{
				map[string]any{"command": "/usr/bin/other", "matcher": "bash", "async": true},
			},
		},
	})

	if err := installCodexHooks("/usr/bin/confire", false); err != nil {
		t.Fatalf("install: %v", err)
	}
	content := readString(t, p)
	for _, want := range []string{`"settings"`, `"debug"`, `/usr/bin/other`, `"matcher"`, `"async"`} {
		if !strings.Contains(content, want) {
			t.Errorf("after install: missing %q", want)
		}
	}

	uninstallCodexHooks(false)
	content = readString(t, p)
	if !strings.Contains(content, `"settings"`) {
		t.Error("settings key dropped by uninstall")
	}
	if !strings.Contains(content, `/usr/bin/other`) {
		t.Error("non-confire entry removed by uninstall")
	}
	if strings.Contains(content, "confire") {
		t.Error("confire entry not removed")
	}
}

func TestCodex_PreservesHooksKeyOrder(t *testing.T) {
	home := redirectHome(t)
	p := filepath.Join(home, ".codex", "hooks.json")
	os.MkdirAll(filepath.Dir(p), 0700)
	// SessionEnd before PreToolUse — non-alphabetical, must stay.
	os.WriteFile(p, []byte(`{"version":1,"hooks":{"SessionEnd":[],"PreToolUse":[]}}`), 0644)

	installCodexHooks("/usr/bin/confire", false)
	content := readString(t, p)

	sessionPos := strings.Index(content, `"SessionEnd"`)
	prePos := strings.Index(content, `"PreToolUse"`)
	if sessionPos < 0 || prePos < 0 || sessionPos > prePos {
		t.Errorf("hooks key order not preserved: SessionEnd=%d PreToolUse=%d\n%s", sessionPos, prePos, content)
	}
}

// ── Cross-client: idempotency ─────────────────────────────────────────────────

func TestAllClients_InstallIdempotent(t *testing.T) {
	home := redirectHome(t)

	cases := []struct {
		name    string
		install func() error
		check   func() bool
	}{
		{
			"cursor",
			func() error {
				return installCursorHooksAt(filepath.Join(home, ".cursor", "hooks.json"), "/usr/bin/confire")
			},
			func() bool { return isCursorHookInstalled(false) },
		},
		{
			"windsurf",
			func() error { return installWindsurfHooks("/usr/bin/confire", false) },
			func() bool  { return isWindsurfHookInstalled(false) },
		},
		{
			"codex",
			func() error { return installCodexHooks("/usr/bin/confire", false) },
			func() bool  { return isCodexHookInstalled(false) },
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Install three times — no duplicates should appear.
			for i := 0; i < 3; i++ {
				if err := tc.install(); err != nil {
					t.Fatalf("install %d: %v", i+1, err)
				}
			}
			if !tc.check() {
				t.Fatal("not installed after 3 installs")
			}
		})
	}
}

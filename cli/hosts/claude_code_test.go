package hosts

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// settingsFixture creates a temp HOME with a .claude/settings.json and returns
// a cleanup function. Tests call it to redirect claudeSettingsPath().
func settingsFixture(t *testing.T, content map[string]interface{}) (settingsPath string, cleanup func()) {
	t.Helper()
	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	if err := os.MkdirAll(claudeDir, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(claudeDir, "settings.json")
	if content != nil {
		data, _ := json.MarshalIndent(content, "", "  ")
		if err := os.WriteFile(path, data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", dir)
	return path, func() { os.Setenv("HOME", origHome) }
}

// ── claudeCodeHookInstalled ───────────────────────────────────────────────

func TestHookInstalled_EmptyFile(t *testing.T) {
	_, cleanup := settingsFixture(t, map[string]interface{}{})
	defer cleanup()

	if claudeCodeHookInstalled() {
		t.Error("expected false for empty settings")
	}
}

func TestHookInstalled_MissingFile(t *testing.T) {
	dir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", dir)
	defer os.Setenv("HOME", origHome)

	if claudeCodeHookInstalled() {
		t.Error("expected false when settings.json does not exist")
	}
}

// Simulates the real on-disk format: binary path is double-quoted.
func TestHookInstalled_QuotedBinaryPath(t *testing.T) {
	_, cleanup := settingsFixture(t, map[string]interface{}{
		"hooks": map[string]interface{}{
			"PostToolUse": []interface{}{
				map[string]interface{}{
					"matcher": ".*",
					"hooks": []interface{}{
						map[string]interface{}{
							"type":    "command",
							"command": `"/usr/local/bin/confire" hook`,
						},
					},
				},
			},
		},
	})
	defer cleanup()

	if !claudeCodeHookInstalled() {
		t.Error("expected true: quoted binary path should be detected")
	}
}

// The bug we fixed: previous code checked "confire hook" (no quote),
// which did not match the stored `"/path/confire" hook` format.
func TestHookInstalled_UnquotedBinaryPath(t *testing.T) {
	_, cleanup := settingsFixture(t, map[string]interface{}{
		"hooks": map[string]interface{}{
			"PostToolUse": []interface{}{
				map[string]interface{}{
					"matcher": ".*",
					"hooks": []interface{}{
						map[string]interface{}{
							"type":    "command",
							"command": "confire hook", // unquoted — still detectable
						},
					},
				},
			},
		},
	})
	defer cleanup()

	if !claudeCodeHookInstalled() {
		t.Error("expected true: unquoted binary name should also be detected")
	}
}

func TestHookInstalled_UnrelatedHooks(t *testing.T) {
	_, cleanup := settingsFixture(t, map[string]interface{}{
		"hooks": map[string]interface{}{
			"PostToolUse": []interface{}{
				map[string]interface{}{
					"matcher": ".*",
					"hooks": []interface{}{
						map[string]interface{}{
							"type":    "command",
							"command": "/usr/local/bin/some-other-tool",
						},
					},
				},
			},
		},
	})
	defer cleanup()

	if claudeCodeHookInstalled() {
		t.Error("expected false: unrelated hook should not match")
	}
}

// ── installClaudeCodeHooks ────────────────────────────────────────────────

func TestInstall_FreshSettings(t *testing.T) {
	_, cleanup := settingsFixture(t, nil) // no file at all
	defer cleanup()

	if err := installClaudeCodeHooks("/usr/local/bin/confire"); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	if !claudeCodeHookInstalled() {
		t.Error("hook should be detected after install")
	}

	// Verify exactly one PostToolUse entry
	data, _ := os.ReadFile(claudeSettingsPath())
	var s map[string]interface{}
	json.Unmarshal(data, &s)
	hooks := s["hooks"].(map[string]interface{})
	post := hooks["PostToolUse"].([]interface{})
	if len(post) != 1 {
		t.Errorf("expected 1 PostToolUse entry, got %d", len(post))
	}
}

func TestInstall_Idempotent_NoDuplicates(t *testing.T) {
	_, cleanup := settingsFixture(t, map[string]interface{}{})
	defer cleanup()

	// Install three times — should still have exactly one entry each
	for i := 0; i < 3; i++ {
		if err := installClaudeCodeHooks("/usr/local/bin/confire"); err != nil {
			t.Fatalf("install %d failed: %v", i+1, err)
		}
	}

	data, _ := os.ReadFile(claudeSettingsPath())
	var s map[string]interface{}
	json.Unmarshal(data, &s)
	hooks := s["hooks"].(map[string]interface{})

	post := hooks["PostToolUse"].([]interface{})
	if len(post) != 1 {
		t.Errorf("expected 1 PostToolUse entry after 3 installs, got %d", len(post))
	}

	pre := hooks["PreToolUse"].([]interface{})
	if len(pre) != 1 {
		t.Errorf("expected 1 PreToolUse entry after 3 installs, got %d", len(pre))
	}
}

func TestInstall_PreservesExistingNonConfire(t *testing.T) {
	_, cleanup := settingsFixture(t, map[string]interface{}{
		"hooks": map[string]interface{}{
			"PostToolUse": []interface{}{
				map[string]interface{}{
					"matcher": "Bash",
					"hooks": []interface{}{
						map[string]interface{}{"type": "command", "command": "/usr/bin/audit-tool"},
					},
				},
			},
		},
	})
	defer cleanup()

	if err := installClaudeCodeHooks("/usr/local/bin/confire"); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	data, _ := os.ReadFile(claudeSettingsPath())
	var s map[string]interface{}
	json.Unmarshal(data, &s)
	hooks := s["hooks"].(map[string]interface{})
	post := hooks["PostToolUse"].([]interface{})

	// Should have 2 entries: the existing audit-tool + new confire entry
	if len(post) != 2 {
		t.Errorf("expected 2 PostToolUse entries (existing + confire), got %d", len(post))
	}

	// Existing tool must be preserved
	found := false
	for _, e := range post {
		m := e.(map[string]interface{})
		for _, h := range toIfaceSlice(m["hooks"]) {
			hm := h.(map[string]interface{})
			if strings.Contains(hm["command"].(string), "audit-tool") {
				found = true
			}
		}
	}
	if !found {
		t.Error("pre-existing non-confire hook was removed during install")
	}
}

func TestInstall_CommandFormat(t *testing.T) {
	_, cleanup := settingsFixture(t, nil)
	defer cleanup()

	if err := installClaudeCodeHooks("/my/path/confire"); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	data, _ := os.ReadFile(claudeSettingsPath())
	var s map[string]interface{}
	json.Unmarshal(data, &s)
	hooks := s["hooks"].(map[string]interface{})
	post := hooks["PostToolUse"].([]interface{})
	m := post[0].(map[string]interface{})
	h := toIfaceSlice(m["hooks"])[0].(map[string]interface{})
	cmd := h["command"].(string)

	if cmd != `"/my/path/confire" hook` {
		t.Errorf("unexpected command format: %q", cmd)
	}
}

// ── Non-hooks fields are preserved ───────────────────────────────────────

// Core safety test: install + uninstall must leave all non-hooks fields intact.
// This guards against the map[string]interface{} re-serialization bug where
// key ordering, value types, or nested structures could be silently changed.
func TestInstall_PreservesNonHooksFields(t *testing.T) {
	original := map[string]interface{}{
		"enabledPlugins": map[string]interface{}{
			"frontend-design@claude-plugins-official": true,
			"swift-lsp@claude-plugins-official":       true,
		},
		"theme": "dark",
		"someNumber": 42,
	}
	_, cleanup := settingsFixture(t, original)
	defer cleanup()

	if err := installClaudeCodeHooks("/usr/local/bin/confire"); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	data, _ := os.ReadFile(claudeSettingsPath())
	var result map[string]interface{}
	json.Unmarshal(data, &result)

	plugins, _ := result["enabledPlugins"].(map[string]interface{})
	if plugins["frontend-design@claude-plugins-official"] != true {
		t.Error("enabledPlugins frontend-design was not preserved")
	}
	if plugins["swift-lsp@claude-plugins-official"] != true {
		t.Error("enabledPlugins swift-lsp was not preserved")
	}
	if result["theme"] != "dark" {
		t.Errorf("theme field not preserved, got: %v", result["theme"])
	}
	// json numbers unmarshal as float64
	if result["someNumber"] != float64(42) {
		t.Errorf("someNumber not preserved, got: %v", result["someNumber"])
	}
}

// Key ordering: top-level keys must survive in their original order.
func TestInstall_PreservesTopLevelKeyOrder(t *testing.T) {
	dir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", dir)
	defer os.Setenv("HOME", origHome)

	// Write JSON with a specific key order that is NOT alphabetical.
	settingsPath := filepath.Join(dir, ".claude", "settings.json")
	os.MkdirAll(filepath.Dir(settingsPath), 0700)
	// "theme" comes before "enabledPlugins" — reverse of alphabetical.
	raw := []byte(`{"theme":"dark","enabledPlugins":{"plug-a":true},"zSetting":"last"}`)
	os.WriteFile(settingsPath, raw, 0600)

	if err := installClaudeCodeHooks("/usr/local/bin/confire"); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	result, _ := os.ReadFile(settingsPath)

	themePos := strings.Index(string(result), `"theme"`)
	pluginsPos := strings.Index(string(result), `"enabledPlugins"`)
	zPos := strings.Index(string(result), `"zSetting"`)
	hooksPos := strings.Index(string(result), `"hooks"`)

	if themePos < 0 || pluginsPos < 0 || zPos < 0 || hooksPos < 0 {
		t.Fatal("one or more keys missing from output")
	}
	// Original order: theme < enabledPlugins < zSetting; hooks was added so it comes after zSetting
	if !(themePos < pluginsPos && pluginsPos < zPos && zPos < hooksPos) {
		t.Errorf("key order not preserved: theme=%d enabledPlugins=%d zSetting=%d hooks=%d",
			themePos, pluginsPos, zPos, hooksPos)
	}
}

func TestUninstall_PreservesNonHooksFields(t *testing.T) {
	original := map[string]interface{}{
		"enabledPlugins": map[string]interface{}{
			"some-plugin@official": true,
		},
		"theme": "dark",
	}
	_, cleanup := settingsFixture(t, original)
	defer cleanup()

	installClaudeCodeHooks("/usr/local/bin/confire")
	if err := uninstallClaudeCodeHooks(); err != nil {
		t.Fatalf("uninstall failed: %v", err)
	}

	data, _ := os.ReadFile(claudeSettingsPath())
	var result map[string]interface{}
	json.Unmarshal(data, &result)

	plugins, _ := result["enabledPlugins"].(map[string]interface{})
	if plugins["some-plugin@official"] != true {
		t.Error("enabledPlugins not preserved after uninstall")
	}
	if result["theme"] != "dark" {
		t.Error("theme field not preserved after uninstall")
	}
}

// ── uninstallClaudeCodeHooks ──────────────────────────────────────────────

func TestUninstall_RemovesConfire(t *testing.T) {
	_, cleanup := settingsFixture(t, nil)
	defer cleanup()

	installClaudeCodeHooks("/usr/local/bin/confire")
	if !claudeCodeHookInstalled() {
		t.Fatal("hook should be installed before uninstall test")
	}

	if err := uninstallClaudeCodeHooks(); err != nil {
		t.Fatalf("uninstall failed: %v", err)
	}

	if claudeCodeHookInstalled() {
		t.Error("hook should be gone after uninstall")
	}
}

func TestUninstall_PreservesOtherHooks(t *testing.T) {
	_, cleanup := settingsFixture(t, map[string]interface{}{
		"hooks": map[string]interface{}{
			"PostToolUse": []interface{}{
				map[string]interface{}{
					"matcher": ".*",
					"hooks":   []interface{}{map[string]interface{}{"type": "command", "command": "/usr/bin/other-tool"}},
				},
			},
		},
	})
	defer cleanup()

	installClaudeCodeHooks("/usr/local/bin/confire")
	uninstallClaudeCodeHooks()

	data, _ := os.ReadFile(claudeSettingsPath())
	var s map[string]interface{}
	json.Unmarshal(data, &s)
	hooks := s["hooks"].(map[string]interface{})
	post := hooks["PostToolUse"].([]interface{})

	if len(post) != 1 {
		t.Errorf("expected 1 remaining entry (other-tool), got %d", len(post))
	}
	m := post[0].(map[string]interface{})
	h := toIfaceSlice(m["hooks"])[0].(map[string]interface{})
	if !strings.Contains(h["command"].(string), "other-tool") {
		t.Error("other-tool hook was removed during uninstall")
	}
}

// plugin_install_test.go verifies Install/IsInstalled/Uninstall for JS bridge plugin clients.
package hosts

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// redirectHome temporarily overrides HOME for hosts that use os.UserHomeDir().
func redirectHome(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	orig := os.Getenv("HOME")
	os.Setenv("HOME", dir)
	t.Cleanup(func() { os.Setenv("HOME", orig) })
	return dir
}

// ── Cline ─────────────────────────────────────────────────────────────────────

func TestCline_InstallWritesPlugin(t *testing.T) {
	home := redirectHome(t)

	// Create the Cline globalStorage path so Detect() would return true.
	globDir := filepath.Join(home, "Library", "Application Support", "Code", "User", "globalStorage", "saoudrizwan.claude-dev")
	if err := os.MkdirAll(globDir, 0755); err != nil {
		t.Fatal(err)
	}

	h := &ClineHost{}
	if err := h.Install(StrategyHooks, InstallOptions{BinaryPath: "/usr/local/bin/confire"}); err != nil {
		t.Fatalf("Install: %v", err)
	}

	pluginPath := clinePluginPath()
	data, err := os.ReadFile(pluginPath)
	if err != nil {
		t.Fatalf("plugin file not found at %s: %v", pluginPath, err)
	}

	content := string(data)
	for _, want := range []string{"confire", "hook", "--host", "cline", "beforeTool", "afterTool"} {
		if !strings.Contains(content, want) {
			t.Errorf("plugin content missing %q", want)
		}
	}
}

func TestCline_IsInstalled(t *testing.T) {
	home := redirectHome(t)
	globDir := filepath.Join(home, "Library", "Application Support", "Code", "User", "globalStorage", "saoudrizwan.claude-dev")
	os.MkdirAll(globDir, 0755)

	h := &ClineHost{}
	if h.IsInstalled(StrategyHooks) {
		t.Fatal("should not be installed before Install()")
	}

	h.Install(StrategyHooks, InstallOptions{})
	if !h.IsInstalled(StrategyHooks) {
		t.Fatal("should be installed after Install()")
	}
}

func TestCline_Uninstall(t *testing.T) {
	home := redirectHome(t)
	globDir := filepath.Join(home, "Library", "Application Support", "Code", "User", "globalStorage", "saoudrizwan.claude-dev")
	os.MkdirAll(globDir, 0755)

	h := &ClineHost{}
	h.Install(StrategyHooks, InstallOptions{})
	if err := h.Uninstall(StrategyHooks); err != nil {
		t.Fatalf("Uninstall: %v", err)
	}
	if h.IsInstalled(StrategyHooks) {
		t.Fatal("should not be installed after Uninstall()")
	}
}

func TestCline_InstallIdempotent(t *testing.T) {
	home := redirectHome(t)
	globDir := filepath.Join(home, "Library", "Application Support", "Code", "User", "globalStorage", "saoudrizwan.claude-dev")
	os.MkdirAll(globDir, 0755)

	h := &ClineHost{}
	for i := 0; i < 3; i++ {
		if err := h.Install(StrategyHooks, InstallOptions{}); err != nil {
			t.Fatalf("Install %d: %v", i+1, err)
		}
	}
	// Content must still be valid — no duplication.
	data, _ := os.ReadFile(clinePluginPath())
	count := strings.Count(string(data), "beforeTool")
	if count != 1 {
		t.Errorf("expected exactly 1 'beforeTool' in plugin, got %d", count)
	}
}

// ── OpenCode ──────────────────────────────────────────────────────────────────

func TestOpenCode_InstallWritesPlugin(t *testing.T) {
	redirectHome(t)
	h := &OpenCodeHost{}

	if err := h.Install(StrategyHooks, InstallOptions{}); err != nil {
		t.Fatalf("Install: %v", err)
	}

	data, err := os.ReadFile(opencodePluginPath(false))
	if err != nil {
		t.Fatalf("plugin not found: %v", err)
	}
	content := string(data)
	for _, want := range []string{"confire", "--host", "opencode", "tool.execute.before", "tool.execute.after"} {
		if !strings.Contains(content, want) {
			t.Errorf("plugin missing %q", want)
		}
	}
}

func TestOpenCode_LocalInstall(t *testing.T) {
	redirectHome(t)
	h := &OpenCodeHost{}

	if err := h.Install(StrategyHooks, InstallOptions{SettingsPath: "."}); err != nil {
		t.Fatalf("Install local: %v", err)
	}

	data, err := os.ReadFile(opencodePluginPath(true))
	if err != nil {
		t.Fatalf("local plugin not found: %v", err)
	}
	if !strings.Contains(string(data), "opencode") {
		t.Error("local plugin missing opencode reference")
	}

	t.Cleanup(func() { os.RemoveAll(".opencode") })
}

func TestOpenCode_IsInstalledAndUninstall(t *testing.T) {
	redirectHome(t)
	h := &OpenCodeHost{}

	if h.IsInstalled(StrategyHooks) {
		t.Fatal("should not be installed before Install()")
	}
	h.Install(StrategyHooks, InstallOptions{})
	if !h.IsInstalled(StrategyHooks) {
		t.Fatal("should be installed after Install()")
	}
	h.Uninstall(StrategyHooks)
	if h.IsInstalled(StrategyHooks) {
		t.Fatal("should not be installed after Uninstall()")
	}
}

// ── OpenClaw ──────────────────────────────────────────────────────────────────

func TestOpenClaw_InstallWritesPlugin(t *testing.T) {
	redirectHome(t)
	h := &OpenClawHost{}

	if err := h.Install(StrategyHooks, InstallOptions{}); err != nil {
		t.Fatalf("Install: %v", err)
	}

	data, err := os.ReadFile(openclawPluginPath(false))
	if err != nil {
		t.Fatalf("plugin not found: %v", err)
	}
	content := string(data)
	for _, want := range []string{"confire", "--host", "openclaw", "before_tool_call", "after_tool_call"} {
		if !strings.Contains(content, want) {
			t.Errorf("plugin missing %q", want)
		}
	}
}

func TestOpenClaw_IsInstalledAndUninstall(t *testing.T) {
	redirectHome(t)
	h := &OpenClawHost{}

	if h.IsInstalled(StrategyHooks) {
		t.Fatal("should not be installed before Install()")
	}
	h.Install(StrategyHooks, InstallOptions{})
	if !h.IsInstalled(StrategyHooks) {
		t.Fatal("should be installed after Install()")
	}
	h.Uninstall(StrategyHooks)
	if h.IsInstalled(StrategyHooks) {
		t.Fatal("should not be installed after Uninstall()")
	}
}

// ── Unsupported strategy ──────────────────────────────────────────────────────

func TestPluginHosts_RejectMCPProxy(t *testing.T) {
	hosts := []Host{&ClineHost{}, &OpenCodeHost{}, &OpenClawHost{}}
	for _, h := range hosts {
		if err := h.Install(StrategyMCPProxy, InstallOptions{}); err == nil {
			t.Errorf("%s: expected error for mcp-proxy strategy, got nil", h.ID())
		}
	}
}

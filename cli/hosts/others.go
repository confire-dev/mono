package hosts

// This file contains Host implementations for non-Claude-Code agents.
// They all use StrategyMCPProxy (the universal strategy for MCP-speaking tools).
// Detection logic is conservative: check for app bundles and MCP config files
// rather than generic binary names that might collide with unrelated tools.
//
// Installation via mcp-proxy is coming in a future release.
// setup.go shows these as "coming soon" until Install() is wired.

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// ── Cursor ────────────────────────────────────────────────────────────────

type CursorHost struct{}

func (h *CursorHost) ID() string    { return "cursor" }
func (h *CursorHost) Label() string { return "Cursor" }
func (h *CursorHost) Strategies() []Strategy  { return []Strategy{StrategyMCPProxy} }
func (h *CursorHost) Preferred() Strategy     { return StrategyMCPProxy }

func (h *CursorHost) Detect() bool {
	home, _ := os.UserHomeDir()
	switch runtime.GOOS {
	case "darwin":
		if _, err := os.Stat("/Applications/Cursor.app"); err == nil {
			return true
		}
	case "linux":
		if _, err := os.Stat(filepath.Join(home, ".cursor")); err == nil {
			return true
		}
	case "windows":
		if _, err := os.Stat(filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "Cursor")); err == nil {
			return true
		}
	}
	// Fallback: Cursor leaves a .cursor directory in home with mcp config
	_, err := os.Stat(filepath.Join(home, ".cursor", "mcp.json"))
	return err == nil
}

func (h *CursorHost) Install(s Strategy, opts InstallOptions) error {
	return fmt.Errorf("cursor mcp-proxy install coming soon — follow confire.dev for updates")
}
func (h *CursorHost) IsInstalled(_ Strategy) bool { return false }
func (h *CursorHost) Uninstall(_ Strategy) error  { return nil }
func (h *CursorHost) ComingSoon() bool             { return true }

// ── Cline (VS Code extension) ─────────────────────────────────────────────

type ClineHost struct{}

func (h *ClineHost) ID() string    { return "cline" }
func (h *ClineHost) Label() string { return "Cline" }
func (h *ClineHost) Strategies() []Strategy { return []Strategy{StrategyMCPProxy} }
func (h *ClineHost) Preferred() Strategy    { return StrategyMCPProxy }

func (h *ClineHost) Detect() bool {
	// Cline stores its config in VS Code's extension settings.
	// Check for the Cline extension data directory.
	home, _ := os.UserHomeDir()
	var paths []string
	switch runtime.GOOS {
	case "darwin":
		paths = []string{
			filepath.Join(home, "Library", "Application Support", "Code", "User", "globalStorage", "saoudrizwan.claude-dev"),
		}
	case "linux":
		paths = []string{
			filepath.Join(home, ".config", "Code", "User", "globalStorage", "saoudrizwan.claude-dev"),
		}
	case "windows":
		paths = []string{
			filepath.Join(os.Getenv("APPDATA"), "Code", "User", "globalStorage", "saoudrizwan.claude-dev"),
		}
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return true
		}
	}
	return false
}

func (h *ClineHost) Install(s Strategy, opts InstallOptions) error {
	return fmt.Errorf("cline mcp-proxy install coming soon — follow confire.dev for updates")
}
func (h *ClineHost) IsInstalled(_ Strategy) bool { return false }
func (h *ClineHost) Uninstall(_ Strategy) error  { return nil }
func (h *ClineHost) ComingSoon() bool             { return true }

// ── Windsurf ──────────────────────────────────────────────────────────────

type WindsurfHost struct{}

func (h *WindsurfHost) ID() string    { return "windsurf" }
func (h *WindsurfHost) Label() string { return "Windsurf" }
func (h *WindsurfHost) Strategies() []Strategy { return []Strategy{StrategyMCPProxy} }
func (h *WindsurfHost) Preferred() Strategy    { return StrategyMCPProxy }

func (h *WindsurfHost) Detect() bool {
	switch runtime.GOOS {
	case "darwin":
		_, err := os.Stat("/Applications/Windsurf.app")
		return err == nil
	case "linux":
		home, _ := os.UserHomeDir()
		_, err := os.Stat(filepath.Join(home, ".windsurf"))
		return err == nil
	}
	return false
}

func (h *WindsurfHost) Install(s Strategy, opts InstallOptions) error {
	return fmt.Errorf("windsurf mcp-proxy install coming soon — follow confire.dev for updates")
}
func (h *WindsurfHost) IsInstalled(_ Strategy) bool { return false }
func (h *WindsurfHost) Uninstall(_ Strategy) error  { return nil }
func (h *WindsurfHost) ComingSoon() bool             { return true }

// ── OpenAI Codex CLI ──────────────────────────────────────────────────────

type CodexHost struct{}

func (h *CodexHost) ID() string    { return "codex" }
func (h *CodexHost) Label() string { return "Codex CLI" }
func (h *CodexHost) Strategies() []Strategy { return []Strategy{StrategyMCPProxy} }
func (h *CodexHost) Preferred() Strategy    { return StrategyMCPProxy }

func (h *CodexHost) Detect() bool {
	// codex binary is the indicator
	home, _ := os.UserHomeDir()
	paths := []string{
		filepath.Join(home, ".local", "bin", "codex"),
		"/usr/local/bin/codex",
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return true
		}
	}
	return false
}

func (h *CodexHost) Install(s Strategy, opts InstallOptions) error {
	return fmt.Errorf("codex mcp-proxy install coming soon — follow confire.dev for updates")
}
func (h *CodexHost) IsInstalled(_ Strategy) bool { return false }
func (h *CodexHost) Uninstall(_ Strategy) error  { return nil }
func (h *CodexHost) ComingSoon() bool             { return true }

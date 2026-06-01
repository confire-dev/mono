package hosts

import (
	"fmt"
	"os"
	"path/filepath"
)

type CodexHost struct{}

func (h *CodexHost) ID() string    { return "codex" }
func (h *CodexHost) Label() string { return "Codex CLI" }
func (h *CodexHost) Strategies() []Strategy { return []Strategy{StrategyMCPProxy} }
func (h *CodexHost) Preferred() Strategy    { return StrategyMCPProxy }
func (h *CodexHost) ComingSoon() bool       { return true }

func (h *CodexHost) Detect() bool {
	home, _ := os.UserHomeDir()
	for _, p := range []string{filepath.Join(home, ".local", "bin", "codex"), "/usr/local/bin/codex"} {
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

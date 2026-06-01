package hosts

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

type WindsurfHost struct{}

func (h *WindsurfHost) ID() string    { return "windsurf" }
func (h *WindsurfHost) Label() string { return "Windsurf" }
func (h *WindsurfHost) Strategies() []Strategy { return []Strategy{StrategyMCPProxy} }
func (h *WindsurfHost) Preferred() Strategy    { return StrategyMCPProxy }
func (h *WindsurfHost) ComingSoon() bool       { return true }

func (h *WindsurfHost) Detect() bool {
	switch runtime.GOOS {
	case "darwin":
		if _, err := os.Stat("/Applications/Windsurf.app"); err == nil {
			return true
		}
	case "linux":
		home, _ := os.UserHomeDir()
		if _, err := os.Stat(filepath.Join(home, ".windsurf")); err == nil {
			return true
		}
	}
	return false
}

func (h *WindsurfHost) Install(s Strategy, opts InstallOptions) error {
	return fmt.Errorf("windsurf mcp-proxy install coming soon — follow confire.dev for updates")
}
func (h *WindsurfHost) IsInstalled(_ Strategy) bool { return false }
func (h *WindsurfHost) Uninstall(_ Strategy) error  { return nil }

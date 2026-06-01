package hosts

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

type ClineHost struct{}

func (h *ClineHost) ID() string    { return "cline" }
func (h *ClineHost) Label() string { return "Cline" }
func (h *ClineHost) Strategies() []Strategy { return []Strategy{StrategyMCPProxy} }
func (h *ClineHost) Preferred() Strategy    { return StrategyMCPProxy }
func (h *ClineHost) ComingSoon() bool       { return true }

func (h *ClineHost) Detect() bool {
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

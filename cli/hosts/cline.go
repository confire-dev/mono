package hosts

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// clinePlugin is the JS bridge plugin embedded into the Confire binary.
// It is written to the Cline plugins directory by `confire install cline`.
// The plugin calls `confire hook --host cline` as a subprocess for each
// tool event, which keeps all policy evaluation in the Go daemon.
const clinePlugin = `// Confire firewall bridge for Cline — installed by confire install cline
// https://confire.dev
const { spawnSync } = require('child_process')

function callConfire(event) {
  const r = spawnSync('confire', ['hook', '--host', 'cline'], {
    input: JSON.stringify(event),
    encoding: 'utf8',
    timeout: 10000,
  })
  if (r.error) return { kind: 'passthrough' }
  try { return JSON.parse(r.stdout || '{}') } catch { return { kind: 'passthrough' } }
}

module.exports = {
  hooks: {
    async beforeTool(context) {
      const r = callConfire({
        host: 'cline', strategy: 'hooks', phase: 'tool.pre',
        session: { id: context.sessionId || '', cwd: context.cwd || '' },
        tool: { name: context.tool, input: context.input },
      })
      if (r.kind === 'block' || r.kind === 'review') {
        return { block: true, blockReason: r.reason || r.context || 'Confire: action blocked by policy' }
      }
    },
    async afterTool(context) {
      const r = callConfire({
        host: 'cline', strategy: 'hooks', phase: 'tool.post',
        session: { id: context.sessionId || '', cwd: context.cwd || '' },
        tool: { name: context.tool, input: context.input, output: context.output },
      })
      if (r.context) {
        context.additionalContext = (context.additionalContext || '') + '\n' + r.context
      }
    },
  },
}
`

type ClineHost struct{}

func (h *ClineHost) ID() string             { return "cline" }
func (h *ClineHost) Label() string          { return "Cline" }
func (h *ClineHost) Strategies() []Strategy { return []Strategy{StrategyHooks} }
func (h *ClineHost) Preferred() Strategy    { return StrategyHooks }
func (h *ClineHost) ComingSoon() bool       { return false }

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
	if s != StrategyHooks {
		return fmt.Errorf("cline: strategy %q not supported", s)
	}
	pluginPath := clinePluginPath()
	if pluginPath == "" {
		return fmt.Errorf("cline: could not determine plugin directory (is Cline installed?)")
	}
	if err := os.MkdirAll(filepath.Dir(pluginPath), 0755); err != nil {
		return err
	}
	return os.WriteFile(pluginPath, []byte(clinePlugin), 0644)
}

func (h *ClineHost) IsInstalled(s Strategy) bool {
	if s != StrategyHooks {
		return false
	}
	p := clinePluginPath()
	if p == "" {
		return false
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return false
	}
	return strings.Contains(string(data), "confire")
}

func (h *ClineHost) Uninstall(s Strategy) error {
	if s != StrategyHooks {
		return nil
	}
	p := clinePluginPath()
	if p == "" {
		return nil
	}
	return os.Remove(p)
}

func clinePluginPath() string {
	home, _ := os.UserHomeDir()
	var base string
	switch runtime.GOOS {
	case "darwin":
		base = filepath.Join(home, "Library", "Application Support", "Code", "User", "globalStorage", "saoudrizwan.claude-dev")
	case "linux":
		base = filepath.Join(home, ".config", "Code", "User", "globalStorage", "saoudrizwan.claude-dev")
	case "windows":
		base = filepath.Join(os.Getenv("APPDATA"), "Code", "User", "globalStorage", "saoudrizwan.claude-dev")
	}
	if base == "" {
		return ""
	}
	return filepath.Join(base, "plugins", "confire.js")
}

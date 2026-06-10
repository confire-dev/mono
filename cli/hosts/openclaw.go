package hosts

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// openclawPlugin is the bridge plugin for OpenClaw's api.on() hook system.
// Ref: https://docs.openclaw.ai/plugins/hooks/
const openclawPlugin = `// Confire firewall bridge for OpenClaw — installed by confire install openclaw
// https://confire.dev
import { spawnSync } from 'node:child_process'

function callConfire(event) {
  const r = spawnSync('confire', ['hook', '--host', 'openclaw'], {
    input: JSON.stringify(event),
    encoding: 'utf8',
    timeout: 10000,
  })
  if (r.error) return { kind: 'passthrough' }
  try { return JSON.parse(r.stdout || '{}') } catch { return { kind: 'passthrough' } }
}

export default function confirePlugin(api) {
  api.on('before_tool_call', async (ctx) => {
    const r = callConfire({
      host: 'openclaw', strategy: 'hooks', phase: 'tool.pre',
      session: { id: ctx.sessionId || '', cwd: ctx.cwd || '' },
      tool: { name: ctx.tool, input: ctx.params },
    })
    if (r.kind === 'block' || r.kind === 'review') {
      return { block: true, blockReason: r.reason || r.context || 'Confire: action blocked by policy' }
    }
  })

  api.on('after_tool_call', async (ctx) => {
    const r = callConfire({
      host: 'openclaw', strategy: 'hooks', phase: 'tool.post',
      session: { id: ctx.sessionId || '', cwd: ctx.cwd || '' },
      tool: { name: ctx.tool, output: ctx.result },
    })
    if (r.context) {
      ctx.additionalContext = (ctx.additionalContext || '') + '\n' + r.context
    }
  })
}
`

type OpenClawHost struct{}

func (h *OpenClawHost) ID() string             { return "openclaw" }
func (h *OpenClawHost) Label() string          { return "OpenClaw" }
func (h *OpenClawHost) Strategies() []Strategy { return []Strategy{StrategyHooks} }
func (h *OpenClawHost) Preferred() Strategy    { return StrategyHooks }
func (h *OpenClawHost) ComingSoon() bool       { return false }

func (h *OpenClawHost) Detect() bool {
	home, _ := os.UserHomeDir()
	for _, p := range []string{
		filepath.Join(home, ".local", "bin", "openclaw"),
		"/usr/local/bin/openclaw",
		".openclaw",
		filepath.Join(home, ".openclaw"),
	} {
		if _, err := os.Stat(p); err == nil {
			return true
		}
	}
	return false
}

func (h *OpenClawHost) Install(s Strategy, opts InstallOptions) error {
	if s != StrategyHooks {
		return fmt.Errorf("openclaw: strategy %q not supported", s)
	}
	p := openclawPluginPath(opts.SettingsPath != "")
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return err
	}
	return os.WriteFile(p, []byte(openclawPlugin), 0644)
}

func (h *OpenClawHost) IsInstalled(s Strategy) bool {
	if s != StrategyHooks {
		return false
	}
	for _, local := range []bool{false, true} {
		data, err := os.ReadFile(openclawPluginPath(local))
		if err == nil && strings.Contains(string(data), "confire") {
			return true
		}
	}
	return false
}

func (h *OpenClawHost) Uninstall(s Strategy) error {
	if s != StrategyHooks {
		return nil
	}
	_ = os.Remove(openclawPluginPath(false))
	_ = os.Remove(openclawPluginPath(true))
	return nil
}

func openclawPluginPath(local bool) string {
	if local {
		return filepath.Join(".openclaw", "plugins", "confire.js")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".openclaw", "plugins", "confire.js")
}

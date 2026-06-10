package hosts

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// opencodePlugin is the ESM bridge plugin for OpenCode.
// Ref: https://opencode.ai/docs/plugins/
const opencodePlugin = `// Confire firewall bridge for OpenCode — installed by confire install opencode
// https://confire.dev
import { spawnSync } from 'node:child_process'

function callConfire(event) {
  const r = spawnSync('confire', ['hook', '--host', 'opencode'], {
    input: JSON.stringify(event),
    encoding: 'utf8',
    timeout: 10000,
  })
  if (r.error) return { kind: 'passthrough' }
  try { return JSON.parse(r.stdout || '{}') } catch { return { kind: 'passthrough' } }
}

export const ConfirePlugin = async (ctx) => ({
  'tool.execute.before': async (input, output) => {
    const r = callConfire({
      host: 'opencode', strategy: 'hooks', phase: 'tool.pre',
      session: { cwd: ctx.directory || '' },
      tool: { name: input.tool, input: output.args },
    })
    if (r.kind === 'block' || r.kind === 'review') {
      output.block = true
      output.blockReason = r.reason || r.context || 'Confire: action blocked by policy'
    }
  },
  'tool.execute.after': async (input, output) => {
    const r = callConfire({
      host: 'opencode', strategy: 'hooks', phase: 'tool.post',
      session: { cwd: ctx.directory || '' },
      tool: { name: input.tool, output: output.result },
    })
    if (r.context) {
      output.additionalContext = (output.additionalContext || '') + '\n' + r.context
    }
  },
})
`

type OpenCodeHost struct{}

func (h *OpenCodeHost) ID() string             { return "opencode" }
func (h *OpenCodeHost) Label() string          { return "OpenCode" }
func (h *OpenCodeHost) Strategies() []Strategy { return []Strategy{StrategyHooks} }
func (h *OpenCodeHost) Preferred() Strategy    { return StrategyHooks }
func (h *OpenCodeHost) ComingSoon() bool       { return false }

func (h *OpenCodeHost) Detect() bool {
	home, _ := os.UserHomeDir()
	for _, p := range []string{
		filepath.Join(home, ".config", "opencode"),
		".opencode",
	} {
		if _, err := os.Stat(p); err == nil {
			return true
		}
	}
	return false
}

func (h *OpenCodeHost) Install(s Strategy, opts InstallOptions) error {
	if s != StrategyHooks {
		return fmt.Errorf("opencode: strategy %q not supported", s)
	}
	p := opencodePluginPath(opts.SettingsPath != "")
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return err
	}
	return os.WriteFile(p, []byte(opencodePlugin), 0644)
}

func (h *OpenCodeHost) IsInstalled(s Strategy) bool {
	if s != StrategyHooks {
		return false
	}
	for _, local := range []bool{false, true} {
		data, err := os.ReadFile(opencodePluginPath(local))
		if err == nil && strings.Contains(string(data), "confire") {
			return true
		}
	}
	return false
}

func (h *OpenCodeHost) Uninstall(s Strategy) error {
	if s != StrategyHooks {
		return nil
	}
	_ = os.Remove(opencodePluginPath(false))
	_ = os.Remove(opencodePluginPath(true))
	return nil
}

func opencodePluginPath(local bool) string {
	if local {
		return filepath.Join(".opencode", "plugins", "confire.js")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "opencode", "plugins", "confire.js")
}

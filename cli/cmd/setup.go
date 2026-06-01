package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/confire-dev/confire/auth"
	"github.com/confire-dev/confire/config"
	"github.com/confire-dev/confire/hosts"
	"github.com/spf13/cobra"
)

var setupLocal  bool
var setupGlobal bool

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Install Confire hooks for your AI coding agent",
	Long: `Interactively installs Confire hooks into your AI agent's settings.

Scope options:
  Global (default)  installs into ~/.claude/settings.json — applies to all projects.
  Local             installs into .claude/settings.json in the nearest git root —
                    applies to this project only, and can override global settings.

Controls:
  ↑ ↓     navigate
  SPACE   toggle selection
  ENTER   confirm
  Q       quit without changes

After installation, restart your AI agent to activate the hook.

Examples:
  confire setup            interactive — asks for scope and agents
  confire setup --local    skip scope question, install locally
  confire setup --global   skip scope question, install globally`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSetup()
	},
}

func init() {
	setupCmd.Flags().BoolVar(&setupLocal,  "local",  false, "install into this project's .claude/settings.json")
	setupCmd.Flags().BoolVar(&setupGlobal, "global", false, "install into ~/.claude/settings.json (default scope)")
	rootCmd.AddCommand(setupCmd)
}

// ── Setup flow ─────────────────────────────────────────────────────────────

func runSetup() error {
	_ = auth.EnsureDeviceID()
	fmt.Printf("\n%s[confire setup]%s\n", bold, reset)

	// 1. Resolve scope ──────────────────────────────────────────────────────
	settingsPath, scopeLabel, err := resolveScope()
	if err != nil {
		fmt.Printf("\n%sSetup cancelled.%s\n\n", dim, reset)
		return nil
	}
	fmt.Printf("\n  %sScope:%s  %s\n", dim, reset, scopeLabel)

	// 2. Detect & select agents ─────────────────────────────────────────────
	registry := hosts.Registry()
	var agentItems []tuiItem
	for _, h := range registry {
		detected := h.Detect()
		switch {
		case h.ComingSoon():
			sub := "coming soon"
			if detected {
				sub = "found · coming soon"
			}
			agentItems = append(agentItems, tuiItem{
				label:    h.Label(),
				sub:      sub,
				disabled: true,
			})
		case !detected:
			agentItems = append(agentItems, tuiItem{
				label:    h.Label(),
				sub:      "not detected",
				disabled: true,
			})
		default:
			alreadyInstalled := false
			if cc, ok := h.(*hosts.ClaudeCodeHost); ok {
				alreadyInstalled = cc.IsInstalledAt(settingsPath)
			} else {
				alreadyInstalled = h.IsInstalled(h.Preferred())
			}
			sub := "found"
			if alreadyInstalled {
				sub = "found · already installed"
			}
			agentItems = append(agentItems, tuiItem{
				label:    h.Label(),
				sub:      sub,
				selected: true, // default: install everything detected
			})
		}
	}

	// Check if there's anything installable
	hasInstallable := false
	for _, it := range agentItems {
		if !it.disabled {
			hasInstallable = true
			break
		}
	}
	if !hasInstallable {
		fmt.Printf("\n  %sNo supported agents detected.%s\n", gray, reset)
		fmt.Printf("  Install Claude Code: https://claude.ai/code\n\n")
		return nil
	}

	agentPicker := newCheckbox("Select agents to configure", agentItems)
	if !agentPicker.run() {
		fmt.Printf("\n%sSetup cancelled.%s\n\n", dim, reset)
		return nil
	}

	// 3. Install ────────────────────────────────────────────────────────────
	fmt.Printf("\n  %sInstalling...%s\n\n", bold, reset)

	confireExe, _ := os.Executable()
	opts := hosts.InstallOptions{
		BinaryPath:   confireExe,
		SettingsPath: settingsPath,
	}

	selectedIdx := agentPicker.selected()
	selectedSet := map[int]bool{}
	for _, i := range selectedIdx {
		selectedSet[i] = true
	}

	anyInstalled := false
	for i, h := range registry {
		if !selectedSet[i] {
			continue
		}
		if h.Preferred() != hosts.StrategyHooks {
			fmt.Printf("  %s○%s  %-18s mcp-proxy coming soon\n", gray, reset, h.Label())
			continue
		}
		if err := h.Install(hosts.StrategyHooks, opts); err != nil {
			fmt.Printf("  %s✗%s  %-18s %s%v%s\n", red, reset, h.Label(), dim, err, reset)
		} else {
			short := shortenPath(settingsPath)
			fmt.Printf("  %s✓%s  %-18s hook installed → %s%s%s\n",
				green, reset, h.Label(), dim, short, reset)
			anyInstalled = true
		}
	}

	// Auto-start the daemon so optimization is active immediately.
	if anyInstalled {
		fmt.Printf("\n  %sStarting optimizer...%s\n\n", bold, reset)
		if isDaemonRunning() {
			fmt.Printf("  %s✓%s  Optimizer already running\n", green, reset)
		} else if err := launchDaemon(); err != nil {
			fmt.Printf("  %s○%s  Could not auto-start optimizer — run %sconfire start%s manually\n",
				gray, reset, cyan, reset)
		} else {
			fmt.Printf("  %s✓%s  Optimizer started\n", green, reset)
		}

		cfg := config.Load()
		cfg.WelcomePending = true
		_ = config.Save(cfg)
	}

	fmt.Printf("\n  %sNext steps:%s\n", bold, reset)
	if anyInstalled {
		fmt.Printf("  • Restart your AI agent to activate the hook.\n")
	}
	if key, _ := auth.LoadKey(); key == "" {
		fmt.Printf("  • Run %sconfire login%s to connect your account.\n", cyan, reset)
	}
	fmt.Printf("  • Run %sconfire status%s to verify everything is running.\n\n", cyan, reset)
	return nil
}

// resolveScope returns the settings path and a display label.
// If --local/--global flags are set it skips the interactive prompt.
func resolveScope() (path, label string, err error) {
	local  := filepath.Clean(localSettingsPath())
	global := filepath.Clean(globalSettingsPath())

	if setupLocal {
		return local, "local  " + shortenPath(local), nil
	}
	if setupGlobal {
		return global, "global  " + shortenPath(global), nil
	}

	p := newRadio("Installation scope", []tuiItem{
		{label: "Global", sub: "all projects · " + shortenPath(global)},
		{label: "Local",  sub: "this project  · " + shortenPath(local)},
	})
	if !p.run() {
		return "", "", fmt.Errorf("cancelled")
	}
	if p.firstSelected() == 1 {
		return local, "local  " + shortenPath(local), nil
	}
	return global, "global  " + shortenPath(global), nil
}

// ── Helpers ────────────────────────────────────────────────────────────────

func shortenPath(p string) string {
	home, _ := os.UserHomeDir()
	if strings.HasPrefix(p, home) {
		return "~" + p[len(home):]
	}
	cwd, _ := os.Getwd()
	if strings.HasPrefix(p, cwd) {
		rel := p[len(cwd):]
		if rel == "" {
			return "."
		}
		return "." + rel
	}
	return p
}

func buildOptimizerList(mcpServers map[string]string) []optimizerItem {
	items := make([]optimizerItem, len(DEFAULT_OPTIMIZERS))
	copy(items, DEFAULT_OPTIMIZERS)
	for name := range mcpServers {
		lower := strings.ToLower(name)
		if lower == "figma" || lower == "figma-desktop" {
			continue
		}
		items = append(items, optimizerItem{
			id:          "mcp-" + lower,
			label:       name + " (MCP) ✦",
			description: "platform-specific optimizer (paid)",
			phase:       "tool.post",
			enabled:     true,
		})
	}
	return items
}

func discoverMCPServers() map[string]string {
	servers := make(map[string]string)
	home, _ := os.UserHomeDir()
	paths := []string{filepath.Join(home, ".claude", "mcp.json")}
	if cwd, err := os.Getwd(); err == nil {
		paths = append(paths, filepath.Join(cwd, ".mcp.json"))
	}
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		var raw struct {
			MCPServers map[string]map[string]interface{} `json:"mcpServers"`
		}
		if err := json.Unmarshal(data, &raw); err != nil {
			continue
		}
		for name, cfg := range raw.MCPServers {
			url, _ := cfg["url"].(string)
			if url == "" {
				url, _ = cfg["command"].(string)
			}
			servers[name] = url
		}
	}
	return servers
}

// ── Optimizer inventory (used by setup display) ────────────────────────────

type optimizerItem struct {
	id          string
	label       string
	description string
	phase       string
	enabled     bool
	isSoon      bool
}

var DEFAULT_OPTIMIZERS = []optimizerItem{
	{id: "bash",        label: "Bash",       description: "trim logs, keep failures (60-95%)",        phase: "tool.post",     enabled: true},
	{id: "read",        label: "Read",       description: "cap large file reads (pre + post)",         phase: "tool.pre+post", enabled: true},
	{id: "webfetch",    label: "WebFetch",   description: "strip HTML/CSS noise (80-90%)",             phase: "tool.post",     enabled: true},
	{id: "generic",     label: "Generic",    description: "universal JSON noise stripping (fallback)", phase: "tool.post",     enabled: true},
	{id: "figma",       label: "Figma ✦",    description: "JSX → section map (98% reduction)",        phase: "tool.post",     enabled: true},
	{id: "mcp-generic", label: "MCP tools ✦", description: "platform-specific optimizers",            phase: "tool.post",     enabled: true},
	{id: "pre-compact", label: "Pre-compact", description: "context compaction",                       phase: "context.pre-compact", enabled: false, isSoon: true},
}

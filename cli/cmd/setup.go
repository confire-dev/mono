package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/confire-dev/confire/auth"
	"github.com/confire-dev/confire/hosts"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Detect AI tools and configure Confire optimization",
	Long: `Scans your system for supported AI coding tools and their MCP servers,
shows what Confire will optimize, and installs the hook.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSetup()
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
}

// ── Optimizer / phase inventory shown in the TUI ──────────────────────────

type optimizerItem struct {
	id          string
	label       string
	description string
	phase       string
	enabled     bool
	isSoon      bool
}

var DEFAULT_OPTIMIZERS = []optimizerItem{
	// Local optimizer (free tier) — runs in the CLI daemon
	{id: "bash",        label: "Bash",      description: "trim logs, keep failures (60-95%)",   phase: "tool.post",     enabled: true},
	{id: "read",        label: "Read",      description: "cap large file reads (pre + post)",    phase: "tool.pre+post", enabled: true},
	{id: "webfetch",    label: "WebFetch",  description: "strip HTML/CSS noise (80-90%)",        phase: "tool.post",     enabled: true},
	{id: "generic",     label: "Generic",   description: "universal JSON noise stripping (fallback)", phase: "tool.post",  enabled: true},
	// Remote optimizer (paid tier) — runs in the Cloudflare Worker
	{id: "figma",       label: "Figma ✦",   description: "JSX → section map (98% reduction)",   phase: "tool.post",     enabled: true},
	{id: "mcp-generic", label: "MCP tools ✦", description: "platform-specific optimizers",      phase: "tool.post",     enabled: true},
	// Coming soon
	{id: "pre-compact", label: "Pre-compact", description: "context compaction",                phase: "context.pre-compact", enabled: false, isSoon: true},
}

func runSetup() error {
	// Ensure device ID exists before anything else.
	_ = auth.EnsureDeviceID()

	fmt.Printf("\n%s[confire setup]%s Scanning your system...\n\n", bold, reset)

	// ── Section 1: Agents (from registry) ─────────────────────────────────
	fmt.Printf("  %sAgents%s\n", bold, reset)
	fmt.Printf("  %s────────────────────────────────────%s\n", dim, reset)

	detected := hosts.Detect()
	registry  := hosts.Registry()

	for _, h := range detected {
		strategy := string(h.Preferred())
		installed := h.IsInstalled(h.Preferred())
		badge := fmt.Sprintf("%sfound ✓%s", green, reset)
		if installed {
			badge = fmt.Sprintf("%sfound ✓  hook installed ✓%s", green, reset)
		}
		fmt.Printf("  %s●%s %-18s %s  [strategy: %s]\n",
			green, reset, h.Label(), badge, strategy)
	}

	// Coming-soon: in registry but not detected, or detected but hooks-only
	detectedIDs := map[string]bool{}
	for _, h := range detected {
		detectedIDs[h.ID()] = true
	}
	var notDetected []hosts.Host
	for _, h := range registry {
		if !detectedIDs[h.ID()] {
			notDetected = append(notDetected, h)
		}
	}
	if len(notDetected) > 0 {
		fmt.Printf("\n  %sComing soon%s\n", dim, reset)
		for _, h := range notDetected {
			fmt.Printf("  %s○%s %s\n", gray, reset, h.Label())
		}
	}

	if len(detected) == 0 {
		fmt.Printf("\n  %sNo supported agents detected.%s\n", gray, reset)
		fmt.Printf("  Install Claude Code: https://claude.ai/code\n\n")
		return nil
	}

	// ── Section 2: MCP Servers ─────────────────────────────────────────────
	mcpServers := discoverMCPServers()
	if len(mcpServers) > 0 {
		fmt.Printf("\n  %sMCP Servers found%s\n", bold, reset)
		fmt.Printf("  %s────────────────────────────────────%s\n", dim, reset)
		for name, url := range mcpServers {
			fmt.Printf("  %s●%s %-20s %s%s%s\n", green, reset, name, dim, url, reset)
		}
	}

	// ── Section 3: Optimizer list ──────────────────────────────────────────
	fmt.Printf("\n  %sOptimizers%s  %s✦ = cloud/paid tier%s\n", bold, reset, dim, reset)
	fmt.Printf("  %s────────────────────────────────────%s\n", dim, reset)

	items := buildOptimizerList(mcpServers)
	for i, item := range items {
		icon := fmt.Sprintf("%s●%s", green, reset)
		suffix := ""
		if item.isSoon {
			icon = fmt.Sprintf("%s○%s", gray, reset)
			suffix = fmt.Sprintf(" %s(soon)%s", dim, reset)
		} else if !item.enabled {
			icon = fmt.Sprintf("%s○%s", gray, reset)
		}
		fmt.Printf("  [%d] %s %-24s %s%s%s%s\n",
			i+1, icon, item.label,
			dim, item.description, reset, suffix,
		)
	}

	fmt.Printf("\n  %sToggle (space-separated numbers, or 'all', enter to keep defaults):%s\n%s> %s",
		bold, reset, dim, reset)

	selected, err := readSetupSelection(len(items))
	if err != nil {
		fmt.Println()
	} else if len(selected) > 0 {
		for _, idx := range selected {
			items[idx].enabled = !items[idx].enabled
		}
	}

	// ── Apply: install hooks for each detected host ────────────────────────
	fmt.Printf("\n%sApplying...%s\n\n", bold, reset)

	confireExe, _ := os.Executable()
	opts := hosts.InstallOptions{BinaryPath: confireExe}

	anyInstalled := false
	for _, h := range detected {
		if h.Preferred() != hosts.StrategyHooks {
			fmt.Printf("  %s○%s  %-18s mcp-proxy install coming soon\n", gray, reset, h.Label())
			continue
		}
		if h.IsInstalled(hosts.StrategyHooks) {
			fmt.Printf("  %s✓%s  %-18s hook already installed\n", green, reset, h.Label())
			anyInstalled = true
			continue
		}
		if err := h.Install(hosts.StrategyHooks, opts); err != nil {
			fmt.Printf("  %s✗%s  %-18s %v\n", "\033[31m", reset, h.Label(), err)
		} else {
			fmt.Printf("  %s✓%s  %-18s hook installed → %s\n",
				green, reset, h.Label(), hookSettingsPath(h.ID()))
			anyInstalled = true
		}
	}

	fmt.Printf("\n%sNext steps:%s\n", bold, reset)
	if anyInstalled {
		fmt.Printf("  Restart Claude Code to pick up the new hook.\n")
	}
	fmt.Printf("  Run %sconfire login%s to connect your account (free tier: 500 req/month).\n\n", cyan, reset)
	return nil
}

// hookSettingsPath returns a human-readable hint for where the hook was written.
func hookSettingsPath(hostID string) string {
	home, _ := os.UserHomeDir()
	switch hostID {
	case "claude-code":
		return "~/.claude/settings.json"
	default:
		return filepath.Join(home, "."+hostID, "settings.json")
	}
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

func readSetupSelection(n int) ([]int, error) {
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, nil
	}
	if strings.ToLower(line) == "all" {
		idxs := make([]int, n)
		for i := range idxs {
			idxs[i] = i
		}
		return idxs, nil
	}
	seen := map[int]bool{}
	var result []int
	for _, part := range strings.Split(line, " ") {
		num, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || num < 1 || num > n {
			continue
		}
		if !seen[num-1] {
			seen[num-1] = true
			result = append(result, num-1)
		}
	}
	return result, nil
}

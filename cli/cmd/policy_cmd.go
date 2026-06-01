package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/confire-dev/confire/config"
	"github.com/confire-dev/confire/intercept"
	"github.com/confire-dev/confire/policy"
	"github.com/spf13/cobra"
)

var policyCmd = &cobra.Command{
	Use:   "policy",
	Short: "Manage Confire firewall policy rules",
}

var policyTestCmd = &cobra.Command{
	Use:   "test <command or tool>",
	Short: "Show what Confire would do for a given command or tool call",
	Long: `Simulates a PreToolUse firewall evaluation and prints the result.

Examples:
  confire policy test 'git push --force'
  confire policy test 'mcp__github__merge_pull_request'
  confire policy test 'gh pr close 42'`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPolicyTest(args[0])
	},
}

var policyStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show active firewall policy status",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPolicyStatus()
	},
}

var policyPullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Fetch custom rules from the Confire dashboard (requires paid plan)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPolicyPull()
	},
}

func init() {
	policyCmd.AddCommand(policyTestCmd)
	policyCmd.AddCommand(policyStatusCmd)
	policyCmd.AddCommand(policyPullCmd)
	rootCmd.AddCommand(policyCmd)
}

// ── policy test ───────────────────────────────────────────────────────────

func runPolicyTest(input string) error {
	cfg := config.Load()
	rules := policy.LoadRules()
	engine := policy.NewEngine(rules)
	mode := policy.Mode(cfg.EffectiveMode())

	// Build a synthetic event from the input string.
	event := syntheticEvent(input)

	result := engine.EvaluatePreTool(event, mode)
	if result == nil {
		fmt.Printf("%sAction:%s   allow\n", bold, reset)
		fmt.Printf("%sReason:%s   No matching rule — tool would proceed normally.\n", bold, reset)
		return nil
	}

	actionColor := colorForAction(result.Action)
	fmt.Printf("%sAction:%s   %s%s%s\n", bold, reset, actionColor, string(result.Action), reset)
	fmt.Printf("%sRule:%s     %s\n", bold, reset, result.Rule.Name)
	fmt.Printf("%sSeverity:%s %s\n", bold, reset, string(result.Rule.Severity))
	fmt.Printf("%sReason:%s   %s\n", bold, reset, result.Message)
	if result.Rule.Source != "" {
		fmt.Printf("%sSource:%s   %s\n", bold, reset, result.Rule.Source)
	}
	return nil
}

// syntheticEvent builds a PreToolUse InterceptEvent from a CLI input string.
// If input looks like an MCP tool name (mcp__*), use it as the tool name.
// Otherwise, assume it's a Bash command.
func syntheticEvent(input string) intercept.InterceptEvent {
	input = strings.TrimSpace(input)
	toolName := "Bash"
	toolInput := map[string]any{"command": input}

	if strings.HasPrefix(input, "mcp__") {
		toolName = input
		toolInput = map[string]any{}
	}

	return intercept.InterceptEvent{
		Host:     "claude-code",
		Strategy: "hooks",
		Phase:    intercept.PhaseToolPre,
		Session:  intercept.Session{ID: "test"},
		Tool: &intercept.Tool{
			Name:  toolName,
			Input: toolInput,
			IsMCP: strings.HasPrefix(toolName, "mcp__"),
		},
	}
}

// ── policy status ─────────────────────────────────────────────────────────

func runPolicyStatus() error {
	cfg := config.Load()
	rules := policy.LoadRules()

	builtinCount := 0
	customCount := 0
	for _, r := range rules {
		if r.Source == "custom" {
			customCount++
		} else {
			builtinCount++
		}
	}

	cacheCount, fetchedAt, cacheVersion := policy.LoadCacheInfo()

	fmt.Printf("\n%s[confire policy]%s\n\n", bold, reset)
	fmt.Printf("  Mode:          %s%s%s\n", bold, cfg.EffectiveMode(), reset)
	fmt.Printf("  Firewall:      %s\n", firewallStatus(cfg))
	fmt.Printf("  Built-in rules: %d\n", builtinCount)
	fmt.Printf("  Custom rules:  %d", customCount)
	if customCount > 0 && cacheCount > 0 {
		age := ""
		if !fetchedAt.IsZero() {
			age = " (fetched " + formatAge(fetchedAt) + ")"
		}
		if cacheVersion != "" {
			age += " [v" + cacheVersion + "]"
		}
		fmt.Print(age)
	}
	fmt.Println()
	fmt.Printf("  Total rules:   %d\n\n", len(rules))
	return nil
}

// ── policy pull ───────────────────────────────────────────────────────────

func runPolicyPull() error {
	// Placeholder: full remote sync is implemented when the backend endpoint is ready.
	// For now, show a clear message.
	fmt.Fprintln(os.Stderr, "[confire] policy pull: remote custom rules require a paid plan.")
	fmt.Fprintln(os.Stderr, "          Visit https://confire.dev to manage custom rules.")
	return nil
}

// ── helpers ───────────────────────────────────────────────────────────────

func firewallStatus(cfg config.Config) string {
	if !cfg.IsFirewallEnabled() {
		return dim + "disabled" + reset
	}
	return green + "enabled" + reset
}

func colorForAction(action policy.RuleAction) string {
	switch action {
	case policy.ActionBlock:
		return red
	case policy.ActionReview:
		return yellow
	case policy.ActionWarn:
		return yellow
	default:
		return green
	}
}

func formatAge(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}

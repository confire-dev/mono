package cmd

import (
	"fmt"
	"os"

	"github.com/confire-dev/confire/config"
	"github.com/confire-dev/confire/policy"
	"github.com/spf13/cobra"
)

var onCmd = &cobra.Command{
	Use:   "on",
	Short: "Enable the Confire firewall (balanced mode)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runFirewallOn()
	},
}

var offCmd = &cobra.Command{
	Use:   "off",
	Short: "Disable the Confire firewall (bypass mode)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runFirewallOff()
	},
}

var bypassNextCmd = &cobra.Command{
	Use:   "bypass-next",
	Short: "Allow the next tool call to skip firewall review (one-shot override)",
	Long: `Writes a one-shot bypass flag that causes the firewall to pass through
the very next PreToolUse event without review or block.

Use this when Confire has reviewed a command and you have confirmed it is safe.
The flag is automatically cleared after one use.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runBypassNext()
	},
}

func init() {
	rootCmd.AddCommand(onCmd)
	rootCmd.AddCommand(offCmd)
	rootCmd.AddCommand(bypassNextCmd)
}

func runFirewallOn() error {
	cfg := config.Load()
	cfg.Mode = string(policy.ModeBalanced)
	t := true
	cfg.FirewallEnabled = &t
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("save config: %w", err)
	}
	fmt.Printf("%s✓ Confire firewall enabled%s (mode: balanced)\n", green, reset)
	fmt.Println("  Restart the Confire daemon for changes to take effect: confire stop && confire start")
	return nil
}

func runFirewallOff() error {
	cfg := config.Load()
	cfg.Mode = string(policy.ModeBypass)
	f := false
	cfg.FirewallEnabled = &f
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("save config: %w", err)
	}
	fmt.Fprintf(os.Stderr, "%s⚠ Confire firewall disabled%s (bypass mode)\n", yellow, reset)
	fmt.Fprintln(os.Stderr, "  Run `confire on` to re-enable. Restart daemon to apply.")
	return nil
}

func runBypassNext() error {
	if err := policy.SetBypassNext(); err != nil {
		return fmt.Errorf("set bypass-next: %w", err)
	}
	fmt.Printf("%s✓ bypass-next set%s — next PreToolUse event will skip firewall review.\n", yellow, reset)
	fmt.Println("  This flag clears automatically after one use.")
	return nil
}

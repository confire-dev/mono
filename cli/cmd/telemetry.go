package cmd

import (
	"fmt"

	"github.com/confire-dev/confire/config"
	"github.com/spf13/cobra"
)

var telemetryCmd = &cobra.Command{
	Use:   "telemetry [on|off]",
	Short: "Manage optional product analytics",
	Long: `Controls whether Confire sends optional product analytics to Amplitude.

Telemetry = optional behavioral analytics (feature adoption, funnel analysis).
Usage accounting required for billing and your dashboard is always active.

Override via environment: CONFIRE_TELEMETRY=0  or  CONFIRE_TELEMETRY=false`,
	RunE: runTelemetry,
}

func init() {
	rootCmd.AddCommand(telemetryCmd)
}

func runTelemetry(cmd *cobra.Command, args []string) error {
	cfg := config.Load()

	if len(args) == 0 {
		// Show status
		fmt.Printf("\n%s[confire telemetry]%s\n\n", bold, reset)
		if cfg.Telemetry {
			fmt.Printf("  Status: %son%s — optional analytics enabled\n", green, reset)
		} else {
			fmt.Printf("  Status: %soff%s — analytics disabled\n", gray, reset)
			fmt.Printf("  %sNote: usage accounting for billing and your dashboard is still active.%s\n", dim, reset)
		}
		fmt.Printf("\n  To change: %sconfire telemetry on|off%s\n", cyan, reset)
		fmt.Printf("  Env override: %sCONFIRE_TELEMETRY=0%s\n\n", dim, reset)
		return nil
	}

	switch args[0] {
	case "on":
		cfg.Telemetry = true
		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("save config: %w", err)
		}
		fmt.Printf("%s✓%s Telemetry enabled — thank you for helping improve Confire\n", green, reset)

	case "off":
		cfg.Telemetry = false
		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("save config: %w", err)
		}
		fmt.Printf("%s✓%s Telemetry disabled\n", green, reset)
		fmt.Printf("  %sUsage accounting for billing and your dashboard remains active.%s\n", dim, reset)

	default:
		return fmt.Errorf("expected 'on' or 'off', got %q", args[0])
	}
	return nil
}

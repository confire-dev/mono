package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var reviewCmd = &cobra.Command{
	Use:   "review",
	Short: "Show recent security events",
	Long: `Lists the last 20 security events in reverse chronological order.

Each event shows: timestamp, tool, event type, risk level, action taken, and pattern matched.

Examples:
  confire review`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runReview()
	},
}

func init() {
	rootCmd.AddCommand(reviewCmd)
}

func runReview() error {
	// TODO: read from local security event log (~/.confire/security.log)
	fmt.Println()
	fmt.Println("  No security events recorded yet.")
	fmt.Println("  Events are logged as Confire reviews tool calls.")
	fmt.Println()
	return nil
}

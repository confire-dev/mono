package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// flagLocal points all network calls at local dev servers (platform: localhost:4321, worker: localhost:8787).
var flagLocal bool

func init() {
	rootCmd.PersistentFlags().BoolVar(&flagLocal, "local", false, "use local dev servers (platform :4321, worker :8787)")
}

var rootCmd = &cobra.Command{
	Use:   "confire",
	Short: "Confire — universal AI agent tool-output optimizer",
	Long: `Confire optimizes tool call outputs before they reach the model context.
Less noise in context = cheaper, faster, sharper AI agents.

Get started:
  confire setup    detect your AI tools and configure optimization
  confire login    connect your account (cloud optimization + stats)
  confire status   show what's running and what's being optimized`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "confire:", err)
		os.Exit(1)
	}
}

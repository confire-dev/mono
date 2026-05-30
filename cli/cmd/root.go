package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

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

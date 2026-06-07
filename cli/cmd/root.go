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
	Short: "Confire — context firewall for AI coding agents",
	Long: `Confire reviews risky tool calls before they run, sanitizes untrusted tool
output before it reaches the model, and labels where context came from.

Get started:
  confire setup    install hooks for your AI agent and start the firewall
  confire login    connect your account (cloud sync + security events)
  confire status   show current state

Firewall control:
  confire start       start the firewall daemon in the background
  confire stop        stop the daemon
  confire bypass-next allow the next tool call without review (one-shot)
  confire review      show recent security events
  confire stats       show activity summary`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "confire:", err)
		os.Exit(1)
	}
}

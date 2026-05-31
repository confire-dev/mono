package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/confire-dev/confire/hosts"
	"github.com/spf13/cobra"
)

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Remove Confire hooks and stop the optimizer (leaves the binary installed)",
	Long: `Removes Confire hooks from your AI agent's settings and stops the optimizer daemon.

Confire remains installed — run 'confire setup' any time to re-enable it.
To remove Confire completely, delete the binary.

Examples:
  confire reset            remove from all settings + stop optimizer
  confire reset --local    remove only from this project
  confire reset --global   remove only from global settings`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runReset()
	},
}

var resetLocalFlag  bool
var resetGlobalFlag bool

func init() {
	resetCmd.Flags().BoolVar(&resetLocalFlag,  "local",  false, "remove from this project's .claude/settings.json only")
	resetCmd.Flags().BoolVar(&resetGlobalFlag, "global", false, "remove from ~/.claude/settings.json only")
	rootCmd.AddCommand(resetCmd)
}

func runReset() error {
	fmt.Printf("\n%s[confire reset]%s\n", bold, reset)

	local  := filepath.Clean(localSettingsPath())
	global := filepath.Clean(globalSettingsPath())

	type target struct{ path, label string }
	var targets []target

	switch {
	case resetLocalFlag:
		targets = []target{{local, "local   " + shortenPath(local)}}
	case resetGlobalFlag:
		targets = []target{{global, "global  " + shortenPath(global)}}
	default:
		targets = []target{
			{global, "global  " + shortenPath(global)},
			{local,  "local   " + shortenPath(local)},
		}
	}

	fmt.Printf("\n  %sRemoving hooks...%s\n\n", bold, reset)

	for _, t := range targets {
		cc := &hosts.ClaudeCodeHost{}
		if !cc.IsInstalledAt(t.path) {
			fmt.Printf("  %s○%s  %s  %snot installed%s\n", gray, reset, t.label, dim, reset)
			continue
		}
		if err := cc.UninstallAt(hosts.StrategyHooks, t.path); err != nil {
			fmt.Printf("  %s✗%s  %s  %v\n", red, reset, t.label, err)
		} else {
			fmt.Printf("  %s✓%s  %s  hook removed\n", green, reset, t.label)
		}
	}

	// Stop the daemon if it's running.
	if isDaemonRunning() {
		fmt.Printf("\n  %sStopping optimizer...%s\n\n", bold, reset)
		if err := stopDaemon(); err != nil {
			fmt.Printf("  %s✗%s  Could not stop optimizer: %v\n", red, reset, err)
		} else {
			fmt.Printf("  %s✓%s  Optimizer stopped\n", green, reset)
		}
	}

	fmt.Printf("\n  Confire is now passive.\n")
	fmt.Printf("  Run %sconfire setup%s to re-enable.\n\n", cyan, reset)
	return nil
}

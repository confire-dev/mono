package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/confire-dev/confire/hosts"
	"github.com/spf13/cobra"
)

var disableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Remove Confire hooks from your AI agent's settings",
	Long: `Removes the Confire hook from Claude Code's settings.json.

Scope options:
  Global (default)  removes from ~/.claude/settings.json
  Local             removes from .claude/settings.json in the nearest git root
  Both              removes from both locations

All other settings are preserved. The original file is backed up as
settings.json.confire-backup before any changes are made.

Controls:
  ↑ ↓     navigate
  ENTER   confirm
  Q       quit without changes

Examples:
  confire disable            interactive — asks which scope to remove from
  confire disable --local    remove only from this project
  confire disable --global   remove only from global settings`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDisable()
	},
}

var disableLocalFlag  bool
var disableGlobalFlag bool

func init() {
	disableCmd.Flags().BoolVar(&disableLocalFlag,  "local",  false, "remove from this project's .claude/settings.json")
	disableCmd.Flags().BoolVar(&disableGlobalFlag, "global", false, "remove from ~/.claude/settings.json")
	rootCmd.AddCommand(disableCmd)
}

func runDisable() error {
	fmt.Printf("\n%s[confire disable]%s\n", bold, reset)

	local  := filepath.Clean(localSettingsPath())
	global := filepath.Clean(globalSettingsPath())

	type target struct{ path, label string }
	var targets []target

	if disableLocalFlag {
		targets = []target{{local, "local  " + shortenPath(local)}}
	} else if disableGlobalFlag {
		targets = []target{{global, "global  " + shortenPath(global)}}
	} else {
		p := newRadio("Remove Confire hooks from", []tuiItem{
			{label: "Global", sub: "all projects · " + shortenPath(global)},
			{label: "Local",  sub: "this project  · " + shortenPath(local)},
			{label: "Both",   sub: "global and local"},
		})
		if !p.run() {
			fmt.Printf("\n%sNo changes made.%s\n\n", dim, reset)
			return nil
		}
		switch p.firstSelected() {
		case 0:
			targets = []target{{global, "global  " + shortenPath(global)}}
		case 1:
			targets = []target{{local, "local  " + shortenPath(local)}}
		case 2:
			targets = []target{
				{global, "global  " + shortenPath(global)},
				{local,  "local   " + shortenPath(local)},
			}
		}
	}

	fmt.Printf("\n  %sRemoving...%s\n\n", bold, reset)

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

	fmt.Printf("\n  Restart your AI agent for changes to take effect.\n\n")
	return nil
}

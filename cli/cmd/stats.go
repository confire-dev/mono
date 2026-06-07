package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show security activity summary",
	Long: `Shows firewall activity from the current session and recent history.

Examples:
  confire stats              summary view`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runStats()
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)
}

func runStats() error {
	printSecuritySummary()
	return nil
}

func printSecuritySummary() {
	w := 45
	line := func(label, value string) {
		if value == "" {
			fmt.Printf("│  %-*s │\n", w-4, label)
		} else {
			fmt.Printf("│  %-28s %*s │\n", label, w-33, value)
		}
	}
	sep := func() { fmt.Printf("├%s┤\n", repeatStr("─", w-2)) }
	top := func() { fmt.Printf("┌%s┐\n", repeatStr("─", w-2)) }
	bot := func() { fmt.Printf("└%s┘\n", repeatStr("─", w-2)) }

	top()
	line("Confire — Activity Summary", "")
	sep()
	line("This session", "")
	line("Tool calls reviewed:", "—")
	line("Actions blocked:", "—")
	line("Actions warned:", "—")
	line("Outputs sanitized:", "—")
	line("Secrets redacted:", "—")
	line("Injections caught:", "—")
	line("Bypasses used:", "—")
	sep()
	line("Trust distribution", "")
	line("Internal:", "—")
	line("External trusted:", "—")
	line("External untrusted:", "—")
	bot()

	fmt.Printf("\n  Run `confire review` to see recent security events.\n\n")
}

func repeatStr(s string, n int) string {
	out := make([]byte, 0, n*len(s))
	for i := 0; i < n; i++ {
		out = append(out, s...)
	}
	return string(out)
}

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/confire-dev/confire/internal/stats"
	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show local optimization stats",
	Long: `Displays token savings from local stats.
Stats are stored locally in ~/.confire/stats.db and never sent anywhere.

Examples:
  confire stats              show all periods
  confire stats --today      today only
  confire stats --month      this month only
  confire stats --tool bash  filter by tool
  confire stats --json       machine-readable output
  confire stats --clear      delete all local stats`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runStats(cmd)
	},
}

var (
	statsFlagToday bool
	statsFlagMonth bool
	statsFlagTool  string
	statsFlagJSON  bool
	statsFlagClear bool
)

func init() {
	statsCmd.Flags().BoolVar(&statsFlagToday, "today", false, "show today's stats only")
	statsCmd.Flags().BoolVar(&statsFlagMonth, "month", false, "show this month's stats only")
	statsCmd.Flags().StringVar(&statsFlagTool, "tool", "", "filter by tool (e.g. bash, figma, github)")
	statsCmd.Flags().BoolVar(&statsFlagJSON, "json", false, "output as JSON")
	statsCmd.Flags().BoolVar(&statsFlagClear, "clear", false, "delete all local stats")
	rootCmd.AddCommand(statsCmd)
}

func statsDBPath() string {
	return filepath.Join(confireDir(), "stats.db")
}

func runStats(cmd *cobra.Command) error {
	db, err := stats.Open(statsDBPath())
	if err != nil {
		return fmt.Errorf("could not open stats database: %w", err)
	}
	defer db.Close()

	if statsFlagClear {
		if err := db.Clear(); err != nil {
			return fmt.Errorf("clear failed: %w", err)
		}
		fmt.Println("✓ Stats cleared.")
		return nil
	}

	today, err := db.Today()
	if err != nil {
		return fmt.Errorf("read stats: %w", err)
	}
	month, err := db.ThisMonth()
	if err != nil {
		return fmt.Errorf("read stats: %w", err)
	}
	allTime, err := db.AllTime()
	if err != nil {
		return fmt.Errorf("read stats: %w", err)
	}
	topTools, err := db.TopToolsThisMonth(4)
	if err != nil {
		// non-fatal
		topTools = nil
	}
	syncCounts, err := db.GetSyncCounts()
	if err != nil {
		syncCounts = stats.SyncCounts{}
	}

	// Session stats: ask the running daemon via a side channel.
	// For now, zero out — session stats live in the daemon process.
	// Future: expose a /stats socket endpoint on the daemon.
	session := stats.SessionStats{}

	if statsFlagJSON {
		stats.PrintJSON(session, today, month, allTime, topTools, syncCounts)
		return nil
	}

	// Narrow output when a flag narrows the period.
	if statsFlagToday {
		fmt.Fprintf(os.Stdout, "Today: %s requests · %s tokens saved\n",
			formatNumber(today.RequestCount), formatNumber(today.TokensSaved))
		return nil
	}
	if statsFlagMonth {
		fmt.Fprintf(os.Stdout, "This month: %s requests · %s tokens saved\n",
			formatNumber(month.RequestCount), formatNumber(month.TokensSaved))
		return nil
	}

	stats.PrintTable(session, today, month, allTime, topTools, syncCounts)
	return nil
}

func formatNumber(n int64) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	var out []byte
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, byte(c))
	}
	return string(out)
}

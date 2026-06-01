package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/confire-dev/confire/internal/stats"
	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show optimization savings",
	Long: `Shows tokens saved and request counts from local stats (this month and all time).

Examples:
  confire stats              summary view
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

func dashboardURL() string {
	return platformURL() + "/dashboard"
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

	var today, month, allTime stats.Stats
	var statsErr error
	if statsFlagTool != "" {
		tool := statsFlagTool
		today, statsErr = db.TodayByTool(tool)
		if statsErr == nil {
			month, statsErr = db.ThisMonthByTool(tool)
		}
		if statsErr == nil {
			allTime, statsErr = db.AllTimeByTool(tool)
		}
	} else {
		today, statsErr = db.Today()
		if statsErr == nil {
			month, statsErr = db.ThisMonth()
		}
		if statsErr == nil {
			allTime, statsErr = db.AllTime()
		}
	}
	if statsErr != nil {
		return fmt.Errorf("read stats: %w", statsErr)
	}

	if statsFlagJSON {
		topTools, _ := db.TopToolsThisMonth(4)
		syncCounts, _ := db.GetSyncCounts()
		stats.PrintJSON(stats.SessionStats{}, today, month, allTime, topTools, syncCounts)
		return nil
	}

	toolLabel := ""
	if statsFlagTool != "" {
		toolLabel = statsFlagTool
	}

	if statsFlagToday {
		printStatsPeriod("Today", today, toolLabel)
		return nil
	}
	if statsFlagMonth {
		printStatsPeriod("This month", month, toolLabel)
		return nil
	}

	printStatsSummary(month, allTime, toolLabel)
	return nil
}

func printStatsSummary(month, allTime stats.Stats, toolFilter string) {
	title := "confire stats"
	if toolFilter != "" {
		title = fmt.Sprintf("confire stats · %s", toolFilter)
	}
	fmt.Printf("\n%s[%s]%s\n\n", bold, title, reset)

	if allTime.RequestCount == 0 {
		fmt.Printf("  %sNo optimizations recorded yet.%s\n", dim, reset)
		fmt.Printf("  %sRun `confire start` and use your agent to begin saving tokens.%s\n\n", dim, reset)
	} else {
		fmt.Printf("  %s%s%s %stokens saved%s\n",
			bold, green, formatNumber(allTime.TokensSaved), dim, reset)
		fmt.Printf("  %sall time%s\n\n", dim, reset)

		printStatsRow("This month", month)
		printStatsRow("All time", allTime)
		fmt.Println()
	}

	fmt.Printf("  %sFull history →%s %s\n\n", dim, reset, dashboardURL())
}

func printStatsPeriod(label string, s stats.Stats, toolFilter string) {
	title := label
	if toolFilter != "" {
		title = fmt.Sprintf("%s · %s", label, toolFilter)
	}
	fmt.Printf("\n%s[%s]%s\n\n", bold, title, reset)

	if s.RequestCount == 0 {
		fmt.Printf("  %sNo data for this period.%s\n\n", dim, reset)
	} else {
		fmt.Printf("  %s%s%s %stokens saved%s\n",
			bold, green, formatNumber(s.TokensSaved), dim, reset)
		fmt.Printf("  %s%d requests%s\n\n", dim, s.RequestCount, reset)
	}

	fmt.Printf("  %sFull history →%s %s\n\n", dim, reset, dashboardURL())
}

func printStatsRow(label string, s stats.Stats) {
	fmt.Printf("  %-12s %s%s%s requests · %s%s%s saved\n",
		label+":",
		cyan, formatNumber(s.RequestCount), reset,
		green, formatNumber(s.TokensSaved), reset)
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

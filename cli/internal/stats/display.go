package stats

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const boxWidth = 41

// SessionStats holds in-memory stats for the current daemon session.
// Passed in from the daemon so the stats command can show live session data.
type SessionStats struct {
	RequestCount int64
	TokensSaved  int64
}

// PrintTable renders the full stats table to stdout.
func PrintTable(session SessionStats, today, month, allTime Stats, tools []ToolStat) {
	line := strings.Repeat("─", boxWidth-2)
	fmt.Printf("┌%s┐\n", line)
	printCentered("Confire Stats")
	fmt.Printf("├%s┤\n", line)

	printSection("This session", session.RequestCount, session.TokensSaved)
	printSection("Today", today.RequestCount, today.TokensSaved)
	printSection("This month", month.RequestCount, month.TokensSaved)
	printSection("All time", allTime.RequestCount, allTime.TokensSaved)

	fmt.Printf("├%s┤\n", line)
	printRow("Top tools this month", "")

	if len(tools) == 0 {
		printRow("  No data yet", "")
	} else {
		for _, t := range tools {
			label := fmt.Sprintf("  %-12s", capitalize(t.ToolName))
			val := fmt.Sprintf("%.0f%% · %.0f%% avg reduction", t.SharePct, t.AvgReduction)
			printRow(label, val)
		}
	}

	fmt.Printf("└%s┘\n", strings.Repeat("─", boxWidth-2))
}

// PrintJSON renders stats as machine-readable JSON.
func PrintJSON(session SessionStats, today, month, allTime Stats, tools []ToolStat) {
	type statJSON struct {
		Requests    int64 `json:"requests"`
		TokensSaved int64 `json:"tokens_saved"`
	}
	type toolJSON struct {
		Tool         string  `json:"tool"`
		Share        float64 `json:"share_pct"`
		AvgReduction float64 `json:"avg_reduction_pct"`
		Requests     int64   `json:"requests"`
	}
	type output struct {
		Session  statJSON   `json:"session"`
		Today    statJSON   `json:"today"`
		Month    statJSON   `json:"this_month"`
		AllTime  statJSON   `json:"all_time"`
		TopTools []toolJSON `json:"top_tools_this_month"`
	}

	var topTools []toolJSON
	for _, t := range tools {
		topTools = append(topTools, toolJSON{
			Tool:         t.ToolName,
			Share:        t.SharePct,
			AvgReduction: t.AvgReduction,
			Requests:     t.RequestCount,
		})
	}

	out := output{
		Session:  statJSON{session.RequestCount, session.TokensSaved},
		Today:    statJSON{today.RequestCount, today.TokensSaved},
		Month:    statJSON{month.RequestCount, month.TokensSaved},
		AllTime:  statJSON{allTime.RequestCount, allTime.TokensSaved},
		TopTools: topTools,
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(out)
}

// ── Helpers ────────────────────────────────────────────────────────────────

func printCentered(s string) {
	pad := (boxWidth - 2 - len(s)) / 2
	fmt.Printf("│%s%s%s│\n",
		strings.Repeat(" ", pad),
		s,
		strings.Repeat(" ", boxWidth-2-pad-len(s)),
	)
}

func printSection(label string, requests, tokensSaved int64) {
	printRow(label, "")
	printRow("  Requests:", formatInt(requests))
	printRow("  Tokens saved:", formatInt(tokensSaved))
	printRow("", "")
}

func printRow(left, right string) {
	inner := boxWidth - 2
	if right == "" {
		padding := inner - len(left)
		if padding < 0 {
			padding = 0
		}
		fmt.Printf("│ %s%s│\n", left, strings.Repeat(" ", padding-1))
		return
	}
	gap := inner - len(left) - len(right) - 1
	if gap < 1 {
		gap = 1
	}
	fmt.Printf("│ %s%s%s│\n", left, strings.Repeat(" ", gap), right)
}

func formatInt(n int64) string {
	// Insert commas every 3 digits.
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

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

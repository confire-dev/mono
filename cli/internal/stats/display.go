package stats

import (
	"encoding/json"
	"os"
)

// PrintJSON renders stats as machine-readable JSON.
func PrintJSON(session SessionStats, today, month, allTime Stats, tools []ToolStat, sync SyncCounts) {
	type statJSON struct {
		Requests int64 `json:"requests"`
	}
	type toolJSON struct {
		Tool     string  `json:"tool"`
		Share    float64 `json:"share_pct"`
		Requests int64   `json:"requests"`
	}
	type syncJSON struct {
		Synced    int64 `json:"synced"`
		Pending   int64 `json:"pending"`
		NoAccount int64 `json:"no_account"`
	}
	type output struct {
		Session  statJSON   `json:"session"`
		Today    statJSON   `json:"today"`
		Month    statJSON   `json:"this_month"`
		AllTime  statJSON   `json:"all_time"`
		TopTools []toolJSON `json:"top_tools_this_month"`
		Sync     syncJSON   `json:"sync"`
	}

	var topTools []toolJSON
	for _, t := range tools {
		topTools = append(topTools, toolJSON{
			Tool:     t.ToolName,
			Share:    t.SharePct,
			Requests: t.RequestCount,
		})
	}

	out := output{
		Session:  statJSON{session.RequestCount},
		Today:    statJSON{today.RequestCount},
		Month:    statJSON{month.RequestCount},
		AllTime:  statJSON{allTime.RequestCount},
		TopTools: topTools,
		Sync:     syncJSON{sync.Synced, sync.Pending, sync.NoAccount},
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(out)
}

// SessionStats holds in-memory stats for the current daemon session.
type SessionStats struct {
	RequestCount int64
}

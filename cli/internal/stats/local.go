// Package stats manages local optimization statistics stored in SQLite.
// Privacy principle: only counts are stored — no content, no code, no responses.
// The database is owned entirely by the user and never read remotely.
package stats

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// DB wraps the local SQLite stats database.
type DB struct {
	db *sql.DB
}

// Open opens (or creates) the stats database at the given path.
func Open(path string) (*DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1) // SQLite is single-writer
	if err := migrate(db); err != nil {
		db.Close()
		return nil, err
	}
	return &DB{db: db}, nil
}

// Close closes the database.
func (s *DB) Close() {
	if s != nil && s.db != nil {
		s.db.Close()
	}
}

// Request is one optimization event to record.
type Request struct {
	EventID     string // UUID shared with Worker for idempotent sync
	ToolName    string // canonical family name: "bash", "figma", "github", …
	Optimizer   string // which optimizer ran: "local/bash", "github", …
	BytesBefore int
	BytesAfter  int
	Host        string // "claude_code", "cursor", …
}

// SyncRow is a pending-sync record returned by PendingSync.
type SyncRow struct {
	EventID       string
	ToolName      string
	Optimizer     string
	TokensBefore  int64
	TokensAfter   int64
	TokensSaved   int64
	Host          string
	CreatedAt     string
}

// RecordRequest writes one optimization event and updates the daily summary.
// Best-effort: errors are returned but callers may ignore them.
func (s *DB) RecordRequest(r Request) error {
	tokensBefore := r.BytesBefore / 4
	tokensAfter := r.BytesAfter / 4
	tokensSaved := tokensBefore - tokensAfter

	if _, err := s.db.Exec(`
		INSERT INTO requests
		  (event_id, tool_name, optimizer, tokens_before, tokens_after, tokens_saved, cost_saved_usd, host)
		VALUES (?, ?, ?, ?, ?, ?, 0.0, ?)`,
		r.EventID, r.ToolName, r.Optimizer, tokensBefore, tokensAfter, tokensSaved, r.Host,
	); err != nil {
		return err
	}

	date := time.Now().Format("2006-01-02")
	_, err := s.db.Exec(`
		INSERT INTO daily_summary (date, request_count, tokens_saved_total, cost_saved_total)
		VALUES (?, 1, ?, 0.0)
		ON CONFLICT(date) DO UPDATE SET
		  request_count      = request_count + 1,
		  tokens_saved_total = tokens_saved_total + excluded.tokens_saved_total`,
		date, tokensSaved,
	)
	return err
}

// PendingSync returns rows that have not yet been confirmed by the Worker.
// Called on daemon startup to retry any events that failed to post.
func (s *DB) PendingSync(limit int) ([]SyncRow, error) {
	rows, err := s.db.Query(`
		SELECT event_id, tool_name, optimizer,
		       tokens_before, tokens_after, tokens_saved, host, created_at
		FROM requests
		WHERE synced_at IS NULL AND event_id != ''
		ORDER BY created_at ASC
		LIMIT ?`, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []SyncRow
	for rows.Next() {
		var r SyncRow
		if err := rows.Scan(&r.EventID, &r.ToolName, &r.Optimizer,
			&r.TokensBefore, &r.TokensAfter, &r.TokensSaved,
			&r.Host, &r.CreatedAt); err != nil {
			continue
		}
		out = append(out, r)
	}
	return out, nil
}

// MarkSynced records that an event was accepted by the Worker.
func (s *DB) MarkSynced(eventID string) error {
	_, err := s.db.Exec(
		`UPDATE requests SET synced_at = datetime('now') WHERE event_id = ?`,
		eventID,
	)
	return err
}

// Stats is an aggregated count for a time period.
type Stats struct {
	RequestCount int64
	TokensSaved  int64
}

// Today returns stats for the current calendar day.
func (s *DB) Today() (Stats, error) {
	date := time.Now().Format("2006-01-02")
	var st Stats
	err := s.db.QueryRow(`
		SELECT COALESCE(request_count,0), COALESCE(tokens_saved_total,0)
		FROM daily_summary WHERE date = ?`, date,
	).Scan(&st.RequestCount, &st.TokensSaved)
	if err == sql.ErrNoRows {
		return Stats{}, nil
	}
	return st, err
}

// ThisMonth returns stats for the current calendar month.
func (s *DB) ThisMonth() (Stats, error) {
	month := time.Now().Format("2006-01")
	var st Stats
	err := s.db.QueryRow(`
		SELECT COALESCE(SUM(request_count),0), COALESCE(SUM(tokens_saved_total),0)
		FROM daily_summary WHERE date LIKE ?`, month+"-%",
	).Scan(&st.RequestCount, &st.TokensSaved)
	return st, err
}

// AllTime returns lifetime stats.
func (s *DB) AllTime() (Stats, error) {
	var st Stats
	err := s.db.QueryRow(`
		SELECT COALESCE(SUM(request_count),0), COALESCE(SUM(tokens_saved_total),0)
		FROM daily_summary`,
	).Scan(&st.RequestCount, &st.TokensSaved)
	return st, err
}

// ToolStat holds per-tool breakdown for a period.
type ToolStat struct {
	ToolName     string
	RequestCount int64
	AvgReduction float64 // 0–100 (percentage)
	SharePct     float64 // share of total requests this period
}

// TopToolsThisMonth returns the top N tools by request count for the current month.
func (s *DB) TopToolsThisMonth(limit int) ([]ToolStat, error) {
	month := time.Now().Format("2006-01")
	rows, err := s.db.Query(`
		SELECT
		  tool_name,
		  COUNT(*)                                                        AS cnt,
		  AVG(CASE WHEN tokens_before > 0
		        THEN CAST(tokens_saved AS REAL) / CAST(tokens_before AS REAL) * 100
		        ELSE 0.0 END)                                             AS avg_reduction
		FROM requests
		WHERE strftime('%Y-%m', created_at) = ?
		GROUP BY tool_name
		ORDER BY cnt DESC
		LIMIT ?`, month, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tools []ToolStat
	var total int64
	for rows.Next() {
		var t ToolStat
		if err := rows.Scan(&t.ToolName, &t.RequestCount, &t.AvgReduction); err != nil {
			continue
		}
		tools = append(tools, t)
		total += t.RequestCount
	}
	for i := range tools {
		if total > 0 {
			tools[i].SharePct = float64(tools[i].RequestCount) / float64(total) * 100
		}
	}
	return tools, nil
}

// Clear deletes all stats data.
func (s *DB) Clear() error {
	_, err := s.db.Exec(`DELETE FROM requests; DELETE FROM daily_summary`)
	return err
}

// ToolFamily maps a raw tool name to its canonical display family.
// Mirrors sanitizeToolType in worker/src/lib/buckets.ts.
func ToolFamily(toolName string) string {
	lower := strings.ToLower(toolName)
	families := []string{
		"figma", "github", "slack", "atlassian", "jira", "confluence",
		"notion", "clickup", "amplitude", "playwright", "zapier",
		"bash", "read", "webfetch", "websearch", "glob", "grep", "edit", "write",
	}
	for _, f := range families {
		if strings.Contains(lower, f) {
			return f
		}
	}
	return "generic"
}

// ── Schema + migrations ────────────────────────────────────────────────────

func migrate(db *sql.DB) error {
	var version int
	db.QueryRow(`PRAGMA user_version`).Scan(&version)

	if version < 1 {
		if _, err := db.Exec(`
			CREATE TABLE IF NOT EXISTS requests (
			  id            INTEGER PRIMARY KEY AUTOINCREMENT,
			  event_id      TEXT    NOT NULL DEFAULT '' UNIQUE,
			  tool_name     TEXT    NOT NULL,
			  optimizer     TEXT    NOT NULL,
			  tokens_before INTEGER NOT NULL,
			  tokens_after  INTEGER NOT NULL,
			  tokens_saved  INTEGER NOT NULL,
			  cost_saved_usd REAL   NOT NULL DEFAULT 0,
			  host          TEXT    NOT NULL DEFAULT '',
			  synced_at     TEXT,
			  created_at    TEXT    NOT NULL DEFAULT (datetime('now'))
			);

			CREATE TABLE IF NOT EXISTS daily_summary (
			  date               TEXT PRIMARY KEY,
			  request_count      INTEGER NOT NULL DEFAULT 0,
			  tokens_saved_total INTEGER NOT NULL DEFAULT 0,
			  cost_saved_total   REAL    NOT NULL DEFAULT 0
			);

			CREATE INDEX IF NOT EXISTS idx_requests_created
			  ON requests (created_at DESC);
			CREATE INDEX IF NOT EXISTS idx_requests_tool_month
			  ON requests (tool_name, created_at DESC);
			CREATE INDEX IF NOT EXISTS idx_requests_unsynced
			  ON requests (synced_at) WHERE synced_at IS NULL;

			PRAGMA user_version = 1;
		`); err != nil {
			return err
		}
	}

	// v1→v2: add sync columns to existing databases (no-op on fresh installs)
	if version == 0 {
		// Existing DB from before v1 schema: add missing columns gracefully.
		// SQLite ignores "duplicate column" errors only via separate statements.
		db.Exec(`ALTER TABLE requests ADD COLUMN event_id TEXT NOT NULL DEFAULT ''`)
		db.Exec(`ALTER TABLE requests ADD COLUMN synced_at TEXT`)
		db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_requests_event_id ON requests (event_id) WHERE event_id != ''`)
		db.Exec(`CREATE INDEX IF NOT EXISTS idx_requests_unsynced ON requests (synced_at) WHERE synced_at IS NULL`)
	}

	return nil
}

// Package stats manages local statistics stored in SQLite.
// Privacy principle: only counts are stored — no content, no code, no responses.
// The database is owned entirely by the user and never read remotely.
package stats

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// Sync status values for the requests table.
const (
	// StatusPending: recorded locally, not yet confirmed by Worker (may be retried).
	StatusPending = "pending"
	// StatusSynced: Worker accepted the event (idempotent — safe to retry).
	StatusSynced = "synced"
	// StatusNoAccount: no API key; will never be synced (local-only use).
	StatusNoAccount = "no_account"
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

// RawDB returns the underlying *sql.DB for callers that need direct access (e.g. MCP stats migration).
func (s *DB) RawDB() *sql.DB {
	if s == nil {
		return nil
	}
	return s.db
}

// NewEventID returns a UUID v4 suitable for idempotent sync with the Worker.
func NewEventID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// Fallback: mix time + pid (collision-resistant enough for our volumes)
		return fmt.Sprintf("%x%x%x", time.Now().UnixNano(), os.Getpid(), time.Now().UnixNano()>>32)
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant bits
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// Request is one firewall event to record.
type Request struct {
	EventID    string // UUID shared with Worker for idempotent sync
	ToolName   string // canonical family name: "bash", "figma", "github", …
	Host       string // "claude_code", "cursor", …
	SyncStatus string // StatusPending | StatusSynced | StatusNoAccount
}

// SyncRow is a pending-sync record returned by PendingSync.
type SyncRow struct {
	EventID   string
	ToolName  string
	Host      string
	CreatedAt string
}

// SyncCounts breaks down requests by sync status.
type SyncCounts struct {
	Synced    int64
	Pending   int64
	NoAccount int64
}

// RecordRequest writes one firewall event.
// Best-effort: errors are returned but callers may ignore them.
func (s *DB) RecordRequest(r Request) error {
	syncStatus := r.SyncStatus
	if syncStatus == "" {
		syncStatus = StatusPending
	}

	if _, err := s.db.Exec(`
		INSERT INTO requests
		  (event_id, tool_name, host, sync_status)
		VALUES (?, ?, ?, ?)`,
		r.EventID, r.ToolName, r.Host, syncStatus,
	); err != nil {
		return err
	}

	date := time.Now().Format("2006-01-02")
	_, err := s.db.Exec(`
		INSERT INTO daily_summary (date, request_count)
		VALUES (?, 1)
		ON CONFLICT(date) DO UPDATE SET
		  request_count = request_count + 1`,
		date,
	)
	return err
}

// MarkSynced marks an event as confirmed by the Worker.
func (s *DB) MarkSynced(eventID string) error {
	_, err := s.db.Exec(
		`UPDATE requests SET sync_status = ?, synced_at = datetime('now') WHERE event_id = ?`,
		StatusSynced, eventID,
	)
	return err
}

// PendingSync returns events with StatusPending that have not yet reached the Worker.
// Called at daemon startup to retry offline events.
func (s *DB) PendingSync(limit int) ([]SyncRow, error) {
	rows, err := s.db.Query(`
		SELECT event_id, tool_name, host, created_at
		FROM requests
		WHERE sync_status = ?
		ORDER BY created_at ASC
		LIMIT ?`, StatusPending, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []SyncRow
	for rows.Next() {
		var r SyncRow
		if err := rows.Scan(&r.EventID, &r.ToolName, &r.Host, &r.CreatedAt); err != nil {
			continue
		}
		out = append(out, r)
	}
	return out, nil
}

// GetSyncCounts returns a breakdown of requests by sync_status.
func (s *DB) GetSyncCounts() (SyncCounts, error) {
	rows, err := s.db.Query(`
		SELECT sync_status, COUNT(*) FROM requests GROUP BY sync_status`,
	)
	if err != nil {
		return SyncCounts{}, err
	}
	defer rows.Close()

	var c SyncCounts
	for rows.Next() {
		var status string
		var count int64
		if err := rows.Scan(&status, &count); err != nil {
			continue
		}
		switch status {
		case StatusSynced:
			c.Synced = count
		case StatusPending:
			c.Pending = count
		case StatusNoAccount:
			c.NoAccount = count
		}
	}
	return c, nil
}

// Stats is an aggregated count for a time period.
type Stats struct {
	RequestCount int64
}

// Today returns stats for the current calendar day.
func (s *DB) Today() (Stats, error) {
	date := time.Now().Format("2006-01-02")
	var st Stats
	err := s.db.QueryRow(`
		SELECT COALESCE(request_count,0)
		FROM daily_summary WHERE date = ?`, date,
	).Scan(&st.RequestCount)
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
		SELECT COALESCE(SUM(request_count),0)
		FROM daily_summary WHERE date LIKE ?`, month+"-%",
	).Scan(&st.RequestCount)
	return st, err
}

// AllTime returns lifetime stats.
func (s *DB) AllTime() (Stats, error) {
	var st Stats
	err := s.db.QueryRow(`
		SELECT COALESCE(SUM(request_count),0)
		FROM daily_summary`,
	).Scan(&st.RequestCount)
	return st, err
}

// TodayByTool returns stats for the current calendar day filtered to one tool family.
func (s *DB) TodayByTool(tool string) (Stats, error) {
	date := time.Now().Format("2006-01-02")
	var st Stats
	err := s.db.QueryRow(`
		SELECT COALESCE(COUNT(*),0)
		FROM requests WHERE date(created_at) = ? AND tool_name = ?`,
		date, tool,
	).Scan(&st.RequestCount)
	return st, err
}

// ThisMonthByTool returns stats for the current calendar month filtered to one tool family.
func (s *DB) ThisMonthByTool(tool string) (Stats, error) {
	month := time.Now().Format("2006-01")
	var st Stats
	err := s.db.QueryRow(`
		SELECT COALESCE(COUNT(*),0)
		FROM requests WHERE strftime('%Y-%m', created_at) = ? AND tool_name = ?`,
		month, tool,
	).Scan(&st.RequestCount)
	return st, err
}

// AllTimeByTool returns lifetime stats filtered to one tool family.
func (s *DB) AllTimeByTool(tool string) (Stats, error) {
	var st Stats
	err := s.db.QueryRow(`
		SELECT COALESCE(COUNT(*),0)
		FROM requests WHERE tool_name = ?`, tool,
	).Scan(&st.RequestCount)
	return st, err
}

// ToolStat holds per-tool breakdown for a period.
type ToolStat struct {
	ToolName     string
	RequestCount int64
	SharePct     float64 // share of total requests this period
}

// TopToolsThisMonth returns the top N tools by request count for the current month.
func (s *DB) TopToolsThisMonth(limit int) ([]ToolStat, error) {
	month := time.Now().Format("2006-01")
	rows, err := s.db.Query(`
		SELECT
		  tool_name,
		  COUNT(*) AS cnt
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
		if err := rows.Scan(&t.ToolName, &t.RequestCount); err != nil {
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

// ── Schema + migrations ─────────────────────────────────────────────────────

const schemaLatest = `
CREATE TABLE IF NOT EXISTS requests (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  event_id    TEXT    NOT NULL DEFAULT '',
  tool_name   TEXT    NOT NULL,
  host        TEXT    NOT NULL DEFAULT '',
  sync_status TEXT    NOT NULL DEFAULT 'pending',
  synced_at   TEXT,
  created_at  TEXT    NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS daily_summary (
  date          TEXT PRIMARY KEY,
  request_count INTEGER NOT NULL DEFAULT 0
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_requests_event_id
  ON requests (event_id) WHERE event_id != '';
CREATE INDEX IF NOT EXISTS idx_requests_created
  ON requests (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_requests_tool_month
  ON requests (tool_name, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_requests_sync_status
  ON requests (sync_status) WHERE sync_status = 'pending';
`

func migrate(db *sql.DB) error {
	var version int
	db.QueryRow(`PRAGMA user_version`).Scan(&version)

	if version == 0 {
		// Fresh install OR pre-versioned DB — apply full schema.
		if _, err := db.Exec(schemaLatest); err != nil {
			return err
		}
		// Best-effort column additions for pre-versioned databases.
		// SQLite's ALTER TABLE ADD COLUMN is idempotent-safe: we ignore errors
		// because "duplicate column name" means it already exists.
		db.Exec(`ALTER TABLE requests ADD COLUMN event_id TEXT NOT NULL DEFAULT ''`)
		db.Exec(`ALTER TABLE requests ADD COLUMN sync_status TEXT NOT NULL DEFAULT 'pending'`)
		db.Exec(`ALTER TABLE requests ADD COLUMN synced_at TEXT`)
		// Rows written before sync_status existed get 'no_account' so they
		// don't appear in the pending retry queue.
		db.Exec(`UPDATE requests SET sync_status = 'no_account' WHERE sync_status = '' OR sync_status IS NULL`)
		if _, err := db.Exec(`PRAGMA user_version = 3`); err != nil {
			return err
		}
	}

	// Future migrations: add `if version < N { ... }` blocks here.

	return nil
}

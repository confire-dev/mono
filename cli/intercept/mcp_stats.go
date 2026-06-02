package intercept

import (
	"database/sql"
	"time"
)

// MCPStatsWriter is a non-blocking writer for MCP tool event stats.
// It is wired into the daemon and receives events after all other handlers.
// Writes are fire-and-forget: a failure never delays a tool call.
type MCPStatsWriter struct {
	db *sql.DB
}

// NewMCPStatsWriter creates a writer backed by the given SQLite *sql.DB.
// The DB must already have mcp_server_stats and mcp_tool_events tables
// (created by MigrateMCPSchema).
func NewMCPStatsWriter(db *sql.DB) *MCPStatsWriter {
	return &MCPStatsWriter{db: db}
}

// MigrateMCPSchema creates the MCP stats tables if they do not exist.
// Call once after opening the stats DB.
func MigrateMCPSchema(db *sql.DB) error {
	_, err := db.Exec(mcpSchema)
	return err
}

const mcpSchema = `
CREATE TABLE IF NOT EXISTS mcp_server_stats (
    server_id          TEXT PRIMARY KEY,
    total_calls        INTEGER DEFAULT 0,
    total_bytes_in     INTEGER DEFAULT 0,
    total_bytes_out    INTEGER DEFAULT 0,
    secrets_redacted   INTEGER DEFAULT 0,
    injections_flagged INTEGER DEFAULT 0,
    blocks             INTEGER DEFAULT 0,
    reviews            INTEGER DEFAULT 0,
    noisiness_score    REAL    DEFAULT 0,
    riskiness_score    REAL    DEFAULT 0,
    last_seen          INTEGER
);

CREATE TABLE IF NOT EXISTS mcp_tool_events (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    ts              INTEGER NOT NULL,
    server_id       TEXT    NOT NULL,
    tool_name       TEXT    NOT NULL,
    phase           TEXT    NOT NULL,
    outcome         TEXT    NOT NULL,
    risk_score      INTEGER,
    bytes_in        INTEGER,
    bytes_out       INTEGER,
    tokens_est_in   INTEGER,
    tokens_est_out  INTEGER,
    secrets_found   INTEGER DEFAULT 0,
    injection_flags INTEGER DEFAULT 0,
    unicode_hits    INTEGER DEFAULT 0,
    redacted        INTEGER DEFAULT 0,
    truncated       INTEGER DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_mcp_events_server_ts ON mcp_tool_events(server_id, ts);
`

// MCPToolEvent is one row to insert into mcp_tool_events.
type MCPToolEvent struct {
	ServerID       string
	ToolName       string
	Phase          Phase
	Outcome        string // "passthrough" | "warn" | "review" | "block" | "sanitize" | "redact"
	RiskScore      int
	BytesIn        int
	BytesOut       int
	SecretsFound   int
	InjectionFlags int
	UnicodeHits    int
	Redacted       bool
	Truncated      bool
}

// RecordAsync writes the event to SQLite in a goroutine.
// The caller is never blocked.
func (w *MCPStatsWriter) RecordAsync(ev MCPToolEvent) {
	if w == nil || w.db == nil {
		return
	}
	go w.record(ev)
}

func (w *MCPStatsWriter) record(ev MCPToolEvent) {
	if ev.ServerID == "" {
		ev.ServerID = "_unknown"
	}
	ts := time.Now().Unix()
	tokIn := ev.BytesIn / 4
	tokOut := ev.BytesOut / 4
	redacted := 0
	if ev.Redacted {
		redacted = 1
	}
	truncated := 0
	if ev.Truncated {
		truncated = 1
	}

	_, _ = w.db.Exec(`
		INSERT INTO mcp_tool_events
		  (ts, server_id, tool_name, phase, outcome, risk_score,
		   bytes_in, bytes_out, tokens_est_in, tokens_est_out,
		   secrets_found, injection_flags, unicode_hits, redacted, truncated)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		ts, ev.ServerID, ev.ToolName, string(ev.Phase), ev.Outcome, ev.RiskScore,
		ev.BytesIn, ev.BytesOut, tokIn, tokOut,
		ev.SecretsFound, ev.InjectionFlags, ev.UnicodeHits, redacted, truncated,
	)

	_, _ = w.db.Exec(`
		INSERT INTO mcp_server_stats (server_id, total_calls, total_bytes_in, total_bytes_out,
		  secrets_redacted, injections_flagged, last_seen)
		VALUES (?, 1, ?, ?, ?, ?, ?)
		ON CONFLICT(server_id) DO UPDATE SET
		  total_calls        = total_calls + 1,
		  total_bytes_in     = total_bytes_in + excluded.total_bytes_in,
		  total_bytes_out    = total_bytes_out + excluded.total_bytes_out,
		  secrets_redacted   = secrets_redacted + excluded.secrets_redacted,
		  injections_flagged = injections_flagged + excluded.injections_flagged,
		  last_seen          = excluded.last_seen`,
		ev.ServerID, ev.BytesIn, ev.BytesOut, ev.SecretsFound, ev.InjectionFlags, ts,
	)
}

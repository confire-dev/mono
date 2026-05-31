package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/confire-dev/confire/auth"
	"github.com/confire-dev/confire/config"
	"github.com/confire-dev/confire/intercept"
	"github.com/confire-dev/confire/transport"
	"github.com/spf13/cobra"
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Start the Confire daemon (long-lived optimizer + HTTP/2 to Worker)",
	Long: `The daemon holds a warm HTTP/2 connection to the Confire Worker,
owns the API key, tracks session stats, and serves the hook shim over a
unix socket. Normally started automatically by the launchd plist.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDaemon()
	},
}

func init() {
	rootCmd.AddCommand(daemonCmd)
}

// ── Session tracking ──────────────────────────────────────────────────────────

type sessionStats struct {
	startedAt      time.Time
	sessionID      string
	integration    string
	cliVersion     string
	totalCalls     int
	optimizedCalls int
	rawBytes       int64
	optimizedBytes int64
	// Biggest single save this session (for the 🔥 summary)
	biggestWinTool string
	biggestWinRaw  int
	biggestWinOpt  int
}

type daemonState struct {
	mu       sync.Mutex
	sessions map[string]*sessionStats
	apiKey   string
	deviceID string
	cfg      config.Config
}

func (ds *daemonState) onSessionStart(event intercept.InterceptEvent) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	if event.Session.ID == "" {
		return
	}
	ds.sessions[event.Session.ID] = &sessionStats{
		startedAt:   time.Now(),
		sessionID:   event.Session.ID,
		integration: string(event.Host),
	}
}

func (ds *daemonState) onToolResult(event intercept.InterceptEvent, result intercept.InterceptResult) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	sess := ds.sessions[event.Session.ID]
	if sess == nil {
		return
	}
	sess.totalCalls++
	if result.Kind == intercept.ResultReplaceOutput && result.Stats != nil {
		sess.optimizedCalls++
		sess.rawBytes += int64(result.Stats.BeforeBytes)
		sess.optimizedBytes += int64(result.Stats.AfterBytes)
		// Track biggest single save for session summary
		savedBytes := result.Stats.BeforeBytes - result.Stats.AfterBytes
		if savedBytes > sess.biggestWinRaw-sess.biggestWinOpt {
			name := event.Tool.GetServerHint()
			if event.Tool != nil {
				name = event.Tool.Name
			}
			sess.biggestWinTool = normalizeToolName(name)
			sess.biggestWinRaw  = result.Stats.BeforeBytes
			sess.biggestWinOpt  = result.Stats.AfterBytes
		}
		go ds.postToolEvent(event, result)
	}
}

func (ds *daemonState) onSessionEnd(event intercept.InterceptEvent) {
	ds.mu.Lock()
	sess := ds.sessions[event.Session.ID]
	delete(ds.sessions, event.Session.ID)
	ds.mu.Unlock()

	if sess == nil {
		return
	}

	// Print session summary using 🔥 formatting
	if summary := sessionSummary(sess, ds.cfg); summary != "" {
		fmt.Fprintf(os.Stderr, "\n%s\n\n", summary)
	}

	go ds.postSessionEnd(sess)
}

// ── Telemetry posting ─────────────────────────────────────────────────────────

type telemetryPayload struct {
	EventID            string `json:"event_id"`
	EventType          string `json:"event_type"`
	CLIVersion         string `json:"cli_version,omitempty"`
	Integration        string `json:"integration,omitempty"`
	SessionID          string `json:"session_id,omitempty"`
	ToolType           string `json:"tool_type,omitempty"`
	Optimizer          string `json:"optimizer,omitempty"`
	RawBytes           int    `json:"raw_bytes,omitempty"`
	OptimizedBytes     int    `json:"optimized_bytes,omitempty"`
	DurationMs         int    `json:"duration_ms,omitempty"`
	WasCached          bool   `json:"was_cached,omitempty"`
	AnalyticsConsented bool   `json:"analytics_consented"`
	// Session end only
	TotalToolCalls  int `json:"total_tool_calls,omitempty"`
	OptimizedCalls  int `json:"optimized_calls,omitempty"`
	TotalRawBytes   int `json:"raw_bytes_total,omitempty"`
	TotalOptBytes   int `json:"optimized_bytes_total,omitempty"`
}

func (ds *daemonState) postToolEvent(event intercept.InterceptEvent, result intercept.InterceptResult) {
	if ds.apiKey == "" {
		return
	}
	payload := telemetryPayload{
		EventID:            randomHex(8),
		EventType:          "tool_call_optimized",
		CLIVersion:         buildVersion,
		Integration:        string(event.Host),
		SessionID:          event.Session.ID,
		AnalyticsConsented: ds.cfg.Telemetry,
	}
	if event.Tool != nil {
		payload.ToolType = resolveToolType(event.Tool)
		payload.DurationMs = event.Tool.DurationMs
	}
	if result.Stats != nil {
		payload.Optimizer       = result.Stats.Optimizer
		payload.RawBytes        = result.Stats.BeforeBytes
		payload.OptimizedBytes  = result.Stats.AfterBytes
	}
	ds.postEvent(payload)
}

func (ds *daemonState) postSessionEnd(sess *sessionStats) {
	if ds.apiKey == "" {
		return
	}
	ds.postEvent(telemetryPayload{
		EventID:            randomHex(8),
		EventType:          "session_end",
		CLIVersion:         buildVersion,
		Integration:        sess.integration,
		SessionID:          sess.sessionID,
		TotalToolCalls:     sess.totalCalls,
		OptimizedCalls:     sess.optimizedCalls,
		TotalRawBytes:      int(sess.rawBytes),
		TotalOptBytes:      int(sess.optimizedBytes),
		AnalyticsConsented: ds.cfg.Telemetry,
	})
}

func (ds *daemonState) postSessionStart(event intercept.InterceptEvent) {
	if ds.apiKey == "" {
		return
	}
	ds.postEvent(telemetryPayload{
		EventID:            randomHex(8),
		EventType:          "session_start",
		CLIVersion:         buildVersion,
		Integration:        string(event.Host),
		SessionID:          event.Session.ID,
		AnalyticsConsented: ds.cfg.Telemetry,
	})
}

func (ds *daemonState) postEvent(payload telemetryPayload) {
	body, err := json.Marshal(payload)
	if err != nil {
		return
	}
	workerURL := workerURLEnv()
	req, err := http.NewRequest(http.MethodPost, workerURL+"/v1/events", bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ds.apiKey)
	req.Header.Set("X-Confire-Device", ds.deviceID)
	for k, v := range transport.SignedHeaders(ds.apiKey, ds.deviceID, body) {
		req.Header.Set(k, v)
	}
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	resp.Body.Close()
}

// ── Daemon main ───────────────────────────────────────────────────────────────

func runDaemon() error {
	socketPath := daemonSocketPath()
	os.Remove(socketPath)

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return fmt.Errorf("listen %s: %w", socketPath, err)
	}
	defer func() {
		listener.Close()
		os.Remove(socketPath)
	}()
	os.Chmod(socketPath, 0600)

	cfg := config.Load()

	apiKey  := resolveAPIKey()
	deviceID := ""
	if d, err := auth.DeviceID(); err == nil {
		deviceID = d
	}

	state := &daemonState{
		sessions: make(map[string]*sessionStats),
		apiKey:   apiKey,
		deviceID: deviceID,
		cfg:      cfg,
	}

	t := buildDaemonTransportWithKey(apiKey, deviceID)

	fmt.Fprintf(os.Stderr, "[confire daemon] v%s listening on %s\n", buildVersion, socketPath)
	if apiKey != "" {
		fmt.Fprintf(os.Stderr, "[confire daemon] cloud mode → %s\n", workerURLEnv())
	} else {
		fmt.Fprintln(os.Stderr, "[confire daemon] local mode — run `confire login` for cloud optimization")
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-sigCh
		fmt.Fprintln(os.Stderr, "[confire daemon] shutting down")
		listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			return nil
		}
		go handleConn(conn, t, state)
	}
}

func handleConn(conn net.Conn, t transport.Transport, state *daemonState) {
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(30 * time.Second))

	var event intercept.InterceptEvent
	if err := json.NewDecoder(conn).Decode(&event); err != nil {
		return
	}

	// Session lifecycle events → track state + async telemetry.
	switch event.Phase {
	case intercept.PhaseSessionStart:
		state.onSessionStart(event)
		go state.postSessionStart(event)
		// Return notification for Claude to see.
		result := intercept.InterceptResult{
			Kind:    intercept.ResultAddContext,
			Context: sessionStartNotification(state),
		}
		json.NewEncoder(conn).Encode(result)
		return

	case intercept.PhaseSessionEnd:
		state.onSessionEnd(event)
		json.NewEncoder(conn).Encode(intercept.InterceptResult{Kind: intercept.ResultPassthrough})
		return
	}

	// Tool calls → optimize + track stats + notify.
	result, _ := t.Send(event)
	state.onToolResult(event, result)

	// Format the 🔥 notification
	toolName := ""
	if event.Tool != nil {
		toolName = event.Tool.Name
	}
	stderrLine, contextLine := notifyResult(result, toolName, state.cfg)

	if stderrLine != "" {
		fmt.Fprintln(os.Stderr, stderrLine)
	}

	// Big save → add one line to Claude's context (visible in conversation)
	if contextLine != "" && result.Kind == intercept.ResultReplaceOutput {
		result.Context = contextLine
		if result.Kind == intercept.ResultReplaceOutput {
			// Upgrade to add-context only if there's no tool output to set
			// (we still send the optimized output + context via hookSpecificOutput)
		}
	}

	json.NewEncoder(conn).Encode(result)
}

func sessionStartNotification(state *daemonState) string {
	if state.apiKey == "" {
		return "⚠️ Confire: not logged in — run `confire login` to enable optimization."
	}
	return fmt.Sprintf("✓ Confire v%s active (cloud optimizer)", buildVersion)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func buildDaemonTransportWithKey(apiKey, deviceID string) transport.Transport {
	if apiKey == "" {
		// No account → no optimization. Local runs as fallback only when the
		// Worker can't help (quota exhausted, offline); it is not a free tier.
		return transport.NewPassthrough()
	}
	worker := transport.NewWorker(workerURLEnv(), apiKey, deviceID)
	return transport.NewFallback(worker, transport.NewLocal(""))
}

// Keep old name for backward compat with other callers.
func buildDaemonTransport() transport.Transport {
	key := resolveAPIKey()
	did, _ := auth.DeviceID()
	return buildDaemonTransportWithKey(key, did)
}

func resolveAPIKey() string {
	if k := apiKeyEnv(); k != "" {
		return k
	}
	k, _ := auth.LoadKey()
	return k
}

func resolveToolType(tool *intercept.Tool) string {
	if tool.IsMCP && tool.MCPServer != "" {
		return tool.MCPServer
	}
	return tool.Name
}

func randomHex(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = "0123456789abcdef"[time.Now().UnixNano()>>uint(i)&0xf]
	}
	return string(b)
}

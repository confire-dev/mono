package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/confire-dev/confire/auth"
	"github.com/confire-dev/confire/config"
	"github.com/confire-dev/confire/guardrail"
	"github.com/confire-dev/confire/intercept"
	"github.com/confire-dev/confire/internal/stats"
	"github.com/confire-dev/confire/policy"
	"github.com/confire-dev/confire/sanitize"
	"github.com/confire-dev/confire/transport"
	"github.com/spf13/cobra"
)

var daemonCmd = &cobra.Command{
	Use:    "daemon",
	Short:  "Start the Confire daemon",
	Hidden: true, // internal — users use `confire start` / `confire stop`
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
	// Firewall stats
	blockedCalls    int
	reviewedCalls   int
	warnedCalls     int
	sanitizedCalls  int
	secretsRedacted int
}

type daemonState struct {
	mu        sync.Mutex
	sessions  map[string]*sessionStats
	apiKey    string
	deviceID  string
	cfg       config.Config
	statsDB   *stats.DB
	guardrail *guardrail.Handler
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
		if event.Session.ID == "" {
			return
		}
		// SessionStart hook not configured — create lazily so stats still record.
		sess = &sessionStats{
			startedAt:   time.Now(),
			sessionID:   event.Session.ID,
			integration: string(event.Host),
		}
		ds.sessions[event.Session.ID] = sess
	}
	sess.totalCalls++
	if result.Kind == intercept.ResultReplaceOutput && result.Stats != nil {
		sess.optimizedCalls++
		sess.rawBytes += int64(result.Stats.BeforeBytes)
		sess.optimizedBytes += int64(result.Stats.AfterBytes)
		// Track biggest single save for session summary
		savedBytes := result.Stats.BeforeBytes - result.Stats.AfterBytes
		if savedBytes > sess.biggestWinRaw-sess.biggestWinOpt {
			name := ""
			if event.Tool != nil {
				name = event.Tool.GetServerHint() // returns MCP server name or lowercase tool name
			}
			sess.biggestWinTool = normalizeToolName(name)
			sess.biggestWinRaw  = result.Stats.BeforeBytes
			sess.biggestWinOpt  = result.Stats.AfterBytes
		}

		// Record locally first (works offline). eventID ties this to the Worker post.
		eventID := stats.NewEventID()
		toolName := ""
		if event.Tool != nil {
			toolName = event.Tool.Name
		}
		// SyncStatus reflects whether this event can ever reach the Worker.
		syncStatus := stats.StatusPending
		if ds.apiKey == "" {
			syncStatus = stats.StatusNoAccount
		}
		if ds.statsDB != nil {
			ds.statsDB.RecordRequest(stats.Request{
				EventID:     eventID,
				ToolName:    stats.ToolFamily(toolName),
				Optimizer:   result.Stats.Optimizer,
				BytesBefore: result.Stats.BeforeBytes,
				BytesAfter:  result.Stats.AfterBytes,
				Host:        string(event.Host),
				SyncStatus:  syncStatus,
			})
		}

		go ds.postToolEvent(event, result, eventID)
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

func (ds *daemonState) postToolEvent(event intercept.InterceptEvent, result intercept.InterceptResult, eventID string) {
	if ds.apiKey == "" {
		return
	}
	payload := telemetryPayload{
		EventID:            eventID,
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
		payload.Optimizer      = result.Stats.Optimizer
		payload.RawBytes       = result.Stats.BeforeBytes
		payload.OptimizedBytes = result.Stats.AfterBytes
	}
	if ds.postEvent(payload) && ds.statsDB != nil {
		ds.statsDB.MarkSynced(eventID)
	}
}

func (ds *daemonState) postSessionEnd(sess *sessionStats) {
	if ds.apiKey == "" {
		return
	}
	ds.postEvent(telemetryPayload{
		EventID:            stats.NewEventID(),
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
	// Flush any pending events now that we know the network is up.
	if ds.statsDB != nil {
		ds.syncPending()
	}
}

func (ds *daemonState) postSessionStart(event intercept.InterceptEvent) {
	if ds.apiKey == "" {
		return
	}
	ds.postEvent(telemetryPayload{
		EventID:            stats.NewEventID(),
		EventType:          "session_start",
		CLIVersion:         buildVersion,
		Integration:        string(event.Host),
		SessionID:          event.Session.ID,
		AnalyticsConsented: ds.cfg.Telemetry,
	})
}

// periodicSync calls syncPending on a fixed interval until the daemon exits.
func (ds *daemonState) periodicSync(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		ds.syncPending()
	}
}

// syncPending retries unsynced local stats rows against the Worker.
func (ds *daemonState) syncPending() {
	pending, err := ds.statsDB.PendingSync(200)
	if err != nil || len(pending) == 0 {
		return
	}
	for _, row := range pending {
		payload := telemetryPayload{
			EventID:            row.EventID,
			EventType:          "tool_call_optimized",
			CLIVersion:         buildVersion,
			Integration:        row.Host,
			ToolType:           row.ToolName,
			Optimizer:          row.Optimizer,
			RawBytes:           int(row.TokensBefore * 4),
			OptimizedBytes:     int(row.TokensAfter * 4),
			AnalyticsConsented: ds.cfg.Telemetry,
		}
		if ds.postEvent(payload) {
			ds.statsDB.MarkSynced(row.EventID)
		}
	}
}

func (ds *daemonState) postEvent(payload telemetryPayload) bool {
	body, err := json.Marshal(payload)
	if err != nil {
		return false
	}
	workerURL := workerURLEnv()
	req, err := http.NewRequest(http.MethodPost, workerURL+"/v1/events", bytes.NewReader(body))
	if err != nil {
		return false
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
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
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
		os.Remove(daemonPIDPath())
	}()
	os.Chmod(socketPath, 0600)

	// Write PID file so `confire stop` can signal us.
	pidPath := daemonPIDPath()
	os.WriteFile(pidPath, []byte(strconv.Itoa(os.Getpid())), 0600)

	cfg := config.Load()

	apiKey := resolveAPIKey()
	deviceID := ""
	if d, err := auth.DeviceID(); err == nil {
		deviceID = d
	}

	statsDB, _ := stats.Open(statsDBPath())

	// Initialize guardrail (PreToolUse firewall).
	policyEngine := policy.NewEngine(policy.LoadRules())
	mode := policy.Mode(cfg.EffectiveMode())

	state := &daemonState{
		sessions:  make(map[string]*sessionStats),
		apiKey:    apiKey,
		deviceID:  deviceID,
		cfg:       cfg,
		statsDB:   statsDB,
		guardrail: guardrail.New(policyEngine, mode),
	}

	if statsDB != nil {
		defer statsDB.Close()
	}

	t := buildDaemonTransportWithKey(apiKey, deviceID)

	// Sync pending events: on startup, then every 5 minutes.
	if statsDB != nil && apiKey != "" {
		go state.syncPending()
		go state.periodicSync(5 * time.Minute)
	}

	fmt.Fprintf(os.Stderr, "[confire daemon] v%s listening on %s\n", buildVersion, socketPath)
	if apiKey != "" {
		fmt.Fprintf(os.Stderr, "[confire daemon] cloud mode → %s\n", workerURLEnv())
	} else {
		fmt.Fprintln(os.Stderr, "[confire daemon] no account — run `confire login` to enable optimization")
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
		ctxMsg, systemMsg := sessionStartNotification(state)
		if ctxMsg != "" {
			json.NewEncoder(conn).Encode(intercept.InterceptResult{
				Kind:          intercept.ResultAddContext,
				Context:       ctxMsg,
				SystemMessage: systemMsg,
			})
		} else {
			json.NewEncoder(conn).Encode(intercept.InterceptResult{
				Kind:          intercept.ResultPassthrough,
				SystemMessage: systemMsg,
			})
		}
		return

	case intercept.PhaseSessionEnd:
		state.onSessionEnd(event)
		json.NewEncoder(conn).Encode(intercept.InterceptResult{Kind: intercept.ResultPassthrough})
		return

	case intercept.PhaseToolPre:
		// Firewall: evaluate before the tool executes.
		// bypass-next flag allows one-shot override.
		if policy.ConsumeBypassNext() {
			json.NewEncoder(conn).Encode(intercept.InterceptResult{Kind: intercept.ResultPassthrough})
			return
		}
		if state.cfg.IsFirewallEnabled() && state.guardrail != nil {
			result, _ := state.guardrail.Run(event)
			state.onFirewallResult(event, result)
			json.NewEncoder(conn).Encode(result)
		} else {
			json.NewEncoder(conn).Encode(intercept.InterceptResult{Kind: intercept.ResultPassthrough})
		}
		return
	}

	// PostToolUse and other tool phases — run sanitization then optimizer.
	var sanitizeReport intercept.SanitizeReport
	if state.cfg.IsFirewallEnabled() && event.Phase == intercept.PhaseToolPost {
		event, sanitizeReport = state.sanitizeOutput(event)
	}

	// Tool calls → optimize + track stats + notify (when enabled for this host).
	var result intercept.InterceptResult
	if shouldOptimizeEvent(event, state.cfg) {
		result, _ = t.Send(event)
	} else {
		result = intercept.InterceptResult{Kind: intercept.ResultPassthrough}
	}
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

	// Standardized post-tool steer for hosts that cannot replace native output.
	caps := state.cfg.CapabilitiesFor(event.Host)
	if caps.PostToolSteer {
		steerCtx := intercept.FormatPostToolSteer(intercept.PostToolSteerInput{
			Host:                event.Host,
			ToolName:            toolName,
			NativeUnreplaceable: !caps.NativeOutputReplaceable && event.Tool != nil && !event.Tool.IsMCP,
			OutputReplaced:      result.Kind == intercept.ResultReplaceOutput,
			Report:              sanitizeReport,
			Stats:               result.Stats,
			ExtraContext:        contextLine,
		})
		if steerCtx != "" {
			result.Context = intercept.JoinContext(steerCtx, result.Context)
			if result.Kind == intercept.ResultPassthrough {
				result.Kind = intercept.ResultAddContext
			}
		} else if contextLine != "" && result.Kind == intercept.ResultReplaceOutput {
			result.Context = intercept.JoinContext(result.Context, contextLine)
		}
	} else if contextLine != "" && result.Kind == intercept.ResultReplaceOutput {
		// Claude Code: big-save line only when output was replaced.
		result.Context = contextLine
	}

	json.NewEncoder(conn).Encode(result)
}

// onFirewallResult tracks PreToolUse firewall decisions.
func (ds *daemonState) onFirewallResult(event intercept.InterceptEvent, result intercept.InterceptResult) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	sess := ds.sessions[event.Session.ID]
	if sess == nil {
		return
	}
	switch result.Kind {
	case intercept.ResultBlock:
		sess.blockedCalls++
	case intercept.ResultReview:
		sess.reviewedCalls++
	case intercept.ResultWarn:
		sess.warnedCalls++
	}
}

// sanitizeOutput runs secret redaction and prompt-injection sanitization
// on the raw tool output before it goes to the optimizer.
// Returns the (potentially modified) event and a report for post-tool steer.
func (ds *daemonState) sanitizeOutput(event intercept.InterceptEvent) (intercept.InterceptEvent, intercept.SanitizeReport) {
	report := intercept.SanitizeReport{}
	if event.Tool == nil || event.Tool.Output == nil {
		return event, report
	}
	text, ok := sanitize.OutputToString(event.Tool.Output)
	if !ok || len(text) < 100 {
		return event, report
	}

	modified := false
	sess := ds.sessions[event.Session.ID]

	// Secret redaction.
	redacted, count, types := sanitize.Redact(text)
	if count > 0 {
		text = redacted
		modified = true
		report.SecretsRedacted = count
		report.SecretTypes = types
		if sess != nil {
			ds.mu.Lock()
			sess.secretsRedacted += count
			sess.sanitizedCalls++
			ds.mu.Unlock()
		}
		fmt.Fprintf(os.Stderr, "[confire] redacted %d secret(s) from %s output\n", count, event.Tool.Name)
	}

	// Prompt-injection sanitization.
	sanitized, found, _ := sanitize.Sanitize(text)
	if found {
		text = sanitized
		modified = true
		report.InjectionFound = true
		if sess != nil {
			ds.mu.Lock()
			sess.sanitizedCalls++
			ds.mu.Unlock()
		}
		fmt.Fprintf(os.Stderr, "[confire] sanitized prompt-injection pattern from %s output\n", event.Tool.Name)
	}

	if modified {
		// Clone the tool to avoid mutating shared state.
		toolCopy := *event.Tool
		toolCopy.Output = text
		event.Tool = &toolCopy
	}
	return event, report
}

func shouldOptimizeEvent(event intercept.InterceptEvent, cfg config.Config) bool {
	if event.Phase != intercept.PhaseToolPost || event.Tool == nil {
		return false
	}
	caps := cfg.CapabilitiesFor(event.Host)
	if event.Tool.IsMCP {
		return caps.OptimizeMCP
	}
	return caps.OptimizeNative
}

// sessionStartNotification returns context (for Claude) and systemMessage (shown in Claude Code).
// A healthy balanced session injects nothing into Claude — saving context is the product.
func sessionStartNotification(state *daemonState) (contextMsg, systemMsg string) {
	mode := state.cfg.EffectiveMode()
	statusLine := sessionStatusLine(state, mode)

	switch {
	case state.apiKey == "":
		msg := "⚠️ Confire: not logged in — cloud optimization disabled. Run `confire login`. Run `confire help` for commands."
		return msg, msg
	case mode == "strict":
		msg := "[Confire] Strict mode active. Dangerous tool calls will be blocked, not just reviewed. Run `confire help` for commands."
		return msg, msg
	case mode == "bypass":
		msg := "[Confire] Bypass mode active. Firewall and optimization are disabled. Run `confire help` for commands."
		return msg, msg
	default:
		if msg, ok := consumeWelcomePending(state); ok {
			return msg, msg
		}
		return "", statusLine
	}
}

func sessionStatusLine(state *daemonState, mode string) string {
	if state.guardrail != nil {
		return fmt.Sprintf("[confire] v%s active — %s mode, %d rules loaded",
			buildVersion, mode, len(state.guardrail.Rules()))
	}
	return fmt.Sprintf("[confire] v%s active — %s mode", buildVersion, mode)
}

func consumeWelcomePending(state *daemonState) (string, bool) {
	cfg := config.Load()
	if !cfg.WelcomePending {
		return "", false
	}
	cfg.WelcomePending = false
	_ = config.Save(cfg)
	state.cfg.WelcomePending = false
	return "🔥 Confire context firewall is active. Run `confire help` for commands or `confire status` to check your setup.", true
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func buildDaemonTransportWithKey(apiKey, deviceID string) transport.Transport {
	if apiKey == "" {
		// No account → no optimization. Local runs as fallback only when the
		// Worker can't help (quota exhausted, offline); it is not a free tier.
		return transport.NewPassthrough()
	}
	worker := transport.NewWorker(workerURLEnv(), apiKey, deviceID)
	// Local runs first (Bash/Read/WebFetch — zero latency, works on free plan).
	// Cloud gets the remainder: MCP tools and anything local passes through.
	return transport.NewLocalFirst(transport.NewLocal(""), worker)
}

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

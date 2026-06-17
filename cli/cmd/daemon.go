package cmd

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
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
	"github.com/confire-dev/confire/firewall"
	"github.com/confire-dev/confire/guardrail"
	"github.com/confire-dev/confire/intercept"
	"github.com/confire-dev/confire/policy"
	"github.com/confire-dev/confire/provenance"
	"github.com/confire-dev/confire/receipt"
	"github.com/confire-dev/confire/sanitize"
	"github.com/confire-dev/confire/transport"
	"github.com/spf13/cobra"
)

func newEventID() string {
	b := make([]byte, 16)
	rand.Read(b) //nolint:errcheck
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return hex.EncodeToString(b[:4]) + "-" + hex.EncodeToString(b[4:6]) + "-" +
		hex.EncodeToString(b[6:8]) + "-" + hex.EncodeToString(b[8:10]) + "-" +
		hex.EncodeToString(b[10:])
}

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
	startedAt   time.Time
	sessionID   string
	integration string
	totalCalls  int
	// Firewall stats
	blockedCalls    int
	reviewedCalls   int
	warnedCalls     int
	sanitizedCalls  int
	secretsRedacted int
	// Provenance trust distribution
	mcpUnknownCalls        int
	externalUntrustedCalls int
	// Ring buffer of recent provenance labels for cross-tool flow detection
	recentLabels []provenance.ProvenanceLabel
	// pendingContextIDs stores daemon-generated context_ids for hosts that don't
	// provide tool_use_id (windsurf, cline, openclaw, opencode). Keyed by tool name.
	pendingContextIDs map[string]string
}

func (s *sessionStats) getRecentLabels() []provenance.ProvenanceLabel {
	if s == nil {
		return nil
	}
	return s.recentLabels
}

type daemonState struct {
	mu           sync.Mutex
	sessions     map[string]*sessionStats
	apiKey       string
	deviceID     string
	cfg          config.Config
	guardrail    *guardrail.Handler
	mcpRisk      *intercept.MCPRiskHandler
	mcpSanitize  *intercept.MCPSanitizeHandler
	receiptStore *receipt.Store
	receiptKey   ed25519.PrivateKey
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

func (ds *daemonState) onToolResult(event intercept.InterceptEvent, report intercept.SanitizeReport) provenance.ProvenanceLabel {
	label := provenance.Classify(event, report)

	ds.mu.Lock()
	defer ds.mu.Unlock()
	sess := ds.sessions[event.Session.ID]
	if sess == nil {
		if event.Session.ID == "" {
			return label
		}
		sess = &sessionStats{
			startedAt:   time.Now(),
			sessionID:   event.Session.ID,
			integration: string(event.Host),
		}
		ds.sessions[event.Session.ID] = sess
	}
	sess.totalCalls++
	if report.SecretsRedacted > 0 {
		sess.secretsRedacted += report.SecretsRedacted
		sess.sanitizedCalls++
	}
	if report.InjectionFound {
		sess.sanitizedCalls++
	}
	// Track trust distribution.
	switch label.TrustLevel {
	case provenance.TrustMCPUnknown:
		sess.mcpUnknownCalls++
	case provenance.TrustExternalUntrusted:
		sess.externalUntrustedCalls++
	}
	// Maintain ring buffer for cross-tool flow detection (cap 10).
	if len(sess.recentLabels) >= 10 {
		sess.recentLabels = sess.recentLabels[1:]
	}
	sess.recentLabels = append(sess.recentLabels, label)
	return label
}

func (ds *daemonState) onSessionEnd(event intercept.InterceptEvent) {
	ds.mu.Lock()
	sess := ds.sessions[event.Session.ID]
	delete(ds.sessions, event.Session.ID)
	ds.mu.Unlock()

	if sess == nil {
		return
	}

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
	// Security event fields
	RiskLevel      string `json:"risk_level,omitempty"`
	ActionTaken    string `json:"action_taken,omitempty"`
	PatternMatched string `json:"pattern_matched,omitempty"`
	Sanitized      bool   `json:"sanitized,omitempty"`
	SecretsRedacted int   `json:"secrets_redacted,omitempty"`
	// Provenance event fields
	TrustLevel  string   `json:"trust_level,omitempty"`
	Flags       []string `json:"flags,omitempty"`
	MCPServer   string   `json:"mcp_server,omitempty"`
	OriginDomain string  `json:"origin_domain,omitempty"`
	AnalyticsConsented bool `json:"analytics_consented"`
	// Session end only
	TotalToolCalls int `json:"total_tool_calls,omitempty"`
	BlockedCalls   int `json:"blocked_calls,omitempty"`
	ReviewedCalls  int `json:"reviewed_calls,omitempty"`
	WarnedCalls    int `json:"warned_calls,omitempty"`
	SanitizedCalls int `json:"sanitized_calls,omitempty"`
	SecretsTotal   int `json:"secrets_total,omitempty"`
}

func (ds *daemonState) postSecurityEvent(event intercept.InterceptEvent, result intercept.InterceptResult) {
	if ds.apiKey == "" {
		return
	}
	payload := telemetryPayload{
		EventID:            newEventID(),
		EventType:          "security_event",
		CLIVersion:         buildVersion,
		Integration:        string(event.Host),
		SessionID:          event.Session.ID,
		AnalyticsConsented: ds.cfg.Telemetry,
	}
	if event.Tool != nil {
		payload.ToolType = resolveToolType(event.Tool)
	}
	switch result.Kind {
	case intercept.ResultBlock:
		payload.RiskLevel = "HIGH"
		payload.ActionTaken = "BLOCKED"
	case intercept.ResultReview:
		payload.RiskLevel = "MEDIUM"
		payload.ActionTaken = "REVIEW_REQUIRED"
	case intercept.ResultWarn:
		payload.RiskLevel = "LOW"
		payload.ActionTaken = "WARNED"
	case intercept.ResultSanitize:
		payload.ActionTaken = "SANITIZED"
		payload.Sanitized = true
	}
	ds.postEvent(payload)
}

func (ds *daemonState) postSessionEnd(sess *sessionStats) {
	if ds.apiKey == "" {
		return
	}
	ds.postEvent(telemetryPayload{
		EventID:            newEventID(),
		EventType:          "session_end",
		CLIVersion:         buildVersion,
		Integration:        sess.integration,
		SessionID:          sess.sessionID,
		TotalToolCalls:     sess.totalCalls,
		BlockedCalls:       sess.blockedCalls,
		ReviewedCalls:      sess.reviewedCalls,
		WarnedCalls:        sess.warnedCalls,
		SanitizedCalls:     sess.sanitizedCalls,
		SecretsTotal:       sess.secretsRedacted,
		AnalyticsConsented: ds.cfg.Telemetry,
	})
}

func (ds *daemonState) postProvenanceEvent(label provenance.ProvenanceLabel) {
	if ds.apiKey == "" {
		return
	}
	sanitized := false
	for _, f := range label.Flags {
		if f == provenance.FlagSanitized {
			sanitized = true
			break
		}
	}
	ds.postEvent(telemetryPayload{
		EventID:            newEventID(),
		EventType:          "provenance_event",
		CLIVersion:         buildVersion,
		SessionID:          label.SessionID,
		Integration:        label.Client,
		ToolType:           label.SourceTool,
		TrustLevel:         string(label.TrustLevel),
		Flags:              label.Flags,
		MCPServer:          label.MCPServer,
		OriginDomain:       label.OriginDomain,
		Sanitized:          sanitized,
		SecretsRedacted:    label.RedactionsCount,
		AnalyticsConsented: ds.cfg.Telemetry,
	})
}

// writeProvenanceLabel appends a provenance label to the session JSONL store.
// Only metadata is written — no raw tool output or secret values.
func writeProvenanceLabel(label provenance.ProvenanceLabel) {
	if label.SessionID == "" {
		return
	}
	data, err := json.Marshal(label)
	if err != nil {
		return
	}
	path := sessionLabelsPath(label.SessionID)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return
	}
	defer f.Close()
	f.Write(append(data, '\n')) //nolint:errcheck
}

func (ds *daemonState) postSessionStart(event intercept.InterceptEvent) {
	if ds.apiKey == "" {
		return
	}
	ds.postEvent(telemetryPayload{
		EventID:            newEventID(),
		EventType:          "session_start",
		CLIVersion:         buildVersion,
		Integration:        string(event.Host),
		SessionID:          event.Session.ID,
		AnalyticsConsented: ds.cfg.Telemetry,
	})
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
	if resp.StatusCode == http.StatusUnauthorized {
		ds.mu.Lock()
		ds.apiKey = ""
		ds.mu.Unlock()
		fmt.Fprintln(os.Stderr, "[confire] API key revoked — run `confire login` to reconnect")
	}
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

	pidPath := daemonPIDPath()
	os.WriteFile(pidPath, []byte(strconv.Itoa(os.Getpid())), 0600)

	cfg := config.Load()

	apiKey := resolveAPIKey()
	deviceID := ""
	if d, err := auth.DeviceID(); err == nil {
		deviceID = d
	}

	receiptPrivKey, err := receipt.LoadOrGenerate(deviceID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[confire] receipt key: %v\n", err)
	}
	var receiptPubB64 string
	if receiptPrivKey != nil {
		receiptPubB64 = base64.RawURLEncoding.EncodeToString(
			receiptPrivKey.Public().(ed25519.PublicKey),
		)
	}
	receiptStore, err := receipt.NewStore(receiptsDirPath(), buildVersion, receiptPubB64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[confire] receipt store: %v\n", err)
		receiptStore = nil
	}

	rules := policy.LoadRules()
	policyEngine := policy.NewEngine(rules)
	mode := policy.Mode(cfg.EffectiveMode())

	state := &daemonState{
		sessions:     make(map[string]*sessionStats),
		apiKey:       apiKey,
		deviceID:     deviceID,
		cfg:          cfg,
		guardrail:    guardrail.New(policyEngine, mode),
		mcpRisk:      intercept.NewMCPRiskHandler(provenance.TrustedMCPServers),
		mcpSanitize:  &intercept.MCPSanitizeHandler{},
		receiptStore: receiptStore,
		receiptKey:   receiptPrivKey,
	}

	go state.emitBundleReceipt(len(rules))

	fmt.Fprintf(os.Stderr, "[confire daemon] v%s listening on %s\n", buildVersion, socketPath)
	if apiKey != "" {
		fmt.Fprintf(os.Stderr, "[confire daemon] firewall active → %s\n", workerURLEnv())
	} else {
		fmt.Fprintln(os.Stderr, "[confire daemon] no account — run `confire login` to enable cloud sync")
	}
	go state.postEvent(telemetryPayload{
		EventID:    newEventID(),
		EventType:  "daemon_started",
		CLIVersion: buildVersion,
	})

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
		go handleConn(conn, state)
	}
}

func handleConn(conn net.Conn, state *daemonState) {
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(30 * time.Second))

	var event intercept.InterceptEvent
	if err := json.NewDecoder(conn).Decode(&event); err != nil {
		return
	}

	// Session lifecycle events.
	switch event.Phase {
	case intercept.PhaseSessionStart:
		state.onSessionStart(event)
		go state.postSessionStart(event)
		state.emitSessionReceipt(event, receipt.PhaseSessionStart)
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
		state.emitSessionReceipt(event, receipt.PhaseSessionEnd)
		json.NewEncoder(conn).Encode(intercept.InterceptResult{Kind: intercept.ResultPassthrough})
		return

	case intercept.PhaseToolPre:
		var result intercept.InterceptResult

		if policy.ConsumeBypassNext() {
			go state.postEvent(telemetryPayload{
				EventID:   newEventID(),
				EventType: "bypass_next_consumed",
				SessionID: event.Session.ID,
			})
			result = intercept.InterceptResult{Kind: intercept.ResultPassthrough}
		} else if state.cfg.IsFirewallEnabled() && state.guardrail != nil {
			// Cross-tool flow detection runs first against recent provenance labels.
			state.mu.Lock()
			recentLabels := append([]provenance.ProvenanceLabel(nil), state.sessions[event.Session.ID].getRecentLabels()...)
			state.mu.Unlock()
			mode := policy.Mode(state.cfg.EffectiveMode())
			if flowResult := firewall.CheckFlowRules(recentLabels, event, mode); flowResult.Kind != intercept.ResultPassthrough {
				state.onFirewallResult(event, flowResult)
				logFirewallResult(event, flowResult)
				go state.postSecurityEvent(event, flowResult)
				result = flowResult
			} else {
				result, _ = state.guardrail.Run(event)
				// MCP risk classifier runs after the policy guardrail.
				if result.Kind == intercept.ResultPassthrough && state.mcpRisk != nil && state.mcpRisk.Matches(event) {
					if r, err := state.mcpRisk.Run(event); err == nil && r.Kind != intercept.ResultPassthrough {
						result = r
					}
				}
				state.onFirewallResult(event, result)
				logFirewallResult(event, result)
				if result.Kind != intercept.ResultPassthrough {
					go state.postSecurityEvent(event, result)
				}
			}
		} else {
			result = intercept.InterceptResult{Kind: intercept.ResultPassthrough}
		}

		json.NewEncoder(conn).Encode(result)
		state.emitToolPreReceipt(event, result)
		return
	}

	// PostToolUse — run security sanitization.
	var sanitizeReport intercept.SanitizeReport
	if state.cfg.IsFirewallEnabled() && event.Phase == intercept.PhaseToolPost {
		if event.Tool == nil || !event.Tool.IsMCP {
			event, sanitizeReport = state.sanitizeOutput(event)
		} else {
			event, sanitizeReport = state.runMCPSanitizePipeline(event)
		}
	}

	label := state.onToolResult(event, sanitizeReport)
	go writeProvenanceLabel(label)
	go state.postProvenanceEvent(label)
	state.emitToolPostReceipt(event, label)

	result := intercept.InterceptResult{Kind: intercept.ResultPassthrough}

	// If we sanitized content, return the cleaned output.
	if sanitizeReport.HasFindings() && event.Tool != nil && event.Tool.Output != nil {
		result = intercept.InterceptResult{
			Kind:       intercept.ResultSanitize,
			ToolOutput: event.Tool.Output,
		}
		go state.postSecurityEvent(event, result)
	}

	// Inject advisory context for PostToolSteer hosts.
	toolName := ""
	if event.Tool != nil {
		toolName = event.Tool.Name
	}
	caps := state.cfg.CapabilitiesFor(event.Host)
	if caps.PostToolSteer {
		steerCtx := intercept.FormatPostToolSteer(intercept.PostToolSteerInput{
			Host:                event.Host,
			ToolName:            toolName,
			NativeUnreplaceable: !caps.NativeOutputReplaceable && event.Tool != nil && !event.Tool.IsMCP,
			OutputReplaced:      result.Kind == intercept.ResultSanitize,
			Report:              sanitizeReport,
			ExtraContext:        "",
		})
		if steerCtx != "" {
			result.Context = intercept.JoinContext(steerCtx, result.Context)
			if result.Kind == intercept.ResultPassthrough {
				result.Kind = intercept.ResultAddContext
			}
		}
	}

	json.NewEncoder(conn).Encode(result)
}

// sanitizeOutput runs secret redaction and prompt-injection sanitization on non-MCP tool output.
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

	redacted, count, types := sanitize.Redact(text)
	if count > 0 {
		text = redacted
		modified = true
		report.SecretsRedacted = count
		report.SecretTypes = types
		fmt.Fprintf(os.Stderr, "[confire] ⚡ redacted %d secret(s) from %s output\n", count, event.Tool.Name)
	}

	sanitized, found, _ := sanitize.Sanitize(text)
	if found {
		text = sanitized
		modified = true
		report.InjectionFound = true
		fmt.Fprintf(os.Stderr, "[confire] ⚡ %s sanitized / prompt injection removed\n", event.Tool.Name)
	}

	if modified {
		toolCopy := *event.Tool
		toolCopy.Output = text
		event.Tool = &toolCopy
	}
	return event, report
}

// runMCPSanitizePipeline applies the MCP-specific PostToolUse security pipeline.
func (ds *daemonState) runMCPSanitizePipeline(event intercept.InterceptEvent) (intercept.InterceptEvent, intercept.SanitizeReport) {
	report := intercept.SanitizeReport{}
	if event.Tool == nil || event.Tool.Output == nil {
		return event, report
	}

	if ds.mcpSanitize != nil && ds.mcpSanitize.Matches(event) {
		if r, err := ds.mcpSanitize.Run(event); err == nil && r.Kind != intercept.ResultPassthrough {
			toolCopy := *event.Tool
			toolCopy.Output = r.ToolOutput
			event.Tool = &toolCopy
			if r.Stats != nil && r.Stats.BeforeBytes > r.Stats.AfterBytes {
				report.SecretsRedacted = 1
			}
			report.InjectionFound = r.Kind == intercept.ResultSanitize
			if report.HasFindings() {
				fmt.Fprintf(os.Stderr, "[confire] ⚡ %s sanitized / source: external\n", event.Tool.Name)
			}
		}
	}

	return event, report
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

// logFirewallResult prints the new-format firewall decision line to stderr.
func logFirewallResult(event intercept.InterceptEvent, result intercept.InterceptResult) {
	toolName := ""
	if event.Tool != nil {
		toolName = event.Tool.Name
	}
	switch result.Kind {
	case intercept.ResultBlock:
		fmt.Fprintf(os.Stderr, "%s[confire] BLOCKED%s %s\n", colorRed, colorReset, toolName)
	case intercept.ResultReview:
		fmt.Fprintf(os.Stderr, "%s[confire] REVIEW REQUIRED%s %s — run: confire bypass-next\n", colorOrange, colorReset, toolName)
	case intercept.ResultWarn:
		fmt.Fprintf(os.Stderr, "%s[confire] warning%s %s\n", colorYellow, colorReset, toolName)
	case intercept.ResultSanitize:
		fmt.Fprintf(os.Stderr, "%s[confire] sanitized%s %s\n", colorGreen, colorReset, toolName)
	case intercept.ResultPassthrough:
		// Log allowed only for MCP tools to avoid noise on every Bash/Read call.
		if event.Tool != nil && event.Tool.IsMCP {
			fmt.Fprintf(os.Stderr, "%s[confire] ✓%s %s — allowed\n", colorGreen, colorReset, toolName)
		}
	}
}

// ── Session notifications ─────────────────────────────────────────────────────

func sessionStartNotification(state *daemonState) (contextMsg, systemMsg string) {
	mode := state.cfg.EffectiveMode()

	switch {
	case state.apiKey == "":
		// Local firewall is active without login — do not inject a warning into the agent context.
		if line := sessionStatusLine(state, mode); line != "" {
			fmt.Fprintf(os.Stderr, "%s\n", line)
		}
		return "", ""
	case mode == "strict":
		msg := "[Confire] Strict mode active. Dangerous tool calls will be blocked. Run `confire help` for commands."
		return msg, msg
	case mode == "bypass":
		msg := "[Confire] Bypass mode active. Firewall is disabled. Run `confire help` for commands."
		return msg, msg
	default:
		if msg, ok := consumeWelcomePending(state); ok {
			return msg, msg
		}
		if line := sessionStatusLine(state, mode); line != "" {
			fmt.Fprintf(os.Stderr, "%s\n", line)
		}
		return "", ""
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

// ── Context ID pairing ────────────────────────────────────────────────────────

// storeContextID saves a daemon-generated context_id for the given session + tool name.
// Used when the host doesn't provide tool_use_id so tool.pre and tool.post can be linked.
func (ds *daemonState) storeContextID(sessionID, toolName, contextID string) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	sess := ds.sessions[sessionID]
	if sess == nil {
		return
	}
	if sess.pendingContextIDs == nil {
		sess.pendingContextIDs = make(map[string]string)
	}
	sess.pendingContextIDs[toolName] = contextID
}

// consumeContextID retrieves and removes the pending context_id for the given session + tool name.
func (ds *daemonState) consumeContextID(sessionID, toolName string) string {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	sess := ds.sessions[sessionID]
	if sess == nil || sess.pendingContextIDs == nil {
		return ""
	}
	id := sess.pendingContextIDs[toolName]
	delete(sess.pendingContextIDs, toolName)
	return id
}

// ── Receipt emission ──────────────────────────────────────────────────────────

// emitReceipt signs and writes r. Best-effort — errors are logged but never fatal.
func (ds *daemonState) emitReceipt(r receipt.Receipt) {
	if ds.receiptStore == nil || ds.receiptKey == nil {
		return
	}
	if err := receipt.Sign(ds.receiptKey, ds.deviceID, &r); err != nil {
		fmt.Fprintf(os.Stderr, "[confire] receipt sign: %v\n", err)
		return
	}
	if err := ds.receiptStore.Write(r); err != nil {
		fmt.Fprintf(os.Stderr, "[confire] receipt write: %v\n", err)
	}
}

func (ds *daemonState) newReceiptBody(sessionID string, phase receipt.Phase) receipt.Body {
	return receipt.Body{
		Version:             receipt.Version,
		ConfireVersion:      buildVersion,
		ReceiptID:           receipt.NewID(),
		PreviousPayloadHash: ds.receiptStore.ChainTip(sessionID),
		Phase:               phase,
		Timestamp:           time.Now().UTC(),
	}
}

func (ds *daemonState) emitSessionReceipt(event intercept.InterceptEvent, phase receipt.Phase) {
	if ds.receiptStore == nil {
		return
	}
	body := ds.newReceiptBody(event.Session.ID, phase)
	body.Event = &receipt.Event{
		SessionID: event.Session.ID,
		Client:    event.Host,
	}
	ds.emitReceipt(receipt.Receipt{Body: body})
}

func (ds *daemonState) emitToolPreReceipt(event intercept.InterceptEvent, result intercept.InterceptResult) {
	if ds.receiptStore == nil || event.Tool == nil {
		return
	}
	sessionID := event.Session.ID
	contextID := event.Tool.UseID
	if contextID == "" {
		contextID = receipt.NewID()
		ds.storeContextID(sessionID, event.Tool.Name, contextID)
	}
	inputHash, _ := receipt.HashAny(event.Tool.Input)
	body := ds.newReceiptBody(sessionID, receipt.PhaseToolPre)
	body.Event = &receipt.Event{
		SessionID: sessionID,
		Client:    event.Host,
		ContextID: contextID,
		ToolName:  event.Tool.Name,
		IsMCP:     event.Tool.IsMCP,
		MCPServer: event.Tool.MCPServer,
		InputHash: inputHash,
		CWDHash:   receipt.HashString(event.Session.CWD),
	}
	body.Decision = &receipt.Decision{
		Action:     string(result.Kind),
		RuleID:     result.RuleID,
		RuleName:   result.RuleName,
		RuleSource: result.RuleSource,
		PolicyMode: ds.cfg.EffectiveMode(),
	}
	ds.emitReceipt(receipt.Receipt{Body: body})
}

func (ds *daemonState) emitToolPostReceipt(event intercept.InterceptEvent, label provenance.ProvenanceLabel) {
	if ds.receiptStore == nil || event.Tool == nil {
		return
	}
	sessionID := event.Session.ID
	contextID := event.Tool.UseID
	if contextID == "" {
		contextID = ds.consumeContextID(sessionID, event.Tool.Name)
	}
	inputHash, _ := receipt.HashAny(event.Tool.Input)

	var outputHash string
	if event.Tool.Output != nil {
		if s, ok := sanitize.OutputToString(event.Tool.Output); ok {
			outputHash = receipt.HashString(s)
		}
	}

	body := ds.newReceiptBody(sessionID, receipt.PhaseToolPost)
	body.Event = &receipt.Event{
		SessionID:  sessionID,
		Client:     event.Host,
		ContextID:  contextID,
		ToolName:   event.Tool.Name,
		IsMCP:      event.Tool.IsMCP,
		MCPServer:  event.Tool.MCPServer,
		DurationMs: event.Tool.DurationMs,
		InputHash:  inputHash,
		OutputHash: outputHash,
		CWDHash:    receipt.HashString(event.Session.CWD),
	}
	body.Provenance = &receipt.Provenance{
		TrustLevel:        string(label.TrustLevel),
		Flags:             label.Flags,
		RedactionsCount:   label.RedactionsCount,
		SanitizationCount: label.SanitizationCount,
		RiskScore:         label.RiskScore,
		OriginDomain:      label.OriginDomain,
	}
	ds.emitReceipt(receipt.Receipt{Body: body})
}

func (ds *daemonState) emitBundleReceipt(ruleCount int) {
	if ds.receiptStore == nil {
		return
	}
	bundleVersion := "builtin"
	fetchedAt := time.Now().UTC()
	groupOverrideCount := 0
	verResult := receipt.VerificationNoSignature

	if cp := policy.LoadCache(); cp != nil {
		if cp.Version != "" {
			bundleVersion = cp.Version
		}
		fetchedAt = cp.FetchedAt
		groupOverrideCount = len(cp.GroupOverrides)
		verResult = receipt.VerificationVerified
	}

	body := receipt.Body{
		Version:             receipt.Version,
		ConfireVersion:      buildVersion,
		ReceiptID:           receipt.NewID(),
		PreviousPayloadHash: ds.receiptStore.ChainTip(receipt.BundleSessionKey),
		Phase:               receipt.PhasePolicyBundleLoaded,
		Timestamp:           time.Now().UTC(),
		Event: &receipt.Event{
			SessionID: receipt.BundleSessionKey,
			Client:    "daemon",
		},
		Bundle: &receipt.Bundle{
			BundleVersion:      bundleVersion,
			FetchedAt:          fetchedAt,
			RuleCount:          ruleCount,
			GroupOverrideCount: groupOverrideCount,
			VerificationResult: verResult,
		},
	}
	ds.emitReceipt(receipt.Receipt{Body: body})
}

// ── Helpers ───────────────────────────────────────────────────────────────────

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

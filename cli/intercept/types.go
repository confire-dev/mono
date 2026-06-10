// Package intercept defines the canonical, host-agnostic envelope contract.
// All optimizers and handlers speak these types — never native hook payloads.
// Mirrors worker/src/types.ts exactly (same JSON field names).
package intercept

// Phase is the canonical event phase, host-agnostic.
// Adapters map native events (PostToolUse, tool-executed…) onto these names.
type Phase string

const (
	PhaseSessionStart  Phase = "session.start"
	PhaseSessionEnd    Phase = "session.end"
	PhaseToolPre       Phase = "tool.pre"
	PhaseToolPost      Phase = "tool.post"
	PhaseToolBatchPost Phase = "tool.batch.post"
	PhaseTurnStop      Phase = "turn.stop"
	PhasePromptSubmit  Phase = "prompt.submit"
	// PhasePreCompact is reserved but DISABLED in v1.
	// Highest blast radius: replaces transcript state. Enable in week 2 after stability.
	PhasePreCompact Phase = "context.pre-compact"

	// PhaseSchemaDrift is a synthetic internal phase used to report unknown fields
	// detected in a client's hook payload. Never sent from client adapters directly —
	// only emitted by the hook dispatcher when unknown JSON keys are found.
	// The daemon routes this to telemetry for Slack alerting.
	PhaseSchemaDrift Phase = "internal.schema-drift"
)

type Host     = string
type Strategy = string

// InterceptEvent is the canonical event envelope sent from a host adapter to the engine.
// It is also the HTTP/2 wire format sent from the Go daemon to the TS Worker.
type InterceptEvent struct {
	Host     Host            `json:"host"`
	Strategy Strategy        `json:"strategy"`
	Phase    Phase           `json:"phase"`
	Session  Session         `json:"session"`
	Tool     *Tool           `json:"tool,omitempty"`
	Prompt   string          `json:"prompt,omitempty"`
	Raw      any             `json:"raw,omitempty"`
}

type Session struct {
	ID             string `json:"id"`
	CWD            string `json:"cwd,omitempty"`
	TranscriptPath string `json:"transcriptPath,omitempty"`
}

type Tool struct {
	Name       string `json:"name"`
	Input      any    `json:"input,omitempty"`
	Output     any    `json:"output,omitempty"`
	UseID      string `json:"useId,omitempty"`
	IsMCP      bool   `json:"isMcp"`
	MCPServer  string `json:"mcpServer,omitempty"`
	DurationMs int    `json:"durationMs,omitempty"`
}

// ResultKind describes what the engine wants the adapter to do.
type ResultKind string

const (
	ResultPassthrough   ResultKind = "passthrough"
	ResultReplaceOutput ResultKind = "replace-output"
	ResultReplaceInput  ResultKind = "replace-input"
	ResultAddContext    ResultKind = "add-context"
	ResultCompact       ResultKind = "compact"

	// Firewall result kinds (PreToolUse).
	ResultBlock  ResultKind = "block"  // prevent tool execution, explain why
	ResultReview ResultKind = "review" // block + invite user to allow/retry
	ResultWarn   ResultKind = "warn"   // allow but inject advisory context

	// Firewall result kinds (PostToolUse).
	ResultSanitize ResultKind = "sanitize" // replace output with sanitized version
	ResultRedact   ResultKind = "redact"   // replace output with redacted version
)

type InterceptResult struct {
	Kind       ResultKind `json:"kind"`
	ToolOutput any        `json:"toolOutput,omitempty"`
	ToolInput  any        `json:"toolInput,omitempty"`
	Context    string     `json:"context,omitempty"`
	SystemMessage string  `json:"systemMessage,omitempty"` // shown to the user in Claude Code (hooks have no TTY)
	Reason     string     `json:"reason,omitempty"` // block/review human-readable explanation
	Stats      *Stats     `json:"stats,omitempty"`
}

type Stats struct {
	BeforeBytes int    `json:"beforeBytes"`
	AfterBytes  int    `json:"afterBytes"`
	Optimizer   string `json:"optimizer"`
}

// Package policy defines the Confire rule schema and evaluation engine.
// Bundled rules live in policy/rules/*.json (embedded at build time).
// Custom rules are cloud-managed, fetched by cache.go, and evaluated locally.
package policy

// Mode controls how strictly the firewall enforces rules.
type Mode string

const (
	ModeObserve  Mode = "observe"  // record only, never block
	ModeBalanced Mode = "balanced" // default: review dangerous, optimize noise
	ModeStrict   Mode = "strict"   // block dangerous, enforce budgets
	ModeBypass   Mode = "bypass"   // installed but fully passive
)

// RuleAction describes what happens when a rule matches.
type RuleAction string

const (
	ActionAllow       RuleAction = "allow"
	ActionWarn        RuleAction = "warn"
	ActionReview      RuleAction = "review"
	ActionBlock       RuleAction = "block"
	ActionOptimize    RuleAction = "optimize"
	ActionSanitize    RuleAction = "sanitize"
	ActionRedact      RuleAction = "redact"
	ActionPassThrough RuleAction = "pass_through"
)

// RulePhase restricts which hook phase a rule applies to.
type RulePhase string

const (
	PhasePreToolUse  RulePhase = "PreToolUse"
	PhasePostToolUse RulePhase = "PostToolUse"
	PhaseAny         RulePhase = ""
)

// Severity is informational — used in review/block messages.
type Severity string

const (
	SeverityLow    Severity = "low"
	SeverityMedium Severity = "medium"
	SeverityHigh   Severity = "high"
)

// Rule is one firewall policy rule.
type Rule struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Enabled  bool      `json:"enabled"`
	Phase    RulePhase `json:"phase"`
	Action   RuleAction `json:"action"`
	Severity Severity  `json:"severity"`
	Match    RuleMatch `json:"match"`
	Message  string    `json:"message"`

	// MinMode is the minimum mode required to activate this rule.
	// e.g. MinMode=strict means the rule only fires in strict mode.
	// Empty means fires in balanced and strict.
	MinMode Mode `json:"min_mode,omitempty"`

	// Source identifies where this rule came from: "builtin" | "custom".
	Source string `json:"source,omitempty"`
}

// RuleMatch describes what conditions trigger a rule.
type RuleMatch struct {
	// ToolName matches the exact tool name (case-insensitive).
	ToolName string `json:"tool_name,omitempty"`

	// ToolNames matches any of these exact tool names.
	ToolNames []string `json:"tool_names,omitempty"`

	// ToolPrefix matches tools whose name starts with this prefix.
	ToolPrefix string `json:"tool_prefix,omitempty"`

	// ToolNameRegex is applied to the tool name (case-insensitive).
	ToolNameRegex string `json:"tool_name_regex,omitempty"`

	// CommandContains checks if any of these strings appear in the Bash command.
	CommandContains []string `json:"command_contains,omitempty"`

	// CommandRegex is a regex applied to the Bash command input.
	CommandRegex string `json:"command_regex,omitempty"`

	// MCPServer restricts to a specific MCP server name.
	MCPServer string `json:"mcp_server,omitempty"`

	// OutputMinBytes only matches if raw output exceeds this byte count.
	OutputMinBytes int `json:"output_min_bytes,omitempty"`

	// OutputSecretScan runs secret detection on tool output.
	OutputSecretScan bool `json:"output_secret_scan,omitempty"`

	// OutputInjectionScan runs prompt-injection detection on tool output.
	OutputInjectionScan bool `json:"output_injection_scan,omitempty"`
}

// MatchResult is returned when a rule fires.
type MatchResult struct {
	Rule    Rule
	Action  RuleAction
	Message string
}

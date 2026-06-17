// Package receipt implements signed, tamper-evident provenance receipts for
// every Confire firewall event. Each receipt is an Ed25519-signed record that
// chains to its predecessor via previous_payload_hash, producing an auditable,
// offline-verifiable log of all tool decisions and context-firewall actions.
package receipt

import "time"

// Version is the current receipt schema version embedded in every receipt body.
const Version = "1"

// GenesisHash is the previous_payload_hash for the first receipt in any chain.
// 64 hex zeros — unambiguous sentinel that requires no prior receipt.
const GenesisHash = "0000000000000000000000000000000000000000000000000000000000000000"

// BundleSessionKey is the internal chain key used for policy.bundle.loaded receipts.
// These form their own independent chain in daemon-policy.jsonl.
const BundleSessionKey = "_daemon_policy_"

// Phase identifies what kind of event a receipt covers.
type Phase string

const (
	PhaseSessionStart       Phase = "session.start"
	PhaseSessionEnd         Phase = "session.end"
	PhaseToolPre            Phase = "tool.pre"
	PhaseToolPost           Phase = "tool.post"
	PhasePolicyBundleLoaded Phase = "policy.bundle.loaded"
)

// VerificationResult describes the outcome of policy bundle signature verification.
type VerificationResult string

const (
	VerificationVerified          VerificationResult = "verified"
	VerificationNoSignature       VerificationResult = "no_signature"
	VerificationInvalidSignature  VerificationResult = "invalid_signature"
	VerificationFallbackToBuiltin VerificationResult = "fallback_to_builtin"
)

// Body is the signable portion of a Receipt.
// It contains every field except payload_hash and signers[].
// Ed25519 signatures are computed over Canonical(Body) and
// payload_hash = hex(SHA-256(Canonical(Body))).
type Body struct {
	Version             string      `json:"version"`
	ConfireVersion      string      `json:"confire_version"`
	ReceiptID           string      `json:"receipt_id"`
	PreviousPayloadHash string      `json:"previous_payload_hash"`
	Phase               Phase       `json:"phase"`
	ParentSessionID     string      `json:"parent_session_id,omitempty"`
	Event               *Event      `json:"event,omitempty"`
	Decision            *Decision   `json:"decision,omitempty"`
	Provenance          *Provenance `json:"provenance,omitempty"`
	Bundle              *Bundle     `json:"bundle,omitempty"`
	Timestamp           time.Time   `json:"timestamp"`
}

// Receipt is the full on-disk record: Body embedded, plus payload_hash and signers.
// JSON embedding promotes Body fields to the top level of the receipt object.
type Receipt struct {
	Body
	PayloadHash string   `json:"payload_hash"`
	Signers     []Signer `json:"signers"`
}

// Event holds event-scoped fields present on most receipt phases.
// Fields not applicable to a phase are omitted (omitempty).
type Event struct {
	SessionID string `json:"session_id"`
	Client    string `json:"client"`

	// ContextID links a tool.pre receipt with its corresponding tool.post receipt
	// for the same tool invocation. Set from Tool.UseID when the host provides it
	// (claude_code, cursor, vscode, codex); for hosts that don't send tool_use_id
	// (windsurf, cline, openclaw) the daemon generates a UUID at tool.pre time and
	// stores it in session state to be consumed by the matching tool.post.
	// Absent on session.start, session.end, and policy.bundle.loaded receipts.
	ContextID string `json:"context_id,omitempty"`

	ToolName   string `json:"tool_name,omitempty"`
	IsMCP      bool   `json:"is_mcp,omitempty"`
	MCPServer  string `json:"mcp_server,omitempty"`
	DurationMs int    `json:"duration_ms,omitempty"`

	// InputHash is the SHA-256 hex of the canonical JSON of tool.Input.
	// Present on tool.pre and tool.post.
	InputHash string `json:"input_hash,omitempty"`

	// OutputHash is the SHA-256 hex of the canonical JSON of the SANITIZED
	// tool output — after secret redaction and injection removal. This is what
	// the model context actually received. Raw (pre-sanitization) output is
	// never hashed into a receipt. Present on tool.post only.
	OutputHash string `json:"output_hash,omitempty"`

	// CWDHash is the SHA-256 hex of the session CWD path string.
	// The path itself is not stored (privacy).
	CWDHash string `json:"cwd_hash,omitempty"`
}

// Decision holds the firewall outcome. Present on tool.pre receipts only.
type Decision struct {
	Action     string `json:"action"`
	RuleID     string `json:"rule_id,omitempty"`
	RuleName   string `json:"rule_name,omitempty"`
	RuleSource string `json:"rule_source,omitempty"` // "builtin" | "custom" | "flow" | "mcp_risk"
	PolicyMode string `json:"policy_mode,omitempty"`
	BypassNext bool   `json:"bypass_next,omitempty"` // true when bypass-next flag was consumed
}

// Provenance holds the context-firewall classification. Present on tool.post receipts only.
type Provenance struct {
	TrustLevel        string   `json:"trust_level"`
	Flags             []string `json:"flags,omitempty"`
	RedactionsCount   int      `json:"redactions_count,omitempty"`
	SanitizationCount int      `json:"sanitization_count,omitempty"`
	RiskScore         int      `json:"risk_score,omitempty"`
	OriginDomain      string   `json:"origin_domain,omitempty"`
}

// Bundle holds policy bundle metadata. Present on policy.bundle.loaded receipts only.
type Bundle struct {
	BundleVersion      string             `json:"bundle_version"`
	FetchedAt          time.Time          `json:"fetched_at"`
	RuleCount          int                `json:"rule_count"`
	GroupOverrideCount int                `json:"group_override_count,omitempty"`
	VerificationResult VerificationResult `json:"verification_result"`
	SignerID           string             `json:"signer_id,omitempty"`
}

// Signer holds one Ed25519 signature entry.
//
// Canonical order: local signer at index 0, remote co-signer (if present) at index 1.
// Verifiers must reject any receipt where a remote signer appears at index 0.
type Signer struct {
	SignerID  string `json:"signer_id"`  // "local:<device_id>" or "remote:confire-cloud"
	Pubkey    string `json:"pubkey"`     // base64url-encoded 32-byte Ed25519 public key
	Algorithm string `json:"algorithm"`  // always "Ed25519"
	Signature string `json:"signature"`  // base64url-encoded 64-byte Ed25519 signature over Canonical(Body)
}

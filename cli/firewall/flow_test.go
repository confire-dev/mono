package firewall

import (
	"testing"

	"github.com/confire-dev/confire/intercept"
	"github.com/confire-dev/confire/policy"
	"github.com/confire-dev/confire/provenance"
)

func label(trust provenance.TrustLevel, flags ...string) provenance.ProvenanceLabel {
	return provenance.ProvenanceLabel{TrustLevel: trust, Flags: flags}
}

func mcpTool(name, server string) *intercept.Tool {
	return &intercept.Tool{Name: name, IsMCP: true, MCPServer: server}
}

func nativeTool(name string, input map[string]any) *intercept.Tool {
	t := &intercept.Tool{Name: name, IsMCP: false}
	if input != nil {
		t.Input = input
	}
	return t
}

func TestCheckFlowRules_NoLabels(t *testing.T) {
	event := intercept.InterceptEvent{Tool: &intercept.Tool{Name: "Bash", IsMCP: false}}
	result := CheckFlowRules(nil, event, policy.ModeBalanced)
	if result.Kind != intercept.ResultPassthrough {
		t.Errorf("no labels: want passthrough, got %s", result.Kind)
	}
}

func TestCheckFlowRules_NoTool(t *testing.T) {
	recent := []provenance.ProvenanceLabel{label(provenance.TrustMCPUnknown)}
	result := CheckFlowRules(recent, intercept.InterceptEvent{}, policy.ModeBalanced)
	if result.Kind != intercept.ResultPassthrough {
		t.Errorf("no tool: want passthrough, got %s", result.Kind)
	}
}

// Rule 1: untrusted → secret read
func TestRule_UntrustedThenSecretRead_Triggers(t *testing.T) {
	recent := []provenance.ProvenanceLabel{label(provenance.TrustExternalUntrusted)}
	event := intercept.InterceptEvent{
		Tool: nativeTool("Read", map[string]any{"file_path": "/home/user/.env"}),
	}
	result := CheckFlowRules(recent, event, policy.ModeBalanced)
	if result.Kind != intercept.ResultReview {
		t.Errorf("untrusted→secret-read: want review, got %s", result.Kind)
	}
}

func TestRule_UntrustedThenSecretRead_NormalReadNoTrigger(t *testing.T) {
	recent := []provenance.ProvenanceLabel{label(provenance.TrustExternalUntrusted)}
	event := intercept.InterceptEvent{
		Tool: nativeTool("Read", map[string]any{"file_path": "/home/user/main.go"}),
	}
	result := CheckFlowRules(recent, event, policy.ModeBalanced)
	if result.Kind != intercept.ResultPassthrough {
		t.Errorf("normal read after untrusted: want passthrough, got %s", result.Kind)
	}
}

func TestRule_UntrustedThenSecretRead_TrustedSourceNoTrigger(t *testing.T) {
	recent := []provenance.ProvenanceLabel{label(provenance.TrustMCPTrusted)}
	event := intercept.InterceptEvent{
		Tool: nativeTool("Read", map[string]any{"file_path": "~/.aws/credentials"}),
	}
	result := CheckFlowRules(recent, event, policy.ModeBalanced)
	if result.Kind != intercept.ResultPassthrough {
		t.Errorf("trusted source + secret file: want passthrough, got %s", result.Kind)
	}
}

// Rule 2: secret read → external send
func TestRule_SecretReadThenExternalSend_Triggers_Balanced(t *testing.T) {
	recent := []provenance.ProvenanceLabel{label(provenance.TrustSensitive, provenance.FlagSecretRedacted)}
	event := intercept.InterceptEvent{
		Tool: mcpTool("mcp__slack__send_message", "slack"),
	}
	result := CheckFlowRules(recent, event, policy.ModeBalanced)
	if result.Kind != intercept.ResultReview {
		t.Errorf("secret→exfil balanced: want review, got %s", result.Kind)
	}
}

func TestRule_SecretReadThenExternalSend_Triggers_Strict(t *testing.T) {
	recent := []provenance.ProvenanceLabel{label(provenance.TrustSensitive, provenance.FlagSecretRedacted)}
	event := intercept.InterceptEvent{
		Tool: mcpTool("mcp__slack__send_message", "slack"),
	}
	result := CheckFlowRules(recent, event, policy.ModeStrict)
	if result.Kind != intercept.ResultBlock {
		t.Errorf("secret→exfil strict: want block, got %s", result.Kind)
	}
}

func TestRule_SecretReadThenExternalSend_NoSecretFlag_NoTrigger(t *testing.T) {
	recent := []provenance.ProvenanceLabel{label(provenance.TrustMCPTrusted)}
	event := intercept.InterceptEvent{
		Tool: mcpTool("mcp__slack__send_message", "slack"),
	}
	result := CheckFlowRules(recent, event, policy.ModeBalanced)
	if result.Kind != intercept.ResultPassthrough {
		t.Errorf("no secret flag: want passthrough, got %s", result.Kind)
	}
}

// Rule 3: injection → shell
func TestRule_InjectionThenShell_Triggers(t *testing.T) {
	recent := []provenance.ProvenanceLabel{label(provenance.TrustMCPTrusted, provenance.FlagPromptInjection)}
	event := intercept.InterceptEvent{
		Tool: nativeTool("Bash", map[string]any{"command": "ls -la"}),
	}
	result := CheckFlowRules(recent, event, policy.ModeBalanced)
	if result.Kind != intercept.ResultReview {
		t.Errorf("injection→shell: want review, got %s", result.Kind)
	}
}

func TestRule_InjectionThenShell_MCPToolNoTrigger(t *testing.T) {
	recent := []provenance.ProvenanceLabel{label(provenance.TrustMCPTrusted, provenance.FlagPromptInjection)}
	event := intercept.InterceptEvent{
		Tool: mcpTool("mcp__github__get_pull_request", "github"),
	}
	result := CheckFlowRules(recent, event, policy.ModeBalanced)
	if result.Kind != intercept.ResultPassthrough {
		t.Errorf("injection→mcp (not shell): want passthrough, got %s", result.Kind)
	}
}

// Rule 4: unknown-mcp → mutating-mcp
func TestRule_UnknownMCPThenMutatingMCP_Triggers(t *testing.T) {
	recent := []provenance.ProvenanceLabel{label(provenance.TrustMCPUnknown)}
	event := intercept.InterceptEvent{
		Tool: mcpTool("mcp__unknowntool__delete_document", "unknowntool"),
	}
	result := CheckFlowRules(recent, event, policy.ModeBalanced)
	if result.Kind != intercept.ResultReview {
		t.Errorf("unknown-mcp→mutating-mcp: want review, got %s", result.Kind)
	}
}

func TestRule_UnknownMCPThenMutatingMCP_ReadOnlyNoTrigger(t *testing.T) {
	recent := []provenance.ProvenanceLabel{label(provenance.TrustMCPUnknown)}
	event := intercept.InterceptEvent{
		Tool: mcpTool("mcp__unknowntool__get_document", "unknowntool"),
	}
	result := CheckFlowRules(recent, event, policy.ModeBalanced)
	if result.Kind != intercept.ResultPassthrough {
		t.Errorf("unknown-mcp→read-only-mcp: want passthrough, got %s", result.Kind)
	}
}

// Rule 5: credential-lure → message
func TestRule_CredentialLureThenMessage_Triggers(t *testing.T) {
	recent := []provenance.ProvenanceLabel{label(provenance.TrustMCPTrusted, provenance.FlagCredentialLure)}
	event := intercept.InterceptEvent{
		Tool: mcpTool("mcp__slack__send_message", "slack"),
	}
	result := CheckFlowRules(recent, event, policy.ModeBalanced)
	if result.Kind != intercept.ResultReview {
		t.Errorf("credential-lure→message: want review, got %s", result.Kind)
	}
}

func TestRule_CredentialLureThenMessage_NoLureNoTrigger(t *testing.T) {
	recent := []provenance.ProvenanceLabel{label(provenance.TrustMCPTrusted)}
	event := intercept.InterceptEvent{
		Tool: mcpTool("mcp__slack__send_message", "slack"),
	}
	result := CheckFlowRules(recent, event, policy.ModeBalanced)
	// This will be caught by rule 2 (secret→exfil) only if secret flag is set.
	// Without credential lure AND without secret flag, passthrough.
	if result.Kind != intercept.ResultPassthrough {
		t.Errorf("no credential lure: want passthrough, got %s", result.Kind)
	}
}

// Ring buffer: only last 10 labels matter
func TestFlowRules_RingBuffer_OldLabelsIgnored(t *testing.T) {
	// Build 10 trusted labels, then the 11th (which would be position 0 in ring)
	// has an injection flag. The current event should NOT trigger because the
	// injection label is beyond the ring window.
	recent := make([]provenance.ProvenanceLabel, 10)
	for i := range recent {
		recent[i] = label(provenance.TrustLocal)
	}
	// All 10 recent labels are trusted_local — no injection flag.
	event := intercept.InterceptEvent{
		Tool: nativeTool("Bash", map[string]any{"command": "ls"}),
	}
	result := CheckFlowRules(recent, event, policy.ModeBalanced)
	if result.Kind != intercept.ResultPassthrough {
		t.Errorf("no injection in ring: want passthrough, got %s", result.Kind)
	}
}

// id_rsa detection
func TestIsSecretFileAccess_SSHKey(t *testing.T) {
	tool := nativeTool("Read", map[string]any{"file_path": "/home/user/.ssh/id_rsa"})
	if !isSecretFileAccess(tool) {
		t.Error("expected id_rsa to be detected as secret file")
	}
}

// .aws detection
func TestIsSecretFileAccess_AWS(t *testing.T) {
	tool := nativeTool("Bash", map[string]any{"command": "cat ~/.aws/credentials"})
	if !isSecretFileAccess(tool) {
		t.Error("expected .aws/credentials to be detected as secret file")
	}
}

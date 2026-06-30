package guardrail

import (
	"testing"

	"github.com/confire-dev/confire/intercept"
	"github.com/confire-dev/confire/policy"
)

func TestGuardrail_PropagatesIrreversible_Warn(t *testing.T) {
	rule := policy.Rule{
		ID:           "test-irrev-warn",
		Name:         "Test Irreversible Warn",
		Enabled:      true,
		Phase:        policy.PhasePreToolUse,
		Action:       policy.ActionWarn,
		Match:        policy.RuleMatch{ToolName: "Bash"},
		Message:      "test warn",
		Irreversible: true,
	}
	h := New(policy.NewEngine([]policy.Rule{rule}), policy.ModeBalanced)
	result, err := h.Run(bashEvent("echo hi"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Kind != intercept.ResultWarn {
		t.Fatalf("expected ResultWarn, got %v", result.Kind)
	}
	if !result.Irreversible {
		t.Fatal("Irreversible must be true when rule carries the flag")
	}
}

func TestGuardrail_PropagatesIrreversible_Review(t *testing.T) {
	rule := policy.Rule{
		ID:           "test-irrev-review",
		Name:         "Test Irreversible Review",
		Enabled:      true,
		Phase:        policy.PhasePreToolUse,
		Action:       policy.ActionReview,
		Match:        policy.RuleMatch{ToolName: "Bash"},
		Message:      "test review",
		Irreversible: true,
	}
	h := New(policy.NewEngine([]policy.Rule{rule}), policy.ModeBalanced)
	result, err := h.Run(bashEvent("rm -rf /"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Kind != intercept.ResultReview {
		t.Fatalf("expected ResultReview, got %v", result.Kind)
	}
	if !result.Irreversible {
		t.Fatal("Irreversible must be true when rule carries the flag")
	}
}

func TestGuardrail_PropagatesIrreversible_Block(t *testing.T) {
	rule := policy.Rule{
		ID:           "test-irrev-block",
		Name:         "Test Irreversible Block",
		Enabled:      true,
		Phase:        policy.PhasePreToolUse,
		Action:       policy.ActionBlock,
		Match:        policy.RuleMatch{ToolName: "Bash"},
		Message:      "test block",
		Irreversible: true,
	}
	h := New(policy.NewEngine([]policy.Rule{rule}), policy.ModeBalanced)
	result, err := h.Run(bashEvent("destroy everything"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Kind != intercept.ResultBlock {
		t.Fatalf("expected ResultBlock, got %v", result.Kind)
	}
	if !result.Irreversible {
		t.Fatal("Irreversible must be true when rule carries the flag")
	}
}

func TestGuardrail_Irreversible_FalseByDefault(t *testing.T) {
	// A rule without Irreversible set must produce Irreversible=false in the result.
	rule := policy.Rule{
		ID:      "test-non-irrev",
		Name:    "Test Non-Irreversible",
		Enabled: true,
		Phase:   policy.PhasePreToolUse,
		Action:  policy.ActionWarn,
		Match:   policy.RuleMatch{ToolName: "Bash"},
		Message: "normal warn",
	}
	h := New(policy.NewEngine([]policy.Rule{rule}), policy.ModeBalanced)
	result, err := h.Run(bashEvent("echo hi"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Irreversible {
		t.Fatal("Irreversible should be false when rule is not marked")
	}
}

// ── helpers ──────────────────────────────────────────────────────────────────

func bashEvent(cmd string) intercept.InterceptEvent {
	return intercept.InterceptEvent{
		Host:     "claude-code",
		Strategy: "hooks",
		Phase:    intercept.PhaseToolPre,
		Session:  intercept.Session{ID: "test-session"},
		Tool: &intercept.Tool{
			Name:  "Bash",
			Input: map[string]any{"command": cmd},
		},
	}
}

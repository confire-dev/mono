package guardrail_test

import (
	"strings"
	"testing"

	"github.com/confire-dev/confire/guardrail"
	"github.com/confire-dev/confire/intercept"
	"github.com/confire-dev/confire/policy"
)

// aiBannedPhrases are phrases that indicate a message targets the AI agent
// rather than the human user sitting at the terminal.
var aiBannedPhrases = []string{
	"Explain why",
	"explain why",
	"Ask the user",
	"Claude is about to",
	"The agent is about to",
	"You are about to",
}

func bashPreEvent(command string) intercept.InterceptEvent {
	return intercept.InterceptEvent{
		Host:     "claude-code",
		Strategy: "hooks",
		Phase:    intercept.PhaseToolPre,
		Session:  intercept.Session{ID: "test"},
		Tool: &intercept.Tool{
			Name:  "Bash",
			Input: map[string]any{"command": command},
			IsMCP: false,
		},
	}
}

// TestReviewMessageIsUserFacing verifies that review messages contain the
// required user-facing sections and no AI-agent-directed language.
func TestReviewMessageIsUserFacing(t *testing.T) {
	h := guardrail.New(policy.NewEngine(policy.BuiltinRules()), policy.ModeBalanced)
	result, err := h.Run(bashPreEvent("git push --force origin main"))
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Kind != intercept.ResultReview {
		t.Fatalf("expected ResultReview, got %s", result.Kind)
	}
	msg := result.Reason

	for _, want := range []string{"CONFIRE REVIEW REQUIRED", "Rule:", "confire bypass-next"} {
		if !strings.Contains(msg, want) {
			t.Errorf("review message missing %q\nfull message:\n%s", want, msg)
		}
	}
	for _, phrase := range aiBannedPhrases {
		if strings.Contains(msg, phrase) {
			t.Errorf("review message contains AI-facing phrase %q\nfull message:\n%s", phrase, msg)
		}
	}
}

// TestBlockMessageDoesNotMentionBypassNext verifies that block messages do not
// offer bypass-next — that option is only for review.
func TestBlockMessageDoesNotMentionBypassNext(t *testing.T) {
	h := guardrail.New(policy.NewEngine(policy.BuiltinRules()), policy.ModeBalanced)
	result, err := h.Run(bashPreEvent("gh repo delete myorg/myrepo --yes"))
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Kind != intercept.ResultBlock {
		t.Fatalf("expected ResultBlock for repo delete, got %s", result.Kind)
	}
	msg := result.Reason

	if strings.Contains(msg, "bypass-next") {
		t.Errorf("block message must not mention bypass-next (only review messages should):\n%s", msg)
	}
	if !strings.Contains(msg, "CONFIRE BLOCKED") {
		t.Errorf("block message missing 'CONFIRE BLOCKED':\n%s", msg)
	}
	for _, phrase := range aiBannedPhrases {
		if strings.Contains(msg, phrase) {
			t.Errorf("block message contains AI-facing phrase %q:\n%s", phrase, msg)
		}
	}
}

// TestWarnMessageIsUserFacing verifies warn messages (observe mode) carry no
// AI-facing language.
func TestWarnMessageIsUserFacing(t *testing.T) {
	h := guardrail.New(policy.NewEngine(policy.BuiltinRules()), policy.ModeObserve)
	result, err := h.Run(bashPreEvent("git push --force origin main"))
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Kind != intercept.ResultWarn {
		t.Fatalf("expected ResultWarn in observe mode, got %s", result.Kind)
	}
	for _, phrase := range aiBannedPhrases {
		if strings.Contains(result.Context, phrase) {
			t.Errorf("warn message contains AI-facing phrase %q:\n%s", phrase, result.Context)
		}
	}
}

// TestReviewMessageHasRiskNotSeverityLabel verifies that the risk field in the
// review message contains the human-readable rule message, not just "high severity".
func TestReviewMessageHasRiskNotSeverityLabel(t *testing.T) {
	h := guardrail.New(policy.NewEngine(policy.BuiltinRules()), policy.ModeBalanced)
	result, _ := h.Run(bashPreEvent("git push --force origin main"))
	if result.Kind != intercept.ResultReview {
		t.Skip("no review triggered")
	}
	if strings.TrimSpace(result.Reason) == "high severity" || result.Reason == "high" {
		t.Errorf("risk field must be a human description, not just a severity label: %q", result.Reason)
	}
}

// TestBypassModePassesEverything verifies that bypass mode disables all policy
// evaluation — the firewall is completely off.
func TestBypassModePassesEverything(t *testing.T) {
	h := guardrail.New(policy.NewEngine(policy.BuiltinRules()), policy.ModeBypass)
	for _, cmd := range []string{
		"git push --force origin main",
		"gh repo delete myorg/myrepo",
		"rm -rf /",
	} {
		result, err := h.Run(bashPreEvent(cmd))
		if err != nil {
			t.Fatalf("Run(%q): %v", cmd, err)
		}
		if result.Kind != intercept.ResultPassthrough {
			t.Errorf("bypass mode: expected passthrough for %q, got %s", cmd, result.Kind)
		}
	}
}

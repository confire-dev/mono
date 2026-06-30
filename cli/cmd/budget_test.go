package cmd

import (
	"strings"
	"testing"

	"github.com/confire-dev/confire/intercept"
)

// warnResult builds a minimal ResultWarn for budget testing.
func warnResult(ctx string, irreversible bool) intercept.InterceptResult {
	return intercept.InterceptResult{
		Kind:         intercept.ResultWarn,
		Context:      ctx,
		Irreversible: irreversible,
	}
}

func TestBudget_SkipsWithAssumption(t *testing.T) {
	// budget=5: first 5 warns should return passthrough (skip-with-assumption).
	budget := 5
	for asksUsed := 0; asksUsed < budget; asksUsed++ {
		out, skipped, exhausted := budgetDecision(warnResult("some context", false), asksUsed, budget)
		if out.Kind != intercept.ResultPassthrough {
			t.Fatalf("call %d: expected ResultPassthrough, got %v", asksUsed+1, out.Kind)
		}
		if !skipped {
			t.Fatalf("call %d: expected skipped=true", asksUsed+1)
		}
		if exhausted {
			t.Fatalf("call %d: expected exhausted=false", asksUsed+1)
		}
	}
}

func TestBudget_FifthWarnReviews(t *testing.T) {
	// budget=4: the 5th warn (asksUsed=4 == budget) must surface as review, not passthrough.
	budget := 4
	out, skipped, exhausted := budgetDecision(warnResult("ctx", false), budget /*asksUsed==budget*/, budget)
	if out.Kind != intercept.ResultReview {
		t.Fatalf("expected ResultReview when budget exhausted, got %v", out.Kind)
	}
	if skipped || !exhausted {
		t.Fatalf("expected skipped=false exhausted=true; got skipped=%v exhausted=%v", skipped, exhausted)
	}
	if out.Reason == "" {
		t.Fatal("exhausted result must have a non-empty Reason")
	}
	if out.Context != "" {
		t.Fatal("exhausted result must have Context cleared")
	}
}

func TestBudget_ReviewNeverSkipped(t *testing.T) {
	// A review-action result must always stay ResultReview — budget never skips it.
	reviewResult := intercept.InterceptResult{Kind: intercept.ResultReview, Reason: "rule fired"}
	out, skipped, exhausted := budgetDecision(reviewResult, 0, 5)
	if out.Kind != intercept.ResultReview {
		t.Fatalf("expected ResultReview unchanged, got %v", out.Kind)
	}
	if skipped || exhausted {
		t.Fatalf("review must not consume budget; got skipped=%v exhausted=%v", skipped, exhausted)
	}
}

func TestBudget_IrreversibleForcesReview(t *testing.T) {
	// Irreversible warn with budget available: defense-in-depth forces review, budget not consumed.
	out, skipped, exhausted := budgetDecision(warnResult("risky op", true), 0, 5)
	if out.Kind != intercept.ResultReview {
		t.Fatalf("expected ResultReview for irreversible warn, got %v", out.Kind)
	}
	if skipped || exhausted {
		t.Fatalf("irreversible path must not set skipped or exhausted; got skipped=%v exhausted=%v", skipped, exhausted)
	}
}

func TestBudget_BlockUnaffected(t *testing.T) {
	// Block results must pass through budgetDecision unchanged — budget never touches blocks.
	blockResult := intercept.InterceptResult{Kind: intercept.ResultBlock, Reason: "blocked by rule"}
	out, skipped, exhausted := budgetDecision(blockResult, 0, 5)
	if out.Kind != intercept.ResultBlock {
		t.Fatalf("expected ResultBlock unchanged, got %v", out.Kind)
	}
	if out.Reason != "blocked by rule" {
		t.Fatal("block Reason must not be modified")
	}
	if skipped || exhausted {
		t.Fatalf("block must not affect budget; got skipped=%v exhausted=%v", skipped, exhausted)
	}
}

func TestBudget_ExhaustionSignal(t *testing.T) {
	// Caller relies on exhausted=true to emit budget_exhaustion telemetry.
	// Verify the signal fires on the first call that exceeds the budget.
	budget := 1
	_, skipped, _ := budgetDecision(warnResult("ctx", false), 0, budget)
	if !skipped {
		t.Fatal("first warn within budget should skip")
	}
	_, _, exhausted := budgetDecision(warnResult("ctx", false), 1 /*asksUsed==budget*/, budget)
	if !exhausted {
		t.Fatal("first warn past budget must set exhausted=true so caller emits telemetry")
	}
}

func TestBudget_NeverSilent(t *testing.T) {
	// budget=0: every warn immediately hits the exhausted path — never passthrough, never silent.
	out, skipped, exhausted := budgetDecision(warnResult("something", false), 0, 0)
	if out.Kind != intercept.ResultReview {
		t.Fatalf("budget=0 must collapse warn to review immediately, got %v", out.Kind)
	}
	if skipped || !exhausted {
		t.Fatalf("budget=0 should set exhausted=true; got skipped=%v exhausted=%v", skipped, exhausted)
	}
}

func TestBudget_ExhaustedMessage_IncludesOriginalContext(t *testing.T) {
	// The exhausted review Reason must contain the original warn context so the user
	// knows what was being flagged — not just a generic "budget exhausted" message.
	out, _, _ := budgetDecision(warnResult("original warn text", false), 0, 0)
	if !strings.Contains(out.Reason, "original warn text") {
		t.Fatalf("exhausted Reason should include original context; got: %q", out.Reason)
	}
}

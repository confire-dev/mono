package firewall_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/confire-dev/confire/firewall"
	"github.com/confire-dev/confire/intercept"
	"github.com/confire-dev/confire/policy"
)

// ── helpers ──────────────────────────────────────────────────────────────────

func defaultThresholds() firewall.RateThresholds {
	return firewall.DefaultRateThresholds()
}

func makeEvent(toolName string, input map[string]any) intercept.InterceptEvent {
	tool := &intercept.Tool{Name: toolName, Input: input}
	return intercept.InterceptEvent{
		Session: intercept.Session{ID: "test-session"},
		Tool:    tool,
	}
}

// buildCalls creates n RecentCall entries with the given sig, all within window.
func buildCalls(sig string, n int, window time.Duration) []firewall.RecentCall {
	now := time.Now()
	calls := make([]firewall.RecentCall, n)
	for i := range calls {
		calls[i] = firewall.RecentCall{Sig: sig, At: now.Add(-time.Duration(i) * time.Second)}
	}
	return calls
}

// ── TestCheckRateRules_RunawayLoop_Fires ─────────────────────────────────────

func TestCheckRateRules_RunawayLoop_Fires(t *testing.T) {
	event := makeEvent("Bash", map[string]any{"command": "echo hello"})
	sig := firewall.CallSig(event)
	// 20 identical sigs in window (the 20th triggers the block).
	calls := buildCalls(sig, 20, firewall.DefaultRunawayLoopWindow)

	result := firewall.CheckRateRules(calls, event, defaultThresholds(), policy.ModeBalanced)
	if result.Kind != intercept.ResultBlock {
		t.Fatalf("expected ResultBlock, got %v", result.Kind)
	}
	if !strings.Contains(result.Reason, firewall.RuleIDRunawayLoop) {
		t.Errorf("reason should contain rule ID %q, got: %s", firewall.RuleIDRunawayLoop, result.Reason)
	}
}

// ── TestCheckRateRules_RunawayLoop_NotFired_19 ───────────────────────────────

func TestCheckRateRules_RunawayLoop_NotFired_19(t *testing.T) {
	event := makeEvent("Bash", map[string]any{"command": "echo hello"})
	sig := firewall.CallSig(event)
	// 19 identical sigs — one below threshold.
	calls := buildCalls(sig, 19, firewall.DefaultRunawayLoopWindow)

	result := firewall.CheckRateRules(calls, event, defaultThresholds(), policy.ModeBalanced)
	if result.Kind != intercept.ResultPassthrough {
		t.Fatalf("expected ResultPassthrough at 19 calls, got %v", result.Kind)
	}
}

// ── TestCheckRateRules_RunawayLoop_WindowExpiry ───────────────────────────────

func TestCheckRateRules_RunawayLoop_WindowExpiry(t *testing.T) {
	event := makeEvent("Bash", map[string]any{"command": "echo hello"})
	sig := firewall.CallSig(event)

	now := time.Now()
	// 11 old calls outside the 5-min window + 9 recent ones (same sig). Total = 20,
	// but only 9 are within the window → should not fire.
	var calls []firewall.RecentCall
	for i := 0; i < 11; i++ {
		calls = append(calls, firewall.RecentCall{Sig: sig, At: now.Add(-6 * time.Minute)})
	}
	for i := 0; i < 9; i++ {
		calls = append(calls, firewall.RecentCall{Sig: sig, At: now.Add(-time.Duration(i+1) * 10 * time.Second)})
	}

	result := firewall.CheckRateRules(calls, event, defaultThresholds(), policy.ModeBalanced)
	if result.Kind != intercept.ResultPassthrough {
		t.Fatalf("expected ResultPassthrough (9 in-window, need 20), got %v", result.Kind)
	}
}

// ── TestCheckRateRules_RunawayLoop_DifferentSigs ─────────────────────────────

func TestCheckRateRules_RunawayLoop_DifferentSigs(t *testing.T) {
	now := time.Now()
	// 20 different sigs (reads of 20 different files) — not a loop.
	var calls []firewall.RecentCall
	for i := 0; i < 20; i++ {
		e := makeEvent("Read", map[string]any{"file_path": "/src/file_" + string(rune('a'+i)) + ".go"})
		calls = append(calls, firewall.RecentCall{Sig: firewall.CallSig(e), At: now.Add(-time.Duration(i) * time.Second)})
	}

	// The current event is one more distinct read.
	event := makeEvent("Read", map[string]any{"file_path": "/src/file_z.go"})
	result := firewall.CheckRateRules(calls, event, defaultThresholds(), policy.ModeBalanced)
	if result.Kind != intercept.ResultPassthrough {
		t.Fatalf("expected ResultPassthrough (different sigs), got %v", result.Kind)
	}
}

// ── TestCheckRateRules_CallCap_Fires ─────────────────────────────────────────

func TestCheckRateRules_CallCap_Fires(t *testing.T) {
	// 101 calls of distinct sigs within 1 min → block.
	// Space at 100ms intervals so all 101 fit within the 60s window.
	now := time.Now()
	calls := make([]firewall.RecentCall, 101)
	for i := range calls {
		calls[i] = firewall.RecentCall{
			Sig: fmt.Sprintf("unique_sig_%d", i),
			At:  now.Add(-time.Duration(i) * 100 * time.Millisecond),
		}
	}

	event := makeEvent("Write", map[string]any{"path": "/tmp/x"})
	result := firewall.CheckRateRules(calls, event, defaultThresholds(), policy.ModeBalanced)
	if result.Kind != intercept.ResultBlock {
		t.Fatalf("expected ResultBlock at 101 calls, got %v", result.Kind)
	}
	if !strings.Contains(result.Reason, firewall.RuleIDCallCap) {
		t.Errorf("reason should contain rule ID %q, got: %s", firewall.RuleIDCallCap, result.Reason)
	}
}

// ── TestCheckRateRules_CallCap_NotFired_100 ───────────────────────────────────

func TestCheckRateRules_CallCap_NotFired_100(t *testing.T) {
	// Exactly 100 distinct calls within 1 min — should not fire (threshold is >100).
	now := time.Now()
	calls := make([]firewall.RecentCall, 100)
	for i := range calls {
		calls[i] = firewall.RecentCall{
			Sig: fmt.Sprintf("sig_%d", i),
			At:  now.Add(-time.Duration(i) * 100 * time.Millisecond),
		}
	}

	event := makeEvent("Write", map[string]any{"path": "/tmp/x"})
	result := firewall.CheckRateRules(calls, event, defaultThresholds(), policy.ModeBalanced)
	if result.Kind != intercept.ResultPassthrough {
		t.Fatalf("expected ResultPassthrough at exactly 100 calls, got %v", result.Kind)
	}
}

// ── TestCheckRateRules_ObserveMode_DowngradesBlock ───────────────────────────

func TestCheckRateRules_ObserveMode_DowngradesBlock(t *testing.T) {
	event := makeEvent("Bash", map[string]any{"command": "make test"})
	sig := firewall.CallSig(event)
	calls := buildCalls(sig, 20, firewall.DefaultRunawayLoopWindow)

	result := firewall.CheckRateRules(calls, event, defaultThresholds(), policy.ModeObserve)
	if result.Kind != intercept.ResultWarn {
		t.Fatalf("observe mode: expected ResultWarn (downgraded from block), got %v", result.Kind)
	}
}

// ── TestCallSig_Stability ─────────────────────────────────────────────────────

func TestCallSig_Stability(t *testing.T) {
	event := makeEvent("Read", map[string]any{"file_path": "/Users/user/project/main.go"})
	sig1 := firewall.CallSig(event)
	sig2 := firewall.CallSig(event)
	if sig1 != sig2 {
		t.Fatalf("CallSig is not stable: %q != %q", sig1, sig2)
	}
	if sig1 == "" {
		t.Fatal("CallSig returned empty string")
	}
}

// ── TestCallSig_InputTruncation ───────────────────────────────────────────────

func TestCallSig_InputTruncation(t *testing.T) {
	// Build a long input string (>256 bytes of JSON).
	longValue := strings.Repeat("x", 500)
	event := makeEvent("Bash", map[string]any{"command": longValue})

	// Same first 256 bytes → same sig regardless of tail content.
	var truncInput map[string]any
	b, _ := json.Marshal(map[string]any{"command": longValue})
	b = b[:256]
	// Reconstruct the truncation that CallSig applies.
	_ = truncInput

	sigFull := firewall.CallSig(event)

	// A call with a longer value but the same first 256 bytes (still differing tail
	// only after byte 256) should ideally produce the same sig. Construct it:
	longerValue := longValue + "extra-tail-that-differs"
	eventLonger := makeEvent("Bash", map[string]any{"command": longerValue})

	// The hash is over the first 256 bytes of the JSON-marshaled input.
	// JSON: {"command":"xxx..."} — first 256 bytes are identical since the value
	// prefix is the same. Verify the sigs match.
	sigLonger := firewall.CallSig(eventLonger)

	if sigFull != sigLonger {
		// Only fail if the first 256 bytes of the JSON representations are identical.
		bFull, _ := json.Marshal(map[string]any{"command": longValue})
		bLonger, _ := json.Marshal(map[string]any{"command": longerValue})
		if len(bFull) >= 256 && len(bLonger) >= 256 && string(bFull[:256]) == string(bLonger[:256]) {
			t.Errorf("sigs differ even though first 256 bytes are identical: %q vs %q", sigFull, sigLonger)
		}
		// If first 256 bytes differ, differing sigs are correct — not an error.
	}
}

// ── TestCallSig_LongPrefixCollision ──────────────────────────────────────────
//
// Inputs whose first 256 bytes of JSON are identical but whose tails differ
// will hash to the same CallSig. This is intentional: the truncated hash is a
// "near-identical call" heuristic, not a perfect identity check. At the loose
// default threshold of 20, a genuine batch over differently-suffixed paths at
// constant high volume will still accumulate enough identical prefixes to fire.
// This test names and asserts that behavior so it is never accidentally removed.

func TestCallSig_LongPrefixCollision(t *testing.T) {
	// JSON structure: {"file_path":"<value>"}
	// Header `{"file_path":"` = 14 bytes. We need 256 - 14 = 242 bytes of identical
	// value prefix so that the first 256 bytes of every JSON blob are the same.
	commonPrefix := strings.Repeat("x", 242)

	now := time.Now()
	calls := make([]firewall.RecentCall, 20)
	for i := range calls {
		e := makeEvent("Read", map[string]any{"file_path": commonPrefix + fmt.Sprintf("/variant%d", i)})
		calls[i] = firewall.RecentCall{Sig: firewall.CallSig(e), At: now.Add(-time.Duration(i) * time.Second)}
	}

	// Assert all 20 calls share the same sig due to the prefix collision.
	firstSig := calls[0].Sig
	for i, c := range calls[1:] {
		if c.Sig != firstSig {
			t.Fatalf("expected all calls to share one sig (first-256-byte collision); call %d differs: %q vs %q", i+1, firstSig, c.Sig)
		}
	}

	// The 20 colliding entries plus the current event (same prefix) fires runaway_loop.
	event := makeEvent("Read", map[string]any{"file_path": commonPrefix + "/variant20"})
	result := firewall.CheckRateRules(calls, event, defaultThresholds(), policy.ModeBalanced)
	if result.Kind != intercept.ResultBlock {
		t.Fatalf("expected ResultBlock: 20 calls with identical 256-byte JSON prefix should trigger runaway_loop (intentional collision behavior), got %v", result.Kind)
	}
}

// ── TestLegitBatch_50DifferentFiles ──────────────────────────────────────────

func TestLegitBatch_50DifferentFiles(t *testing.T) {
	// 50 reads of 50 different paths within 1 min: not a loop, under the call cap.
	now := time.Now()
	calls := make([]firewall.RecentCall, 50)
	for i := range calls {
		e := makeEvent("Read", map[string]any{"file_path": "/src/" + string(rune('a'+i%26)) + string(rune('0'+i%10)) + ".go"})
		calls[i] = firewall.RecentCall{Sig: firewall.CallSig(e), At: now.Add(-time.Duration(i) * time.Second)}
	}

	// Current event is one more distinct read.
	event := makeEvent("Read", map[string]any{"file_path": "/src/main.go"})
	result := firewall.CheckRateRules(calls, event, defaultThresholds(), policy.ModeBalanced)
	if result.Kind != intercept.ResultPassthrough {
		t.Fatalf("legitimate batch of 50 distinct reads should not fire: got %v", result.Kind)
	}
}

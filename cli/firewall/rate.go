// Package firewall — rate.go implements the rate-policy rules:
//
//   rate.runaway_loop  same call (tool + input hash) repeated N times in M minutes → block
//   rate.call_cap      more than N total calls per minute → block
//
// These rules are programmatic (unlike JSON-driven policy rules) because they
// require cross-call state that matchesPreTool() does not have access to.
// They share the firewall package with the cross-tool flow rules (flow.go).
package firewall

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"strings"
	"time"

	"github.com/confire-dev/confire/intercept"
	"github.com/confire-dev/confire/policy"
)

// Rule ID constants — used in telemetry PatternMatched and block messages.
const (
	RuleIDRunawayLoop = "rate.runaway_loop"
	RuleIDCallCap     = "rate.call_cap"
)

// Default thresholds. Deliberately loose so they never fire on legitimate
// batch work (test runs, refactors, sequential file reads).
const (
	DefaultRunawayLoopN      = 20              // same-sig calls within window → block
	DefaultRunawayLoopWindow = 5 * time.Minute
	DefaultCallCapN          = 100             // total calls within window → block at N+1
	DefaultCallCapWindow     = 1 * time.Minute
	InputHashPrefixBytes     = 256

	// MaxRateWindow is the maximum look-back window used by any rate rule.
	// The daemon prunes recentCalls entries older than this.
	MaxRateWindow = DefaultRunawayLoopWindow // 5 minutes
)

// RecentCall is one entry in the per-session call history used by rate rules.
// It is separate from the provenance ring buffer (which is capped at 10 entries
// for cross-tool flow detection and is too small for time-windowed counting).
type RecentCall struct {
	Sig string    // CallSig() output: "toolname:fnv64hex"
	At  time.Time
}

// RateThresholds holds the configurable limits for both rate rules.
// Build one from config at daemon startup and pass it through to CheckRateRules.
type RateThresholds struct {
	RunawayLoopN      int
	RunawayLoopWindow time.Duration
	CallCapN          int
	CallCapWindow     time.Duration
}

// DefaultRateThresholds returns the hardcoded loose defaults.
// Free users always get these; paid users can tighten via Worker config endpoint.
func DefaultRateThresholds() RateThresholds {
	return RateThresholds{
		RunawayLoopN:      DefaultRunawayLoopN,
		RunawayLoopWindow: DefaultRunawayLoopWindow,
		CallCapN:          DefaultCallCapN,
		CallCapWindow:     DefaultCallCapWindow,
	}
}

// CallSig returns a stable, content-sensitive identifier for a tool call.
// It hashes the tool name + first InputHashPrefixBytes bytes of the JSON-marshaled
// input so that near-identical retry calls produce the same sig, while calls
// operating on different targets (different files, different commands) produce
// different sigs.
func CallSig(event intercept.InterceptEvent) string {
	if event.Tool == nil {
		return ""
	}
	b, _ := json.Marshal(event.Tool.Input)
	if len(b) > InputHashPrefixBytes {
		b = b[:InputHashPrefixBytes]
	}
	h := fnv.New64a()
	h.Write([]byte(strings.ToLower(event.Tool.Name)))
	h.Write(b)
	return fmt.Sprintf("%s:%016x", strings.ToLower(event.Tool.Name), h.Sum64())
}

// CheckRateRules evaluates both rate-policy rules against the recent call history.
// calls must already include the current call (the daemon appends before calling).
// Returns a non-passthrough result if a rate limit is exceeded.
// Respects bypass and observe modes the same way flow rules do.
func CheckRateRules(
	calls []RecentCall,
	event intercept.InterceptEvent,
	thresholds RateThresholds,
	mode policy.Mode,
) intercept.InterceptResult {
	if mode == policy.ModeBypass || event.Tool == nil {
		return intercept.InterceptResult{Kind: intercept.ResultPassthrough}
	}

	now := time.Now()
	sig := CallSig(event)

	// rate.runaway_loop: count same-sig calls within the loop window.
	loopCount := 0
	loopCutoff := now.Add(-thresholds.RunawayLoopWindow)
	for _, c := range calls {
		if c.At.After(loopCutoff) && c.Sig == sig {
			loopCount++
		}
	}
	if loopCount >= thresholds.RunawayLoopN {
		kind := intercept.ResultBlock
		if mode == policy.ModeObserve {
			kind = intercept.ResultWarn
		}
		return intercept.InterceptResult{
			Kind:   kind,
			Reason: runawayLoopMessage(event.Tool.Name, loopCount, thresholds),
		}
	}

	// rate.call_cap: count all calls within the cap window.
	capCount := 0
	capCutoff := now.Add(-thresholds.CallCapWindow)
	for _, c := range calls {
		if c.At.After(capCutoff) {
			capCount++
		}
	}
	if capCount > thresholds.CallCapN {
		kind := intercept.ResultBlock
		if mode == policy.ModeObserve {
			kind = intercept.ResultWarn
		}
		return intercept.InterceptResult{
			Kind:   kind,
			Reason: callCapMessage(capCount, thresholds),
		}
	}

	return intercept.InterceptResult{Kind: intercept.ResultPassthrough}
}

// ── Message copy ──────────────────────────────────────────────────────────────

func runawayLoopMessage(toolName string, count int, t RateThresholds) string {
	return fmt.Sprintf(`[Confire %s] %s has been called %d times with the same input in the past %s — this looks like a runaway retry loop.

To continue immediately: run `+"`confire off`"+` to disable the firewall for this batch, then `+"`confire on`"+` when done.
To raise the threshold permanently: add "runaway_loop_threshold": 50 to ~/.confire/config.json, then run `+"`confire stop && confire start`"+` to apply.`,
		RuleIDRunawayLoop, toolName, count, formatDuration(t.RunawayLoopWindow))
}

func callCapMessage(count int, t RateThresholds) string {
	return fmt.Sprintf(`[Confire %s] %d tool calls in the past %s exceeds the rate limit (>%d). Session blocked to prevent runaway agent behavior.

To continue immediately: run `+"`confire off`"+` to disable the firewall for this batch, then `+"`confire on`"+` when done.
To raise the threshold permanently: add "call_rate_threshold": 200 to ~/.confire/config.json, then run `+"`confire stop && confire start`"+` to apply.`,
		RuleIDCallCap, count, formatDuration(t.CallCapWindow), t.CallCapN)
}

func formatDuration(d time.Duration) string {
	switch d {
	case time.Minute:
		return "1 minute"
	case 5 * time.Minute:
		return "5 minutes"
	default:
		return d.String()
	}
}

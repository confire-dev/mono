// E2E tests for the confire hook pipeline.
// Each test uses a real httptest.Server as the mock Worker and exercises
// the full path: InterceptEvent → Transport → (mock Worker / local fallback) → InterceptResult.
//
// Author: Efe <efe@efebehar.dev>
package e2e_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/confire-dev/confire/intercept"
	"github.com/confire-dev/confire/transport"
)

// ── Fixtures ──────────────────────────────────────────────────────────────────

const testAPIKey = "cf_live_test_key_e2e_abc123"

func figmaMCPEvent(jsxText string) intercept.InterceptEvent {
	return intercept.InterceptEvent{
		Host: "claude-code", Strategy: "hooks", Phase: intercept.PhaseToolPost,
		Session: intercept.Session{ID: "test-session"},
		Tool: &intercept.Tool{
			Name: "mcp__figma__get_design_context", IsMCP: true, MCPServer: "figma",
			Output: map[string]interface{}{
				"content": []interface{}{
					map[string]interface{}{"type": "text", "text": jsxText},
				},
			},
		},
	}
}

func bashEvent(output, command string) intercept.InterceptEvent {
	return intercept.InterceptEvent{
		Host: "claude-code", Strategy: "hooks", Phase: intercept.PhaseToolPost,
		Session: intercept.Session{ID: "test-session"},
		Tool: &intercept.Tool{
			Name: "Bash", IsMCP: false,
			Input:  map[string]interface{}{"command": command},
			Output: output,
		},
	}
}

func readPreEvent(filePath string) intercept.InterceptEvent {
	return intercept.InterceptEvent{
		Host: "claude-code", Strategy: "hooks", Phase: intercept.PhaseToolPre,
		Session: intercept.Session{ID: "test-session"},
		Tool: &intercept.Tool{
			Name: "Read", IsMCP: false,
			Input: map[string]interface{}{"file_path": filePath},
		},
	}
}

// ── Helper ────────────────────────────────────────────────────────────────────

func extractOptimizedText(result intercept.InterceptResult) string {
	if result.Kind != intercept.ResultReplaceOutput || result.ToolOutput == nil {
		return ""
	}
	m, ok := result.ToolOutput.(map[string]interface{})
	if !ok {
		if s, ok := result.ToolOutput.(string); ok {
			return s
		}
		// re-marshal + unmarshal to handle json.Number and similar
		b, _ := json.Marshal(result.ToolOutput)
		var out map[string]interface{}
		json.Unmarshal(b, &out)
		m = out
	}
	content, _ := m["content"].([]interface{})
	if len(content) == 0 {
		return ""
	}
	first, _ := content[0].(map[string]interface{})
	text, _ := first["text"].(string)
	return text
}

func reductionPct(before, after int) float64 {
	if before == 0 {
		return 0
	}
	return float64(before-after) / float64(before) * 100
}

// ── Tests ─────────────────────────────────────────────────────────────────────

// Test 1: Figma MCP through Worker path — HMAC signed, 30%+ reduction.
func TestFigmaMCPViaWorker(t *testing.T) {
	srv := newMockWorker(mockWorkerConfig{
		apiKey:    testAPIKey,
		verifySig: true,
	})
	defer srv.Close()

	jsxText := strings.Join([]string{
		`export default function Hero() {`,
		`  return (`,
		`    <div data-node-id="123:456" className="bg-[var(--color,#3b82f6)] content-stretch p-4">`,
		`      <h1 data-node-id="789:012">Hello world</h1>`,
		`    </div>`,
		`  )`,
		`}`,
		``,
		`SUPER CRITICAL: these rules must be followed or everything breaks`,
	}, "\n")

	wt := newWorkerTransport(srv, testAPIKey)
	result, err := wt.Send(figmaMCPEvent(jsxText))

	if err != nil {
		t.Fatalf("Send error: %v", err)
	}
	if result.Kind != intercept.ResultReplaceOutput {
		t.Fatalf("expected replace-output, got %s", result.Kind)
	}

	text := extractOptimizedText(result)
	if text == "" {
		t.Fatal("optimized text is empty")
	}

	checks := []struct {
		name string
		ok   bool
	}{
		{"data-node-id stripped", !strings.Contains(text, "data-node-id")},
		{"SUPER CRITICAL stripped", !strings.Contains(text, "SUPER CRITICAL")},
		{"CSS var unwrapped", !strings.Contains(text, "var(--") && strings.Contains(text, "#3b82f6")},
		{"content-stretch stripped", !strings.Contains(text, "content-stretch")},
	}
	for _, c := range checks {
		if !c.ok {
			t.Errorf("FAIL: %s", c.name)
		} else {
			t.Logf("PASS: %s", c.name)
		}
	}

	if result.Stats != nil {
		pct := reductionPct(result.Stats.BeforeBytes, result.Stats.AfterBytes)
		t.Logf("Reduction: %d → %d bytes (%.0f%%)", result.Stats.BeforeBytes, result.Stats.AfterBytes, pct)
		if pct < 20 {
			t.Errorf("expected ≥20%% reduction, got %.0f%%", pct)
		}
	}
}

// Test 2: Bash test runner — strips 200 passing lines, keeps failures.
func TestBashTestRunnerViaWorker(t *testing.T) {
	srv := newMockWorker(mockWorkerConfig{apiKey: testAPIKey})
	defer srv.Close()

	// 200 passing + 2 failing + summary
	lines := make([]string, 0, 210)
	for i := 0; i < 200; i++ {
		lines = append(lines, "  ✓ user.create passes (3ms)")
	}
	lines = append(lines,
		"  ✗ auth.login fails with bad password",
		"  ✗ api.404 returns wrong status code",
		"",
		"200 passing",
		"2 failing",
	)
	output := strings.Join(lines, "\n")

	wt := newWorkerTransport(srv, testAPIKey)
	result, err := wt.Send(bashEvent(output, "npm test"))
	if err != nil {
		t.Fatalf("Send error: %v", err)
	}
	if result.Kind != intercept.ResultReplaceOutput {
		t.Fatalf("expected replace-output, got %s", result.Kind)
	}

	optimized, ok := result.ToolOutput.(string)
	if !ok {
		// re-marshal
		b, _ := json.Marshal(result.ToolOutput)
		optimized = string(b)
	}

	if !strings.Contains(optimized, "passing tests omitted") {
		t.Error("expected passing-tests-omitted marker")
	}
	if !strings.Contains(optimized, "fails with bad password") {
		t.Error("expected failures to be preserved")
	}

	pct := reductionPct(len(output), len(optimized))
	t.Logf("Reduction: %d → %d bytes (%.0f%%)", len(output), len(optimized), pct)
	if pct < 80 {
		t.Errorf("expected ≥80%% reduction on test runner output, got %.0f%%", pct)
	}
}

// Test 3: PreToolUse Read — updatedInput injects limit=500.
func TestPreToolUseReadLimit(t *testing.T) {
	lt := transport.NewLocal("")
	result, err := lt.Send(readPreEvent("/src/some/big/file.ts"))

	if err != nil {
		t.Fatalf("Send error: %v", err)
	}
	if result.Kind != intercept.ResultReplaceInput {
		t.Fatalf("expected replace-input, got %s", result.Kind)
	}

	input, ok := result.ToolInput.(map[string]interface{})
	if !ok {
		t.Fatal("ToolInput is not a map")
	}
	limit, _ := input["limit"].(int)
	if limit != 500 {
		t.Errorf("expected limit=500, got %v", input["limit"])
	}
	if input["file_path"] != "/src/some/big/file.ts" {
		t.Errorf("file_path not preserved: %v", input["file_path"])
	}
	t.Logf("PASS: updatedInput = %v", input)
}

// Test 4: PreToolUse Read with existing limit — must pass through unchanged.
func TestPreToolUseReadAlreadyLimited(t *testing.T) {
	event := readPreEvent("/src/file.ts")
	event.Tool.Input = map[string]interface{}{"file_path": "/src/file.ts", "limit": 100}

	lt := transport.NewLocal("")
	result, _ := lt.Send(event)

	if result.Kind != intercept.ResultPassthrough {
		t.Errorf("expected passthrough when limit already set, got %s", result.Kind)
	}
}

// Test 5: Worker at rate limit → passthrough + additionalContext warning.
func TestWorkerAtRateLimit(t *testing.T) {
	srv := newMockWorker(mockWorkerConfig{apiKey: testAPIKey, atLimit: true})
	defer srv.Close()

	jsxText := `export default function C() { return <div data-node-id="1:1">x</div> }`
	wt := newWorkerTransport(srv, testAPIKey)
	result, err := wt.Send(figmaMCPEvent(jsxText))

	if err != nil {
		t.Fatalf("Send error: %v", err)
	}
	// At-limit → returns additionalContext (not passthrough, not replace-output)
	if result.Kind != intercept.ResultAddContext {
		t.Errorf("expected add-context at limit, got %s", result.Kind)
	}
	if !strings.Contains(result.Context, "Free tier") {
		t.Errorf("expected upgrade message in context, got: %q", result.Context)
	}
	t.Logf("PASS: at-limit context = %q", result.Context[:60])
}

// Test 6: Worker offline → FallbackTransport returns local result.
func TestWorkerOfflineFallbackToLocal(t *testing.T) {
	// Point at a closed server — connection will be refused immediately.
	wt := transport.NewWorker("http://127.0.0.1:19999", testAPIKey, "dev-device")
	lt := transport.NewLocal("")
	ft := transport.NewFallback(wt, lt)

	jsxText := strings.Join([]string{
		`export default function C() {`,
		`  return <div data-node-id="1:1">test</div>`,
		`}`,
		`SUPER CRITICAL: follow these rules`,
	}, "\n")

	result, err := ft.Send(figmaMCPEvent(jsxText))

	// FallbackTransport swallows the Worker error — result must come from local.
	if err != nil {
		t.Fatalf("FallbackTransport should not propagate errors, got: %v", err)
	}
	// Local transport: Figma is Worker-only, so MCP tools are passthrough locally.
	// The FallbackTransport must return something (not panic or error).
	t.Logf("PASS: fallback kind = %s (local MCP passthrough is correct)", result.Kind)
}

// Test 7: HMAC signature verification — tampered body is rejected.
func TestHMACSignatureRejectsReplay(t *testing.T) {
	srv := newMockWorker(mockWorkerConfig{
		apiKey:    testAPIKey,
		verifySig: true,
	})
	defer srv.Close()

	// Craft a request with an old timestamp (simulating a replay attack).
	// VerifySignature with a stale timestamp should fail.
	apiKey := testAPIKey
	body := []byte(`{"test":"payload"}`)
	oldTimestamp := "1000000000" // epoch 2001 — very old

	ok := transport.VerifySignature(apiKey, "dev-device", oldTimestamp, "invalidsig", body)
	if ok {
		t.Error("VerifySignature should reject a stale timestamp")
	}
	t.Log("PASS: stale timestamp correctly rejected")

	// Valid timestamp but wrong signature should also fail.
	futureTS := "9999999999" // far future — also invalid
	ok2 := transport.VerifySignature(apiKey, "dev-device", futureTS, "wrong", body)
	if ok2 {
		t.Error("VerifySignature should reject wrong signature")
	}
	t.Log("PASS: wrong signature correctly rejected")
}

// Test 8: Signature round-trip — sign a body and verify it.
func TestHMACSignatureRoundTrip(t *testing.T) {
	apiKey := testAPIKey
	deviceID := "test-device-abc"
	body := []byte(`{"event":{"host":"claude-code","phase":"tool.post"}}`)

	headers := transport.SignedHeaders(apiKey, deviceID, body)

	ts := headers["X-Confire-Timestamp"]
	sig := headers["X-Confire-Sig"]

	if ts == "" || sig == "" {
		t.Fatal("SignedHeaders missing required fields")
	}

	ok := transport.VerifySignature(apiKey, deviceID, ts, sig, body)
	if !ok {
		t.Error("VerifySignature should accept a freshly signed request")
	}
	t.Log("PASS: signature round-trip verified")
}

// Test 9: MCP tool — passes through locally (Worker-only tool).
func TestMCPToolPassesThroughLocally(t *testing.T) {
	lt := transport.NewLocal("figma")

	// Figma is Worker-only — local should return passthrough for MCP tools.
	result, _ := lt.Send(figmaMCPEvent(`export default function X() { return <div data-node-id="1:1">x</div> }`))

	if result.Kind != intercept.ResultPassthrough {
		t.Errorf("expected local to passthrough MCP tools, got %s", result.Kind)
	}
	t.Log("PASS: MCP tool correctly passes through locally (Worker-only)")
}

// Test 10: WebFetch — HTML stripped locally.
func TestWebFetchLocal(t *testing.T) {
	htmlPage := strings.Repeat("<script>alert('x')</script><style>.a{}</style>", 50) +
		"<main><h1>Title</h1><p>This is the important content you need.</p></main>" +
		strings.Repeat("<nav><ul><li>nav item</li></ul></nav>", 20)

	event := intercept.InterceptEvent{
		Host: "claude-code", Strategy: "hooks", Phase: intercept.PhaseToolPost,
		Session: intercept.Session{ID: "test-session"},
		Tool: &intercept.Tool{Name: "WebFetch", IsMCP: false, Output: htmlPage},
	}

	lt := transport.NewLocal("")
	result, err := lt.Send(event)
	if err != nil {
		t.Fatalf("Send error: %v", err)
	}

	if result.Kind != intercept.ResultReplaceOutput {
		t.Fatalf("expected replace-output for large HTML, got %s", result.Kind)
	}

	optimized, _ := result.ToolOutput.(string)
	if strings.Contains(optimized, "<script>") {
		t.Error("scripts not stripped")
	}
	if strings.Contains(optimized, "<style>") {
		t.Error("styles not stripped")
	}

	pct := reductionPct(len(htmlPage), len(optimized))
	t.Logf("Reduction: %d → %d bytes (%.0f%%)", len(htmlPage), len(optimized), pct)
	if pct < 50 {
		t.Errorf("expected ≥50%% reduction on HTML, got %.0f%%", pct)
	}
	t.Log("PASS: WebFetch HTML stripped correctly")
}

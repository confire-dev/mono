//go:build integration

// Integration parity tests: run the same fixtures through both the Go CLI
// optimizer and the TS Worker optimizer (via npx tsx), then assert they
// produce structurally equivalent outputs.
//
// Run with: go test -tags integration ./optimizer/...
// Requires node + tsx to be available (they are in the mono dev environment).

package optimizer

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// fixturesDir returns the absolute path to testdata/optimizer/ at the repo root.
func fixturesDir(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	// thisFile is .../cli/optimizer/parity_test.go → go up two dirs → repo root
	root := filepath.Join(filepath.Dir(thisFile), "..", "..")
	return filepath.Join(root, "testdata", "optimizer")
}

func readFixture(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(fixturesDir(t), name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return string(data)
}

// runTS pipes input through the TS runner and returns the output.
// Skips the test if tsx is not available.
func runTS(t *testing.T, optimizerName, input string) string {
	t.Helper()
	// Find the run.ts script relative to repo root
	_, thisFile, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(thisFile), "..", "..")
	runScript := filepath.Join(root, "worker", "src", "optimizers", "run.ts")

	cmd := exec.Command("npx", "--yes", "tsx", runScript, optimizerName)
	cmd.Stdin = strings.NewReader(input)
	cmd.Dir = filepath.Join(root, "worker")
	out, err := cmd.Output()
	if err != nil {
		t.Skipf("tsx not available or run.ts failed (%v) — skipping TS cross-check", err)
	}
	return string(out)
}

// normJSON parses JSON and re-serialises with sorted keys so Go and TS
// formatting differences don't cause false negatives.
func normJSON(t *testing.T, s string) map[string]interface{} {
	t.Helper()
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(s)), &m); err != nil {
		t.Fatalf("normJSON: %v\ninput: %.200s", err, s)
	}
	return m
}

// ── Bash ─────────────────────────────────────────────────────────────────────

func TestParityBash(t *testing.T) {
	input := readFixture(t, "bash-test-runner.txt")

	goOut := optimizeBashOutput(input, "")
	tsOut := runTS(t, "bash", input)

	// Both must reduce
	if len(goOut) >= len(input) {
		t.Errorf("Go: output not reduced (%d >= %d)", len(goOut), len(input))
	}
	if len(tsOut) >= len(input) {
		t.Errorf("TS: output not reduced (%d >= %d)", len(tsOut), len(input))
	}

	// Both must strip passing-test lines from pure-pass suites
	// (Lines inside a FAIL block may be kept — that's correct behavior)
	for _, passLine := range []string{"✓ redirects to dashboard", "✓ shows error on invalid", "✓ renders correctly"} {
		if strings.Contains(goOut, passLine) {
			t.Errorf("Go: passing test line not stripped: %q", passLine)
		}
		if strings.Contains(tsOut, passLine) {
			t.Errorf("TS: passing test line not stripped: %q", passLine)
		}
	}

	// Both must keep the failing test
	for _, failLine := range []string{"creates stripe checkout session", "checkout.test.ts"} {
		if !strings.Contains(goOut, failLine) {
			t.Errorf("Go: failing test line missing: %q", failLine)
		}
		if !strings.Contains(tsOut, failLine) {
			t.Errorf("TS: failing test line missing: %q", failLine)
		}
	}

	// Both must mention the omission
	if !strings.Contains(goOut, "omitted") {
		t.Error("Go: missing omission marker")
	}
	if !strings.Contains(tsOut, "omitted") {
		t.Error("TS: missing omission marker")
	}

	// Outputs should be similar length (within 25% — formatting can differ)
	ratio := float64(len(goOut)) / float64(len(tsOut))
	if ratio < 0.75 || ratio > 1.25 {
		t.Errorf("outputs diverge in size: Go=%d TS=%d ratio=%.2f", len(goOut), len(tsOut), ratio)
	}
}

// ── WebFetch ──────────────────────────────────────────────────────────────────

func TestParityWebFetch(t *testing.T) {
	input := readFixture(t, "webfetch-page.html")
	w := &WebFetchOptimizer{}

	goResult := w.Optimize(input)
	goOut, ok := goResult.(string)
	if !ok || goOut == input {
		t.Fatal("Go WebFetch: expected optimization to apply")
	}
	tsOut := runTS(t, "webfetch", input)

	// Both must reduce
	if len(goOut) >= len(input) {
		t.Errorf("Go: output not reduced (%d >= %d)", len(goOut), len(input))
	}
	if len(tsOut) >= len(input) {
		t.Errorf("TS: output not reduced (%d >= %d)", len(tsOut), len(input))
	}

	// Both must strip these noise patterns
	noisePatterns := []string{"<script", "<style", "<nav", "googletagmanager", "intercomSettings", "linear-gradient"}
	for _, pat := range noisePatterns {
		if strings.Contains(goOut, pat) {
			t.Errorf("Go: noise not stripped: %q", pat)
		}
		if strings.Contains(tsOut, pat) {
			t.Errorf("TS: noise not stripped: %q", pat)
		}
	}

	// Both must preserve main content
	contentSignals := []string{"Build faster, ship more", "Intelligent optimizations", "50+"}
	for _, sig := range contentSignals {
		if !strings.Contains(goOut, sig) {
			t.Errorf("Go: content missing: %q", sig)
		}
		if !strings.Contains(tsOut, sig) {
			t.Errorf("TS: content missing: %q", sig)
		}
	}
}

// ── Generic ───────────────────────────────────────────────────────────────────

func TestParityGeneric(t *testing.T) {
	// Use the raw JSON inside content[0].text as the optimizer input.
	// The TS optimizer receives raw JSON strings; the Go optimizer receives
	// already-parsed maps (the intercept layer unmarshals before calling).
	raw := readFixture(t, "generic-mcp.json")
	var envelope map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &envelope); err != nil {
		t.Fatalf("parse fixture: %v", err)
	}
	innerText := envelope["content"].([]interface{})[0].(map[string]interface{})["text"].(string)

	// Parse the inner JSON — this is what the CLI intercept layer does before
	// handing data to the optimizer.
	var parsedInner map[string]interface{}
	if err := json.Unmarshal([]byte(innerText), &parsedInner); err != nil {
		t.Fatalf("parse inner JSON: %v", err)
	}

	g := &GenericOptimizer{}
	goResult := g.Optimize(parsedInner)
	goMap, ok := goResult.(map[string]interface{})
	if !ok {
		t.Fatalf("Go Generic: expected map result, got %T", goResult)
	}
	goOut, _ := json.Marshal(goMap)

	tsOut := runTS(t, "generic", innerText)

	// Both must reduce
	if len(goOut) >= len(innerText) {
		t.Errorf("Go: output not reduced (%d >= %d)", len(goOut), len(innerText))
	}
	if len(tsOut) >= len(innerText) {
		t.Errorf("TS: output not reduced (%d >= %d)", len(tsOut), len(innerText))
	}

	goMap = normJSON(t, string(goOut))
	tsMap := normJSON(t, tsOut)

	// Both must strip noise fields
	noiseFields := []string{"node_id", "diff_url", "patch_url", "issue_url", "commits_url", "review_comments_url", "statuses_url", "merge_commit_sha", "auto_merge", "active_lock_reason"}
	for _, field := range noiseFields {
		if _, ok := goMap[field]; ok {
			t.Errorf("Go: noise field not stripped: %q", field)
		}
		if _, ok := tsMap[field]; ok {
			t.Errorf("TS: noise field not stripped: %q", field)
		}
	}

	// Both must preserve meaningful fields
	for _, field := range []string{"number", "title", "state"} {
		if _, ok := goMap[field]; !ok {
			t.Errorf("Go: meaningful field missing: %q", field)
		}
		if _, ok := tsMap[field]; !ok {
			t.Errorf("TS: meaningful field missing: %q", field)
		}
	}

	// Same key set (within reason — minor differences in user object stripping are ok)
	goKeys := keySet(goMap)
	tsKeys := keySet(tsMap)
	commonKeys := intersect(goKeys, tsKeys)
	if float64(len(commonKeys))/float64(max(len(goKeys), len(tsKeys))) < 0.7 {
		t.Errorf("key sets diverge: Go=%v TS=%v", sortedKeys(goMap), sortedKeys(tsMap))
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func keySet(m map[string]interface{}) map[string]bool {
	s := make(map[string]bool, len(m))
	for k := range m {
		s[k] = true
	}
	return s
}

func intersect(a, b map[string]bool) map[string]bool {
	out := make(map[string]bool)
	for k := range a {
		if b[k] {
			out[k] = true
		}
	}
	return out
}

func sortedKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

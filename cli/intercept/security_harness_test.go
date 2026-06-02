package intercept

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// attackCase mirrors testdata/security/attacks/**/*.json
type attackCase struct {
	ID          string `json:"id"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Payload     string `json:"payload"`
	Risk        string `json:"risk"`
	ShouldCatch bool   `json:"should_catch"`
	SkipUntil   string `json:"skip_until"`
	Notes       string `json:"notes"`
}

// benignCase mirrors testdata/security/benign/**/*.json
type benignCase struct {
	ID      string `json:"id"`
	Tool    string `json:"tool"`
	Content string `json:"content"`
	Notes   string `json:"notes"`
}

func TestSecurityHarness_Attacks(t *testing.T) {
	engine := NewEngine(ContentScanHandlers()...)
	attacks := loadSecurityFixtures[attackCase](t, "attacks")

	var caught, missed int
	for _, a := range attacks {
		a := a
		t.Run(a.ID, func(t *testing.T) {
			if a.SkipUntil != "" {
				t.Skipf("deferred to %s: %s", a.SkipUntil, a.Notes)
			}
			result, err := engine.Handle(mcpPostEvent(a.Payload))
			if err != nil {
				t.Fatalf("handler error: %v", err)
			}
			detected := result.Kind != ResultPassthrough
			if a.ShouldCatch && !detected {
				t.Errorf("missed attack\ncategory: %s\ndesc:     %s\npayload:  %.120s",
					a.Category, a.Description, a.Payload)
				missed++
			} else {
				caught++
			}
		})
	}

	total := caught + missed
	if total > 0 {
		t.Logf("detection: %d/%d caught (%.1f%%)", caught, total, float64(caught)/float64(total)*100)
	}
}

func TestSecurityHarness_FalsePositives(t *testing.T) {
	engine := NewEngine(ContentScanHandlers()...)
	benign := loadSecurityFixtures[benignCase](t, "benign")

	var clean, flagged int
	for _, b := range benign {
		b := b
		t.Run(b.ID, func(t *testing.T) {
			result, err := engine.Handle(mcpPostEvent(b.Content))
			if err != nil {
				t.Fatalf("handler error: %v", err)
			}
			if result.Kind != ResultPassthrough {
				t.Errorf("false positive\ntool: %s\nnotes: %s\ncontent: %.120s",
					b.Tool, b.Notes, b.Content)
				flagged++
			} else {
				clean++
			}
		})
	}

	total := clean + flagged
	if total == 0 {
		t.Fatal("no benign fixtures found")
	}
	rate := float64(flagged) / float64(total) * 100
	t.Logf("false positive rate: %.2f%% (%d/%d flagged)", rate, flagged, total)
	if rate > 1.0 {
		t.Errorf("false positive rate %.2f%% exceeds 1%% threshold", rate)
	}
}

// mcpPostEvent wraps a raw string payload in a tool.post MCP InterceptEvent.
// The payload is passed as the tool output so content-scanning handlers see it.
func mcpPostEvent(payload string) InterceptEvent {
	return InterceptEvent{
		Host:     "claude-code",
		Strategy: "hooks",
		Phase:    PhaseToolPost,
		Session:  Session{ID: "security-harness"},
		Tool: &Tool{
			Name:   "mcp__test__get_content",
			Output: payload,
			IsMCP:  true,
		},
	}
}

// loadSecurityFixtures walks testdata/security/<subdir>/**/*.json and
// unmarshals every file as a JSON array of T.
func loadSecurityFixtures[T any](t *testing.T, subdir string) []T {
	t.Helper()
	root := securityFixturesDir(t)
	dir := filepath.Join(root, subdir)

	var results []T
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var cases []T
		if err := json.Unmarshal(data, &cases); err != nil {
			t.Fatalf("parse error in %s: %v", path, err)
		}
		results = append(results, cases...)
		return nil
	})
	if err != nil {
		t.Fatalf("walking %s: %v", dir, err)
	}
	if len(results) == 0 {
		t.Fatalf("no fixtures found in %s", dir)
	}
	return results
}

func securityFixturesDir(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	// file = cli/intercept/security_harness_test.go → repo root is ../../
	return filepath.Join(filepath.Dir(file), "..", "..", "testdata", "security")
}

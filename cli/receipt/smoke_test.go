package receipt

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"os"
	"testing"
	"time"
)

// TestSmokeFullPipeline is an end-to-end walkthrough of the complete receipt
// lifecycle: key gen → sign → write → daemon restart → chain verify.
//
// Run with:
//
//	go test ./receipt/... -run TestSmokeFullPipeline -v
func TestSmokeFullPipeline(t *testing.T) {
	dir := t.TempDir()

	// ── 1. Generate a signing key ─────────────────────────────────────────────
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	pub := priv.Public().(ed25519.PublicKey)
	pubB64 := base64.RawURLEncoding.EncodeToString(pub)
	t.Logf("Key fingerprint : %s", Fingerprint(pub))

	// ── 2. Open store (daemon startup) ───────────────────────────────────────
	store, err := NewStore(dir, "1.4.2", pubB64)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	sessionID := "smoke-" + NewID()[:8]
	t.Logf("Session         : %s", sessionID)

	// ── 3. session.start ─────────────────────────────────────────────────────
	r0 := Receipt{Body: Body{
		Version:             Version,
		ConfireVersion:      "1.4.2",
		ReceiptID:           NewID(),
		PreviousPayloadHash: store.ChainTip(sessionID),
		Phase:               PhaseSessionStart,
		Timestamp:           time.Now().UTC(),
		Event:               &Event{SessionID: sessionID, Client: "claude_code"},
	}}
	mustSign(t, priv, "device-smoke", &r0)
	mustWrite(t, store, r0)
	t.Logf("session.start   : %s…", r0.PayloadHash[:16])

	// ── 4. tool.pre — bash blocked by policy rule ─────────────────────────────
	ctxID := NewID()
	inputHash, _ := HashAny(map[string]any{"command": "git push --force origin main"})
	r1 := Receipt{Body: Body{
		Version:             Version,
		ConfireVersion:      "1.4.2",
		ReceiptID:           NewID(),
		PreviousPayloadHash: store.ChainTip(sessionID),
		Phase:               PhaseToolPre,
		Timestamp:           time.Now().UTC(),
		Event: &Event{
			SessionID: sessionID,
			Client:    "claude_code",
			ContextID: ctxID,
			ToolName:  "Bash",
			InputHash: inputHash,
		},
		Decision: &Decision{
			Action:     "review",
			RuleID:     "builtin/git.force-push",
			RuleName:   "Destructive git push",
			RuleSource: "builtin",
			PolicyMode: "balanced",
		},
	}}
	mustSign(t, priv, "device-smoke", &r1)
	mustWrite(t, store, r1)
	t.Logf("tool.pre        : %s…  ctx=%s…", r1.PayloadHash[:16], ctxID[:8])

	// ── 5. Simulate daemon restart ────────────────────────────────────────────
	t.Log("--- daemon restart ---")
	store2, err := NewStore(dir, "1.4.2", pubB64)
	if err != nil {
		t.Fatalf("NewStore after restart: %v", err)
	}
	if tip := store2.ChainTip(sessionID); tip != r1.PayloadHash {
		t.Errorf("chain tip not restored: got %s…, want %s…", tip[:16], r1.PayloadHash[:16])
	}
	t.Logf("chain tip restored: %s…", store2.ChainTip(sessionID)[:16])

	// ── 6. tool.post — continue chain after restart ───────────────────────────
	outputHash := HashString("sanitized bash output — secrets already redacted")
	r2 := Receipt{Body: Body{
		Version:             Version,
		ConfireVersion:      "1.4.2",
		ReceiptID:           NewID(),
		PreviousPayloadHash: store2.ChainTip(sessionID),
		Phase:               PhaseToolPost,
		Timestamp:           time.Now().UTC(),
		Event: &Event{
			SessionID:  sessionID,
			Client:     "claude_code",
			ContextID:  ctxID, // same as tool.pre — links the pair
			ToolName:   "Bash",
			InputHash:  inputHash,
			OutputHash: outputHash,
		},
		Provenance: &Provenance{
			TrustLevel:      "trusted_local",
			Flags:           []string{"secret_redacted", "sanitized"},
			RedactionsCount: 1,
		},
	}}
	mustSign(t, priv, "device-smoke", &r2)
	mustWrite(t, store2, r2)
	t.Logf("tool.post       : %s…  ctx=%s…", r2.PayloadHash[:16], ctxID[:8])

	// ── 7. session.end ────────────────────────────────────────────────────────
	r3 := Receipt{Body: Body{
		Version:             Version,
		ConfireVersion:      "1.4.2",
		ReceiptID:           NewID(),
		PreviousPayloadHash: store2.ChainTip(sessionID),
		Phase:               PhaseSessionEnd,
		Timestamp:           time.Now().UTC(),
		Event:               &Event{SessionID: sessionID, Client: "claude_code"},
	}}
	mustSign(t, priv, "device-smoke", &r3)
	mustWrite(t, store2, r3)
	t.Logf("session.end     : %s…", r3.PayloadHash[:16])

	// ── 8. Verify the full chain ──────────────────────────────────────────────
	receipts, err := store2.ReadChain(sessionID)
	if err != nil {
		t.Fatalf("ReadChain: %v", err)
	}
	t.Logf("--- verifying %d receipts ---", len(receipts))

	report := VerifyChain(receipts)
	for _, v := range report.Receipts {
		mark := "✓"
		if !v.Valid {
			mark = "✗"
		}
		t.Logf("  [%d] %s %-22s ctx=%s", v.Index, mark, v.Phase, shortCtx(v.ContextID))
		for _, e := range v.Errors {
			t.Logf("      ERROR: %s", e)
		}
	}
	if !report.ChainIntact() {
		t.Errorf("chain not intact: %v", report.Errors)
	} else {
		t.Logf("chain intact ✓  (%d receipts, %d cloud co-sigs)",
			report.ReceiptCount, report.CloudCoSignCount)
	}

	// ── 9. Pretty-print the tool.pre receipt ─────────────────────────────────
	pretty, _ := json.MarshalIndent(receipts[1], "", "  ")
	t.Logf("sample receipt (tool.pre):\n%s", pretty)

	// ── 10. Show what was written to disk ─────────────────────────────────────
	entries, _ := os.ReadDir(dir)
	t.Log("files on disk:")
	for _, e := range entries {
		info, _ := e.Info()
		t.Logf("  %-35s %d bytes", e.Name(), info.Size())
	}
}

func mustSign(t *testing.T, priv ed25519.PrivateKey, deviceID string, r *Receipt) {
	t.Helper()
	if err := Sign(priv, deviceID, r); err != nil {
		t.Fatalf("Sign: %v", err)
	}
}

func mustWrite(t *testing.T, s *Store, r Receipt) {
	t.Helper()
	if err := s.Write(r); err != nil {
		t.Fatalf("Store.Write: %v", err)
	}
}

func shortCtx(id string) string {
	if id == "" {
		return "(none)    "
	}
	return id[:8] + "…"
}

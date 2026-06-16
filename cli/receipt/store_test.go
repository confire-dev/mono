package receipt

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func testStore(t *testing.T) (*Store, string) {
	t.Helper()
	dir := t.TempDir()
	s, err := NewStore(dir, "1.0.0", "test-pubkey")
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	return s, dir
}

func signedReceipt(t *testing.T, priv ed25519.PrivateKey, phase Phase, sessionID, prevHash string) Receipt {
	t.Helper()
	r := Receipt{
		Body: Body{
			Version:             Version,
			ConfireVersion:      "1.0.0",
			ReceiptID:           NewID(),
			PreviousPayloadHash: prevHash,
			Phase:               phase,
			Timestamp:           time.Now().UTC(),
			Event:               &Event{SessionID: sessionID, Client: "claude_code"},
		},
	}
	if err := Sign(priv, "device-test", &r); err != nil {
		t.Fatalf("Sign: %v", err)
	}
	return r
}

func testKey(t *testing.T) ed25519.PrivateKey {
	t.Helper()
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	return priv
}

// ── NewStore ──────────────────────────────────────────────────────────────────

func TestNewStore_CreatesDirectory(t *testing.T) {
	parent := t.TempDir()
	dir := filepath.Join(parent, "receipts")

	s, err := NewStore(dir, "1.0.0", "pk")
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	_ = s

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected directory, got file")
	}
}

func TestNewStore_DirectoryPermissions(t *testing.T) {
	dir := t.TempDir()
	newDir := filepath.Join(dir, "receipts")

	if _, err := NewStore(newDir, "1.0.0", "pk"); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(newDir)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0700 {
		t.Errorf("receipt dir mode %o, want 0700", info.Mode().Perm())
	}
}

// ── ChainTip ─────────────────────────────────────────────────────────────────

func TestChainTip_ReturnsGenesisHashBeforeAnyWrite(t *testing.T) {
	s, _ := testStore(t)
	tip := s.ChainTip("new-session")
	if tip != GenesisHash {
		t.Errorf("ChainTip before write = %q, want GenesisHash", tip)
	}
}

func TestChainTip_UpdatesAfterWrite(t *testing.T) {
	s, _ := testStore(t)
	priv := testKey(t)

	r := signedReceipt(t, priv, PhaseSessionStart, "sess-A", GenesisHash)
	if err := s.Write(r); err != nil {
		t.Fatal(err)
	}

	tip := s.ChainTip("sess-A")
	if tip != r.PayloadHash {
		t.Errorf("ChainTip after write = %q, want %q", tip, r.PayloadHash)
	}
}

func TestChainTip_IndependentPerSession(t *testing.T) {
	s, _ := testStore(t)
	priv := testKey(t)

	rA := signedReceipt(t, priv, PhaseSessionStart, "sess-A", GenesisHash)
	rB := signedReceipt(t, priv, PhaseSessionStart, "sess-B", GenesisHash)
	s.Write(rA) //nolint:errcheck
	s.Write(rB) //nolint:errcheck

	if s.ChainTip("sess-A") != rA.PayloadHash {
		t.Error("chain tip for sess-A is wrong")
	}
	if s.ChainTip("sess-B") != rB.PayloadHash {
		t.Error("chain tip for sess-B is wrong")
	}
	if s.ChainTip("sess-A") == s.ChainTip("sess-B") {
		t.Error("different sessions should have different chain tips")
	}
}

// ── Write ─────────────────────────────────────────────────────────────────────

func TestWrite_CreatesJSONLFile(t *testing.T) {
	s, dir := testStore(t)
	priv := testKey(t)

	r := signedReceipt(t, priv, PhaseSessionStart, "sess-1", GenesisHash)
	if err := s.Write(r); err != nil {
		t.Fatalf("Write: %v", err)
	}

	path := filepath.Join(dir, "sess-1.jsonl")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("JSONL file not created: %v", err)
	}
}

func TestWrite_JSONLContainsValidReceipt(t *testing.T) {
	s, dir := testStore(t)
	priv := testKey(t)

	r := signedReceipt(t, priv, PhaseSessionStart, "sess-1", GenesisHash)
	if err := s.Write(r); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "sess-1.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	var got Receipt
	if err := json.Unmarshal(data[:len(data)-1], &got); err != nil { // strip trailing newline
		t.Fatalf("invalid JSON in JSONL: %v", err)
	}
	if got.ReceiptID != r.ReceiptID {
		t.Errorf("receipt_id = %q, want %q", got.ReceiptID, r.ReceiptID)
	}
}

func TestWrite_AppendsMultipleReceipts(t *testing.T) {
	s, dir := testStore(t)
	priv := testKey(t)

	r1 := signedReceipt(t, priv, PhaseSessionStart, "sess-1", GenesisHash)
	r2 := signedReceipt(t, priv, PhaseToolPre, "sess-1", r1.PayloadHash)
	s.Write(r1) //nolint:errcheck
	s.Write(r2) //nolint:errcheck

	receipts, err := ReadChainFromFile(filepath.Join(dir, "sess-1.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if len(receipts) != 2 {
		t.Errorf("expected 2 receipts in JSONL, got %d", len(receipts))
	}
}

func TestWrite_MetaFileCreated(t *testing.T) {
	s, dir := testStore(t)
	priv := testKey(t)

	r := signedReceipt(t, priv, PhaseSessionStart, "sess-meta", GenesisHash)
	if err := s.Write(r); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, "sess-meta.meta")); err != nil {
		t.Fatalf("meta file not created: %v", err)
	}
}

func TestWrite_MetaContainsCorrectChainTip(t *testing.T) {
	s, dir := testStore(t)
	priv := testKey(t)

	r := signedReceipt(t, priv, PhaseSessionStart, "sess-1", GenesisHash)
	s.Write(r) //nolint:errcheck

	data, err := os.ReadFile(filepath.Join(dir, "sess-1.meta"))
	if err != nil {
		t.Fatal(err)
	}
	var m Meta
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("invalid meta JSON: %v", err)
	}
	if m.ChainTip != r.PayloadHash {
		t.Errorf("meta chain_tip = %q, want %q", m.ChainTip, r.PayloadHash)
	}
}

// ── Daemon restart resilience ─────────────────────────────────────────────────

func TestDaemonRestartResumesChainTip(t *testing.T) {
	dir := t.TempDir()
	priv := testKey(t)

	// First daemon instance.
	s1, err := NewStore(dir, "1.0.0", "pk")
	if err != nil {
		t.Fatal(err)
	}
	r1 := signedReceipt(t, priv, PhaseSessionStart, "sess-restart", GenesisHash)
	r2 := signedReceipt(t, priv, PhaseToolPre, "sess-restart", r1.PayloadHash)
	s1.Write(r1) //nolint:errcheck
	s1.Write(r2) //nolint:errcheck

	// Simulate restart: open a new Store on the same directory.
	s2, err := NewStore(dir, "1.0.0", "pk")
	if err != nil {
		t.Fatal(err)
	}
	restoredTip := s2.ChainTip("sess-restart")
	if restoredTip != r2.PayloadHash {
		t.Errorf("chain tip not restored after restart: got %q, want %q",
			restoredTip, r2.PayloadHash)
	}
}

func TestDaemonRestartChainRemainsValid(t *testing.T) {
	dir := t.TempDir()
	priv := testKey(t)

	s1, _ := NewStore(dir, "1.0.0", "pk")
	r1 := signedReceipt(t, priv, PhaseSessionStart, "sess-valid", GenesisHash)
	r2 := signedReceipt(t, priv, PhaseToolPre, "sess-valid", r1.PayloadHash)
	s1.Write(r1) //nolint:errcheck
	s1.Write(r2) //nolint:errcheck

	// New Store instance, write a third receipt.
	s2, _ := NewStore(dir, "1.0.0", "pk")
	r3 := signedReceipt(t, priv, PhaseToolPost, "sess-valid", s2.ChainTip("sess-valid"))
	s2.Write(r3) //nolint:errcheck

	// Verify the full 3-receipt chain is intact.
	receipts, err := s2.ReadChain("sess-valid")
	if err != nil {
		t.Fatal(err)
	}
	if len(receipts) != 3 {
		t.Fatalf("expected 3 receipts, got %d", len(receipts))
	}
	report := VerifyChain(receipts)
	if !report.ChainIntact() {
		t.Errorf("chain broken after daemon restart: %v", report.Errors)
	}
}

// ── Bundle receipts ───────────────────────────────────────────────────────────

func TestWrite_BundleReceiptGoesToDaemonPolicyJSONL(t *testing.T) {
	s, dir := testStore(t)
	priv := testKey(t)

	r := Receipt{
		Body: Body{
			Version:             Version,
			ConfireVersion:      "1.0.0",
			ReceiptID:           NewID(),
			PreviousPayloadHash: GenesisHash,
			Phase:               PhasePolicyBundleLoaded,
			Timestamp:           time.Now().UTC(),
			Event:               &Event{SessionID: "_daemon_policy_", Client: "daemon"},
			Bundle: &Bundle{
				BundleVersion:      "v1",
				FetchedAt:          time.Now().UTC(),
				RuleCount:          12,
				VerificationResult: VerificationNoSignature,
			},
		},
	}
	if err := Sign(priv, "device-test", &r); err != nil {
		t.Fatal(err)
	}
	if err := s.Write(r); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, "daemon-policy.jsonl")); err != nil {
		t.Fatalf("daemon-policy.jsonl not created: %v", err)
	}
}

func TestChainTip_BundleKeyIsIndependentFromSessions(t *testing.T) {
	s, _ := testStore(t)
	priv := testKey(t)

	bundleReceipt := Receipt{
		Body: Body{
			Version:             Version,
			ConfireVersion:      "1.0.0",
			ReceiptID:           NewID(),
			PreviousPayloadHash: GenesisHash,
			Phase:               PhasePolicyBundleLoaded,
			Timestamp:           time.Now().UTC(),
			Event:               &Event{SessionID: BundleSessionKey, Client: "daemon"},
			Bundle: &Bundle{
				BundleVersion:      "v1",
				FetchedAt:          time.Now().UTC(),
				RuleCount:          5,
				VerificationResult: VerificationNoSignature,
			},
		},
	}
	Sign(priv, "device-test", &bundleReceipt) //nolint:errcheck
	s.Write(bundleReceipt)                    //nolint:errcheck

	// Normal session should still return GenesisHash.
	if tip := s.ChainTip("some-session"); tip != GenesisHash {
		t.Errorf("session chain tip affected by bundle write: got %q, want GenesisHash", tip)
	}
	// Bundle chain tip should reflect the bundle receipt.
	if tip := s.ChainTip(BundleSessionKey); tip != bundleReceipt.PayloadHash {
		t.Errorf("bundle chain tip = %q, want %q", tip, bundleReceipt.PayloadHash)
	}
}

// ── ReadChainFromFile ─────────────────────────────────────────────────────────

func TestReadChainFromFile_EmptyFileReturnsNil(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.jsonl")
	os.WriteFile(path, []byte(""), 0600) //nolint:errcheck

	receipts, err := ReadChainFromFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(receipts) != 0 {
		t.Errorf("expected 0 receipts, got %d", len(receipts))
	}
}

func TestReadChainFromFile_NonExistentFileReturnsError(t *testing.T) {
	_, err := ReadChainFromFile("/tmp/does-not-exist-confire-test.jsonl")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

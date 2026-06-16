package receipt

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
	"time"
)

// genKey generates a fresh Ed25519 keypair for testing.
func genKey(t *testing.T) ed25519.PrivateKey {
	t.Helper()
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	return priv
}

// minimalBody returns a valid Body for testing with all required fields set.
func minimalBody(phase Phase) Body {
	return Body{
		Version:             Version,
		ConfireVersion:      "1.4.2",
		ReceiptID:           NewID(),
		PreviousPayloadHash: GenesisHash,
		Phase:               phase,
		Timestamp:           time.Now().UTC(),
		Event: &Event{
			SessionID: "test-session",
			Client:    "claude_code",
		},
	}
}

func TestNewID_IsUUIDv4Format(t *testing.T) {
	id := NewID()
	// UUID v4: xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx
	parts := strings.Split(id, "-")
	if len(parts) != 5 {
		t.Fatalf("NewID %q: expected 5 dash-separated parts, got %d", id, len(parts))
	}
	if len(parts[0]) != 8 || len(parts[1]) != 4 || len(parts[2]) != 4 ||
		len(parts[3]) != 4 || len(parts[4]) != 12 {
		t.Errorf("NewID %q: unexpected part lengths", id)
	}
	// Version nibble must be '4'.
	if parts[2][0] != '4' {
		t.Errorf("NewID %q: version nibble is %c, want '4'", id, parts[2][0])
	}
	// Variant nibble must be 8, 9, a, or b.
	v := parts[3][0]
	if v != '8' && v != '9' && v != 'a' && v != 'b' {
		t.Errorf("NewID %q: variant nibble is %c, want 8/9/a/b", id, v)
	}
}

func TestNewID_UniqueEachCall(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id := NewID()
		if seen[id] {
			t.Fatalf("NewID produced duplicate after %d calls: %s", i, id)
		}
		seen[id] = true
	}
}

func TestSign_PopulatesPayloadHash(t *testing.T) {
	priv := genKey(t)
	r := Receipt{Body: minimalBody(PhaseSessionStart)}

	if err := Sign(priv, "device-1", &r); err != nil {
		t.Fatalf("Sign: %v", err)
	}
	if r.PayloadHash == "" {
		t.Error("PayloadHash is empty after Sign")
	}
	if len(r.PayloadHash) != 64 {
		t.Errorf("PayloadHash length %d, want 64 hex chars", len(r.PayloadHash))
	}
}

func TestSign_PayloadHashMatchesCanonicalBodySHA256(t *testing.T) {
	priv := genKey(t)
	r := Receipt{Body: minimalBody(PhaseToolPre)}

	if err := Sign(priv, "device-1", &r); err != nil {
		t.Fatal(err)
	}

	// Recompute independently.
	body, err := CanonicalBody(r)
	if err != nil {
		t.Fatal(err)
	}
	h := sha256.Sum256(body)
	want := hex.EncodeToString(h[:])

	if r.PayloadHash != want {
		t.Errorf("PayloadHash = %s, want %s", r.PayloadHash, want)
	}
}

func TestSign_LocalSignerIsAtIndex0(t *testing.T) {
	priv := genKey(t)
	r := Receipt{Body: minimalBody(PhaseToolPost)}

	if err := Sign(priv, "device-42", &r); err != nil {
		t.Fatal(err)
	}
	if len(r.Signers) != 1 {
		t.Fatalf("expected 1 signer, got %d", len(r.Signers))
	}
	if r.Signers[0].SignerID != "local:device-42" {
		t.Errorf("signers[0].signer_id = %q, want %q", r.Signers[0].SignerID, "local:device-42")
	}
	if r.Signers[0].Algorithm != "Ed25519" {
		t.Errorf("signers[0].algorithm = %q, want Ed25519", r.Signers[0].Algorithm)
	}
}

func TestSign_SignatureIsVerifiable(t *testing.T) {
	priv := genKey(t)
	r := Receipt{Body: minimalBody(PhaseToolPre)}

	if err := Sign(priv, "device-1", &r); err != nil {
		t.Fatal(err)
	}
	if errs := VerifyReceipt(r); len(errs) != 0 {
		t.Errorf("VerifyReceipt after Sign: %v", errs)
	}
}

func TestSign_LocalSignerPrependedBeforeExistingSigners(t *testing.T) {
	// If Signers already has an entry (simulating a future Phase-2 remote signer
	// that was pre-populated), Sign must insert the local signer at index 0.
	priv := genKey(t)
	r := Receipt{Body: minimalBody(PhaseToolPre)}
	r.Signers = []Signer{{SignerID: "remote:confire-cloud"}} // pre-populated

	if err := Sign(priv, "device-1", &r); err != nil {
		t.Fatal(err)
	}
	if r.Signers[0].SignerID != "local:device-1" {
		t.Errorf("signers[0] = %q, want local signer first", r.Signers[0].SignerID)
	}
}

func TestSign_MutatingBodyInvalidatesSignature(t *testing.T) {
	priv := genKey(t)
	r := Receipt{Body: minimalBody(PhaseToolPre)}

	if err := Sign(priv, "device-1", &r); err != nil {
		t.Fatal(err)
	}

	// Tamper with the receipt phase after signing.
	r.Phase = PhaseSessionEnd

	errs := VerifyReceipt(r)
	if len(errs) == 0 {
		t.Error("expected verification error after body mutation, got none")
	}
}

func TestHashAny_NilReturnsEmpty(t *testing.T) {
	h, err := HashAny(nil)
	if err != nil {
		t.Fatal(err)
	}
	if h != "" {
		t.Errorf("HashAny(nil) = %q, want empty string", h)
	}
}

func TestHashAny_ConsistentForSameInput(t *testing.T) {
	input := map[string]any{"command": "git push --force", "cwd": "/tmp"}
	h1, err := HashAny(input)
	if err != nil {
		t.Fatal(err)
	}
	h2, err := HashAny(input)
	if err != nil {
		t.Fatal(err)
	}
	if h1 != h2 {
		t.Errorf("HashAny returned different hashes for same input: %q vs %q", h1, h2)
	}
}

func TestHashAny_DifferentInputsDifferentHashes(t *testing.T) {
	h1, _ := HashAny(map[string]any{"a": 1})
	h2, _ := HashAny(map[string]any{"a": 2})
	if h1 == h2 {
		t.Error("HashAny returned same hash for different inputs")
	}
}

func TestHashString_EmptyReturnsEmpty(t *testing.T) {
	if h := HashString(""); h != "" {
		t.Errorf("HashString(\"\") = %q, want empty", h)
	}
}

func TestHashString_Is64HexChars(t *testing.T) {
	h := HashString("/Users/efe/dev/mono")
	if len(h) != 64 {
		t.Errorf("HashString length %d, want 64", len(h))
	}
}

func TestHashString_Deterministic(t *testing.T) {
	s := "/some/working/directory"
	if HashString(s) != HashString(s) {
		t.Error("HashString is not deterministic")
	}
}

func TestCanonicalBody_ExcludesPayloadHashAndSigners(t *testing.T) {
	priv := genKey(t)
	r := Receipt{Body: minimalBody(PhaseToolPre)}

	if err := Sign(priv, "device-1", &r); err != nil {
		t.Fatal(err)
	}

	body, err := CanonicalBody(r)
	if err != nil {
		t.Fatal(err)
	}
	bodyStr := string(body)

	// Check for the exact JSON key (quoted) so we don't match "previous_payload_hash".
	if strings.Contains(bodyStr, `"payload_hash":`) {
		t.Error("canonical body must not contain the payload_hash key")
	}
	if strings.Contains(bodyStr, `"signers":`) {
		t.Error("canonical body must not contain the signers key")
	}
	if strings.Contains(bodyStr, `"signature":`) {
		t.Error("canonical body must not contain the signature key")
	}
	// previous_payload_hash is EXPECTED to be present.
	if !strings.Contains(bodyStr, `"previous_payload_hash":`) {
		t.Error("canonical body must contain previous_payload_hash")
	}
}

func TestChainLinkage_PreviousPayloadHash(t *testing.T) {
	priv := genKey(t)

	r1 := Receipt{Body: minimalBody(PhaseSessionStart)}
	if err := Sign(priv, "device-1", &r1); err != nil {
		t.Fatal(err)
	}

	// Second receipt must chain to first.
	r2 := Receipt{Body: minimalBody(PhaseToolPre)}
	r2.PreviousPayloadHash = r1.PayloadHash
	if err := Sign(priv, "device-1", &r2); err != nil {
		t.Fatal(err)
	}

	if r2.PreviousPayloadHash != r1.PayloadHash {
		t.Errorf("chain broken: r2.previous_payload_hash = %s, want %s",
			r2.PreviousPayloadHash, r1.PayloadHash)
	}
}

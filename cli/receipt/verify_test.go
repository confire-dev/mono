package receipt

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"testing"
	"time"
)

// buildChain creates n signed receipts forming a valid chain using the given key.
func buildChain(t *testing.T, priv ed25519.PrivateKey, n int, sessionID string) []Receipt {
	t.Helper()
	phases := []Phase{
		PhaseSessionStart,
		PhaseToolPre,
		PhaseToolPost,
		PhaseSessionEnd,
	}
	var receipts []Receipt
	prevHash := GenesisHash
	for i := 0; i < n; i++ {
		phase := phases[i%len(phases)]
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
		if err := Sign(priv, "device-1", &r); err != nil {
			t.Fatalf("receipt %d: Sign: %v", i, err)
		}
		prevHash = r.PayloadHash
		receipts = append(receipts, r)
	}
	return receipts
}

// ── Happy paths ───────────────────────────────────────────────────────────────

func TestVerifyChain_SingleReceiptValid(t *testing.T) {
	_, priv, _ := ed25519.GenerateKey(rand.Reader)
	receipts := buildChain(t, priv, 1, "sess-1")
	report := VerifyChain(receipts)
	if !report.ChainIntact() {
		t.Errorf("single receipt chain not intact: %v", report.Errors)
	}
}

func TestVerifyChain_MultipleReceiptsValid(t *testing.T) {
	_, priv, _ := ed25519.GenerateKey(rand.Reader)
	receipts := buildChain(t, priv, 8, "sess-multi")
	report := VerifyChain(receipts)
	if !report.ChainIntact() {
		t.Errorf("8-receipt chain not intact: %v", report.Errors)
	}
	if report.ReceiptCount != 8 {
		t.Errorf("ReceiptCount = %d, want 8", report.ReceiptCount)
	}
	if report.BrokenAt != -1 {
		t.Errorf("BrokenAt = %d, want -1", report.BrokenAt)
	}
}

func TestVerifyChain_TimestampsTracked(t *testing.T) {
	_, priv, _ := ed25519.GenerateKey(rand.Reader)
	receipts := buildChain(t, priv, 3, "sess-ts")
	report := VerifyChain(receipts)

	if report.EarliestTimestamp.IsZero() {
		t.Error("EarliestTimestamp should not be zero for a valid chain")
	}
	if report.LatestTimestamp.IsZero() {
		t.Error("LatestTimestamp should not be zero for a valid chain")
	}
	if report.LatestTimestamp.Before(report.EarliestTimestamp) {
		t.Error("LatestTimestamp before EarliestTimestamp")
	}
}

func TestVerifyChain_FirstReceiptUsesGenesisHash(t *testing.T) {
	_, priv, _ := ed25519.GenerateKey(rand.Reader)
	receipts := buildChain(t, priv, 2, "sess-genesis")
	if receipts[0].PreviousPayloadHash != GenesisHash {
		t.Errorf("first receipt previous_payload_hash = %q, want GenesisHash",
			receipts[0].PreviousPayloadHash)
	}
}

// ── Empty / degenerate ────────────────────────────────────────────────────────

func TestVerifyChain_EmptySlice(t *testing.T) {
	report := VerifyChain(nil)
	if report.ChainIntact() {
		t.Error("empty chain should not be intact")
	}
	if len(report.Errors) == 0 {
		t.Error("expected errors for empty chain")
	}
}

// ── Tampered receipts ─────────────────────────────────────────────────────────

func TestVerifyChain_TamperedPayloadHash(t *testing.T) {
	_, priv, _ := ed25519.GenerateKey(rand.Reader)
	receipts := buildChain(t, priv, 3, "sess-tamper-hash")

	// Tamper with the stored payload_hash of receipt 1 (not the body, just the field).
	receipts[1].PayloadHash = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

	report := VerifyChain(receipts)
	if report.ChainIntact() {
		t.Error("expected chain to be broken after payload_hash tamper")
	}
	if report.BrokenAt != 1 {
		t.Errorf("BrokenAt = %d, want 1", report.BrokenAt)
	}
}

func TestVerifyChain_TamperedBody(t *testing.T) {
	_, priv, _ := ed25519.GenerateKey(rand.Reader)
	receipts := buildChain(t, priv, 3, "sess-tamper-body")

	// Mutate the body of receipt 2 — this invalidates both payload_hash and signature.
	receipts[2].Phase = PhaseSessionEnd

	report := VerifyChain(receipts)
	if report.ChainIntact() {
		t.Error("expected chain to be broken after body mutation")
	}
	found := false
	for _, v := range report.Receipts {
		if v.Index == 2 && !v.Valid {
			found = true
			break
		}
	}
	if !found {
		t.Error("receipt at index 2 should be marked invalid")
	}
}

func TestVerifyChain_BrokenChainLink(t *testing.T) {
	_, priv, _ := ed25519.GenerateKey(rand.Reader)
	receipts := buildChain(t, priv, 4, "sess-broken-link")

	// Sever the chain link at receipt 2 by forging a wrong previous_payload_hash.
	// We must also re-sign so the signature still validates (testing chain-only breakage).
	receipts[2].PreviousPayloadHash = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	// Clear existing signers and re-sign so signature is valid but chain link is wrong.
	receipts[2].Signers = nil
	Sign(priv, "device-1", &receipts[2]) //nolint:errcheck

	report := VerifyChain(receipts)
	if report.ChainIntact() {
		t.Error("expected broken chain due to wrong previous_payload_hash")
	}
	if report.BrokenAt != 2 {
		t.Errorf("BrokenAt = %d, want 2", report.BrokenAt)
	}
}

// ── Signature errors ──────────────────────────────────────────────────────────

func TestVerifyChain_WrongSignature(t *testing.T) {
	_, priv, _ := ed25519.GenerateKey(rand.Reader)
	receipts := buildChain(t, priv, 2, "sess-wrong-sig")

	// Replace the signature bytes with garbage (keep same length).
	garbled := make([]byte, 64)
	rand.Read(garbled) //nolint:errcheck
	receipts[1].Signers[0].Signature = base64.RawURLEncoding.EncodeToString(garbled)

	report := VerifyChain(receipts)
	if report.ChainIntact() {
		t.Error("expected verification failure for wrong signature")
	}
}

func TestVerifyChain_WrongPublicKey(t *testing.T) {
	_, priv, _ := ed25519.GenerateKey(rand.Reader)
	receipts := buildChain(t, priv, 2, "sess-wrong-pubkey")

	// Replace the public key with a random 32-byte value.
	fakeKey := make([]byte, 32)
	rand.Read(fakeKey) //nolint:errcheck
	receipts[1].Signers[0].Pubkey = base64.RawURLEncoding.EncodeToString(fakeKey)

	report := VerifyChain(receipts)
	if report.ChainIntact() {
		t.Error("expected verification failure for wrong public key")
	}
}

func TestVerifyChain_InvalidBase64Signature(t *testing.T) {
	_, priv, _ := ed25519.GenerateKey(rand.Reader)
	receipts := buildChain(t, priv, 1, "sess-bad-b64")
	receipts[0].Signers[0].Signature = "not-valid-base64!!!"

	report := VerifyChain(receipts)
	if report.ChainIntact() {
		t.Error("expected failure for invalid base64 signature")
	}
}

// ── Signer order ─────────────────────────────────────────────────────────────

func TestVerifyChain_RemoteSignerAtIndex0Rejected(t *testing.T) {
	_, priv, _ := ed25519.GenerateKey(rand.Reader)
	receipts := buildChain(t, priv, 1, "sess-order")

	// Simulate a receipt where remote signer was (incorrectly) placed first.
	receipts[0].Signers = append(
		[]Signer{{
			SignerID:  "remote:confire-cloud",
			Pubkey:   receipts[0].Signers[0].Pubkey,
			Algorithm: "Ed25519",
			Signature: receipts[0].Signers[0].Signature,
		}},
		receipts[0].Signers...,
	)

	report := VerifyChain(receipts)
	if report.ChainIntact() {
		t.Error("expected signer order violation to break the chain")
	}
	found := false
	for _, v := range report.Receipts {
		for _, e := range v.Errors {
			if containsStr(e, "signer order") {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected signer order violation error in receipt verdicts")
	}
}

// ── Cloud co-signature counting ───────────────────────────────────────────────

func TestVerifyChain_CountsCloudCoSignatures(t *testing.T) {
	_, priv, _ := ed25519.GenerateKey(rand.Reader)
	receipts := buildChain(t, priv, 3, "sess-cosign")

	// Add a valid remote co-signer to receipt 1 (simulating Phase 2 co-signing).
	// We use the same key for simplicity — in production this would be the cloud key.
	_, remotePriv, _ := ed25519.GenerateKey(rand.Reader)
	remotePub := remotePriv.Public().(ed25519.PublicKey)

	body, _ := CanonicalBody(receipts[1])
	remoteSig := ed25519.Sign(remotePriv, body)
	receipts[1].Signers = append(receipts[1].Signers, Signer{
		SignerID:  "remote:confire-cloud",
		Pubkey:   base64.RawURLEncoding.EncodeToString(remotePub),
		Algorithm: "Ed25519",
		Signature: base64.RawURLEncoding.EncodeToString(remoteSig),
	})

	report := VerifyChain(receipts)
	if !report.ChainIntact() {
		t.Errorf("chain with cloud co-signer should be intact: %v", report.Errors)
	}
	if report.CloudCoSignCount != 1 {
		t.Errorf("CloudCoSignCount = %d, want 1", report.CloudCoSignCount)
	}
}

// ── Timestamp gap warning ─────────────────────────────────────────────────────

func TestVerifyChain_TimestampGapWarning(t *testing.T) {
	_, priv, _ := ed25519.GenerateKey(rand.Reader)
	receipts := buildChain(t, priv, 2, "sess-gap")

	// Backdate the first receipt so there's a >5 minute gap.
	receipts[0].Timestamp = time.Now().UTC().Add(-10 * time.Minute)
	// Re-sign receipt 0 since we changed the body.
	receipts[0].Signers = nil
	receipts[0].PayloadHash = ""
	Sign(priv, "device-1", &receipts[0]) //nolint:errcheck

	// Fix the chain link on receipt 1 to point to the re-signed receipt 0.
	receipts[1].PreviousPayloadHash = receipts[0].PayloadHash
	receipts[1].Signers = nil
	receipts[1].PayloadHash = ""
	Sign(priv, "device-1", &receipts[1]) //nolint:errcheck

	report := VerifyChain(receipts)
	if len(report.Warnings) == 0 {
		t.Error("expected timestamp gap warning, got none")
	}
	found := false
	for _, w := range report.Warnings {
		if containsStr(w, "timestamp gap") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'timestamp gap' warning, got: %v", report.Warnings)
	}
}

// ── VerifyReceipt (isolated) ──────────────────────────────────────────────────

func TestVerifyReceipt_ValidReceipt(t *testing.T) {
	_, priv, _ := ed25519.GenerateKey(rand.Reader)
	r := Receipt{
		Body: Body{
			Version:             Version,
			ConfireVersion:      "1.0.0",
			ReceiptID:           NewID(),
			PreviousPayloadHash: GenesisHash,
			Phase:               PhaseToolPre,
			Timestamp:           time.Now().UTC(),
			Event:               &Event{SessionID: "s1", Client: "claude_code"},
		},
	}
	Sign(priv, "device-1", &r) //nolint:errcheck

	if errs := VerifyReceipt(r); len(errs) != 0 {
		t.Errorf("VerifyReceipt returned errors for valid receipt: %v", errs)
	}
}

func TestVerifyReceipt_TamperedBody(t *testing.T) {
	_, priv, _ := ed25519.GenerateKey(rand.Reader)
	r := Receipt{
		Body: Body{
			Version:             Version,
			ConfireVersion:      "1.0.0",
			ReceiptID:           NewID(),
			PreviousPayloadHash: GenesisHash,
			Phase:               PhaseToolPre,
			Timestamp:           time.Now().UTC(),
			Event:               &Event{SessionID: "s1", Client: "claude_code"},
		},
	}
	Sign(priv, "device-1", &r) //nolint:errcheck

	r.Phase = PhaseSessionEnd // mutate body post-signing
	if errs := VerifyReceipt(r); len(errs) == 0 {
		t.Error("expected errors after body mutation, got none")
	}
}

func TestVerifyReceipt_RemoteSignerAtIndex0(t *testing.T) {
	_, priv, _ := ed25519.GenerateKey(rand.Reader)
	r := Receipt{
		Body: Body{
			Version:             Version,
			ConfireVersion:      "1.0.0",
			ReceiptID:           NewID(),
			PreviousPayloadHash: GenesisHash,
			Phase:               PhaseToolPre,
			Timestamp:           time.Now().UTC(),
			Event:               &Event{SessionID: "s1", Client: "claude_code"},
		},
	}
	Sign(priv, "device-1", &r) //nolint:errcheck

	// Prepend a remote signer (violates canonical order).
	r.Signers = append([]Signer{{
		SignerID:  "remote:confire-cloud",
		Pubkey:   r.Signers[0].Pubkey,
		Algorithm: "Ed25519",
		Signature: r.Signers[0].Signature,
	}}, r.Signers...)

	errs := VerifyReceipt(r)
	found := false
	for _, e := range errs {
		if containsStr(e, "signer order") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected signer order violation error, got: %v", errs)
	}
}

package receipt

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// ChainReport is the result of verifying an ordered sequence of receipts.
type ChainReport struct {
	ReceiptCount      int
	CloudCoSignCount  int    // number of receipts that carry a valid remote co-signature
	BrokenAt         int    // index of first invalid receipt; -1 means chain is intact
	Receipts         []ReceiptVerdict
	EarliestTimestamp time.Time
	LatestTimestamp   time.Time
	Errors           []string // chain-level errors
	Warnings         []string // non-fatal observations
}

// ChainIntact reports whether the chain is complete and all receipts are valid.
func (r ChainReport) ChainIntact() bool {
	return r.BrokenAt == -1 && len(r.Errors) == 0
}

// ReceiptVerdict holds the per-receipt result within a ChainReport.
type ReceiptVerdict struct {
	Index     int
	ReceiptID string
	Phase     Phase
	ContextID string
	Valid     bool
	Errors    []string
}

// VerifyChain verifies an ordered slice of receipts as a chain.
//
// For each receipt it checks:
//   - payload_hash = hex(SHA-256(Canonical(body)))
//   - previous_payload_hash == payload_hash of the prior receipt (or GenesisHash for receipt 0)
//   - all Ed25519 signatures are valid
//   - local signer is at index 0 (remote signer must not appear before local signer)
func VerifyChain(receipts []Receipt) ChainReport {
	report := ChainReport{
		ReceiptCount: len(receipts),
		BrokenAt:     -1,
	}
	if len(receipts) == 0 {
		report.Errors = append(report.Errors, "chain is empty")
		return report
	}

	expectedPrev := GenesisHash

	for i, r := range receipts {
		verdict := ReceiptVerdict{
			Index:     i,
			ReceiptID: r.ReceiptID,
			Phase:     r.Phase,
			Valid:     true,
		}
		if r.Event != nil {
			verdict.ContextID = r.Event.ContextID
		}

		body, err := CanonicalBody(r)
		if err != nil {
			verdict.Valid = false
			verdict.Errors = append(verdict.Errors, fmt.Sprintf("canonical JSON: %v", err))
			markBroken(&report, i)
			report.Receipts = append(report.Receipts, verdict)
			break
		}

		// Check payload_hash.
		h := sha256.Sum256(body)
		computed := hex.EncodeToString(h[:])
		if computed != r.PayloadHash {
			verdict.Valid = false
			verdict.Errors = append(verdict.Errors, fmt.Sprintf(
				"payload_hash mismatch: computed %s, stored %s", computed, r.PayloadHash))
			markBroken(&report, i)
		}

		// Check chain linkage.
		if r.PreviousPayloadHash != expectedPrev {
			verdict.Valid = false
			verdict.Errors = append(verdict.Errors, fmt.Sprintf(
				"previous_payload_hash mismatch: expected %s, got %s",
				expectedPrev, r.PreviousPayloadHash))
			markBroken(&report, i)
		}

		// Enforce signer order: local signer must be at index 0.
		for j, sig := range r.Signers {
			if j == 0 && strings.HasPrefix(sig.SignerID, "remote:") {
				verdict.Valid = false
				verdict.Errors = append(verdict.Errors,
					"signer order violation: remote signer at index 0 (local signer must be first)")
				markBroken(&report, i)
			}
		}

		// Verify each Ed25519 signature over the canonical body.
		for _, sig := range r.Signers {
			if err := verifySig(sig, body); err != nil {
				verdict.Valid = false
				verdict.Errors = append(verdict.Errors, fmt.Sprintf("signer %s: %v", sig.SignerID, err))
				markBroken(&report, i)
			}
			if strings.HasPrefix(sig.SignerID, "remote:") {
				report.CloudCoSignCount++
			}
		}

		// Timestamp gap warning (> 5 minutes between consecutive receipts).
		if i > 0 && !receipts[i-1].Timestamp.IsZero() && !r.Timestamp.IsZero() {
			gap := r.Timestamp.Sub(receipts[i-1].Timestamp)
			if gap > 5*time.Minute {
				report.Warnings = append(report.Warnings, fmt.Sprintf(
					"receipt %d: timestamp gap of %s (possible clock skew or daemon restart)",
					i, gap.Round(time.Second)))
			}
		}

		// Track timestamp range.
		if verdict.Valid {
			if report.EarliestTimestamp.IsZero() {
				report.EarliestTimestamp = r.Timestamp
			}
			report.LatestTimestamp = r.Timestamp
		}

		expectedPrev = r.PayloadHash
		report.Receipts = append(report.Receipts, verdict)
	}

	return report
}

// VerifyReceipt checks a single receipt in isolation — no chain context.
// It verifies payload_hash consistency and all Ed25519 signatures, but does
// not check previous_payload_hash linkage.
// Returns a slice of error strings; empty means the receipt is valid.
func VerifyReceipt(r Receipt) []string {
	var errs []string

	body, err := CanonicalBody(r)
	if err != nil {
		return []string{fmt.Sprintf("canonical JSON: %v", err)}
	}

	h := sha256.Sum256(body)
	computed := hex.EncodeToString(h[:])
	if computed != r.PayloadHash {
		errs = append(errs, fmt.Sprintf("payload_hash mismatch: computed %s, stored %s",
			computed, r.PayloadHash))
	}

	for j, sig := range r.Signers {
		if j == 0 && strings.HasPrefix(sig.SignerID, "remote:") {
			errs = append(errs, "signer order violation: remote signer at index 0")
		}
		if err := verifySig(sig, body); err != nil {
			errs = append(errs, fmt.Sprintf("signer %s: %v", sig.SignerID, err))
		}
	}
	return errs
}

// ── internal ──────────────────────────────────────────────────────────────────

func verifySig(sig Signer, body []byte) error {
	pub, err := base64.RawURLEncoding.DecodeString(sig.Pubkey)
	if err != nil {
		return fmt.Errorf("invalid pubkey base64: %w", err)
	}
	if len(pub) != ed25519.PublicKeySize {
		return fmt.Errorf("invalid pubkey length %d (want %d)", len(pub), ed25519.PublicKeySize)
	}
	rawSig, err := base64.RawURLEncoding.DecodeString(sig.Signature)
	if err != nil {
		return fmt.Errorf("invalid signature base64: %w", err)
	}
	if !ed25519.Verify(pub, body, rawSig) {
		return fmt.Errorf("signature verification failed")
	}
	return nil
}

func markBroken(r *ChainReport, i int) {
	if r.BrokenAt == -1 {
		r.BrokenAt = i
	}
}

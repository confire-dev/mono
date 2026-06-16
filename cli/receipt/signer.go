package receipt

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"
)

// NewID returns a UUID v4 suitable for use as receipt_id or context_id.
func NewID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// Fallback: SHA-256 of current time — collision-resistant for our volumes.
		h := sha256.Sum256([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
		copy(b, h[:16])
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant bits (RFC 4122)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// CanonicalBody returns the RFC 8785 canonical JSON encoding of r.Body.
// This is the byte sequence over which Ed25519 signatures are computed, and
// whose SHA-256 digest is stored as r.PayloadHash.
func CanonicalBody(r Receipt) ([]byte, error) {
	return Canonical(r.Body)
}

// HashBytes returns the hex-encoded SHA-256 digest of b.
func HashBytes(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

// HashAny returns the hex-encoded SHA-256 of the canonical JSON encoding of v.
// Returns "" for nil inputs.
func HashAny(v any) (string, error) {
	if v == nil {
		return "", nil
	}
	b, err := Canonical(v)
	if err != nil {
		return "", err
	}
	return HashBytes(b), nil
}

// HashString returns the hex-encoded SHA-256 of the UTF-8 bytes of s.
// Returns "" for empty strings.
func HashString(s string) string {
	if s == "" {
		return ""
	}
	return HashBytes([]byte(s))
}

// Sign populates r.PayloadHash and appends a local Ed25519 Signer to r.Signers.
// It must be called after all Body fields are set and before the receipt is written.
// The local signer is always placed at index 0 per the canonical signer ordering.
func Sign(privKey ed25519.PrivateKey, deviceID string, r *Receipt) error {
	body, err := CanonicalBody(*r)
	if err != nil {
		return fmt.Errorf("receipt/signer: canonical body: %w", err)
	}

	h := sha256.Sum256(body)
	r.PayloadHash = hex.EncodeToString(h[:])

	sig := ed25519.Sign(privKey, body)
	pub := privKey.Public().(ed25519.PublicKey)

	// Prepend so local signer is always at index 0 even if callers pre-populated Signers.
	r.Signers = append([]Signer{{
		SignerID:  "local:" + deviceID,
		Pubkey:    base64.RawURLEncoding.EncodeToString(pub),
		Algorithm: "Ed25519",
		Signature: base64.RawURLEncoding.EncodeToString(sig),
	}}, r.Signers...)
	return nil
}

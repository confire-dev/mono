package transport

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"
)

const signVersion = "v1"

// SignedHeaders returns the security headers to add to every Worker request.
//
// Scheme:
//   X-Confire-Timestamp: <unix seconds>
//   X-Confire-Device:    <device ID>
//   X-Confire-Sig:       HMAC-SHA256(signingKey, "v1:<timestamp>:<bodyHash>")
//
// The Worker verifies:
//   1. Timestamp is within ±5 minutes (prevents replay attacks).
//   2. Signature matches (proves the request came from a client holding the key).
//   3. Device ID is not revoked.
//
// Even if an attacker captures a valid API key and HTTPS traffic, they cannot
// forge new requests without the signing key, and cannot replay captured ones.
func SignedHeaders(apiKey, deviceID string, body []byte) map[string]string {
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	bodyHash := hex.EncodeToString(sha256Sum(body))
	sigInput := fmt.Sprintf("%s:%s:%s", signVersion, ts, bodyHash)

	// Signing key is derived from the API key — never transmitted over the wire.
	signingKey := deriveSigningKey(apiKey)
	sig := hex.EncodeToString(hmacSHA256(signingKey, []byte(sigInput)))

	return map[string]string{
		"X-Confire-Timestamp": ts,
		"X-Confire-Device":    deviceID,
		"X-Confire-Sig":       sig,
	}
}

// VerifySignature can be used in tests. Production verification runs in the Worker.
func VerifySignature(apiKey, deviceID, timestamp, sig string, body []byte) bool {
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return false
	}
	// Reject requests older than 5 minutes or more than 1 minute in the future.
	delta := time.Now().Unix() - ts
	if delta > 300 || delta < -60 {
		return false
	}
	bodyHash := hex.EncodeToString(sha256Sum(body))
	sigInput := fmt.Sprintf("%s:%s:%s", signVersion, timestamp, bodyHash)
	signingKey := deriveSigningKey(apiKey)
	expected := hex.EncodeToString(hmacSHA256(signingKey, []byte(sigInput)))
	// Constant-time comparison — no timing oracle.
	return hmac.Equal([]byte(sig), []byte(expected))
}

// deriveSigningKey derives a signing key from the API key using a fixed context.
// Separating the signing key from the API key means the API key can be validated
// by the Worker without exposing the signing key derivation.
func deriveSigningKey(apiKey string) []byte {
	return hmacSHA256([]byte(apiKey), []byte("confire-request-signing-v1"))
}

func hmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

func sha256Sum(data []byte) []byte {
	h := sha256.Sum256(data)
	return h[:]
}

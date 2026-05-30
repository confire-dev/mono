package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

// PKCEPair holds the code_verifier and code_challenge for an OAuth PKCE flow.
// RFC 7636: verifier is random; challenge = BASE64URL(SHA256(verifier)).
type PKCEPair struct {
	Verifier  string
	Challenge string
}

// NewPKCE generates a fresh PKCE pair.
func NewPKCE() (PKCEPair, error) {
	// 32 random bytes → 43-char base64url verifier (within RFC spec 43-128)
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return PKCEPair{}, err
	}
	verifier := base64.RawURLEncoding.EncodeToString(b)

	h := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(h[:])

	return PKCEPair{Verifier: verifier, Challenge: challenge}, nil
}

package receipt

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

// KeyDir returns the directory where local signing keys are stored.
func KeyDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".confire", "keys")
}

// SeedPath returns the path to the 32-byte Ed25519 seed file.
// The file is mode 0600 and never transmitted off the device.
func SeedPath() string {
	return filepath.Join(KeyDir(), "local.ed25519.seed")
}

// PubKeyPath returns the path to the human-readable public key file (mode 0644).
func PubKeyPath() string {
	return filepath.Join(KeyDir(), "local.ed25519.pub")
}

// LoadOrGenerate loads the local Ed25519 private key from disk.
// If no valid seed exists, a new keypair is generated and persisted.
// deviceID is embedded in the public key file header for identification.
func LoadOrGenerate(deviceID string) (ed25519.PrivateKey, error) {
	seed, err := os.ReadFile(SeedPath())
	if err == nil && len(seed) == ed25519.SeedSize {
		return ed25519.NewKeyFromSeed(seed), nil
	}

	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("receipt/keystore: generate key: %w", err)
	}

	dir := KeyDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("receipt/keystore: create key dir: %w", err)
	}
	if err := os.WriteFile(SeedPath(), priv.Seed(), 0600); err != nil {
		return nil, fmt.Errorf("receipt/keystore: write seed: %w", err)
	}
	// Public key write is best-effort; not fatal if it fails.
	_ = writePubKey(priv.Public().(ed25519.PublicKey), deviceID)
	return priv, nil
}

// LoadPubKey returns the local Ed25519 public key derived from the seed file.
func LoadPubKey() (ed25519.PublicKey, error) {
	seed, err := os.ReadFile(SeedPath())
	if err != nil {
		return nil, fmt.Errorf("receipt/keystore: read seed: %w", err)
	}
	if len(seed) != ed25519.SeedSize {
		return nil, fmt.Errorf("receipt/keystore: invalid seed length %d (want %d)",
			len(seed), ed25519.SeedSize)
	}
	return ed25519.NewKeyFromSeed(seed).Public().(ed25519.PublicKey), nil
}

// PubKeyB64 returns the base64url-encoded local Ed25519 public key.
func PubKeyB64() (string, error) {
	pub, err := LoadPubKey()
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(pub), nil
}

// Fingerprint returns the first 16 hex characters (8 bytes) of the SHA-256
// of the public key — a short, human-readable key identifier.
func Fingerprint(pub ed25519.PublicKey) string {
	h := sha256.Sum256(pub)
	return hex.EncodeToString(h[:8])
}

func writePubKey(pub ed25519.PublicKey, deviceID string) error {
	content := fmt.Sprintf("# confire-local %s\n%s\n",
		deviceID,
		base64.RawURLEncoding.EncodeToString(pub),
	)
	return os.WriteFile(PubKeyPath(), []byte(content), 0644)
}

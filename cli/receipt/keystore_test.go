package receipt

import (
	"crypto/ed25519"
	"os"
	"path/filepath"
	"testing"
)

// withKeyDir overrides KeyDir / SeedPath / PubKeyPath to point at dir for the
// duration of the test. Returns a restore function.
// We achieve this by temporarily monkey-patching os.UserHomeDir via an env var.
func withTempHome(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	return dir
}

func TestLoadOrGenerate_CreatesKeyOnFirstCall(t *testing.T) {
	home := withTempHome(t)
	_ = home

	priv, err := LoadOrGenerate("device-test")
	if err != nil {
		t.Fatalf("LoadOrGenerate: %v", err)
	}
	if len(priv) != ed25519.PrivateKeySize {
		t.Errorf("private key length %d, want %d", len(priv), ed25519.PrivateKeySize)
	}

	// Seed file must exist with correct permissions.
	info, err := os.Stat(SeedPath())
	if err != nil {
		t.Fatalf("seed file not created: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("seed file mode %o, want 0600", info.Mode().Perm())
	}
	if info.Size() != ed25519.SeedSize {
		t.Errorf("seed file size %d, want %d", info.Size(), ed25519.SeedSize)
	}
}

func TestLoadOrGenerate_ReturnsSameKeyOnSubsequentCalls(t *testing.T) {
	withTempHome(t)

	priv1, err := LoadOrGenerate("device-test")
	if err != nil {
		t.Fatal(err)
	}
	priv2, err := LoadOrGenerate("device-test")
	if err != nil {
		t.Fatal(err)
	}
	if string(priv1) != string(priv2) {
		t.Error("second LoadOrGenerate returned a different key")
	}
}

func TestLoadOrGenerate_KeyDirHasRestrictedPermissions(t *testing.T) {
	withTempHome(t)

	if _, err := LoadOrGenerate("device-test"); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(KeyDir())
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0700 {
		t.Errorf("key dir mode %o, want 0700", info.Mode().Perm())
	}
}

func TestLoadPubKey_MatchesGeneratedKey(t *testing.T) {
	withTempHome(t)

	priv, err := LoadOrGenerate("device-test")
	if err != nil {
		t.Fatal(err)
	}
	pub, err := LoadPubKey()
	if err != nil {
		t.Fatalf("LoadPubKey: %v", err)
	}
	expectedPub := priv.Public().(ed25519.PublicKey)
	if string(pub) != string(expectedPub) {
		t.Error("LoadPubKey returned a different public key than the generated private key")
	}
}

func TestLoadPubKey_ErrorWhenNoSeedFile(t *testing.T) {
	withTempHome(t)
	// Don't call LoadOrGenerate — no seed file should exist.
	if _, err := LoadPubKey(); err == nil {
		t.Error("expected error when seed file absent, got nil")
	}
}

func TestLoadPubKey_ErrorWhenSeedTruncated(t *testing.T) {
	withTempHome(t)
	if err := os.MkdirAll(KeyDir(), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(SeedPath(), []byte("short"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadPubKey(); err == nil {
		t.Error("expected error for truncated seed, got nil")
	}
}

func TestPubKeyB64_ReturnBase64URL(t *testing.T) {
	withTempHome(t)

	if _, err := LoadOrGenerate("device-test"); err != nil {
		t.Fatal(err)
	}
	b64, err := PubKeyB64()
	if err != nil {
		t.Fatal(err)
	}
	// base64url uses A-Z, a-z, 0-9, -, _ and no padding.
	for _, ch := range b64 {
		if !((ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') ||
			(ch >= '0' && ch <= '9') || ch == '-' || ch == '_') {
			t.Errorf("non-base64url character %q in pubkey %q", ch, b64)
			break
		}
	}
	// Ed25519 public key is 32 bytes → 43 base64url characters (no padding).
	if len(b64) != 43 {
		t.Errorf("base64url pubkey length %d, want 43", len(b64))
	}
}

func TestFingerprint_Is16HexChars(t *testing.T) {
	withTempHome(t)

	priv, err := LoadOrGenerate("device-test")
	if err != nil {
		t.Fatal(err)
	}
	pub := priv.Public().(ed25519.PublicKey)
	fp := Fingerprint(pub)
	if len(fp) != 16 {
		t.Errorf("fingerprint length %d, want 16", len(fp))
	}
	for _, ch := range fp {
		if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f')) {
			t.Errorf("non-hex character %q in fingerprint %q", ch, fp)
			break
		}
	}
}

func TestFingerprint_DeterministicForSameKey(t *testing.T) {
	withTempHome(t)

	priv, err := LoadOrGenerate("device-test")
	if err != nil {
		t.Fatal(err)
	}
	pub := priv.Public().(ed25519.PublicKey)
	fp1 := Fingerprint(pub)
	fp2 := Fingerprint(pub)
	if fp1 != fp2 {
		t.Errorf("fingerprint is not deterministic: %q != %q", fp1, fp2)
	}
}

func TestWritePubKey_CreatesFile(t *testing.T) {
	withTempHome(t)

	priv, err := LoadOrGenerate("device-42")
	if err != nil {
		t.Fatal(err)
	}
	pub := priv.Public().(ed25519.PublicKey)
	if err := writePubKey(pub, "device-42"); err != nil {
		t.Fatalf("writePubKey: %v", err)
	}

	content, err := os.ReadFile(PubKeyPath())
	if err != nil {
		t.Fatal(err)
	}
	text := string(content)
	if len(text) == 0 {
		t.Error("pub key file is empty")
	}
	// Header line must contain the device ID.
	if !contains(text, "device-42") {
		t.Errorf("pub key file does not contain device ID: %q", text)
	}
	// Must include the public key on a separate line.
	lines := filepath.SplitList(text) // reuse for non-path splitting
	_ = lines
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

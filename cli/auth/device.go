package auth

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/zalando/go-keyring"
)

const deviceIDField = "device_id"

// DeviceID returns a stable unique ID for this installation.
// Generated once on first call, stored in the OS keychain alongside the API key.
// Included in every Worker request so the server can track and revoke individual
// devices without invalidating the whole API key.
func DeviceID() (string, error) {
	existing, err := keyring.Get(service, deviceIDField)
	if err == nil && existing != "" {
		return existing, nil
	}
	// First time — generate a 16-byte random ID.
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	id := hex.EncodeToString(b)
	if err := keyring.Set(service, deviceIDField, id); err != nil {
		return id, nil // non-fatal: return the ID even if we can't persist it
	}
	return id, nil
}

// EnsureDeviceID ensures a device ID exists, creating it if needed.
// Called during setup so the ID is ready before login.
func EnsureDeviceID() error {
	_, err := DeviceID()
	return err
}

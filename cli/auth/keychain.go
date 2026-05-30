// Package auth handles OS keychain storage and OAuth PKCE flow.
package auth

import "github.com/zalando/go-keyring"

const (
	service  = "confire"
	keyField = "api_key"
	emailField = "email"
)

// StoreKey stores the API key securely in the OS keychain.
// macOS: Keychain, Linux: libsecret, Windows: Credential Manager.
func StoreKey(apiKey string) error {
	return keyring.Set(service, keyField, apiKey)
}

// LoadKey retrieves the stored API key.
// Returns ("", nil) if not set.
func LoadKey() (string, error) {
	key, err := keyring.Get(service, keyField)
	if err == keyring.ErrNotFound {
		return "", nil
	}
	return key, err
}

// DeleteKey removes the stored API key.
func DeleteKey() error {
	return keyring.Delete(service, keyField)
}

// StoreEmail caches the logged-in email for display in whoami/status.
func StoreEmail(email string) error {
	return keyring.Set(service, emailField, email)
}

// LoadEmail retrieves the cached email.
func LoadEmail() (string, error) {
	email, err := keyring.Get(service, emailField)
	if err == keyring.ErrNotFound {
		return "", nil
	}
	return email, err
}

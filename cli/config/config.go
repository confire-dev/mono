// Package config provides persistent CLI configuration via ~/.confire/config.json.
// Environment variables always override the file.
//
// Author: Efe <efe@efebehar.dev>
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Config is the user-editable CLI configuration.
// All fields have safe defaults so a missing or empty file works fine.
type Config struct {
	// Telemetry controls optional product analytics forwarding (Amplitude).
	// true = send analytics events (default)
	// false = skip Amplitude; usage accounting for billing still runs.
	Telemetry bool `json:"telemetry"`

	// WorkerURL overrides the default Worker endpoint.
	// Leave empty to use the production Worker.
	WorkerURL string `json:"worker_url,omitempty"`
}

const defaultTelemetry = true

// Load reads the config file and applies env-var overrides.
// Never returns an error — missing/corrupt config uses safe defaults.
func Load() Config {
	cfg := Config{Telemetry: defaultTelemetry}

	if data, err := os.ReadFile(Path()); err == nil {
		_ = json.Unmarshal(data, &cfg)
	}

	// CONFIRE_TELEMETRY=0 or CONFIRE_TELEMETRY=false → disable
	if v := os.Getenv("CONFIRE_TELEMETRY"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			cfg.Telemetry = b
		} else if strings.ToLower(v) == "0" || strings.ToLower(v) == "off" {
			cfg.Telemetry = false
		}
	}

	return cfg
}

// Save writes the config to disk (chmod 600).
func Save(cfg Config) error {
	if err := os.MkdirAll(filepath.Dir(Path()), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(Path(), data, 0600)
}

// Path returns the absolute path to the config file.
func Path() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".confire", "config.json")
}

// Package config provides persistent CLI configuration via ~/.confire/config.json.
// Authoritative source for preferences: Supabase user_preferences table.
// Local file is the offline fallback; synced via `confire config set`.
//
// Author: Efe <efe@efebehar.dev>
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/confire-dev/confire/hosts"
)

// NotificationConfig controls how Confire reports savings to the developer.
type NotificationConfig struct {
	// Enabled: false = silent (no 🔥 lines at all)
	Enabled bool `json:"enabled"`

	// Style: "brand" (🔥 with context) | "minimal" ([confire] compact) | "off" (silent)
	Style string `json:"style"`

	// MinSavedTokens: saves smaller than this emit nothing. Default 5000.
	MinSavedTokens int `json:"min_saved_tokens"`

	// BigSaveTokens: saves ≥ this get an additionalContext line (visible to Claude).
	// Smaller saves only write to stderr. Default 50000.
	BigSaveTokens int `json:"big_save_tokens"`
}

// Config is the user-editable CLI configuration.
// All fields have safe defaults so a missing or empty file works fine.
type Config struct {
	// Telemetry: legacy alias kept for backward compat.
	// Controls Amplitude analytics forwarding. Does NOT disable usage accounting.
	Telemetry bool `json:"telemetry"`

	// Analytics is the canonical name (Telemetry is kept for backward compat).
	Analytics bool `json:"analytics"`

	// WorkerURL overrides the default Worker endpoint.
	WorkerURL string `json:"worker_url,omitempty"`

	// Notifications controls 🔥 save notifications.
	Notifications NotificationConfig `json:"notifications"`

	// Mode controls firewall behavior: "observe"|"balanced"|"strict"|"bypass".
	// Default: "balanced".
	Mode string `json:"mode,omitempty"`

	// WelcomePending: show a one-time SessionStart note in Claude Code after setup.
	WelcomePending bool `json:"welcome_pending,omitempty"`

	// FirewallEnabled controls whether the tool/context firewall is active.
	// nil means true (default on). Use pointer so we can distinguish unset from false.
	FirewallEnabled *bool `json:"firewall_enabled,omitempty"`

	// Hosts overrides per-agent capabilities (merged onto hosts.DefaultCapabilities).
	// Example: {"cursor": {"post_tool_steer": false}}
	Hosts map[string]HostSettings `json:"hosts,omitempty"`

	// DashboardSync controls metadata event sync for logged-in users.
	DashboardSync DashboardSyncConfig `json:"dashboard_sync,omitempty"`
}

// HostSettings overrides capability defaults for one host ID (cursor, claude-code, …).
type HostSettings struct {
	PostToolSteer *bool `json:"post_tool_steer,omitempty"`
}

// DashboardSyncConfig controls whether metadata events are synced to the dashboard.
type DashboardSyncConfig struct {
	Enabled *bool `json:"enabled,omitempty"`
}

// IsDashboardSyncEnabled returns true (default) unless explicitly disabled.
func (d *DashboardSyncConfig) IsDashboardSyncEnabled() bool {
	if d.Enabled != nil {
		return *d.Enabled
	}
	return true
}

// CapabilitiesFor returns effective capabilities for a host, merging config overrides.
func (c Config) CapabilitiesFor(hostID string) hosts.Capabilities {
	caps := hosts.DefaultCapabilities(hostID)
	if o, ok := c.Hosts[hostID]; ok {
		if o.PostToolSteer != nil {
			caps.PostToolSteer = *o.PostToolSteer
		}
	}
	return caps
}

// IsFirewallEnabled returns true unless the firewall has been explicitly disabled
// or mode is set to bypass.
func (c *Config) IsFirewallEnabled() bool {
	if c.FirewallEnabled != nil && !*c.FirewallEnabled {
		return false
	}
	if c.Mode == "bypass" {
		return false
	}
	return true
}

// EffectiveMode returns the firewall mode, defaulting to "balanced".
func (c *Config) EffectiveMode() string {
	if c.Mode == "" {
		return "balanced"
	}
	return c.Mode
}

func defaults() Config {
	return Config{
		Telemetry: true,
		Analytics: true,
		Notifications: NotificationConfig{
			Enabled:        true,
			Style:          "brand",
			MinSavedTokens: 200,
			BigSaveTokens:  5_000,
		},
	}
}

// Load reads the config file and applies env-var overrides.
// Never returns an error — missing/corrupt config uses safe defaults.
func Load() Config {
	cfg := defaults()

	if data, err := os.ReadFile(Path()); err == nil {
		_ = json.Unmarshal(data, &cfg)
		// Ensure notification defaults for older config files
		if cfg.Notifications.BigSaveTokens == 0 {
			cfg.Notifications = defaults().Notifications
		}
	}

	// CONFIRE_TELEMETRY=0 or CONFIRE_ANALYTICS=0 env overrides
	for _, env := range []string{"CONFIRE_TELEMETRY", "CONFIRE_ANALYTICS"} {
		if v := os.Getenv(env); v != "" {
			if b, err := strconv.ParseBool(v); err == nil {
				cfg.Telemetry = b
				cfg.Analytics = b
			} else if strings.ToLower(v) == "0" || strings.ToLower(v) == "off" {
				cfg.Telemetry = false
				cfg.Analytics = false
			}
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

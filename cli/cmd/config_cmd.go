package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/confire-dev/confire/auth"
	"github.com/confire-dev/confire/config"
	"github.com/confire-dev/confire/hosts"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config [get|set] [key[=value]]",
	Short: "Read or write CLI configuration",
	Long: `Read or write CLI preferences.
Changes are saved to ~/.confire/config.json and synced to Supabase when online.

Keys:
  notifications.enabled          true | false
  notifications.style            brand | minimal | off
  notifications.min_saved_tokens <N>       (tokens, default 5000)
  notifications.big_save_tokens  <N>       (tokens, default 50000)
  analytics.enabled              true | false
  worker_url                     <URL>
  hosts.<id>.optimize_native     true | false   (per host, e.g. cursor)
  hosts.<id>.optimize_mcp        true | false
  hosts.<id>.post_tool_steer     true | false

Examples:
  confire config                                 # show all
  confire config get notifications.style
  confire config set notifications.enabled=false
  confire config set hosts.cursor.optimize_native=false
  confire config set notifications.big_save_tokens=20000`,
	RunE: runConfig,
}

func init() {
	rootCmd.AddCommand(configCmd)
}

func runConfig(cmd *cobra.Command, args []string) error {
	cfg := config.Load()

	// No args → show everything
	if len(args) == 0 {
		return showConfig(cfg)
	}

	switch args[0] {
	case "get":
		if len(args) < 2 {
			return showConfig(cfg)
		}
		return getConfigKey(cfg, args[1])

	case "set":
		if len(args) < 2 {
			return fmt.Errorf("usage: confire config set <key>=<value>")
		}
		return setConfigKey(&cfg, args[1])

	default:
		// Bare "confire config key" → get
		if strings.Contains(args[0], "=") {
			return setConfigKey(&cfg, args[0])
		}
		return getConfigKey(cfg, args[0])
	}
}

func showConfig(cfg config.Config) error {
	out, _ := json.MarshalIndent(cfg, "", "  ")
	fmt.Println(string(out))
	return nil
}

func getConfigKey(cfg config.Config, key string) error {
	val, err := getField(cfg, key)
	if err != nil {
		return err
	}
	fmt.Println(val)
	return nil
}

func setConfigKey(cfg *config.Config, kv string) error {
	parts := strings.SplitN(kv, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("expected key=value, got %q", kv)
	}
	key, val := parts[0], parts[1]

	if err := setField(cfg, key, val); err != nil {
		return err
	}
	if err := config.Save(*cfg); err != nil {
		return fmt.Errorf("save: %w", err)
	}

	fmt.Printf("%s✓%s  %s = %s\n", green, reset, key, val)

	// Best-effort sync to Supabase
	go syncPrefsToSupabase(*cfg)
	return nil
}

// ── Field get/set ──────────────────────────────────────────────────────────

func getField(cfg config.Config, key string) (string, error) {
	switch key {
	case "analytics.enabled", "telemetry":
		return boolStr(cfg.Analytics), nil
	case "notifications.enabled":
		return boolStr(cfg.Notifications.Enabled), nil
	case "notifications.style":
		return cfg.Notifications.Style, nil
	case "notifications.min_saved_tokens":
		return strconv.Itoa(cfg.Notifications.MinSavedTokens), nil
	case "notifications.big_save_tokens":
		return strconv.Itoa(cfg.Notifications.BigSaveTokens), nil
	case "worker_url":
		if cfg.WorkerURL == "" {
			return "(default)", nil
		}
		return cfg.WorkerURL, nil
	default:
		if val, ok := getHostField(cfg, key); ok {
			return val, nil
		}
		return "", fmt.Errorf("unknown key %q — run `confire config` to see all keys", key)
	}
}

func setField(cfg *config.Config, key, val string) error {
	switch key {
	case "analytics.enabled", "telemetry":
		b, err := parseBool(val)
		if err != nil {
			return err
		}
		cfg.Analytics = b
		cfg.Telemetry = b

	case "notifications.enabled":
		b, err := parseBool(val)
		if err != nil {
			return err
		}
		cfg.Notifications.Enabled = b

	case "notifications.style":
		switch val {
		case "brand", "minimal", "off":
			cfg.Notifications.Style = val
		default:
			return fmt.Errorf("notifications.style must be brand|minimal|off, got %q", val)
		}

	case "notifications.min_saved_tokens":
		n, err := strconv.Atoi(val)
		if err != nil || n < 0 {
			return fmt.Errorf("notifications.min_saved_tokens must be a non-negative integer")
		}
		cfg.Notifications.MinSavedTokens = n

	case "notifications.big_save_tokens":
		n, err := strconv.Atoi(val)
		if err != nil || n < 0 {
			return fmt.Errorf("notifications.big_save_tokens must be a non-negative integer")
		}
		cfg.Notifications.BigSaveTokens = n

	case "worker_url":
		cfg.WorkerURL = val

	default:
		if err := setHostField(cfg, key, val); err == nil {
			return nil
		}
		return fmt.Errorf("unknown key %q — run `confire config` to see all keys", key)
	}
	return nil
}

func getHostField(cfg config.Config, key string) (string, bool) {
	if !strings.HasPrefix(key, "hosts.") {
		return "", false
	}
	parts := strings.Split(key, ".")
	if len(parts) != 3 {
		return "", false
	}
	hostID, field := parts[1], parts[2]
	o, ok := cfg.Hosts[hostID]
	if !ok {
		return "(default)", true
	}
	switch field {
	case "optimize_native":
		if o.OptimizeNative == nil {
			return boolStr(hosts.DefaultCapabilities(hostID).OptimizeNative), true
		}
		return boolStr(*o.OptimizeNative), true
	case "optimize_mcp":
		if o.OptimizeMCP == nil {
			return boolStr(hosts.DefaultCapabilities(hostID).OptimizeMCP), true
		}
		return boolStr(*o.OptimizeMCP), true
	case "post_tool_steer":
		if o.PostToolSteer == nil {
			return boolStr(hosts.DefaultCapabilities(hostID).PostToolSteer), true
		}
		return boolStr(*o.PostToolSteer), true
	default:
		return "", false
	}
}

func setHostField(cfg *config.Config, key, val string) error {
	if !strings.HasPrefix(key, "hosts.") {
		return fmt.Errorf("unknown key")
	}
	parts := strings.Split(key, ".")
	if len(parts) != 3 {
		return fmt.Errorf("unknown key")
	}
	hostID, field := parts[1], parts[2]
	b, err := parseBool(val)
	if err != nil {
		return err
	}
	if cfg.Hosts == nil {
		cfg.Hosts = map[string]config.HostSettings{}
	}
	o := cfg.Hosts[hostID]
	switch field {
	case "optimize_native":
		o.OptimizeNative = &b
	case "optimize_mcp":
		o.OptimizeMCP = &b
	case "post_tool_steer":
		o.PostToolSteer = &b
	default:
		return fmt.Errorf("unknown key")
	}
	cfg.Hosts[hostID] = o
	return nil
}

// ── Supabase sync ──────────────────────────────────────────────────────────

func syncPrefsToSupabase(cfg config.Config) {
	apiKey, _ := auth.LoadKey()
	if apiKey == "" {
		return // not logged in, nothing to sync
	}
	// TODO: POST to Worker /api/preferences with the notification + analytics settings
	// Worker upserts user_preferences in Supabase.
	// This is a fire-and-forget best-effort sync.
}

// ── Helpers ────────────────────────────────────────────────────────────────

func parseBool(s string) (bool, error) {
	switch strings.ToLower(s) {
	case "true", "1", "yes", "on":
		return true, nil
	case "false", "0", "no", "off":
		return false, nil
	}
	return false, fmt.Errorf("expected true/false, got %q", s)
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

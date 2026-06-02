package policy

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// CachedPolicy is the on-disk format for custom rules fetched from the cloud.
type CachedPolicy struct {
	FetchedAt      time.Time       `json:"fetched_at"`
	Version        string          `json:"version,omitempty"`
	Rules          []Rule          `json:"rules"`
	// GroupOverrides maps group IDs to enabled state.
	// false = user toggled that group off in the dashboard.
	// Sent by the worker alongside custom rules; applied after merge.
	GroupOverrides map[string]bool `json:"group_overrides,omitempty"`
}

// CachePath returns the path to the local custom-rule cache.
func CachePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".confire", "policies", "cache.json")
}

// LoadRules returns the merged rule set: bundled rules + cached custom rules,
// with group overrides applied. Never returns an error — missing or corrupt
// cache silently uses built-ins only.
func LoadRules() []Rule {
	builtin := BuiltinRules()
	cp := loadCache()
	var custom []Rule
	var overrides map[string]bool
	if cp != nil {
		custom = cp.Rules
		overrides = cp.GroupOverrides
	}
	rules := mergeRules(builtin, custom)
	return applyGroupOverrides(rules, overrides)
}

// LoadGroupOverrides returns the current group override map from cache.
// Returns nil if no cache or no overrides set.
func LoadGroupOverrides() map[string]bool {
	cp := loadCache()
	if cp == nil {
		return nil
	}
	return cp.GroupOverrides
}

// SaveCache persists a fetched custom-rule set and group overrides to disk.
func SaveCache(rules []Rule, version string, overrides map[string]bool) error {
	path := CachePath()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	cp := CachedPolicy{
		FetchedAt:      time.Now().UTC(),
		Version:        version,
		Rules:          rules,
		GroupOverrides: overrides,
	}
	data, err := json.MarshalIndent(cp, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// LoadCacheInfo returns metadata about the cached custom rules (for CLI display).
// Returns zero value if no cache exists.
func LoadCacheInfo() (count int, fetchedAt time.Time, version string) {
	cp := loadCache()
	if cp == nil {
		return 0, time.Time{}, ""
	}
	return len(cp.Rules), cp.FetchedAt, cp.Version
}

// BypassNextPath returns the path to the one-shot bypass flag file.
func BypassNextPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".confire", ".bypass-next")
}

// ConsumeBypassNext checks and atomically clears the bypass-next flag.
// Returns true if the flag was set (and clears it).
func ConsumeBypassNext() bool {
	path := BypassNextPath()
	if _, err := os.Stat(path); err != nil {
		return false
	}
	_ = os.Remove(path)
	return true
}

// SetBypassNext writes the one-shot bypass flag.
func SetBypassNext() error {
	path := BypassNextPath()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	return os.WriteFile(path, []byte("1"), 0600)
}

// ── internal ──────────────────────────────────────────────────────────────

func loadCachedCustomRules() []Rule {
	cp := loadCache()
	if cp == nil {
		return nil
	}
	return cp.Rules
}

func loadCache() *CachedPolicy {
	data, err := os.ReadFile(CachePath())
	if err != nil {
		return nil
	}
	var cp CachedPolicy
	if err := json.Unmarshal(data, &cp); err != nil {
		return nil
	}
	return &cp
}

// ApplyGroupOverrides disables all rules whose Group appears in overrides with value false.
// Rules with no group, or whose group is not in the map, are unaffected.
// This is the exported form used by tests and by the CLI sync path.
func ApplyGroupOverrides(rules []Rule, overrides map[string]bool) []Rule {
	return applyGroupOverrides(rules, overrides)
}

// applyGroupOverrides is the internal form called during rule loading.
func applyGroupOverrides(rules []Rule, overrides map[string]bool) []Rule {
	if len(overrides) == 0 {
		return rules
	}
	out := make([]Rule, len(rules))
	copy(out, rules)
	for i, r := range out {
		if r.Group == "" {
			continue
		}
		if enabled, ok := overrides[r.Group]; ok && !enabled {
			out[i].Enabled = false
		}
	}
	return out
}

// mergeRules merges custom rules on top of built-in rules.
// If a custom rule has the same ID as a built-in, the custom rule wins.
// New IDs from custom rules are appended.
func mergeRules(builtin, custom []Rule) []Rule {
	if len(custom) == 0 {
		return builtin
	}
	// Index built-ins by ID.
	result := make([]Rule, 0, len(builtin)+len(custom))
	seen := make(map[string]int) // id → index in result
	for _, r := range builtin {
		seen[r.ID] = len(result)
		result = append(result, r)
	}
	// Overlay custom rules.
	for _, r := range custom {
		if r.Source == "" {
			r.Source = "custom"
		}
		if idx, ok := seen[r.ID]; ok {
			result[idx] = r // override built-in
		} else {
			seen[r.ID] = len(result)
			result = append(result, r)
		}
	}
	return result
}

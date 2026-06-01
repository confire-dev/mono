package policy

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// CachedPolicy is the on-disk format for custom rules fetched from the cloud.
type CachedPolicy struct {
	FetchedAt time.Time `json:"fetched_at"`
	Version   string    `json:"version,omitempty"`
	Rules     []Rule    `json:"rules"`
}

// CachePath returns the path to the local custom-rule cache.
func CachePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".confire", "policies", "cache.json")
}

// LoadRules returns the merged rule set: bundled rules + cached custom rules.
// Bundled rules always load. Custom rules overlay on top (same ID = override).
// Never returns an error — missing or corrupt cache silently uses built-ins only.
func LoadRules() []Rule {
	builtin := BuiltinRules()
	custom := loadCachedCustomRules()
	return mergeRules(builtin, custom)
}

// SaveCache persists a fetched custom-rule set to disk.
func SaveCache(rules []Rule, version string) error {
	path := CachePath()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	cp := CachedPolicy{
		FetchedAt: time.Now().UTC(),
		Version:   version,
		Rules:     rules,
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

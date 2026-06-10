// json_util.go — shared JSON helpers for all hook config file operations.
//
// All install/uninstall functions must use these utilities so that:
//   - Unknown top-level keys in any config file are never dropped.
//   - Key ordering is preserved (orderedMap tracks insertion order).
//   - Unknown fields in existing entries are never dropped (raw JSON).
//   - Non-confire entries are never removed or modified.
package hosts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// ── orderedMap ────────────────────────────────────────────────────────────────
// A JSON object that preserves key insertion order on both read and write.
// Values are stored as raw JSON so unknown fields are never unmarshaled.

type orderedMap struct {
	keys []string
	vals map[string]json.RawMessage
}

func newOrderedMap() *orderedMap {
	return &orderedMap{vals: make(map[string]json.RawMessage)}
}

func (m *orderedMap) get(key string) (json.RawMessage, bool) {
	v, ok := m.vals[key]
	return v, ok
}

func (m *orderedMap) set(key string, val json.RawMessage) {
	if _, exists := m.vals[key]; !exists {
		m.keys = append(m.keys, key)
	}
	m.vals[key] = val
}

func (m *orderedMap) UnmarshalJSON(data []byte) error {
	if m.vals == nil {
		m.vals = make(map[string]json.RawMessage)
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	if t, err := dec.Token(); err != nil || t != json.Delim('{') {
		return fmt.Errorf("expected JSON object")
	}
	for dec.More() {
		t, err := dec.Token()
		if err != nil {
			return err
		}
		key, _ := t.(string)
		var val json.RawMessage
		if err := dec.Decode(&val); err != nil {
			return err
		}
		m.set(key, val)
	}
	_, err := dec.Token() // consume '}'
	return err
}

func (m *orderedMap) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, k := range m.keys {
		if i > 0 {
			buf.WriteByte(',')
		}
		kb, err := json.Marshal(k)
		if err != nil {
			return nil, err
		}
		buf.Write(kb)
		buf.WriteByte(':')
		buf.Write(m.vals[k])
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

// ── Load / save ───────────────────────────────────────────────────────────────

func loadOrderedMap(path string) (*orderedMap, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return newOrderedMap(), nil
		}
		return nil, err
	}
	m := newOrderedMap()
	if len(bytes.TrimSpace(raw)) > 0 {
		if err := json.Unmarshal(raw, m); err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
	}
	return m, nil
}

func saveOrderedMap(path string, m *orderedMap) error {
	out, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	// Back up the original file once — only if no backup exists yet.
	// The backup captures the pre-Confire state so the user can always restore it.
	// Subsequent install/uninstall runs leave the backup untouched.
	backupOnce(path)

	// Write to a temp file in the same directory, then rename atomically.
	// This prevents a partial write from corrupting the config if the process
	// is killed between write and close.
	tmp := path + ".confire-tmp"
	if err := os.WriteFile(tmp, append(out, '\n'), 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// backupOnce copies path → path+".confire-backup" if:
//   - path exists (there is something to back up), and
//   - the backup does not already exist (first modification only).
//
// Errors are silently ignored — a missing backup is inconvenient, not fatal.
func backupOnce(path string) {
	backupPath := path + ".confire-backup"
	if _, err := os.Stat(backupPath); err == nil {
		return // backup already exists — leave it (preserves the original)
	}
	data, err := os.ReadFile(path)
	if err != nil || len(data) == 0 {
		return // file doesn't exist or is empty — nothing worth backing up
	}
	_ = os.WriteFile(backupPath, data, 0644)
}

// ── Entry helpers ─────────────────────────────────────────────────────────────

// isConfireEntry reports whether a raw JSON hook entry belongs to confire.
func isConfireEntry(raw json.RawMessage) bool {
	s := string(raw)
	return strings.Contains(s, "confire") && strings.Contains(s, "hook")
}

// dedupAppend removes existing confire entries from arr then appends newEntry.
// All non-confire entries are kept verbatim — their fields are never touched.
func dedupAppend(arr []json.RawMessage, newEntry json.RawMessage) []json.RawMessage {
	kept := make([]json.RawMessage, 0, len(arr))
	for _, e := range arr {
		if !isConfireEntry(e) {
			kept = append(kept, e)
		}
	}
	return append(kept, newEntry)
}

// removeConfire removes all confire entries from arr, preserving everything else.
func removeConfire(arr []json.RawMessage) []json.RawMessage {
	kept := make([]json.RawMessage, 0, len(arr))
	for _, e := range arr {
		if !isConfireEntry(e) {
			kept = append(kept, e)
		}
	}
	return kept
}

// ── Generic patch helpers ─────────────────────────────────────────────────────

// patchHooksField modifies hook event arrays inside a "hooks" sub-object of top.
// All other top-level keys, their order, and all fields in non-confire entries
// are preserved. Unknown event names inside the hooks object are untouched.
func patchHooksField(
	top *orderedMap,
	events []string,
	fn func(event string, arr []json.RawMessage) []json.RawMessage,
) error {
	hooksMap := newOrderedMap()
	if raw, ok := top.get("hooks"); ok {
		if err := json.Unmarshal(raw, hooksMap); err != nil {
			return err
		}
	}
	for _, event := range events {
		var arr []json.RawMessage
		if raw, ok := hooksMap.get(event); ok {
			_ = json.Unmarshal(raw, &arr)
		}
		arr = fn(event, arr)
		b, err := json.Marshal(arr)
		if err != nil {
			return err
		}
		hooksMap.set(event, b)
	}
	b, err := json.Marshal(hooksMap)
	if err != nil {
		return err
	}
	top.set("hooks", b)
	return nil
}

// patchFlatHooksField modifies hook event arrays that are direct keys of top
// (Windsurf format: the file IS the hooks map, no "hooks" wrapper).
func patchFlatHooksField(
	top *orderedMap,
	events []string,
	fn func(event string, arr []json.RawMessage) []json.RawMessage,
) {
	for _, event := range events {
		var arr []json.RawMessage
		if raw, ok := top.get(event); ok {
			_ = json.Unmarshal(raw, &arr)
		}
		arr = fn(event, arr)
		b, _ := json.Marshal(arr)
		top.set(event, b)
	}
}

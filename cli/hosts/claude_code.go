// Package hosts — Claude Code host adapter + Hook encode/decode.
package hosts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/confire-dev/confire/intercept"
)

// ── Host interface implementation ─────────────────────────────────────────

type ClaudeCodeHost struct{}

func (h *ClaudeCodeHost) ID() string    { return "claude-code" }
func (h *ClaudeCodeHost) Label() string { return "Claude Code" }

func (h *ClaudeCodeHost) Detect() bool {
	// Check for the claude binary in PATH.
	if _, err := exec.LookPath("claude"); err == nil {
		return true
	}
	// Fallback: check the native macOS install location.
	home, _ := os.UserHomeDir()
	if _, err := os.Stat(filepath.Join(home, ".local", "share", "claude")); err == nil {
		return true
	}
	return false
}

func (h *ClaudeCodeHost) Strategies() []Strategy {
	return []Strategy{StrategyHooks, StrategyMCPProxy}
}

func (h *ClaudeCodeHost) Preferred() Strategy { return StrategyHooks }

func (h *ClaudeCodeHost) Install(s Strategy, opts InstallOptions) error {
	if s != StrategyHooks {
		return fmt.Errorf("%s strategy not yet implemented for Claude Code", s)
	}
	path := opts.SettingsPath
	if path == "" {
		path = claudeSettingsPath()
	}
	return installClaudeCodeHooksAt(path, opts.BinaryPath)
}

func (h *ClaudeCodeHost) IsInstalled(s Strategy) bool {
	if s != StrategyHooks {
		return false
	}
	return claudeCodeHookInstalled()
}

// IsInstalledAt reports whether confire hooks are present at a specific path.
func (h *ClaudeCodeHost) IsInstalledAt(path string) bool {
	return claudeCodeHookInstalledAt(path)
}

func (h *ClaudeCodeHost) ComingSoon() bool { return false }

func (h *ClaudeCodeHost) Uninstall(s Strategy) error {
	return h.UninstallAt(s, claudeSettingsPath())
}

func (h *ClaudeCodeHost) UninstallAt(s Strategy, path string) error {
	if s != StrategyHooks {
		return nil
	}
	return uninstallClaudeCodeHooksAt(path)
}

// ── Settings.json manipulation ────────────────────────────────────────────

func claudeSettingsPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude", "settings.json")
}

func claudeCodeHookInstalled() bool {
	return claudeCodeHookInstalledAt(claudeSettingsPath())
}

func claudeCodeHookInstalledAt(settingsPath string) bool {
	top, err := loadSettings(settingsPath)
	if err != nil {
		return false
	}
	rawHooks, ok := top.get("hooks")
	if !ok {
		return false
	}
	hooksMap := newOrderedMap()
	if err := json.Unmarshal(rawHooks, hooksMap); err != nil {
		return false
	}
	rawPost, ok := hooksMap.get("PostToolUse")
	if !ok {
		return false
	}
	var entries []json.RawMessage
	if err := json.Unmarshal(rawPost, &entries); err != nil {
		return false
	}
	for _, e := range entries {
		if isConfireEntry(e) {
			return true
		}
	}
	return false
}

// ── Ordered JSON object ───────────────────────────────────────────────────
// orderedMap preserves JSON object key order on read and write.
// Only the fields we modify are re-serialized; everything else stays raw.

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

// ── Settings I/O ──────────────────────────────────────────────────────────

func loadSettings(settingsPath string) (*orderedMap, error) {
	raw, err := os.ReadFile(settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return newOrderedMap(), nil
		}
		return nil, err
	}
	m := newOrderedMap()
	if err := json.Unmarshal(raw, m); err != nil {
		return nil, fmt.Errorf("parse settings: %w", err)
	}
	return m, nil
}

func saveSettings(settingsPath string, top *orderedMap) error {
	// MarshalJSON produces compact output; json.MarshalIndent re-indents it.
	out, err := json.MarshalIndent(top, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(settingsPath, out, 0644)
}

// hookEntryJSON returns a pre-serialized hook entry with a fixed key order:
// matcher → hooks → type → command. No map involved so no key reordering.
func hookEntryJSON(matcher, cmd string) json.RawMessage {
	m, _ := json.Marshal(matcher)
	c, _ := json.Marshal(cmd)
	return json.RawMessage(fmt.Sprintf(`{"matcher":%s,"hooks":[{"type":"command","command":%s}]}`, m, c))
}

// isConfireEntry reports whether a raw hook entry belongs to confire.
func isConfireEntry(raw json.RawMessage) bool {
	s := string(raw)
	return strings.Contains(s, "confire") && strings.Contains(s, "hook")
}

// dedupAppend removes existing confire entries from arr then appends newEntry.
func dedupAppend(arr []json.RawMessage, newEntry json.RawMessage) []json.RawMessage {
	kept := make([]json.RawMessage, 0, len(arr))
	for _, e := range arr {
		if !isConfireEntry(e) {
			kept = append(kept, e)
		}
	}
	return append(kept, newEntry)
}

// patchHooks reads the hooks orderedMap, applies fn to each named phase array,
// and writes the result back into top — leaving all other fields untouched.
func patchHooks(top *orderedMap, fn func(phase string, arr []json.RawMessage) []json.RawMessage) error {
	hooksMap := newOrderedMap()
	if rawHooks, ok := top.get("hooks"); ok {
		if err := json.Unmarshal(rawHooks, hooksMap); err != nil {
			return err
		}
	}

	for _, phase := range []string{"PostToolUse", "PreToolUse"} {
		var arr []json.RawMessage
		if raw, ok := hooksMap.get(phase); ok {
			_ = json.Unmarshal(raw, &arr)
		}
		arr = fn(phase, arr)
		arrBytes, err := json.Marshal(arr)
		if err != nil {
			return err
		}
		hooksMap.set(phase, arrBytes)
	}

	hooksBytes, err := json.Marshal(hooksMap)
	if err != nil {
		return err
	}
	top.set("hooks", hooksBytes)
	return nil
}

// ── Install / Uninstall ───────────────────────────────────────────────────

func installClaudeCodeHooks(binaryPath string) error {
	return installClaudeCodeHooksAt(claudeSettingsPath(), binaryPath)
}

func installClaudeCodeHooksAt(settingsPath, binaryPath string) error {
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0700); err != nil {
		return err
	}

	if orig, err := os.ReadFile(settingsPath); err == nil && len(orig) > 0 {
		_ = os.WriteFile(settingsPath+".confire-backup", orig, 0644)
	}

	top, err := loadSettings(settingsPath)
	if err != nil {
		return err
	}

	if binaryPath == "" {
		binaryPath, _ = os.Executable()
	}
	hookCmd := fmt.Sprintf(`"%s" hook`, binaryPath)

	if err := patchHooks(top, func(phase string, arr []json.RawMessage) []json.RawMessage {
		var matcher string
		switch phase {
		case "PostToolUse":
			matcher = ".*"
		case "PreToolUse":
			matcher = "Read"
		}
		return dedupAppend(arr, hookEntryJSON(matcher, hookCmd))
	}); err != nil {
		return err
	}

	return saveSettings(settingsPath, top)
}

func uninstallClaudeCodeHooks() error {
	return uninstallClaudeCodeHooksAt(claudeSettingsPath())
}

func uninstallClaudeCodeHooksAt(settingsPath string) error {

	top, err := loadSettings(settingsPath)
	if err != nil || len(top.keys) == 0 {
		return nil
	}

	if err := patchHooks(top, func(_ string, arr []json.RawMessage) []json.RawMessage {
		kept := make([]json.RawMessage, 0, len(arr))
		for _, e := range arr {
			if !isConfireEntry(e) {
				kept = append(kept, e)
			}
		}
		return kept
	}); err != nil {
		return err
	}

	return saveSettings(settingsPath, top)
}


func toIfaceSlice(v interface{}) []interface{} {
	if s, ok := v.([]interface{}); ok {
		return s
	}
	return nil
}

// HookInput is the exact JSON structure Claude Code sends to a hook on stdin.
// Field names verified from the Claude Code 2.1.158 binary (not public docs).
type HookInput struct {
	SessionID      string `json:"session_id"`
	TranscriptPath string `json:"transcript_path"`
	CWD            string `json:"cwd"`
	PermissionMode string `json:"permission_mode"`
	HookEventName  string `json:"hook_event_name"`
	ToolName       string `json:"tool_name"`
	ToolInput      any    `json:"tool_input"`
	ToolResponse   any    `json:"tool_response"`
	ToolUseID      string `json:"tool_use_id"`
	DurationMs     int    `json:"duration_ms"`
}

// HookOutput is what confire prints to stdout for PostToolUse.
// Claude Code reads this and applies hookSpecificOutput.updatedToolOutput
// before passing the result to the model.
type HookOutput struct {
	HookSpecificOutput HookSpecificOutput `json:"hookSpecificOutput"`
}

type HookSpecificOutput struct {
	HookEventName     string `json:"hookEventName"`
	UpdatedToolOutput any    `json:"updatedToolOutput,omitempty"`
	UpdatedInput      any    `json:"updatedInput,omitempty"`
	AdditionalContext string `json:"additionalContext,omitempty"`
}

// DecodeHookInput translates a ClaudeCode HookInput into a canonical InterceptEvent.
func DecodeHookInput(input HookInput) intercept.InterceptEvent {
	e := intercept.InterceptEvent{
		Host:     "claude-code",
		Strategy: "hooks",
		Phase:    MapPhase(input.HookEventName),
		Session: intercept.Session{
			ID:             input.SessionID,
			CWD:            input.CWD,
			TranscriptPath: input.TranscriptPath,
		},
		Raw: input,
	}

	if input.ToolName != "" {
		e.Tool = &intercept.Tool{
			Name:       input.ToolName,
			Input:      input.ToolInput,
			Output:     input.ToolResponse,
			UseID:      input.ToolUseID,
			IsMCP:      IsMCPTool(input.ToolName),
			DurationMs: input.DurationMs,
		}
		if e.Tool.IsMCP {
			e.Tool.MCPServer = MCPServerName(input.ToolName)
		}
	}
	return e
}

// EncodeResult translates an InterceptResult into the HookOutput Claude Code expects.
func EncodeResult(result intercept.InterceptResult, eventName string) (HookOutput, bool) {
	out := HookOutput{
		HookSpecificOutput: HookSpecificOutput{HookEventName: eventName},
	}

	switch result.Kind {
	case intercept.ResultReplaceOutput:
		out.HookSpecificOutput.UpdatedToolOutput = result.ToolOutput
		return out, true
	case intercept.ResultReplaceInput:
		out.HookSpecificOutput.UpdatedInput = result.ToolInput
		return out, true
	case intercept.ResultAddContext:
		out.HookSpecificOutput.AdditionalContext = result.Context
		return out, true
	default:
		return HookOutput{}, false
	}
}

// MapPhase maps Claude Code hook event names to canonical phases.
func MapPhase(hookEventName string) intercept.Phase {
	switch hookEventName {
	case "SessionStart":      return intercept.PhaseSessionStart
	case "SessionEnd":        return intercept.PhaseSessionEnd
	case "PreToolUse":        return intercept.PhaseToolPre
	case "PostToolUse":       return intercept.PhaseToolPost
	case "PostToolBatch":     return intercept.PhaseToolBatchPost
	case "Stop", "StopFailure": return intercept.PhaseTurnStop
	case "PreCompact":        return intercept.PhasePreCompact
	case "UserPromptSubmit":  return intercept.PhasePromptSubmit
	default:                  return intercept.Phase(hookEventName)
	}
}

// IsMCPTool returns true for tools namespaced as mcp__server__tool.
func IsMCPTool(name string) bool {
	return strings.HasPrefix(name, "mcp__")
}

// MCPServerName extracts the server name from "mcp__server__tool" → "server".
func MCPServerName(toolName string) string {
	parts := strings.SplitN(toolName, "__", 3)
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}

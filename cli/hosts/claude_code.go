// Package hosts — Claude Code host adapter + Hook encode/decode.
package hosts

import (
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
	return installClaudeCodeHooks(opts.BinaryPath)
}

func (h *ClaudeCodeHost) IsInstalled(s Strategy) bool {
	if s != StrategyHooks {
		return false
	}
	return claudeCodeHookInstalled()
}

func (h *ClaudeCodeHost) Uninstall(s Strategy) error {
	if s != StrategyHooks {
		return nil
	}
	return uninstallClaudeCodeHooks()
}

// ── Settings.json manipulation ────────────────────────────────────────────

func claudeSettingsPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude", "settings.json")
}

func claudeCodeHookInstalled() bool {
	data, err := os.ReadFile(claudeSettingsPath())
	if err != nil {
		return false
	}
	var s map[string]interface{}
	if err := json.Unmarshal(data, &s); err != nil {
		return false
	}
	hooks, _ := s["hooks"].(map[string]interface{})
	post, _ := hooks["PostToolUse"].([]interface{})
	for _, e := range post {
		m, _ := e.(map[string]interface{})
		for _, h := range toIfaceSlice(m["hooks"]) {
			hm, _ := h.(map[string]interface{})
			if strings.Contains(fmt.Sprint(hm["command"]), "confire hook") {
				return true
			}
		}
	}
	return false
}

func installClaudeCodeHooks(binaryPath string) error {
	settingsPath := claudeSettingsPath()
	raw, _ := os.ReadFile(settingsPath)
	var settings map[string]interface{}
	if len(raw) == 0 {
		settings = make(map[string]interface{})
	} else if err := json.Unmarshal(raw, &settings); err != nil {
		return fmt.Errorf("parse settings: %w", err)
	}

	// Backup
	if len(raw) > 0 {
		_ = os.WriteFile(settingsPath+".confire-backup", raw, 0644)
	}

	if binaryPath == "" {
		binaryPath, _ = os.Executable()
	}

	hookCmd := fmt.Sprintf(`"%s" hook`, binaryPath)

	hooks, _ := settings["hooks"].(map[string]interface{})
	if hooks == nil {
		hooks = make(map[string]interface{})
	}

	// PostToolUse — matches all tools
	postEntry := map[string]interface{}{
		"matcher": ".*",
		"hooks":   []interface{}{map[string]interface{}{"type": "command", "command": hookCmd}},
	}
	if existing, ok := hooks["PostToolUse"].([]interface{}); ok {
		if !claudeCodeHookInstalled() {
			hooks["PostToolUse"] = append(existing, postEntry)
		}
	} else {
		hooks["PostToolUse"] = []interface{}{postEntry}
	}

	// PreToolUse — Read only (inject limit before file is read)
	preEntry := map[string]interface{}{
		"matcher": "Read",
		"hooks":   []interface{}{map[string]interface{}{"type": "command", "command": hookCmd}},
	}
	if existing, ok := hooks["PreToolUse"].([]interface{}); ok {
		hooks["PreToolUse"] = append(existing, preEntry)
	} else {
		hooks["PreToolUse"] = []interface{}{preEntry}
	}

	settings["hooks"] = hooks
	out, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(settingsPath, out, 0644)
}

func uninstallClaudeCodeHooks() error {
	settingsPath := claudeSettingsPath()
	raw, err := os.ReadFile(settingsPath)
	if err != nil {
		return nil
	}
	var settings map[string]interface{}
	if err := json.Unmarshal(raw, &settings); err != nil {
		return err
	}
	hooks, _ := settings["hooks"].(map[string]interface{})
	if hooks == nil {
		return nil
	}
	for _, phase := range []string{"PostToolUse", "PreToolUse"} {
		entries, _ := hooks[phase].([]interface{})
		var kept []interface{}
		for _, e := range entries {
			m, _ := e.(map[string]interface{})
			hasConfire := false
			for _, h := range toIfaceSlice(m["hooks"]) {
				hm, _ := h.(map[string]interface{})
				if strings.Contains(fmt.Sprint(hm["command"]), "confire hook") {
					hasConfire = true
					break
				}
			}
			if !hasConfire {
				kept = append(kept, e)
			}
		}
		hooks[phase] = kept
	}
	settings["hooks"] = hooks
	out, _ := json.MarshalIndent(settings, "", "  ")
	return os.WriteFile(settingsPath, out, 0644)
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

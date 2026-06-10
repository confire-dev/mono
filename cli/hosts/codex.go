package hosts

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/confire-dev/confire/intercept"
)

// ── CodexHost ─────────────────────────────────────────────────────────────────

type CodexHost struct{}

func (h *CodexHost) ID() string             { return "codex" }
func (h *CodexHost) Label() string          { return "Codex CLI" }
func (h *CodexHost) Strategies() []Strategy { return []Strategy{StrategyHooks} }
func (h *CodexHost) Preferred() Strategy    { return StrategyHooks }
func (h *CodexHost) ComingSoon() bool       { return false }

func (h *CodexHost) Detect() bool {
	home, _ := os.UserHomeDir()
	// Check binary locations.
	for _, p := range []string{
		filepath.Join(home, ".local", "bin", "codex"),
		"/usr/local/bin/codex",
	} {
		if _, err := os.Stat(p); err == nil {
			return true
		}
	}
	// Check config directory.
	_, err := os.Stat(filepath.Join(home, ".codex"))
	return err == nil
}

func (h *CodexHost) Install(s Strategy, opts InstallOptions) error {
	if s != StrategyHooks {
		return fmt.Errorf("codex: strategy %q not supported", s)
	}
	local := opts.SettingsPath != ""
	return installCodexHooks(opts.BinaryPath, local)
}

func (h *CodexHost) IsInstalled(s Strategy) bool {
	if s != StrategyHooks {
		return false
	}
	return isCodexHookInstalled(false) || isCodexHookInstalled(true)
}

func (h *CodexHost) Uninstall(s Strategy) error {
	if s != StrategyHooks {
		return nil
	}
	_ = uninstallCodexHooks(false)
	_ = uninstallCodexHooks(true)
	return nil
}

// ── Codex hook types ──────────────────────────────────────────────────────────

// CodexHookInput is the JSON Codex CLI sends to a hook command via stdin.
// Ref: https://developers.openai.com/codex/hooks
type CodexHookInput struct {
	SessionID      string `json:"session_id"`
	CWD            string `json:"cwd"`
	HookEventName  string `json:"hook_event_name"`
	Model          string `json:"model"`
	ToolName       string `json:"tool_name"`
	ToolInput      any    `json:"tool_input"`
	ToolResponse   any    `json:"tool_response"`
	ToolUseID      string `json:"tool_use_id,omitempty"`
	TranscriptPath string `json:"transcript_path,omitempty"`
}

// CodexHookOutput is the envelope Codex CLI expects back from a hook command.
type CodexHookOutput struct {
	Continue           *bool               `json:"continue,omitempty"`
	StopReason         string              `json:"stopReason,omitempty"`
	SystemMessage      string              `json:"systemMessage,omitempty"`
	HookSpecificOutput *codexSpecificOut   `json:"hookSpecificOutput,omitempty"`
}

type codexSpecificOut struct {
	HookEventName            string `json:"hookEventName"`
	AdditionalContext        string `json:"additionalContext,omitempty"`
	PermissionDecision       string `json:"permissionDecision,omitempty"`
	PermissionDecisionReason string `json:"permissionDecisionReason,omitempty"`
	UpdatedToolOutput        any    `json:"updatedToolOutput,omitempty"`
}

func boolPtr(b bool) *bool { return &b }

// DecodeCodexHookInput converts a Codex hook input to a canonical InterceptEvent.
func DecodeCodexHookInput(input CodexHookInput) intercept.InterceptEvent {
	e := intercept.InterceptEvent{
		Host:     "codex",
		Strategy: "hooks",
		Phase:    mapCodexPhase(input.HookEventName),
		Session: intercept.Session{
			ID:             input.SessionID,
			CWD:            input.CWD,
			TranscriptPath: input.TranscriptPath,
		},
		Raw: input,
	}
	if input.ToolName != "" {
		e.Tool = &intercept.Tool{
			Name:   input.ToolName,
			Input:  input.ToolInput,
			Output: input.ToolResponse,
			UseID:  input.ToolUseID,
			IsMCP:  IsMCPTool(input.ToolName),
		}
		if e.Tool.IsMCP {
			e.Tool.MCPServer = MCPServerName(input.ToolName)
		}
	}
	return e
}

// EncodeCodexPreToolResult maps firewall decisions to Codex PreToolUse output.
// Returns output, whether to write, and whether to exit 2 (block).
func EncodeCodexPreToolResult(result intercept.InterceptResult, eventName string) (CodexHookOutput, bool, bool) {
	spec := &codexSpecificOut{HookEventName: eventName}
	cont := boolPtr(true)

	switch result.Kind {
	case intercept.ResultBlock, intercept.ResultReview:
		spec.PermissionDecision = "deny"
		spec.PermissionDecisionReason = result.Reason
		return CodexHookOutput{Continue: cont, HookSpecificOutput: spec}, true, true
	case intercept.ResultWarn:
		if result.Context == "" {
			return CodexHookOutput{}, false, false
		}
		spec.PermissionDecision = "allow"
		spec.AdditionalContext = result.Context
		return CodexHookOutput{Continue: cont, HookSpecificOutput: spec}, true, false
	case intercept.ResultReplaceInput:
		spec.PermissionDecision = "allow"
		return CodexHookOutput{Continue: cont, HookSpecificOutput: spec}, true, false
	default:
		return CodexHookOutput{}, false, false
	}
}

// EncodeCodexResult translates an InterceptResult for PostToolUse / SessionStart.
func EncodeCodexResult(result intercept.InterceptResult, eventName string) (CodexHookOutput, bool) {
	cont := boolPtr(true)
	spec := &codexSpecificOut{HookEventName: eventName}
	hasOutput := false

	if result.Context != "" {
		spec.AdditionalContext = result.Context
		hasOutput = true
	}

	// Codex supports updatedToolOutput for native tools.
	if result.Kind == intercept.ResultReplaceOutput && result.ToolOutput != nil {
		spec.UpdatedToolOutput = result.ToolOutput
		hasOutput = true
	}

	if !hasOutput {
		return CodexHookOutput{}, false
	}
	return CodexHookOutput{Continue: cont, HookSpecificOutput: spec}, true
}

func mapCodexPhase(event string) intercept.Phase {
	switch strings.ToLower(event) {
	case "sessionstart":
		return intercept.PhaseSessionStart
	case "sessionend", "stop":
		return intercept.PhaseSessionEnd
	case "pretooluse":
		return intercept.PhaseToolPre
	case "posttooluse":
		return intercept.PhaseToolPost
	default:
		return intercept.Phase(event)
	}
}

// ── Install ───────────────────────────────────────────────────────────────────

var codexHookEvents = []string{"PreToolUse", "PostToolUse", "SessionStart", "SessionEnd"}

func codexHooksPath(local bool) string {
	if local {
		return filepath.Join(".codex", "hooks.json")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".codex", "hooks.json")
}

// codexEntryJSON builds a Codex hook entry with fixed key order.
func codexEntryJSON(cmd string) json.RawMessage {
	c, _ := json.Marshal(cmd)
	return json.RawMessage(`{"command":` + string(c) + `,"timeout":10}`)
}

func installCodexHooks(binaryPath string, local bool) error {
	p := codexHooksPath(local)
	if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
		return err
	}
	if binaryPath == "" {
		binaryPath, _ = os.Executable()
	}
	hookCmd := `"` + binaryPath + `" hook --host codex`

	top, err := loadOrderedMap(p)
	if err != nil {
		return err
	}
	// Ensure version:1 is present when creating a new file.
	if _, ok := top.get("version"); !ok {
		vb, _ := json.Marshal(1)
		top.set("version", vb)
	}

	entry := codexEntryJSON(hookCmd)
	if err := patchHooksField(top, codexHookEvents, func(_ string, arr []json.RawMessage) []json.RawMessage {
		return dedupAppend(arr, entry)
	}); err != nil {
		return err
	}
	return saveOrderedMap(p, top)
}

func uninstallCodexHooks(local bool) error {
	p := codexHooksPath(local)
	top, err := loadOrderedMap(p)
	if err != nil || len(top.keys) == 0 {
		return nil
	}
	if err := patchHooksField(top, codexHookEvents, func(_ string, arr []json.RawMessage) []json.RawMessage {
		return removeConfire(arr)
	}); err != nil {
		return err
	}
	return saveOrderedMap(p, top)
}

func isCodexHookInstalled(local bool) bool {
	data, err := os.ReadFile(codexHooksPath(local))
	if err != nil {
		return false
	}
	return strings.Contains(string(data), "confire")
}

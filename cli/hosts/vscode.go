package hosts

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/confire-dev/confire/intercept"
)

// ── VSCodeHost ────────────────────────────────────────────────────────────────

type VSCodeHost struct{}

func (h *VSCodeHost) ID() string    { return "vscode" }
func (h *VSCodeHost) Label() string { return "VS Code (Copilot)" }
func (h *VSCodeHost) Strategies() []Strategy { return []Strategy{StrategyHooks} }
func (h *VSCodeHost) Preferred() Strategy    { return StrategyHooks }
func (h *VSCodeHost) ComingSoon() bool       { return false }

func (h *VSCodeHost) Detect() bool {
	home, _ := os.UserHomeDir()
	var paths []string
	switch runtime.GOOS {
	case "darwin":
		paths = []string{
			"/Applications/Visual Studio Code.app",
			filepath.Join(home, "Library", "Application Support", "Code"),
		}
	case "linux":
		paths = []string{
			filepath.Join(home, ".config", "Code"),
			"/usr/share/code",
		}
	case "windows":
		paths = []string{
			filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "Microsoft VS Code"),
			filepath.Join(os.Getenv("APPDATA"), "Code"),
		}
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return true
		}
	}
	return false
}

func (h *VSCodeHost) Install(s Strategy, opts InstallOptions) error {
	switch s {
	case StrategyHooks:
		local := opts.SettingsPath != ""
		return installVSCodeHooks(opts.BinaryPath, local)
	default:
		return fmt.Errorf("vscode: strategy %q not supported", s)
	}
}

func (h *VSCodeHost) IsInstalled(s Strategy) bool {
	if s != StrategyHooks {
		return false
	}
	return isVSCodeHookInstalled(false) || isVSCodeHookInstalled(true)
}

func (h *VSCodeHost) Uninstall(s Strategy) error {
	if s != StrategyHooks {
		return nil
	}
	_ = uninstallVSCodeHooks(false)
	_ = uninstallVSCodeHooks(true)
	return nil
}

// ── VSCode hook types ─────────────────────────────────────────────────────────

// VSCodeHookInput is the JSON Copilot sends to a hook command via stdin.
// Ref: https://code.visualstudio.com/docs/copilot/customization/hooks
type VSCodeHookInput struct {
	Timestamp      string `json:"timestamp"`
	CWD            string `json:"cwd"`
	SessionID      string `json:"sessionId"` // camelCase, unlike Claude Code's session_id
	HookEventName  string `json:"hookEventName"`
	TranscriptPath string `json:"transcript_path"`
	ToolName       string `json:"tool_name"`
	ToolInput      any    `json:"tool_input"`
	ToolUseID      string `json:"tool_use_id"`
	ToolResponse   string `json:"tool_response"` // plain string (not object)
}

// VSCodeHookOutput is what confire returns to Copilot hooks.
// VS Code does NOT support updatedToolOutput — additionalContext only.
type VSCodeHookOutput struct {
	Continue             bool                    `json:"continue"`
	HookSpecificOutput   *vsCodeSpecificOutput   `json:"hookSpecificOutput,omitempty"`
}

type vsCodeSpecificOutput struct {
	HookEventName            string `json:"hookEventName"`
	AdditionalContext        string `json:"additionalContext,omitempty"`
	PermissionDecision       string `json:"permissionDecision,omitempty"`
	PermissionDecisionReason string `json:"permissionDecisionReason,omitempty"`
	UpdatedInput             any    `json:"updatedInput,omitempty"`
}

// IsVSCodeHook returns true when the raw JSON looks like a VS Code Copilot hook.
// VS Code uses camelCase sessionId (not session_id like Claude Code,
// not conversation_id like Cursor).
func IsVSCodeHook(raw map[string]any) bool {
	_, hasSessionID := raw["sessionId"]
	_, noSessionSnake := raw["session_id"]
	_, noConversation := raw["conversation_id"]
	return hasSessionID && !noSessionSnake && !noConversation
}

// DecodeVSCodeHookInput converts a VS Code hook input to a canonical InterceptEvent.
func DecodeVSCodeHookInput(input VSCodeHookInput) intercept.InterceptEvent {
	e := intercept.InterceptEvent{
		Host:     "vscode",
		Strategy: "hooks",
		Phase:    mapVSCodePhase(input.HookEventName),
		Session: intercept.Session{
			ID:             input.SessionID,
			CWD:            input.CWD,
			TranscriptPath: input.TranscriptPath,
		},
		Raw: input,
	}

	if input.ToolName != "" {
		toolName := vscodeCanonicalToolName(input.ToolName)
		e.Tool = &intercept.Tool{
			Name:   toolName,
			Input:  input.ToolInput,
			Output: input.ToolResponse,
			UseID:  input.ToolUseID,
			IsMCP:  IsVSCodeMCPTool(toolName),
		}
		if e.Tool.IsMCP {
			e.Tool.MCPServer = MCPServerName(input.ToolName)
		}
	}
	return e
}

func vscodeCanonicalToolName(name string) string {
	switch strings.ToLower(name) {
	case "runterminalcommand":
		return "Bash"
	default:
		return name
	}
}

// EncodeVSCodePreToolResult maps firewall decisions to VS Code PreToolUse output.
// Ref: https://code.visualstudio.com/docs/copilot/customization/hooks#_pretooluse
func EncodeVSCodePreToolResult(result intercept.InterceptResult, eventName string) (VSCodeHookOutput, bool) {
	out := VSCodeHookOutput{Continue: true}
	spec := vsCodeSpecificOutput{HookEventName: eventName}

	switch result.Kind {
	case intercept.ResultBlock, intercept.ResultReview:
		spec.PermissionDecision = "deny"
		spec.PermissionDecisionReason = result.Reason
	case intercept.ResultWarn:
		if result.Context == "" {
			return VSCodeHookOutput{Continue: true}, false
		}
		spec.PermissionDecision = "allow"
		spec.AdditionalContext = result.Context
	case intercept.ResultReplaceInput:
		spec.PermissionDecision = "allow"
		spec.UpdatedInput = result.ToolInput
	default:
		return VSCodeHookOutput{Continue: true}, false
	}

	out.HookSpecificOutput = &spec
	return out, true
}

// EncodeVSCodeContext wraps text in hookSpecificOutput for SessionStart/PostToolUse.
func EncodeVSCodeContext(context, eventName string) VSCodeHookOutput {
	return VSCodeHookOutput{
		Continue: true,
		HookSpecificOutput: &vsCodeSpecificOutput{
			HookEventName:     eventName,
			AdditionalContext: context,
		},
	}
}

// EncodeVSCodeResult translates an InterceptResult into VS Code hook output.
// VS Code cannot replace tool output — only additionalContext is supported.
func EncodeVSCodeResult(result intercept.InterceptResult, eventName string) (VSCodeHookOutput, bool) {
	out := VSCodeHookOutput{Continue: true}

	ctx := result.Context
	if ctx == "" {
		switch result.Kind {
		case intercept.ResultReplaceOutput, intercept.ResultAddContext:
			// Legacy fallback when daemon did not attach steer context.
			if result.Stats != nil && result.Stats.BeforeBytes > 0 {
				pct := float64(result.Stats.BeforeBytes-result.Stats.AfterBytes) /
					float64(result.Stats.BeforeBytes) * 100
				ctx = fmt.Sprintf("[Confire] Output was %dB → would compress to %dB (%.0f%% smaller with cloud optimizer)",
					result.Stats.BeforeBytes, result.Stats.AfterBytes, pct)
			}
		default:
			return VSCodeHookOutput{Continue: true}, false
		}
	}

	if ctx == "" {
		return VSCodeHookOutput{Continue: true}, false
	}
	out.HookSpecificOutput = &vsCodeSpecificOutput{
		HookEventName:     eventName,
		AdditionalContext: ctx,
	}
	return out, true
}

// IsVSCodeMCPTool returns true for tools using the mcp__ prefix convention.
// VS Code Copilot forwards MCP tool names as-is.
func IsVSCodeMCPTool(name string) bool {
	return strings.HasPrefix(name, "mcp__")
}

func mapVSCodePhase(event string) intercept.Phase {
	switch strings.ToLower(event) {
	case "sessionstart":
		return intercept.PhaseSessionStart
	case "stop":
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

var vsCodeHookEvents = []string{"PreToolUse", "PostToolUse", "SessionStart", "Stop"}

func vsCodeHooksPath(local bool) string {
	if local {
		return filepath.Join(".github", "hooks", "confire.json")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".copilot", "hooks", "confire.json")
}

// vsCodeEntryJSON builds a VS Code hook entry with fixed key order.
func vsCodeEntryJSON(cmd string) json.RawMessage {
	c, _ := json.Marshal(cmd)
	return json.RawMessage(`{"type":"command","command":` + string(c) + `,"timeout":10}`)
}

func installVSCodeHooks(binaryPath string, local bool) error {
	p := vsCodeHooksPath(local)
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return err
	}
	if binaryPath == "" {
		binaryPath, _ = os.Executable()
	}
	hookCmd := `"` + binaryPath + `" hook`

	top, err := loadOrderedMap(p)
	if err != nil {
		return err
	}
	entry := vsCodeEntryJSON(hookCmd)
	if err := patchHooksField(top, vsCodeHookEvents, func(_ string, arr []json.RawMessage) []json.RawMessage {
		return dedupAppend(arr, entry)
	}); err != nil {
		return err
	}
	return saveOrderedMap(p, top)
}

func uninstallVSCodeHooks(local bool) error {
	p := vsCodeHooksPath(local)
	top, err := loadOrderedMap(p)
	if err != nil || len(top.keys) == 0 {
		return nil
	}
	if err := patchHooksField(top, vsCodeHookEvents, func(_ string, arr []json.RawMessage) []json.RawMessage {
		return removeConfire(arr)
	}); err != nil {
		return err
	}
	return saveOrderedMap(p, top)
}

func isVSCodeHookInstalled(local bool) bool {
	data, err := os.ReadFile(vsCodeHooksPath(local))
	if err != nil {
		return false
	}
	return strings.Contains(string(data), "confire")
}

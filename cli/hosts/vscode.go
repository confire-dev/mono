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
func (h *VSCodeHost) ComingSoon() bool       { return true }

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
	HookEventName     string `json:"hookEventName"`
	AdditionalContext string `json:"additionalContext,omitempty"`
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
		e.Tool = &intercept.Tool{
			Name:   input.ToolName,
			Input:  input.ToolInput,
			Output: input.ToolResponse,
			UseID:  input.ToolUseID,
			IsMCP:  IsVSCodeMCPTool(input.ToolName),
		}
		if e.Tool.IsMCP {
			e.Tool.MCPServer = MCPServerName(input.ToolName)
		}
	}
	return e
}

// EncodeVSCodeResult translates an InterceptResult into VS Code hook output.
// VS Code cannot replace tool output — only additionalContext is supported.
func EncodeVSCodeResult(result intercept.InterceptResult, eventName string) (VSCodeHookOutput, bool) {
	out := VSCodeHookOutput{Continue: true}

	var ctx string
	switch result.Kind {
	case intercept.ResultReplaceOutput:
		// Can't replace output in VS Code — inject savings as context instead.
		if result.Stats != nil && result.Stats.BeforeBytes > 0 {
			pct := float64(result.Stats.BeforeBytes-result.Stats.AfterBytes) /
				float64(result.Stats.BeforeBytes) * 100
			ctx = fmt.Sprintf("[Confire] Output was %dB → would compress to %dB (%.0f%% smaller with cloud optimizer)",
				result.Stats.BeforeBytes, result.Stats.AfterBytes, pct)
		}
	case intercept.ResultAddContext:
		ctx = result.Context
	default:
		return VSCodeHookOutput{Continue: true}, false
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
	switch event {
	case "SessionStart":  return intercept.PhaseSessionStart
	case "Stop":          return intercept.PhaseSessionEnd
	case "PreToolUse":    return intercept.PhaseToolPre
	case "PostToolUse":   return intercept.PhaseToolPost
	default:              return intercept.Phase(event)
	}
}

// ── Install ───────────────────────────────────────────────────────────────────

func vsCodeHooksPath(local bool) string {
	if local {
		return filepath.Join(".github", "hooks", "confire.json")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".copilot", "hooks", "confire.json")
}

type vsCodeHookEntry struct {
	Type    string `json:"type"`
	Command string `json:"command"`
	Timeout int    `json:"timeout,omitempty"`
}

type vsCodeHooksFile struct {
	Hooks map[string][]vsCodeHookEntry `json:"hooks"`
}

func installVSCodeHooks(binaryPath string, local bool) error {
	path := vsCodeHooksPath(local)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	if binaryPath == "" {
		binaryPath, _ = os.Executable()
	}
	hookCmd := `"` + binaryPath + `" hook`

	current := vsCodeHooksFile{Hooks: map[string][]vsCodeHookEntry{}}
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &current)
	}
	if current.Hooks == nil {
		current.Hooks = map[string][]vsCodeHookEntry{}
	}

	for _, event := range []string{"PostToolUse", "SessionStart"} {
		kept := current.Hooks[event][:0]
		for _, e := range current.Hooks[event] {
			if !strings.Contains(e.Command, "confire") {
				kept = append(kept, e)
			}
		}
		current.Hooks[event] = append(kept, vsCodeHookEntry{
			Type:    "command",
			Command: hookCmd,
			Timeout: 10,
		})
	}

	data, err := json.MarshalIndent(current, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0644)
}

func uninstallVSCodeHooks(local bool) error {
	path := vsCodeHooksPath(local)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var current vsCodeHooksFile
	if err := json.Unmarshal(data, &current); err != nil {
		return nil
	}
	for event, entries := range current.Hooks {
		kept := entries[:0]
		for _, e := range entries {
			if !strings.Contains(e.Command, "confire") {
				kept = append(kept, e)
			}
		}
		current.Hooks[event] = kept
	}
	out, err := json.MarshalIndent(current, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(out, '\n'), 0644)
}

func isVSCodeHookInstalled(local bool) bool {
	data, err := os.ReadFile(vsCodeHooksPath(local))
	if err != nil {
		return false
	}
	return strings.Contains(string(data), "confire")
}

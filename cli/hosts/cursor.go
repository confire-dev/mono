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

// ── CursorHost ────────────────────────────────────────────────────────────────

type CursorHost struct{}

func (h *CursorHost) ID() string    { return "cursor" }
func (h *CursorHost) Label() string { return "Cursor" }
func (h *CursorHost) Strategies() []Strategy { return []Strategy{StrategyHooks, StrategyMCPProxy} }
func (h *CursorHost) Preferred() Strategy    { return StrategyHooks }
func (h *CursorHost) ComingSoon() bool       { return false }

func (h *CursorHost) Detect() bool {
	home, _ := os.UserHomeDir()
	switch runtime.GOOS {
	case "darwin":
		if _, err := os.Stat("/Applications/Cursor.app"); err == nil {
			return true
		}
	case "linux":
		if _, err := os.Stat(filepath.Join(home, ".cursor")); err == nil {
			return true
		}
	case "windows":
		if _, err := os.Stat(filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "Cursor")); err == nil {
			return true
		}
	}
	_, err := os.Stat(filepath.Join(home, ".cursor", "mcp.json"))
	return err == nil
}

func (h *CursorHost) Install(s Strategy, opts InstallOptions) error {
	switch s {
	case StrategyHooks:
		local := opts.SettingsPath != ""
		return installCursorHooks(opts.BinaryPath, local)
	default:
		return fmt.Errorf("cursor: strategy %q not supported", s)
	}
}

func (h *CursorHost) IsInstalled(s Strategy) bool {
	if s != StrategyHooks {
		return false
	}
	return isCursorHookInstalled(false) || isCursorHookInstalled(true)
}

func (h *CursorHost) Uninstall(s Strategy) error {
	if s != StrategyHooks {
		return nil
	}
	_ = uninstallCursorHooks(false)
	_ = uninstallCursorHooks(true)
	return nil
}

// CursorHookInput is the JSON structure Cursor sends to a hook command via stdin.
// Ref: https://cursor.com/docs/hooks#command-based-hooks
type CursorHookInput struct {
	// Common fields (all events)
	ConversationID string   `json:"conversation_id"`
	GenerationID   string   `json:"generation_id"`
	HookEventName  string   `json:"hook_event_name"`
	Model          string   `json:"model"`
	CursorVersion  string   `json:"cursor_version"`
	WorkspaceRoots []string `json:"workspace_roots"`
	TranscriptPath string   `json:"transcript_path"`
	UserEmail      string   `json:"user_email"`
	// preToolUse / postToolUse
	ToolName   string `json:"tool_name"`
	ToolInput  any    `json:"tool_input"`
	ToolOutput string `json:"tool_output"` // JSON-stringified for MCP tools
	ToolUseID  string `json:"tool_use_id"`
	CWD        string `json:"cwd"`
	Duration   int    `json:"duration"`
}

// CursorHookOutput is what confire writes to stdout for Cursor postToolUse.
// updated_mcp_tool_output replaces the output the model sees (MCP tools only).
// additional_context is injected into the conversation after the tool result.
type CursorHookOutput struct {
	UpdatedMCPToolOutput any    `json:"updated_mcp_tool_output,omitempty"`
	AdditionalContext    string `json:"additional_context,omitempty"`
}

// IsCursorHook returns true when the raw decoded JSON looks like a Cursor hook
// input (has conversation_id) rather than a Claude Code hook (has session_id).
func IsCursorHook(raw map[string]any) bool {
	_, ok := raw["conversation_id"]
	return ok
}

// DecodeCursorHookInput converts a Cursor hook input into a canonical InterceptEvent.
func DecodeCursorHookInput(input CursorHookInput) intercept.InterceptEvent {
	cwd := input.CWD
	if cwd == "" && len(input.WorkspaceRoots) > 0 {
		cwd = input.WorkspaceRoots[0]
	}

	e := intercept.InterceptEvent{
		Host:     "cursor",
		Strategy: "hooks",
		Phase:    mapCursorPhase(input.HookEventName),
		Session: intercept.Session{
			ID:             input.ConversationID,
			CWD:            cwd,
			TranscriptPath: input.TranscriptPath,
		},
		Raw: input,
	}

	if input.ToolName != "" {
		// tool_output is a JSON string — unmarshal it so optimizers see a real object.
		var output any = input.ToolOutput
		if input.ToolOutput != "" {
			var parsed any
			if err := json.Unmarshal([]byte(input.ToolOutput), &parsed); err == nil {
				output = parsed
			}
		}

		// In Cursor, postToolUse is fired for MCP tools; native tools (shell, file)
		// have their own hooks (afterShellExecution, afterFileEdit) which are
		// observational-only. So everything reaching postToolUse here is MCP.
		isMCP := strings.EqualFold(input.HookEventName, "postToolUse")

		e.Tool = &intercept.Tool{
			Name:   input.ToolName,
			Input:  input.ToolInput,
			Output: output,
			UseID:  input.ToolUseID,
			IsMCP:  isMCP,
		}
		if isMCP {
			e.Tool.MCPServer = cursorMCPServer(input.ToolName)
		}
	}
	return e
}

// EncodeCursorResult converts an InterceptResult into the Cursor hook output format.
func EncodeCursorResult(result intercept.InterceptResult) (CursorHookOutput, bool) {
	switch result.Kind {
	case intercept.ResultReplaceOutput:
		out := CursorHookOutput{UpdatedMCPToolOutput: result.ToolOutput}
		if result.Context != "" {
			out.AdditionalContext = result.Context
		}
		return out, true
	case intercept.ResultAddContext:
		return CursorHookOutput{AdditionalContext: result.Context}, true
	default:
		return CursorHookOutput{}, false
	}
}

func mapCursorPhase(event string) intercept.Phase {
	switch strings.ToLower(event) {
	case "sessionstart":
		return intercept.PhaseSessionStart
	case "sessionend":
		return intercept.PhaseSessionEnd
	case "pretooluse":
		return intercept.PhaseToolPre
	case "posttooluse":
		return intercept.PhaseToolPost
	default:
		return intercept.Phase(event)
	}
}

func cursorMCPServer(toolName string) string {
	lower := strings.ToLower(toolName)
	for _, k := range []string{"figma", "github", "slack", "linear", "jira", "confluence", "notion", "clickup", "amplitude"} {
		if strings.HasPrefix(lower, k) || strings.Contains(lower, k+"_") {
			return k
		}
	}
	return ""
}

// ── Install ───────────────────────────────────────────────────────────────────

func cursorHooksPath(local bool) string {
	if local {
		return filepath.Join(".cursor", "hooks.json")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cursor", "hooks.json")
}

func installCursorHooks(binaryPath string, local bool) error {
	return installCursorHooksAt(cursorHooksPath(local), binaryPath)
}

func uninstallCursorHooks(local bool) error {
	return uninstallCursorHooksAt(cursorHooksPath(local))
}

func isCursorHookInstalled(local bool) bool {
	data, err := os.ReadFile(cursorHooksPath(local))
	if err != nil {
		return false
	}
	return strings.Contains(string(data), "confire")
}

type cursorHookEntry struct {
	Command string `json:"command"`
}

type cursorHooksFile struct {
	Version int                          `json:"version"`
	Hooks   map[string][]cursorHookEntry `json:"hooks"`
}

func installCursorHooksAt(settingsPath, binaryPath string) error {
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0700); err != nil {
		return err
	}
	if binaryPath == "" {
		binaryPath, _ = os.Executable()
	}
	hookCmd := `"` + binaryPath + `" hook`

	current := cursorHooksFile{Version: 1, Hooks: map[string][]cursorHookEntry{}}
	if data, err := os.ReadFile(settingsPath); err == nil {
		json.Unmarshal(data, &current)
	}
	if current.Hooks == nil {
		current.Hooks = map[string][]cursorHookEntry{}
	}

	for _, event := range []string{"postToolUse", "sessionStart", "sessionEnd"} {
		kept := current.Hooks[event][:0]
		for _, e := range current.Hooks[event] {
			if !strings.Contains(e.Command, "confire") {
				kept = append(kept, e)
			}
		}
		current.Hooks[event] = append(kept, cursorHookEntry{Command: hookCmd})
	}

	data, err := json.MarshalIndent(current, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(settingsPath, append(data, '\n'), 0644)
}

func uninstallCursorHooksAt(settingsPath string) error {
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return nil // nothing to remove
	}
	var current cursorHooksFile
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
	return os.WriteFile(settingsPath, append(out, '\n'), 0644)
}

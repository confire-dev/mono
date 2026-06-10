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
	SessionID      string   `json:"session_id"`
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

// CursorHookOutput is what confire writes to stdout for Cursor postToolUse / sessionStart.
// updated_mcp_tool_output replaces MCP output only; native tools cannot be replaced.
type CursorHookOutput struct {
	UpdatedMCPToolOutput any    `json:"updated_mcp_tool_output,omitempty"`
	AdditionalContext    string `json:"additional_context,omitempty"`
}

// CursorPreToolOutput is returned for preToolUse (firewall + Read limit injection).
type CursorPreToolOutput struct {
	Permission   string `json:"permission,omitempty"`
	UserMessage  string `json:"user_message,omitempty"`
	AgentMessage string `json:"agent_message,omitempty"`
	UpdatedInput any    `json:"updated_input,omitempty"`
}

// IsCursorHook returns true when the raw decoded JSON looks like a Cursor hook input.
func IsCursorHook(raw map[string]any) bool {
	_, ok := raw["cursor_version"]
	return ok
}

// DecodeCursorHookInput converts a Cursor hook input into a canonical InterceptEvent.
func DecodeCursorHookInput(input CursorHookInput) intercept.InterceptEvent {
	cwd := input.CWD
	if cwd == "" && len(input.WorkspaceRoots) > 0 {
		cwd = input.WorkspaceRoots[0]
	}

	sessionID := input.ConversationID
	if sessionID == "" {
		sessionID = input.SessionID
	}

	e := intercept.InterceptEvent{
		Host:     "cursor",
		Strategy: "hooks",
		Phase:    mapCursorPhase(input.HookEventName),
		Session: intercept.Session{
			ID:             sessionID,
			CWD:            cwd,
			TranscriptPath: input.TranscriptPath,
		},
		Raw: input,
	}

	if input.ToolName != "" {
		toolName := cursorCanonicalToolName(input.ToolName)

		// tool_output is a JSON string — unmarshal it so optimizers see a real object.
		var output any = input.ToolOutput
		if input.ToolOutput != "" {
			var parsed any
			if err := json.Unmarshal([]byte(input.ToolOutput), &parsed); err == nil {
				output = parsed
			}
		}

		isMCP := cursorIsMCPTool(input.ToolName)

		e.Tool = &intercept.Tool{
			Name:   toolName,
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

func cursorCanonicalToolName(name string) string {
	if strings.EqualFold(name, "Shell") {
		return "Bash"
	}
	return name
}

func cursorIsMCPTool(toolName string) bool {
	return strings.HasPrefix(toolName, "MCP:") || strings.HasPrefix(toolName, "mcp:")
}

// EncodeCursorPreToolResult converts firewall / Read-limit results for preToolUse.
// The second return value is true when the hook should exit with code 2.
func EncodeCursorPreToolResult(result intercept.InterceptResult) (CursorPreToolOutput, bool, bool) {
	switch result.Kind {
	case intercept.ResultBlock, intercept.ResultReview:
		return CursorPreToolOutput{
			Permission:   "deny",
			UserMessage:  result.Reason,
			AgentMessage: result.Reason,
		}, true, true
	case intercept.ResultWarn:
		if result.Context == "" {
			return CursorPreToolOutput{}, false, false
		}
		return CursorPreToolOutput{
			Permission:   "allow",
			AgentMessage: result.Context,
		}, true, false
	case intercept.ResultReplaceInput:
		return CursorPreToolOutput{
			Permission:   "allow",
			UpdatedInput: result.ToolInput,
		}, true, false
	default:
		return CursorPreToolOutput{}, false, false
	}
}

// EncodeCursorResult converts an InterceptResult into the Cursor hook output format.
func EncodeCursorResult(result intercept.InterceptResult, toolName string) (CursorHookOutput, bool) {
	out := CursorHookOutput{}
	hasOutput := false

	if result.Kind == intercept.ResultReplaceOutput && cursorIsMCPTool(toolName) {
		out.UpdatedMCPToolOutput = result.ToolOutput
		hasOutput = true
	}

	if result.Context != "" {
		out.AdditionalContext = result.Context
		hasOutput = true
	}

	return out, hasOutput
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
	// Cursor colon format: "MCP: server/tool" or "mcp:server/tool"
	if idx := strings.Index(lower, ":"); idx >= 0 {
		rest := strings.TrimSpace(lower[idx+1:])
		if slash := strings.Index(rest, "/"); slash >= 0 {
			return rest[:slash]
		}
		if rest != "" {
			return rest
		}
	}
	// Fallback: canonical mcp__server__tool format
	for _, k := range []string{"figma", "github", "slack", "linear", "jira", "confluence", "notion", "clickup", "amplitude"} {
		if strings.HasPrefix(lower, k) || strings.Contains(lower, k+"_") {
			return k
		}
	}
	return ""
}

// ── Install ───────────────────────────────────────────────────────────────────

var cursorHookEvents = []string{"preToolUse", "postToolUse", "sessionStart", "sessionEnd"}

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

// cursorEntryJSON builds a minimal Cursor hook entry with a fixed key order.
func cursorEntryJSON(cmd string) json.RawMessage {
	c, _ := json.Marshal(cmd)
	return json.RawMessage(`{"command":` + string(c) + `}`)
}

func installCursorHooksAt(settingsPath, binaryPath string) error {
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0700); err != nil {
		return err
	}
	if binaryPath == "" {
		binaryPath, _ = os.Executable()
	}
	hookCmd := `"` + binaryPath + `" hook`

	top, err := loadOrderedMap(settingsPath)
	if err != nil {
		return err
	}
	// Ensure version:1 is present when creating a new file.
	if _, ok := top.get("version"); !ok {
		vb, _ := json.Marshal(1)
		top.set("version", vb)
	}

	entry := cursorEntryJSON(hookCmd)
	if err := patchHooksField(top, cursorHookEvents, func(_ string, arr []json.RawMessage) []json.RawMessage {
		return dedupAppend(arr, entry)
	}); err != nil {
		return err
	}
	return saveOrderedMap(settingsPath, top)
}

func uninstallCursorHooksAt(settingsPath string) error {
	top, err := loadOrderedMap(settingsPath)
	if err != nil || len(top.keys) == 0 {
		return nil
	}
	if err := patchHooksField(top, cursorHookEvents, func(_ string, arr []json.RawMessage) []json.RawMessage {
		return removeConfire(arr)
	}); err != nil {
		return err
	}
	return saveOrderedMap(settingsPath, top)
}

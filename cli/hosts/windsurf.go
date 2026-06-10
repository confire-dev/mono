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

// ── WindsurfHost ──────────────────────────────────────────────────────────────

type WindsurfHost struct{}

func (h *WindsurfHost) ID() string               { return "windsurf" }
func (h *WindsurfHost) Label() string            { return "Windsurf" }
func (h *WindsurfHost) Strategies() []Strategy   { return []Strategy{StrategyHooks} }
func (h *WindsurfHost) Preferred() Strategy      { return StrategyHooks }
func (h *WindsurfHost) ComingSoon() bool         { return false }

func (h *WindsurfHost) Detect() bool {
	home, _ := os.UserHomeDir()
	switch runtime.GOOS {
	case "darwin":
		if _, err := os.Stat("/Applications/Windsurf.app"); err == nil {
			return true
		}
	case "linux":
		if _, err := os.Stat(filepath.Join(home, ".windsurf")); err == nil {
			return true
		}
	case "windows":
		if _, err := os.Stat(filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "Windsurf")); err == nil {
			return true
		}
	}
	// Fallback: config directory present.
	_, err := os.Stat(filepath.Join(home, ".windsurf"))
	return err == nil
}

func (h *WindsurfHost) Install(s Strategy, opts InstallOptions) error {
	if s != StrategyHooks {
		return fmt.Errorf("windsurf: strategy %q not supported", s)
	}
	local := opts.SettingsPath != ""
	return installWindsurfHooks(opts.BinaryPath, local)
}

func (h *WindsurfHost) IsInstalled(s Strategy) bool {
	if s != StrategyHooks {
		return false
	}
	return isWindsurfHookInstalled(false) || isWindsurfHookInstalled(true)
}

func (h *WindsurfHost) Uninstall(s Strategy) error {
	if s != StrategyHooks {
		return nil
	}
	_ = uninstallWindsurfHooks(false)
	_ = uninstallWindsurfHooks(true)
	return nil
}

// ── Windsurf hook types ───────────────────────────────────────────────────────

// WindsurfHookInput is the JSON Windsurf sends to a hook command via stdin.
// Ref: https://docs.devin.ai/cli/extensibility/hooks/overview
type WindsurfHookInput struct {
	HookEventName  string `json:"hook_event_name"`
	ToolName       string `json:"tool_name"`
	ToolInput      any    `json:"tool_input"`
	ToolOutput     any    `json:"tool_output"`
	SessionID      string `json:"session_id"`
	CWD            string `json:"cwd"`
	TranscriptPath string `json:"transcript_path,omitempty"`
}

// WindsurfHookOutput is returned to Windsurf for all hook phases.
// Windsurf uses a decision/reason envelope; no output replacement supported.
type WindsurfHookOutput struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason,omitempty"`
}

// DecodeWindsurfHookInput converts a Windsurf hook input to a canonical InterceptEvent.
func DecodeWindsurfHookInput(input WindsurfHookInput) intercept.InterceptEvent {
	e := intercept.InterceptEvent{
		Host:     "windsurf",
		Strategy: "hooks",
		Phase:    mapWindsurfPhase(input.HookEventName),
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
			Output: input.ToolOutput,
			IsMCP:  IsMCPTool(input.ToolName),
		}
		if e.Tool.IsMCP {
			e.Tool.MCPServer = MCPServerName(input.ToolName)
		}
	}
	return e
}

// EncodeWindsurfPreToolResult maps firewall decisions to Windsurf PreToolUse output.
// Returns the output, whether to write it, and whether to exit 2 (block).
func EncodeWindsurfPreToolResult(result intercept.InterceptResult) (WindsurfHookOutput, bool, bool) {
	switch result.Kind {
	case intercept.ResultBlock, intercept.ResultReview:
		return WindsurfHookOutput{Decision: "block", Reason: result.Reason}, true, true
	case intercept.ResultWarn:
		if result.Context == "" {
			return WindsurfHookOutput{}, false, false
		}
		return WindsurfHookOutput{Decision: "approve", Reason: result.Context}, true, false
	default:
		return WindsurfHookOutput{}, false, false
	}
}

// EncodeWindsurfResult translates an InterceptResult for PostToolUse.
// Windsurf PostToolUse only supports decision/reason — no output replacement.
func EncodeWindsurfResult(result intercept.InterceptResult) (WindsurfHookOutput, bool) {
	if result.Context == "" {
		return WindsurfHookOutput{}, false
	}
	return WindsurfHookOutput{Decision: "approve", Reason: result.Context}, true
}

func mapWindsurfPhase(event string) intercept.Phase {
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

// ── Install ───────────────────────────────────────────────────────────────────

// Windsurf's hooks.json is a flat object: the file IS the hooks map.
// There is no "hooks" wrapper key — event names are top-level keys.
var windsurfHookEvents = []string{"PreToolUse", "PostToolUse", "SessionStart", "SessionEnd"}

func windsurfHooksPath(local bool) string {
	if local {
		return filepath.Join(".windsurf", "hooks.json")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".windsurf", "hooks.json")
}

// windsurfEntryJSON builds a Windsurf hook entry with fixed key order.
func windsurfEntryJSON(cmd string) json.RawMessage {
	c, _ := json.Marshal(cmd)
	return json.RawMessage(`{"command":` + string(c) + `,"timeout":10}`)
}

func installWindsurfHooks(binaryPath string, local bool) error {
	p := windsurfHooksPath(local)
	if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
		return err
	}
	if binaryPath == "" {
		binaryPath, _ = os.Executable()
	}
	hookCmd := `"` + binaryPath + `" hook --host windsurf`

	top, err := loadOrderedMap(p)
	if err != nil {
		return err
	}
	entry := windsurfEntryJSON(hookCmd)
	patchFlatHooksField(top, windsurfHookEvents, func(_ string, arr []json.RawMessage) []json.RawMessage {
		return dedupAppend(arr, entry)
	})
	return saveOrderedMap(p, top)
}

func uninstallWindsurfHooks(local bool) error {
	p := windsurfHooksPath(local)
	top, err := loadOrderedMap(p)
	if err != nil || len(top.keys) == 0 {
		return nil
	}
	patchFlatHooksField(top, windsurfHookEvents, func(_ string, arr []json.RawMessage) []json.RawMessage {
		return removeConfire(arr)
	})
	return saveOrderedMap(p, top)
}

func isWindsurfHookInstalled(local bool) bool {
	data, err := os.ReadFile(windsurfHooksPath(local))
	if err != nil {
		return false
	}
	return strings.Contains(string(data), "confire")
}

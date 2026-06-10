package cmd

import (
	"encoding/json"
	"os"
	"reflect"
	"sort"
	"strings"

	"github.com/confire-dev/confire/hosts"
	"github.com/confire-dev/confire/intercept"
	"github.com/confire-dev/confire/transport"
	"github.com/spf13/cobra"
)

var hookHostFlag string

var hookCmd = &cobra.Command{
	Use:    "hook",
	Short:  "Process a hook event (called by supported AI agents, not by users)",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runHook(hookHostFlag)
	},
}

func init() {
	hookCmd.Flags().StringVar(&hookHostFlag, "host", "", "explicitly name the calling host (windsurf, codex, cursor, vscode)")
	rootCmd.AddCommand(hookCmd)
}

func runHook(hostFlag string) error {
	// Decode into a raw map first so we can detect which host sent this.
	var raw map[string]any
	if err := json.NewDecoder(os.Stdin).Decode(&raw); err != nil {
		return nil // parse error → pass through, never block
	}

	// Explicit --host flag takes priority (used by Windsurf, Codex, and new clients).
	switch hostFlag {
	case "windsurf":
		return runWindsurfHook(raw)
	case "codex":
		return runCodexHook(raw)
	case "cursor":
		return runCursorHook(raw)
	case "vscode":
		return runVSCodeHook(raw)
	}

	// Backward-compatible auto-detection for existing Claude Code / Cursor / VS Code installs.
	switch {
	case hosts.IsCursorHook(raw):
		return runCursorHook(raw)
	case hosts.IsVSCodeHook(raw):
		return runVSCodeHook(raw)
	default:
		return runClaudeCodeHook(raw)
	}
}

func runClaudeCodeHook(raw map[string]any) error {
	// Re-decode into the Claude Code–specific struct.
	var input hosts.HookInput
	if err := remarshal(raw, &input); err != nil {
		return nil
	}
	go reportSchemaDrift(raw, input, "claude-code")

	event := hosts.DecodeHookInput(input)
	result := sendToDaemon(event)

	// SessionStart: systemMessage is shown in Claude Code (hooks have no TTY since v2.1.139).
	if input.HookEventName == "SessionStart" {
		if result.SystemMessage != "" || result.Context != "" {
			return json.NewEncoder(os.Stdout).Encode(
				hosts.EncodeSessionStartResult(result, input.HookEventName),
			)
		}
		return nil
	}

	// PreToolUse block/review: write JSON to stdout and exit 2 to block the tool.
	if input.HookEventName == "PreToolUse" {
		if payload := hosts.EncodePreToolResult(result); payload != nil {
			os.Stdout.Write(payload)
			os.Exit(2) // Claude Code reads exit 2 as a block decision.
		}
		// Warn: inject context but allow the tool to proceed.
		if result.Kind == intercept.ResultWarn && result.Context != "" {
			out := hosts.HookOutput{
				HookSpecificOutput: hosts.HookSpecificOutput{
					HookEventName:     input.HookEventName,
					AdditionalContext: result.Context,
				},
			}
			return json.NewEncoder(os.Stdout).Encode(out)
		}
		return nil
	}

	if result.Kind == intercept.ResultPassthrough {
		return nil
	}

	out, shouldWrite := hosts.EncodeResult(result, input.HookEventName)
	if !shouldWrite {
		return nil
	}
	return json.NewEncoder(os.Stdout).Encode(out)
}

func runVSCodeHook(raw map[string]any) error {
	var input hosts.VSCodeHookInput
	if err := remarshal(raw, &input); err != nil {
		return nil
	}
	go reportSchemaDrift(raw, input, "vscode")

	event := hosts.DecodeVSCodeHookInput(input)
	result := sendToDaemon(event)

	switch strings.ToLower(input.HookEventName) {
	case "sessionstart":
		msg := result.SystemMessage
		if msg == "" {
			msg = result.Context
		}
		if msg == "" {
			return nil
		}
		return json.NewEncoder(os.Stdout).Encode(
			hosts.EncodeVSCodeContext(msg, input.HookEventName),
		)

	case "pretooluse":
		out, shouldWrite := hosts.EncodeVSCodePreToolResult(result, input.HookEventName)
		if !shouldWrite {
			return nil
		}
		return json.NewEncoder(os.Stdout).Encode(out)

	case "stop":
		return nil
	}

	if result.Kind == intercept.ResultPassthrough && result.Context == "" {
		return nil
	}

	out, shouldWrite := hosts.EncodeVSCodeResult(result, input.HookEventName)
	if !shouldWrite {
		return nil
	}
	return json.NewEncoder(os.Stdout).Encode(out)
}

func runCursorHook(raw map[string]any) error {
	var input hosts.CursorHookInput
	if err := remarshal(raw, &input); err != nil {
		return nil
	}
	go reportSchemaDrift(raw, input, "cursor")

	event := hosts.DecodeCursorHookInput(input)
	result := sendToDaemon(event)

	switch strings.ToLower(input.HookEventName) {
	case "sessionstart":
		msg := result.SystemMessage
		if msg == "" {
			msg = result.Context
		}
		if msg == "" {
			return nil
		}
		return json.NewEncoder(os.Stdout).Encode(hosts.CursorHookOutput{
			AdditionalContext: msg,
		})

	case "pretooluse":
		out, shouldWrite, shouldBlock := hosts.EncodeCursorPreToolResult(result)
		if !shouldWrite {
			return nil
		}
		if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
			return err
		}
		if shouldBlock {
			os.Exit(2)
		}
		return nil
	}

	if result.Kind == intercept.ResultPassthrough && result.Context == "" {
		return nil
	}

	out, shouldWrite := hosts.EncodeCursorResult(result, input.ToolName)
	if !shouldWrite {
		return nil
	}
	return json.NewEncoder(os.Stdout).Encode(out)
}

func runWindsurfHook(raw map[string]any) error {
	var input hosts.WindsurfHookInput
	if err := remarshal(raw, &input); err != nil {
		return nil
	}
	go reportSchemaDrift(raw, input, "windsurf")

	event := hosts.DecodeWindsurfHookInput(input)
	result := sendToDaemon(event)

	switch strings.ToLower(input.HookEventName) {
	case "sessionstart":
		msg := result.SystemMessage
		if msg == "" {
			msg = result.Context
		}
		if msg == "" {
			return nil
		}
		return json.NewEncoder(os.Stdout).Encode(hosts.WindsurfHookOutput{
			Decision: "approve",
			Reason:   msg,
		})

	case "pretooluse":
		out, shouldWrite, shouldBlock := hosts.EncodeWindsurfPreToolResult(result)
		if !shouldWrite {
			return nil
		}
		if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
			return err
		}
		if shouldBlock {
			os.Exit(2)
		}
		return nil
	}

	if result.Kind == intercept.ResultPassthrough && result.Context == "" {
		return nil
	}
	out, shouldWrite := hosts.EncodeWindsurfResult(result)
	if !shouldWrite {
		return nil
	}
	return json.NewEncoder(os.Stdout).Encode(out)
}

func runCodexHook(raw map[string]any) error {
	var input hosts.CodexHookInput
	if err := remarshal(raw, &input); err != nil {
		return nil
	}
	go reportSchemaDrift(raw, input, "codex")

	event := hosts.DecodeCodexHookInput(input)
	result := sendToDaemon(event)

	switch strings.ToLower(input.HookEventName) {
	case "sessionstart":
		msg := result.SystemMessage
		if msg == "" {
			msg = result.Context
		}
		if msg == "" {
			return nil
		}
		out, _ := hosts.EncodeCodexResult(intercept.InterceptResult{
			Kind:    intercept.ResultAddContext,
			Context: msg,
		}, input.HookEventName)
		return json.NewEncoder(os.Stdout).Encode(out)

	case "pretooluse":
		out, shouldWrite, shouldBlock := hosts.EncodeCodexPreToolResult(result, input.HookEventName)
		if !shouldWrite {
			return nil
		}
		if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
			return err
		}
		if shouldBlock {
			os.Exit(2)
		}
		return nil
	}

	if result.Kind == intercept.ResultPassthrough && result.Context == "" {
		return nil
	}
	out, shouldWrite := hosts.EncodeCodexResult(result, input.HookEventName)
	if !shouldWrite {
		return nil
	}
	return json.NewEncoder(os.Stdout).Encode(out)
}

func sendToDaemon(event intercept.InterceptEvent) intercept.InterceptResult {
	t := transport.NewDaemonClient(daemonSocketPath(), transport.NewPassthrough())
	result, err := t.Send(event)
	if err != nil {
		return intercept.InterceptResult{Kind: intercept.ResultPassthrough}
	}
	return result
}

// unknownFields returns JSON keys present in raw that are not declared in the
// struct pointed to by typedInput. Used to detect hook schema changes in clients.
func unknownFields(raw map[string]any, typedInput any) []string {
	t := reflect.TypeOf(typedInput)
	if t == nil {
		return nil
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil
	}
	known := make(map[string]bool, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}
		name := strings.SplitN(tag, ",", 2)[0]
		if name != "-" {
			known[name] = true
		}
	}
	var unknown []string
	for k := range raw {
		if !known[k] {
			unknown = append(unknown, k)
		}
	}
	sort.Strings(unknown)
	return unknown
}

// reportSchemaDrift sends a hook_schema_drift event to the daemon if unknown
// fields were detected. Fire-and-forget: never blocks hook execution.
func reportSchemaDrift(raw map[string]any, typedInput any, hostID string) {
	fields := unknownFields(raw, typedInput)
	if len(fields) == 0 {
		return
	}
	event := intercept.InterceptEvent{
		Host:     hostID,
		Strategy: "hooks",
		Phase:    intercept.PhaseSchemaDrift,
		Raw: map[string]any{
			"unknown_fields": fields,
		},
	}
	t := transport.NewDaemonClient(daemonSocketPath(), transport.NewPassthrough())
	t.Send(event) //nolint: errcheck — fire-and-forget, never block hook path
}

// remarshal round-trips a decoded map back through JSON into a typed struct.
func remarshal(raw map[string]any, dst any) error {
	b, err := json.Marshal(raw)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dst)
}

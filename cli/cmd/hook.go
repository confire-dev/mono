package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/confire-dev/confire/hosts"
	"github.com/confire-dev/confire/intercept"
	"github.com/confire-dev/confire/transport"
	"github.com/spf13/cobra"
)

var hookCmd = &cobra.Command{
	Use:    "hook",
	Short:  "Process a hook event (called by Claude Code or Cursor, not by users)",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runHook()
	},
}

func init() {
	rootCmd.AddCommand(hookCmd)
}

func runHook() error {
	// Decode into a raw map first so we can detect which host sent this.
	var raw map[string]any
	if err := json.NewDecoder(os.Stdin).Decode(&raw); err != nil {
		return nil // parse error → pass through, never block
	}

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
	logSavings(input.ToolName, result)
	return json.NewEncoder(os.Stdout).Encode(out)
}

func runVSCodeHook(raw map[string]any) error {
	var input hosts.VSCodeHookInput
	if err := remarshal(raw, &input); err != nil {
		return nil
	}

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
	logSavings(input.ToolName, result)
	return json.NewEncoder(os.Stdout).Encode(out)
}

func runCursorHook(raw map[string]any) error {
	var input hosts.CursorHookInput
	if err := remarshal(raw, &input); err != nil {
		return nil
	}

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
	logSavings(input.ToolName, result)
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

func logSavings(toolName string, result intercept.InterceptResult) {
	if result.Stats == nil || result.Stats.BeforeBytes == 0 {
		return
	}
	pct := float64(result.Stats.BeforeBytes-result.Stats.AfterBytes) /
		float64(result.Stats.BeforeBytes) * 100
	fmt.Fprintf(os.Stderr, "[confire] %s: %d → %d bytes (%.0f%%) [%s]\n",
		toolName, result.Stats.BeforeBytes, result.Stats.AfterBytes,
		pct, result.Stats.Optimizer)
}

// remarshal round-trips a decoded map back through JSON into a typed struct.
func remarshal(raw map[string]any, dst any) error {
	b, err := json.Marshal(raw)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dst)
}

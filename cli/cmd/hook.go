package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/confire-dev/confire/hosts"
	"github.com/confire-dev/confire/intercept"
	"github.com/confire-dev/confire/transport"
	"github.com/spf13/cobra"
)

var hookCmd = &cobra.Command{
	Use:    "hook",
	Short:  "Process a Claude Code hook event (called by Claude Code, not by users)",
	Hidden: true, // not shown in help — internal command called by Claude Code
	RunE: func(cmd *cobra.Command, args []string) error {
		return runHook()
	},
}

func init() {
	rootCmd.AddCommand(hookCmd)
}

func runHook() error {
	var input hosts.HookInput
	if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
		// Parse error → pass through unchanged. Never block the developer.
		return nil
	}

	event := hosts.DecodeHookInput(input)

	hint := ""
	if event.Tool != nil {
		hint = event.Tool.GetServerHint()
	}

	// Try daemon first (warm HTTP/2 + cloud optimizer).
	// If it's not running, fall through to LocalTransport — never block the developer.
	local := transport.NewLocal(hint)
	t := transport.NewDaemonClient(daemonSocketPath(), local)

	result, err := t.Send(event)
	if err != nil || result.Kind == intercept.ResultPassthrough {
		return nil // exit 0 = pass through
	}

	out, shouldWrite := hosts.EncodeResult(result, input.HookEventName)
	if !shouldWrite {
		return nil
	}

	// Log savings to stderr only — never stdout (stdout is the hook result).
	if result.Stats != nil && result.Stats.BeforeBytes > 0 {
		reduction := float64(result.Stats.BeforeBytes-result.Stats.AfterBytes) /
			float64(result.Stats.BeforeBytes) * 100
		fmt.Fprintf(os.Stderr, "[confire] %s: %d → %d bytes (%.0f%%) [%s]\n",
			input.ToolName,
			result.Stats.BeforeBytes, result.Stats.AfterBytes,
			reduction, result.Stats.Optimizer,
		)
	}

	return json.NewEncoder(os.Stdout).Encode(out)
}

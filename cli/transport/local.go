package transport

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/confire-dev/confire/intercept"
	"github.com/confire-dev/confire/optimizer"
)

// LocalTransport runs the bundled Go optimizer core in-process.
// Zero network, ~1ms latency, works offline. Covers the FREE TIER:
//   - Bash: trim logs, keep failures
//   - Read: cap large file reads
//   - WebFetch: strip HTML noise
//   - Generic: universal JSON noise stripping (fallback)
//
// Platform-specific MCP optimizers (Figma, GitHub, Slack, etc.) run in the
// Cloudflare Worker only via WorkerTransport / FallbackTransport.
type LocalTransport struct {
	reg *optimizer.Registry
}

func NewLocal(_ string) *LocalTransport {
	return &LocalTransport{reg: optimizer.NewRegistry("")}
}

func (t *LocalTransport) Send(event intercept.InterceptEvent) (intercept.InterceptResult, error) {
	if event.Phase == intercept.PhasePreCompact {
		return intercept.InterceptResult{Kind: intercept.ResultPassthrough}, nil
	}

	switch event.Phase {
	case intercept.PhaseToolPost:
		return t.handleToolPost(event)
	case intercept.PhaseToolPre:
		return t.handleToolPre(event)
	default:
		return intercept.InterceptResult{Kind: intercept.ResultPassthrough}, nil
	}
}

func (t *LocalTransport) handleToolPost(event intercept.InterceptEvent) (intercept.InterceptResult, error) {
	if event.Tool == nil || event.Tool.Output == nil {
		return intercept.InterceptResult{Kind: intercept.ResultPassthrough}, nil
	}

	// MCP tools are Worker-only — local only handles native tools.
	if event.Tool.IsMCP {
		return intercept.InterceptResult{Kind: intercept.ResultPassthrough}, nil
	}

	toolName := normalizeLocalToolName(event.Tool.Name)

	// Write/Edit/Glob have structured output shapes Claude Code requires preserved
	// (structuredPatch, originalFile, etc.) — never touch them.
	switch toolName {
	case "write", "edit", "glob":
		return intercept.InterceptResult{Kind: intercept.ResultPassthrough}, nil
	}

	// For Bash, Claude Code sends tool_response as {"stdout":"...","stderr":"...",...}.
	// Extract stdout, optimize it, then rebuild the full object with updated stdout.
	if toolName == "bash" {
		stdout, obj := bashStdout(event.Tool.Output)
		if stdout != "" {
			cmd := bashCommand(event.Tool.Input)
			optimized := optimizer.OptimizeBashWithCmd(stdout, cmd)
			if optimized != stdout && len(optimized) < len(stdout) {
				var toolOutput any = optimized
				if obj != nil {
					rebuilt := make(map[string]any, len(obj))
					for k, v := range obj {
						rebuilt[k] = v
					}
					rebuilt["stdout"] = optimized
					toolOutput = rebuilt
				}
				return intercept.InterceptResult{
					Kind:       intercept.ResultReplaceOutput,
					ToolOutput: toolOutput,
					Stats: &intercept.Stats{
						BeforeBytes: len(stdout),
						AfterBytes:  len(optimized),
						Optimizer:   "local/bash",
					},
				}, nil
			}
		}
		return intercept.InterceptResult{Kind: intercept.ResultPassthrough}, nil
	}

	opt := t.reg.Resolve(toolName)
	optimized := opt.Optimize(event.Tool.Output)

	before, _ := json.Marshal(event.Tool.Output)
	after, err := json.Marshal(optimized)
	if err != nil || len(after) >= len(before) {
		return intercept.InterceptResult{Kind: intercept.ResultPassthrough}, nil
	}

	return intercept.InterceptResult{
		Kind:       intercept.ResultReplaceOutput,
		ToolOutput: optimized,
		Stats: &intercept.Stats{
			BeforeBytes: len(before),
			AfterBytes:  len(after),
			Optimizer:   fmt.Sprintf("local/%T", opt),
		},
	}, nil
}

func (t *LocalTransport) handleToolPre(event intercept.InterceptEvent) (intercept.InterceptResult, error) {
	if event.Tool == nil || event.Tool.IsMCP {
		return intercept.InterceptResult{Kind: intercept.ResultPassthrough}, nil
	}

	// Read: inject line limit before the file is even read
	if normalizeLocalToolName(event.Tool.Name) == "read" {
		input, ok := event.Tool.Input.(map[string]interface{})
		if !ok {
			return intercept.InterceptResult{Kind: intercept.ResultPassthrough}, nil
		}
		if _, has := input["limit"]; has {
			return intercept.InterceptResult{Kind: intercept.ResultPassthrough}, nil
		}
		if _, has := input["offset"]; has {
			return intercept.InterceptResult{Kind: intercept.ResultPassthrough}, nil
		}
		newInput := make(map[string]interface{}, len(input)+1)
		for k, v := range input {
			newInput[k] = v
		}
		newInput["limit"] = 500
		return intercept.InterceptResult{
			Kind:      intercept.ResultReplaceInput,
			ToolInput: newInput,
			Stats:     &intercept.Stats{Optimizer: "local/read-pre"},
		}, nil
	}
	return intercept.InterceptResult{Kind: intercept.ResultPassthrough}, nil
}

func bashCommand(input any) string {
	if m, ok := input.(map[string]interface{}); ok {
		if cmd, ok := m["command"].(string); ok {
			return cmd
		}
	}
	return ""
}

// bashStdout extracts the stdout text from a Bash tool_response.
// Claude Code sends it as {"stdout":"...","stderr":"...",...} not a plain string.
// Returns the stdout string and the original map (nil if input was a plain string).
func bashStdout(output any) (string, map[string]any) {
	if s, ok := output.(string); ok {
		return s, nil
	}
	if m, ok := output.(map[string]interface{}); ok {
		if s, ok := m["stdout"].(string); ok {
			return s, m
		}
	}
	return "", nil
}

func (t *LocalTransport) Mode() OptimizerMode { return OptimizerModeLocal }

func normalizeLocalToolName(name string) string {
	switch strings.ToLower(name) {
	case "shell":
		return "bash"
	default:
		return strings.ToLower(name)
	}
}

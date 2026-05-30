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

	toolName := strings.ToLower(event.Tool.Name)

	// For Bash, pass the original command to the optimizer for better heuristics.
	if toolName == "bash" {
		if s, ok := event.Tool.Output.(string); ok {
			cmd := bashCommand(event.Tool.Input)
			optimized := optimizer.OptimizeBashWithCmd(s, cmd)
			if optimized != s && len(optimized) < len(s) {
				return intercept.InterceptResult{
					Kind:       intercept.ResultReplaceOutput,
					ToolOutput: optimized,
					Stats: &intercept.Stats{
						BeforeBytes: len(s),
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
	if event.Tool.Name == "Read" {
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

func (t *LocalTransport) Mode() OptimizerMode { return OptimizerModeLocal }

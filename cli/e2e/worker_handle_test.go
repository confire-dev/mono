package e2e_test

import (
	"encoding/json"
	"strings"

	"github.com/confire-dev/confire/intercept"
	"github.com/confire-dev/confire/optimizer"
)

// workerHandle simulates what the real Cloudflare Worker does.
// It uses NewWorkerRegistry (includes Figma + all MCP optimizers) so the
// mock gives the same results as the deployed Worker.
func workerHandle(event intercept.InterceptEvent) intercept.InterceptResult {
	if event.Tool == nil || event.Tool.Output == nil {
		return intercept.InterceptResult{Kind: intercept.ResultPassthrough}
	}

	toolName := strings.ToLower(event.Tool.Name)
	if event.Tool.IsMCP && event.Tool.MCPServer != "" {
		toolName = strings.ToLower(event.Tool.MCPServer)
	}

	reg := optimizer.NewWorkerRegistry(toolName)
	opt := reg.Resolve(toolName)
	optimized := opt.Optimize(event.Tool.Output)

	before, _ := json.Marshal(event.Tool.Output)
	after, err := json.Marshal(optimized)
	if err != nil || len(after) >= len(before) {
		return intercept.InterceptResult{Kind: intercept.ResultPassthrough}
	}

	return intercept.InterceptResult{
		Kind:       intercept.ResultReplaceOutput,
		ToolOutput: optimized,
		Stats: &intercept.Stats{
			BeforeBytes: len(before),
			AfterBytes:  len(after),
			Optimizer:   "mock-worker/" + toolName,
		},
	}
}

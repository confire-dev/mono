package intercept

import (
	"encoding/json"
	"fmt"
	"strings"
)

// MCPRiskHandler implements the PreToolUse risk classifier for MCP tools.
// It scores tool name + input parameter names and returns review/warn/passthrough.
type MCPRiskHandler struct{}

func (h *MCPRiskHandler) ID() string     { return "mcp.risk_classifier" }
func (h *MCPRiskHandler) Phases() []Phase { return []Phase{PhaseToolPre} }

func (h *MCPRiskHandler) Matches(e InterceptEvent) bool {
	return e.Tool != nil && e.Tool.IsMCP
}

func (h *MCPRiskHandler) Run(e InterceptEvent) (InterceptResult, error) {
	score, reason := scoreMCPRisk(e.Tool)
	switch {
	case score >= 50:
		return InterceptResult{
			Kind:   ResultReview,
			Reason: reason,
			Context: fmt.Sprintf("[Confire mcp.risk] score=%d — %s", score, reason),
		}, nil
	case score >= 20:
		return InterceptResult{
			Kind:    ResultWarn,
			Context: fmt.Sprintf("[Confire mcp.risk] score=%d — %s", score, reason),
		}, nil
	default:
		return InterceptResult{Kind: ResultPassthrough}, nil
	}
}

// scoreMCPRisk returns a 0-100 risk score and the first reason that drove it up.
func scoreMCPRisk(tool *Tool) (int, string) {
	name := strings.ToLower(tool.Name)
	score := 0
	reason := ""

	// Tier 1 — shell/exec names (40 pts)
	if matchesAny(name, shellNamePatterns) {
		score += 40
		reason = "tool name suggests shell or code execution"
	}

	// Tier 2 — write/delete/destroy names (30 pts, cumulative)
	if matchesAny(name, writeNamePatterns) {
		score += 30
		if reason == "" {
			reason = "tool name suggests destructive or write operation"
		}
	}

	// Tier 3 — exfiltration names (20 pts, cumulative)
	if matchesAny(name, exfilNamePatterns) {
		score += 20
		if reason == "" {
			reason = "tool name suggests data may be sent externally"
		}
	}

	// Input param scan (5 pts per dangerous param name, up to 20)
	paramHits := scanInputParams(tool.Input)
	if paramHits > 0 {
		pts := paramHits * 5
		if pts > 20 {
			pts = 20
		}
		score += pts
		if reason == "" {
			reason = "input parameters include credential or shell-command names"
		}
	}

	if score > 100 {
		score = 100
	}
	if reason == "" {
		reason = "no specific risk signals"
	}
	return score, reason
}

// scanInputParams counts how many input keys match dangerous parameter names.
func scanInputParams(input any) int {
	if input == nil {
		return 0
	}
	// Re-marshal to map to handle any underlying type.
	var m map[string]any
	switch v := input.(type) {
	case map[string]any:
		m = v
	default:
		b, err := json.Marshal(input)
		if err != nil {
			return 0
		}
		if err := json.Unmarshal(b, &m); err != nil {
			return 0
		}
	}
	hits := 0
	for k := range m {
		k = strings.ToLower(k)
		if matchesAny(k, credentialParamPatterns) {
			hits++
		}
	}
	return hits
}

func matchesAny(s string, patterns []string) bool {
	for _, p := range patterns {
		if strings.Contains(s, p) {
			return true
		}
	}
	return false
}

var shellNamePatterns = []string{
	"bash", "shell", "exec", "execute", "run_command", "run_cmd",
	"terminal", "eval", "spawn", "cmd", "powershell", "invoke",
}

var writeNamePatterns = []string{
	"write", "delete", "remove", "unlink", "overwrite",
	"destroy", "drop", "truncate", "wipe", "purge",
}

var exfilNamePatterns = []string{
	"send", "email", "post_message", "upload", "transfer",
	"share", "export", "forward", "webhook", "notify",
}

var credentialParamPatterns = []string{
	"token", "secret", "password", "passwd", "api_key", "apikey",
	"private_key", "auth", "credential", "access_key", "bearer",
}

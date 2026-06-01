// Package guardrail implements the PreToolUse firewall handler.
// It evaluates policy rules before a tool executes and returns
// block/review/warn/passthrough decisions.
package guardrail

import (
	"fmt"
	"strings"

	"github.com/confire-dev/confire/intercept"
	"github.com/confire-dev/confire/policy"
)

// Handler is the PreToolUse firewall handler.
type Handler struct {
	engine *policy.Engine
	mode   policy.Mode
}

// New creates a GuardrailHandler with the given policy engine and mode.
func New(engine *policy.Engine, mode policy.Mode) *Handler {
	return &Handler{engine: engine, mode: mode}
}

func (h *Handler) ID() string    { return "guardrail" }
func (h *Handler) Rules() []policy.Rule { return h.engine.Rules() }

func (h *Handler) Phases() []intercept.Phase {
	return []intercept.Phase{intercept.PhaseToolPre}
}

func (h *Handler) Matches(e intercept.InterceptEvent) bool {
	return e.Tool != nil
}

func (h *Handler) Run(e intercept.InterceptEvent) (intercept.InterceptResult, error) {
	match := h.engine.EvaluatePreTool(e, h.mode)
	if match == nil {
		return intercept.InterceptResult{Kind: intercept.ResultPassthrough}, nil
	}

	switch match.Action {
	case policy.ActionBlock:
		return intercept.InterceptResult{
			Kind:   intercept.ResultBlock,
			Reason: formatBlockMessage(match, e),
		}, nil

	case policy.ActionReview:
		return intercept.InterceptResult{
			Kind:   intercept.ResultReview,
			Reason: formatReviewMessage(match, e),
		}, nil

	case policy.ActionWarn:
		return intercept.InterceptResult{
			Kind:    intercept.ResultWarn,
			Context: formatWarnMessage(match, e),
		}, nil

	default:
		return intercept.InterceptResult{Kind: intercept.ResultPassthrough}, nil
	}
}

// ── message formatting ────────────────────────────────────────────────────

func formatBlockMessage(m *policy.MatchResult, e intercept.InterceptEvent) string {
	return fmt.Sprintf(`CONFIRE BLOCKED TOOL CALL

Rule:    %s
Blocked: %s — %s
Reason:  %s`,
		m.Rule.Name,
		e.Tool.Name,
		inputExcerpt(e.Tool, 120),
		m.Message,
	)
}

func formatReviewMessage(m *policy.MatchResult, e intercept.InterceptEvent) string {
	return fmt.Sprintf(`CONFIRE REVIEW REQUIRED

Rule:              %s
Claude is about to run: %s — %s
Risk:              %s severity
Why this matters:  %s

ACTION REQUIRED — ask the user:
"Confire flagged this command. Do you want me to run it anyway?
If yes: run 'confire bypass-next' in your terminal, then tell me to retry."`,
		m.Rule.Name,
		e.Tool.Name,
		inputExcerpt(e.Tool, 120),
		string(m.Rule.Severity),
		m.Message,
	)
}

func formatWarnMessage(m *policy.MatchResult, e intercept.InterceptEvent) string {
	return fmt.Sprintf(`[Confire warning] %s: %s — %s`,
		m.Rule.Name,
		e.Tool.Name,
		m.Message,
	)
}

// inputExcerpt returns a short human-readable excerpt of the tool input.
func inputExcerpt(tool *intercept.Tool, maxLen int) string {
	if tool == nil {
		return ""
	}
	var s string
	switch v := tool.Input.(type) {
	case string:
		s = v
	case map[string]any:
		if cmd, ok := v["command"].(string); ok {
			s = cmd
		} else {
			s = fmt.Sprintf("%v", v)
		}
	default:
		s = fmt.Sprintf("%v", tool.Input)
	}
	s = strings.TrimSpace(s)
	if len(s) > maxLen {
		s = s[:maxLen] + "…"
	}
	return s
}

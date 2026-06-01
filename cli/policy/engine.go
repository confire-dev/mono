package policy

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/confire-dev/confire/intercept"
)

// Engine evaluates a merged rule set against intercept events.
type Engine struct {
	rules []Rule
}

// NewEngine creates an engine with the given rules.
// Use LoadRules() to get the merged built-in + custom rule set.
func NewEngine(rules []Rule) *Engine {
	return &Engine{rules: rules}
}

// EvaluatePreTool evaluates PreToolUse rules for the given event and mode.
// Returns the highest-priority match, or nil if nothing fires.
func (e *Engine) EvaluatePreTool(event intercept.InterceptEvent, mode Mode) *MatchResult {
	if mode == ModeBypass {
		return nil
	}
	if event.Tool == nil {
		return nil
	}

	// Collect all matching rules, then pick highest priority action.
	var results []MatchResult
	for _, rule := range e.rules {
		if !rule.Enabled {
			continue
		}
		if rule.Phase != PhasePreToolUse && rule.Phase != PhaseAny {
			continue
		}
		if rule.MinMode == ModeStrict && mode != ModeStrict {
			continue
		}
		if matchesPreTool(rule, event) {
			action := effectiveAction(rule.Action, mode)
			results = append(results, MatchResult{Rule: rule, Action: action, Message: rule.Message})
		}
	}
	return highestPriority(results)
}

// EvaluatePostTool evaluates PostToolUse rules for the given event and mode.
func (e *Engine) EvaluatePostTool(event intercept.InterceptEvent, mode Mode) *MatchResult {
	if mode == ModeBypass {
		return nil
	}
	if event.Tool == nil {
		return nil
	}

	var results []MatchResult
	for _, rule := range e.rules {
		if !rule.Enabled {
			continue
		}
		if rule.Phase != PhasePostToolUse && rule.Phase != PhaseAny {
			continue
		}
		if matchesPostTool(rule, event) {
			action := effectiveAction(rule.Action, mode)
			results = append(results, MatchResult{Rule: rule, Action: action, Message: rule.Message})
		}
	}
	return highestPriority(results)
}

// Rules returns all rules in the engine (for introspection / CLI commands).
func (e *Engine) Rules() []Rule {
	return e.rules
}

// ── matching ──────────────────────────────────────────────────────────────

func matchesPreTool(rule Rule, event intercept.InterceptEvent) bool {
	tool := event.Tool
	m := rule.Match

	// Tool name / prefix checks.
	if !matchesToolName(m, tool) {
		return false
	}

	// Command-level checks (Bash-specific).
	cmd := bashCommand(tool)
	if len(m.CommandContains) > 0 {
		matched := false
		for _, s := range m.CommandContains {
			if strings.Contains(cmd, s) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	if m.CommandRegex != "" {
		re, err := regexp.Compile("(?i)" + m.CommandRegex)
		if err != nil {
			return false
		}
		// For Bash, match against the command string.
		// For Read/other tools, match against the full input JSON string.
		target := cmd
		if target == "" {
			target = fmt.Sprintf("%v", tool.Input)
		}
		if !re.MatchString(target) {
			return false
		}
	}

	return true
}

func matchesPostTool(rule Rule, event intercept.InterceptEvent) bool {
	m := rule.Match
	// Output secret/injection scan rules don't filter by tool name —
	// they apply to all tool outputs. Return true here; the actual
	// scanning runs in the sanitize pipeline.
	if m.OutputSecretScan || m.OutputInjectionScan {
		return true
	}
	return matchesToolName(m, event.Tool)
}

func matchesToolName(m RuleMatch, tool *intercept.Tool) bool {
	name := normalizeToolName(tool.Name)

	if m.ToolName != "" && strings.ToLower(m.ToolName) != name {
		return false
	}
	if len(m.ToolNames) > 0 {
		found := false
		for _, tn := range m.ToolNames {
			if strings.ToLower(tn) == name {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if m.ToolPrefix != "" && !strings.HasPrefix(name, strings.ToLower(m.ToolPrefix)) {
		return false
	}
	if m.ToolNameRegex != "" {
		re, err := regexp.Compile("(?i)" + m.ToolNameRegex)
		if err != nil {
			return false
		}
		if !re.MatchString(name) {
			return false
		}
	}
	return true
}

// normalizeToolName maps host-specific aliases onto canonical names used in rules.
func normalizeToolName(name string) string {
	switch strings.ToLower(name) {
	case "shell":
		return "bash"
	case "runterminalcommand":
		return "bash"
	default:
		return strings.ToLower(name)
	}
}

// bashCommand extracts the command string from a Bash/Shell tool input.
func bashCommand(tool *intercept.Tool) string {
	if tool == nil || normalizeToolName(tool.Name) != "bash" {
		return ""
	}
	switch v := tool.Input.(type) {
	case string:
		return v
	case map[string]any:
		if cmd, ok := v["command"].(string); ok {
			return cmd
		}
	}
	return fmt.Sprintf("%v", tool.Input)
}

// ── action priority ───────────────────────────────────────────────────────

// actionPriority maps actions to priority (higher = wins).
var actionPriority = map[RuleAction]int{
	ActionBlock:   4,
	ActionReview:  3,
	ActionWarn:    2,
	ActionRedact:  2,
	ActionSanitize: 2,
	ActionOptimize: 1,
	ActionAllow:   0,
	ActionPassThrough: 0,
}

func highestPriority(results []MatchResult) *MatchResult {
	if len(results) == 0 {
		return nil
	}
	best := results[0]
	for _, r := range results[1:] {
		if actionPriority[r.Action] > actionPriority[best.Action] {
			best = r
		}
	}
	return &best
}

// effectiveAction downgrades block/review to warn in observe mode.
func effectiveAction(action RuleAction, mode Mode) RuleAction {
	if mode == ModeObserve {
		switch action {
		case ActionBlock, ActionReview:
			return ActionWarn
		}
	}
	return action
}

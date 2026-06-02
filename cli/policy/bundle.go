package policy

import (
	"embed"
	"encoding/json"
	"fmt"
	"regexp"
)

//go:embed rules/*.json
var bundledRuleFS embed.FS

// ruleBundle is the on-disk format for bundled rule packs.
// Rules with a "patterns" array expand into one Rule per pattern at load time,
// so new checks are data — not Go code.
type ruleBundle struct {
	Version int           `json:"version"`
	Rules   []bundledRule `json:"rules"`
}

type bundledRule struct {
	ID       string     `json:"id"`
	Name     string     `json:"name"`
	Enabled  bool       `json:"enabled"`
	Phase    RulePhase  `json:"phase"`
	Action   RuleAction `json:"action"`
	Severity Severity   `json:"severity"`
	MinMode  Mode       `json:"min_mode,omitempty"`
	Group    string     `json:"group,omitempty"`
	Match    RuleMatch  `json:"match"`
	Message  string     `json:"message"`
	Patterns []pattern  `json:"patterns,omitempty"`
}

type pattern struct {
	ID       string `json:"id"`
	Regex    string `json:"regex,omitempty"`
	Contains string `json:"contains,omitempty"`
	Message  string `json:"message"`
}

// BuiltinRules returns bundled rules embedded in the CLI binary.
// Custom cloud rules (cache.go) merge on top of these.
func BuiltinRules() []Rule {
	rules, err := loadBundledRules()
	if err != nil {
		panic(fmt.Sprintf("policy: invalid bundled rules: %v", err))
	}
	return rules
}

func loadBundledRules() ([]Rule, error) {
	var out []Rule
	for _, name := range []string{"rules/builtin.json", "rules/mcp.json"} {
		data, err := bundledRuleFS.ReadFile(name)
		if err != nil {
			return nil, err
		}
		var bundle ruleBundle
		if err := json.Unmarshal(data, &bundle); err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		if bundle.Version == 0 {
			return nil, fmt.Errorf("%s: missing bundle version", name)
		}
		for _, br := range bundle.Rules {
			expanded, err := expandBundledRule(br)
			if err != nil {
				return nil, fmt.Errorf("%s rule %q: %w", name, br.ID, err)
			}
			out = append(out, expanded...)
		}
	}
	return out, nil
}

func expandBundledRule(br bundledRule) ([]Rule, error) {
	if len(br.Patterns) == 0 {
		r := br.toRule(br.ID, br.Message, br.Match)
		if err := validateRuleMatch(r.Match); err != nil {
			return nil, err
		}
		return []Rule{r}, nil
	}

	out := make([]Rule, 0, len(br.Patterns))
	for _, p := range br.Patterns {
		if p.ID == "" {
			return nil, fmt.Errorf("pattern missing id")
		}
		if p.Message == "" {
			return nil, fmt.Errorf("pattern %q missing message", p.ID)
		}
		match := br.Match
		match.CommandRegex = ""
		match.CommandContains = nil
		switch {
		case p.Regex != "":
			match.CommandRegex = p.Regex
		case p.Contains != "":
			match.CommandContains = []string{p.Contains}
		default:
			return nil, fmt.Errorf("pattern %q needs regex or contains", p.ID)
		}
		if err := validateRuleMatch(match); err != nil {
			return nil, fmt.Errorf("pattern %q: %w", p.ID, err)
		}
		out = append(out, br.toRule(br.ID+"/"+p.ID, p.Message, match))
	}
	return out, nil
}

func (br bundledRule) toRule(id, message string, match RuleMatch) Rule {
	return Rule{
		ID:       id,
		Name:     br.Name,
		Enabled:  br.Enabled,
		Phase:    br.Phase,
		Action:   br.Action,
		Severity: br.Severity,
		MinMode:  br.MinMode,
		Group:    br.Group,
		Match:    match,
		Message:  message,
		Source:   "builtin",
	}
}

func validateRuleMatch(m RuleMatch) error {
	if m.CommandRegex != "" {
		if _, err := regexp.Compile("(?i)" + m.CommandRegex); err != nil {
			return fmt.Errorf("invalid command_regex: %w", err)
		}
	}
	if m.ToolNameRegex != "" {
		if _, err := regexp.Compile("(?i)" + m.ToolNameRegex); err != nil {
			return fmt.Errorf("invalid tool_name_regex: %w", err)
		}
	}
	return nil
}

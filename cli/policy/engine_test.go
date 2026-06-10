package policy

import (
	"testing"

	"github.com/confire-dev/confire/intercept"
)

func TestEvaluatePreTool_BuiltinRules(t *testing.T) {
	engine := NewEngine(BuiltinRules())

	cases := []struct {
		name     string
		toolName string
		input    any
		mode     Mode
		wantAct  RuleAction // "" means passthrough (nil result)
	}{
		// Should review
		{"git force push", "Bash", bash("git push --force"), ModeBalanced, ActionReview},
		{"git force push with-lease", "Bash", bash("git push --force-with-lease origin main"), ModeBalanced, ActionReview},
		{"git reset hard", "Bash", bash("git reset --hard HEAD~1"), ModeBalanced, ActionReview},
		{"git clean fd", "Bash", bash("git clean -fd"), ModeBalanced, ActionReview},
		{"git rebase main", "Bash", bash("git rebase main"), ModeBalanced, ActionReview},
		{"git rebase master", "Bash", bash("git rebase master"), ModeBalanced, ActionReview},
		{"git rebase origin/main", "Bash", bash("git rebase origin/main"), ModeBalanced, ActionReview},
		{"git branch delete", "Bash", bash("git branch -D my-feature"), ModeBalanced, ActionReview},
		{"gh pr close", "Bash", bash("gh pr close 42"), ModeBalanced, ActionReview},
		{"gh pr merge", "Bash", bash("gh pr merge 42"), ModeBalanced, ActionReview},
		{"rm -rf", "Bash", bash("rm -rf /tmp/foo"), ModeBalanced, ActionReview},
		{"rm -fr", "Bash", bash("rm -fr /tmp/foo"), ModeBalanced, ActionReview},
		{"drop database", "Bash", bash("drop database mydb;"), ModeBalanced, ActionReview},
		{"truncate table", "Bash", bash("truncate table users;"), ModeBalanced, ActionReview},
		{"delete without where", "Bash", bash("delete from users;"), ModeBalanced, ActionReview},
		{"supabase db reset", "Bash", bash("supabase db reset"), ModeBalanced, ActionReview},
		{"npm publish", "Bash", bash("npm publish"), ModeBalanced, ActionReview},
		{"pnpm publish", "Bash", bash("pnpm publish"), ModeBalanced, ActionReview},
		{"vercel prod", "Bash", bash("vercel --prod"), ModeBalanced, ActionReview},
		{"fly deploy", "Bash", bash("fly deploy"), ModeBalanced, ActionReview},
		{"kubectl apply", "Bash", bash("kubectl apply -f deployment.yaml"), ModeBalanced, ActionReview},
		{"kubectl delete", "Bash", bash("kubectl delete pod mypod"), ModeBalanced, ActionReview},
		{"terraform apply", "Bash", bash("terraform apply"), ModeBalanced, ActionReview},
		{"terraform destroy", "Bash", bash("terraform destroy"), ModeBalanced, ActionReview},
		{"mcp create", "mcp__github__create_pull_request", nil, ModeBalanced, ActionReview},
		{"mcp merge", "mcp__github__merge_pull_request", nil, ModeBalanced, ActionReview},
		{"mcp delete", "mcp__linear__delete_issue", nil, ModeBalanced, ActionReview},
		{"mcp send", "mcp__slack__send_message", nil, ModeBalanced, ActionReview},

		// Should block
		{"gh repo delete", "Bash", bash("gh repo delete myorg/myrepo"), ModeBalanced, ActionBlock},

		// Should allow (passthrough)
		{"git status", "Bash", bash("git status"), ModeBalanced, ""},
		{"git diff", "Bash", bash("git diff HEAD"), ModeBalanced, ""},
		{"git log", "Bash", bash("git log --oneline -10"), ModeBalanced, ""},
		{"gh pr view", "Bash", bash("gh pr view 42"), ModeBalanced, ""},
		{"ls", "Bash", bash("ls -la"), ModeBalanced, ""},
		{"cat package.json", "Bash", bash("cat package.json"), ModeBalanced, ""},
		{"mcp get", "mcp__github__get_pull_request", nil, ModeBalanced, ""},
		{"mcp list", "mcp__linear__list_issues", nil, ModeBalanced, ""},
		{"mcp search", "mcp__figma__search_design_system", nil, ModeBalanced, ""},

		// .env in git commit message — must not trigger secret-file-read rule
		{"env in commit message", "Bash", bash(`git commit -m "ignore .env file"`), ModeBalanced, ""},
		{"env in commit add", "Bash", bash("git add .env.example"), ModeBalanced, ""},

		// .env file access — real secrets reviewed, safe variants allowed
		{"env file review", "Bash", bash("cat .env"), ModeBalanced, ActionReview},
		{"env.local review", "Bash", bash("cat .env.local"), ModeBalanced, ActionReview},
		{"env.production review", "Bash", bash("cat .env.production"), ModeBalanced, ActionReview},
		{"env.example allowed", "Bash", bash("cat .env.example"), ModeBalanced, ""},
		{"env.sample allowed", "Bash", bash("cat .env.sample"), ModeBalanced, ""},

		// .env access tiers — raw dump vs. redacted preview vs. key-names-only
		// all still require review, but classify differently (see secret-file-read-* rules)
		{"env grep raw review", "Bash", bash(`grep -i "SUPABASE" .env`), ModeBalanced, ActionReview},
		{"env redacted preview review", "Bash", bash(`grep -i "SUPABASE" .env | sed 's/=.*/=<redacted>/'`), ModeBalanced, ActionReview},
		{"env awk redacted review", "Bash", bash(`awk -F= '{print $1"=<redacted>"}' .env`), ModeBalanced, ActionReview},
		{"env key names only review", "Bash", bash(`cut -d= -f1 .env`), ModeBalanced, ActionReview},
		{"env key names grep review", "Bash", bash(`grep -E '^[A-Z_]+=' .env | cut -d= -f1`), ModeBalanced, ActionReview},

		// Observe mode downgrades block→warn, review→warn
		{"observe mode downgrades review", "Bash", bash("git push --force"), ModeObserve, ActionWarn},
		{"observe mode downgrades block", "Bash", bash("gh repo delete myorg/x"), ModeObserve, ActionWarn},

		// Bypass mode — everything passthrough
		{"bypass mode", "Bash", bash("git push --force"), ModeBypass, ""},
		{"cursor shell alias", "Shell", bash("git push --force"), ModeBalanced, ActionReview},
		{"vscode terminal alias", "runTerminalCommand", bash("git push --force"), ModeBalanced, ActionReview},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			event := makeEvent(tc.toolName, tc.input)
			result := engine.EvaluatePreTool(event, tc.mode)
			if tc.wantAct == "" {
				if result != nil {
					t.Errorf("expected passthrough, got action=%s rule=%s", result.Action, result.Rule.ID)
				}
				return
			}
			if result == nil {
				t.Errorf("expected action=%s, got nil (passthrough)", tc.wantAct)
				return
			}
			if result.Action != tc.wantAct {
				t.Errorf("expected action=%s, got action=%s (rule=%s)", tc.wantAct, result.Action, result.Rule.ID)
			}
		})
	}
}

func TestSecretFileReadTiers(t *testing.T) {
	engine := NewEngine(BuiltinRules())

	cases := []struct {
		name     string
		command  string
		wantRule string
		wantSev  Severity
	}{
		{"raw cat", "cat .env", "secret-file-read-raw", SeverityHigh},
		{"raw grep", `grep -i "SUPABASE" .env`, "secret-file-read-raw", SeverityHigh},
		{"redacted sed", `grep -i "SUPABASE" .env | sed 's/=.*/=<redacted>/'`, "secret-file-read-redacted", SeverityLow},
		{"redacted awk", `awk -F= '{print $1"=<redacted>"}' .env`, "secret-file-read-redacted", SeverityLow},
		{"key names cut", "cut -d= -f1 .env", "secret-file-read-keynames", SeverityLow},
		{"key names grep+cut", `grep -E '^[A-Z_]+=' .env | cut -d= -f1`, "secret-file-read-keynames", SeverityLow},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			event := makeEvent("Bash", bash(tc.command))
			result := engine.EvaluatePreTool(event, ModeBalanced)
			if result == nil {
				t.Fatalf("expected a match, got passthrough")
			}
			if result.Rule.ID != tc.wantRule {
				t.Errorf("expected rule=%s, got rule=%s (action=%s)", tc.wantRule, result.Rule.ID, result.Action)
			}
			if result.Rule.Severity != tc.wantSev {
				t.Errorf("expected severity=%s, got severity=%s", tc.wantSev, result.Rule.Severity)
			}
			if result.Action != ActionReview {
				t.Errorf("expected action=review, got action=%s", result.Action)
			}
		})
	}
}

func TestMergeRules(t *testing.T) {
	builtin := []Rule{{ID: "a", Name: "built-in A", Enabled: true, Source: "builtin"}}
	custom := []Rule{
		{ID: "a", Name: "override A", Enabled: false, Source: "custom"}, // override
		{ID: "b", Name: "new B", Enabled: true, Source: "custom"},       // new
	}
	merged := mergeRules(builtin, custom)
	if len(merged) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(merged))
	}
	if merged[0].Name != "override A" {
		t.Errorf("expected custom override for ID=a, got %q", merged[0].Name)
	}
	if merged[1].ID != "b" {
		t.Errorf("expected new rule ID=b at index 1, got %q", merged[1].ID)
	}
}

// ── helpers ───────────────────────────────────────────────────────────────

func bash(cmd string) any {
	return map[string]any{"command": cmd}
}

func makeEvent(toolName string, input any) intercept.InterceptEvent {
	isMCP := len(toolName) > 5 && toolName[:5] == "mcp__"
	return intercept.InterceptEvent{
		Host:     "claude-code",
		Strategy: "hooks",
		Phase:    intercept.PhaseToolPre,
		Session:  intercept.Session{ID: "test"},
		Tool: &intercept.Tool{
			Name:  toolName,
			Input: input,
			IsMCP: isMCP,
		},
	}
}

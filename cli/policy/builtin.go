package policy

// BuiltinRules returns the built-in rules bundled with the CLI binary.
// These are always available regardless of plan tier or network access.
// Custom cloud rules (fetched by cache.go) merge on top of these.
func BuiltinRules() []Rule {
	t := true
	_ = t
	return []Rule{
		// ── Git destructive ────────────────────────────────────────────────
		{
			ID: "git-force-push", Name: "Review force push", Enabled: true,
			Phase: PhasePreToolUse, Action: ActionReview, Severity: SeverityHigh,
			Match:   RuleMatch{ToolName: "Bash", CommandRegex: `\bgit\s+push\b.*--force`},
			Message: "Force push can rewrite remote branch history and affect open PRs.",
			Source:  "builtin",
		},
		{
			ID: "git-reset-hard", Name: "Review hard reset", Enabled: true,
			Phase: PhasePreToolUse, Action: ActionReview, Severity: SeverityHigh,
			Match:   RuleMatch{ToolName: "Bash", CommandContains: []string{"git reset --hard"}},
			Message: "Hard reset discards local commits and staged changes permanently.",
			Source:  "builtin",
		},
		{
			ID: "git-clean-force", Name: "Review force clean", Enabled: true,
			Phase: PhasePreToolUse, Action: ActionReview, Severity: SeverityMedium,
			Match:   RuleMatch{ToolName: "Bash", CommandRegex: `\bgit\s+clean\s+.*-[a-zA-Z]*f`},
			Message: "git clean -f deletes untracked files which cannot be recovered.",
			Source:  "builtin",
		},
		{
			ID: "git-rebase", Name: "Review rebase onto main branch", Enabled: true,
			Phase: PhasePreToolUse, Action: ActionReview, Severity: SeverityMedium,
			Match:   RuleMatch{ToolName: "Bash", CommandRegex: `\bgit\s+rebase\s+(main|master|origin/main|origin/master)\b`},
			Message: "Rebasing onto main rewrites commit history and may cause conflicts if pushed.",
			Source:  "builtin",
		},
		{
			ID: "git-branch-delete", Name: "Review branch deletion", Enabled: true,
			Phase: PhasePreToolUse, Action: ActionReview, Severity: SeverityMedium,
			Match:   RuleMatch{ToolName: "Bash", CommandRegex: `\bgit\s+branch\s+-D\b`},
			Message: "Deleting a branch with -D is not recoverable if it has no remote.",
			Source:  "builtin",
		},
		// ── GitHub CLI ────────────────────────────────────────────────────
		{
			ID: "gh-pr-close", Name: "Review closing a PR", Enabled: true,
			Phase: PhasePreToolUse, Action: ActionReview, Severity: SeverityMedium,
			Match:   RuleMatch{ToolName: "Bash", CommandContains: []string{"gh pr close"}},
			Message: "Closing a PR discards all review comments and may surprise collaborators.",
			Source:  "builtin",
		},
		{
			ID: "gh-pr-merge", Name: "Review merging a PR", Enabled: true,
			Phase: PhasePreToolUse, Action: ActionReview, Severity: SeverityHigh,
			Match:   RuleMatch{ToolName: "Bash", CommandContains: []string{"gh pr merge"}},
			Message: "Merging a PR modifies the shared branch and triggers CI/CD pipelines.",
			Source:  "builtin",
		},
		{
			ID: "gh-repo-delete", Name: "Block repo deletion", Enabled: true,
			Phase: PhasePreToolUse, Action: ActionBlock, Severity: SeverityHigh,
			Match:   RuleMatch{ToolName: "Bash", CommandContains: []string{"gh repo delete"}},
			Message: "Repository deletion is irreversible. Use the GitHub web UI with explicit confirmation.",
			Source:  "builtin",
		},
		// ── Filesystem ───────────────────────────────────────────────────
		{
			ID: "rm-rf", Name: "Review recursive force delete", Enabled: true,
			Phase: PhasePreToolUse, Action: ActionReview, Severity: SeverityHigh,
			Match:   RuleMatch{ToolName: "Bash", CommandRegex: `\brm\s+-[a-zA-Z]*r[a-zA-Z]*f\b|\brm\s+-[a-zA-Z]*f[a-zA-Z]*r\b`},
			Message: "Recursive force delete permanently removes files with no trash recovery.",
			Source:  "builtin",
		},
		{
			ID: "chmod-777", Name: "Review world-writable chmod", Enabled: true,
			Phase: PhasePreToolUse, Action: ActionReview, Severity: SeverityMedium,
			Match:   RuleMatch{ToolName: "Bash", CommandRegex: `\bchmod\s+(-R\s+)?777\b`},
			Message: "chmod 777 makes files world-writable, which is a security risk.",
			Source:  "builtin",
		},
		// ── Database ─────────────────────────────────────────────────────
		{
			ID: "db-drop", Name: "Review DROP DATABASE/TABLE", Enabled: true,
			Phase: PhasePreToolUse, Action: ActionReview, Severity: SeverityHigh,
			Match:   RuleMatch{ToolName: "Bash", CommandRegex: `(?i)\b(drop\s+(database|table)|truncate\s+table)\b`},
			Message: "Dropping or truncating a database table is destructive and may be hard to reverse.",
			Source:  "builtin",
		},
		{
			ID: "db-delete-no-where", Name: "Review DELETE without WHERE", Enabled: true,
			Phase: PhasePreToolUse, Action: ActionReview, Severity: SeverityHigh,
			Match:   RuleMatch{ToolName: "Bash", CommandRegex: `(?i)\bdelete\s+from\s+\w[\w.]*\s*;`},
			Message: "DELETE without a WHERE clause removes all rows from the table.",
			Source:  "builtin",
		},
		{
			ID: "supabase-reset", Name: "Review Supabase DB reset", Enabled: true,
			Phase: PhasePreToolUse, Action: ActionReview, Severity: SeverityHigh,
			Match:   RuleMatch{ToolName: "Bash", CommandContains: []string{"supabase db reset"}},
			Message: "supabase db reset wipes and recreates the local database.",
			Source:  "builtin",
		},
		{
			ID: "supabase-push", Name: "Review Supabase DB push", Enabled: true,
			Phase: PhasePreToolUse, Action: ActionReview, Severity: SeverityMedium,
			Match:   RuleMatch{ToolName: "Bash", CommandContains: []string{"supabase db push"}},
			Message: "supabase db push applies migrations to a remote database.",
			Source:  "builtin",
		},
		{
			ID: "prisma-reset", Name: "Review Prisma migrate reset", Enabled: true,
			Phase: PhasePreToolUse, Action: ActionReview, Severity: SeverityHigh,
			Match:   RuleMatch{ToolName: "Bash", CommandContains: []string{"prisma migrate reset"}},
			Message: "prisma migrate reset drops and recreates the database, deleting all data.",
			Source:  "builtin",
		},
		// ── Package publishing ────────────────────────────────────────────
		{
			ID: "pkg-publish", Name: "Review package publish", Enabled: true,
			Phase: PhasePreToolUse, Action: ActionReview, Severity: SeverityHigh,
			Match:   RuleMatch{ToolName: "Bash", CommandRegex: `\b(npm|pnpm|yarn|bun)\s+publish\b`},
			Message: "Publishing a package to a registry is public and irreversible.",
			Source:  "builtin",
		},
		// ── Production deployments ────────────────────────────────────────
		{
			ID: "deploy-prod", Name: "Review production deployment", Enabled: true,
			Phase: PhasePreToolUse, Action: ActionReview, Severity: SeverityHigh,
			Match: RuleMatch{
				ToolName:     "Bash",
				CommandRegex: `vercel\s+--prod|netlify\s+deploy\s+--prod|fly\s+deploy|railway\s+up\b`,
			},
			Message: "This command deploys to production and affects live users.",
			Source:  "builtin",
		},
		{
			ID: "docker-push", Name: "Review Docker push", Enabled: true,
			Phase: PhasePreToolUse, Action: ActionReview, Severity: SeverityMedium,
			Match:   RuleMatch{ToolName: "Bash", CommandContains: []string{"docker push"}},
			Message: "Pushing a Docker image publishes it to a registry and may overwrite a tag.",
			Source:  "builtin",
		},
		{
			ID: "kubectl-destructive", Name: "Review kubectl apply/delete", Enabled: true,
			Phase: PhasePreToolUse, Action: ActionReview, Severity: SeverityHigh,
			Match:   RuleMatch{ToolName: "Bash", CommandRegex: `\bkubectl\s+(apply|delete)\b`},
			Message: "kubectl apply/delete modifies live Kubernetes workloads.",
			Source:  "builtin",
		},
		{
			ID: "terraform-apply", Name: "Review Terraform apply/destroy", Enabled: true,
			Phase: PhasePreToolUse, Action: ActionReview, Severity: SeverityHigh,
			Match:   RuleMatch{ToolName: "Bash", CommandRegex: `\bterraform\s+(apply|destroy)\b`},
			Message: "terraform apply/destroy modifies real cloud infrastructure.",
			Source:  "builtin",
		},
		// ── Secret file access ────────────────────────────────────────────
		// Matches real secret files. Explicitly excludes .env.example / .env.sample
		// which are intentionally committed and inspected by developers.
		{
			ID: "secret-file-read", Name: "Review reading secret files", Enabled: true,
			Phase: PhasePreToolUse, Action: ActionReview, Severity: SeverityMedium,
			Match: RuleMatch{
				ToolNames: []string{"Bash", "Read"},
				// Match .env / .env.local / .env.production etc. but NOT .env.example or .env.sample.
				CommandRegex: `(^|[\s/])\.env(\.(local|prod|production|staging|test))?(\s|$|")|[\s/](id_rsa|id_ed25519|\.aws/credentials|\.kube/config|\.npmrc|\.pypirc)(\s|$|")`,
			},
			Message: "This accesses a file that likely contains real secrets or credentials.",
			Source:  "builtin",
		},
		// ── MCP mutation tools ────────────────────────────────────────────
		{
			ID: "mcp-mutation", Name: "Review mutating MCP tool", Enabled: true,
			Phase: PhasePreToolUse, Action: ActionReview, Severity: SeverityMedium,
			Match: RuleMatch{
				ToolPrefix:    "mcp__",
				ToolNameRegex: `__(create|update|delete|remove|send|post|publish|merge|approve|refund|charge|transfer|invite|execute|run|write|deploy|close|archive)(_|\z)`,
			},
			Message: "This MCP tool appears to mutate external state.",
			Source:  "builtin",
		},
		// ── PostToolUse: secret scanning ─────────────────────────────────
		{
			ID: "secret-scan-output", Name: "Redact secrets from tool output", Enabled: true,
			Phase: PhasePostToolUse, Action: ActionRedact, Severity: SeverityHigh,
			Match:   RuleMatch{OutputSecretScan: true},
			Message: "Common secret-like values were redacted before entering context.",
			Source:  "builtin",
		},
		// ── PostToolUse: prompt-injection scanning ────────────────────────
		{
			ID: "injection-scan-output", Name: "Sanitize prompt-injection from tool output", Enabled: true,
			Phase: PhasePostToolUse, Action: ActionSanitize, Severity: SeverityHigh,
			Match:   RuleMatch{OutputInjectionScan: true},
			Message: "Suspicious instruction-like text was removed from untrusted tool output.",
			Source:  "builtin",
		},
	}
}

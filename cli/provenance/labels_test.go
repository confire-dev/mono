package provenance

import (
	"testing"

	"github.com/confire-dev/confire/intercept"
)

func TestClassify_NativeLocalTool(t *testing.T) {
	event := intercept.InterceptEvent{
		Host:    "claude-code",
		Session: intercept.Session{ID: "sess1"},
		Tool:    &intercept.Tool{Name: "Bash", IsMCP: false},
	}
	label := Classify(event, intercept.SanitizeReport{})
	if label.TrustLevel != TrustLocal {
		t.Errorf("Bash: want trusted_local, got %s", label.TrustLevel)
	}
}

func TestClassify_WebFetch(t *testing.T) {
	event := intercept.InterceptEvent{
		Host:    "claude-code",
		Session: intercept.Session{ID: "sess1"},
		Tool: &intercept.Tool{
			Name:  "WebFetch",
			IsMCP: false,
			Input: map[string]any{"url": "https://example.com/page"},
		},
	}
	label := Classify(event, intercept.SanitizeReport{})
	if label.TrustLevel != TrustExternalUntrusted {
		t.Errorf("WebFetch: want external_untrusted, got %s", label.TrustLevel)
	}
	if label.OriginDomain != "example.com" {
		t.Errorf("WebFetch domain: want example.com, got %q", label.OriginDomain)
	}
}

func TestClassify_KnownMCP(t *testing.T) {
	event := intercept.InterceptEvent{
		Host:    "claude-code",
		Session: intercept.Session{ID: "sess1"},
		Tool:    &intercept.Tool{Name: "mcp__github__get_pull_request", IsMCP: true, MCPServer: "github"},
	}
	label := Classify(event, intercept.SanitizeReport{})
	if label.TrustLevel != TrustMCPTrusted {
		t.Errorf("github MCP: want mcp_trusted, got %s", label.TrustLevel)
	}
}

func TestClassify_UnknownMCP(t *testing.T) {
	event := intercept.InterceptEvent{
		Host:    "claude-code",
		Session: intercept.Session{ID: "sess1"},
		Tool:    &intercept.Tool{Name: "mcp__unknowntool__do_something", IsMCP: true, MCPServer: "unknowntool"},
	}
	label := Classify(event, intercept.SanitizeReport{})
	if label.TrustLevel != TrustMCPUnknown {
		t.Errorf("unknown MCP: want mcp_unknown, got %s", label.TrustLevel)
	}
	if !containsFlag(label.Flags, FlagUnknownMCP) {
		t.Error("unknown MCP: expected unknown_mcp flag")
	}
}

func TestClassify_SecretUpgradesToSensitive(t *testing.T) {
	event := intercept.InterceptEvent{
		Host:    "claude-code",
		Session: intercept.Session{ID: "sess1"},
		Tool:    &intercept.Tool{Name: "Read", IsMCP: false},
	}
	report := intercept.SanitizeReport{SecretsRedacted: 2, SecretTypes: []string{"aws_access_key"}}
	label := Classify(event, report)
	if label.TrustLevel != TrustSensitive {
		t.Errorf("secret found: want sensitive, got %s", label.TrustLevel)
	}
	if !containsFlag(label.Flags, FlagSecretRedacted) {
		t.Error("expected secret_redacted flag")
	}
	if label.RedactionsCount != 2 {
		t.Errorf("expected RedactionsCount=2, got %d", label.RedactionsCount)
	}
}

func TestClassify_InjectionFlags(t *testing.T) {
	event := intercept.InterceptEvent{
		Host:    "claude-code",
		Session: intercept.Session{ID: "sess1"},
		Tool:    &intercept.Tool{Name: "mcp__slack__get_messages", IsMCP: true, MCPServer: "slack"},
	}
	report := intercept.SanitizeReport{InjectionFound: true}
	label := Classify(event, report)
	if !containsFlag(label.Flags, FlagPromptInjection) {
		t.Error("expected prompt_injection_detected flag")
	}
	if !containsFlag(label.Flags, FlagSanitized) {
		t.Error("expected sanitized flag")
	}
}

func TestIsAgentAuthoredTool(t *testing.T) {
	authored := []string{"Write", "Edit", "NotebookEdit", "TodoWrite", "write", "EDIT"}
	for _, name := range authored {
		if !IsAgentAuthoredTool(name) {
			t.Errorf("IsAgentAuthoredTool(%q) = false, want true", name)
		}
	}

	notAuthored := []string{"Bash", "Read", "Grep", "Glob", "WebFetch", "mcp__github__get_pull_request", ""}
	for _, name := range notAuthored {
		if IsAgentAuthoredTool(name) {
			t.Errorf("IsAgentAuthoredTool(%q) = true, want false", name)
		}
	}
}

func TestExtractDomain(t *testing.T) {
	cases := []struct{ url, want string }{
		{"https://example.com/path", "example.com"},
		{"http://api.example.com:8080/v1", "api.example.com"},
		{"https://github.com", "github.com"},
	}
	for _, c := range cases {
		got := extractDomain(c.url)
		if got != c.want {
			t.Errorf("extractDomain(%q): want %q, got %q", c.url, c.want, got)
		}
	}
}

func containsFlag(flags []string, flag string) bool {
	for _, f := range flags {
		if f == flag {
			return true
		}
	}
	return false
}

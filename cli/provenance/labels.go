// Package provenance classifies tool outputs with trust levels and context labels.
// It answers "where did this context come from, and how trusted is it?"
package provenance

import (
	"strings"
	"time"

	"github.com/confire-dev/confire/intercept"
)

// TrustLevel describes how much the source of a tool output should be trusted.
type TrustLevel string

const (
	TrustLocal             TrustLevel = "trusted_local"
	TrustWorkspace         TrustLevel = "workspace"
	TrustExternalUntrusted TrustLevel = "external_untrusted"
	TrustMCPUnknown        TrustLevel = "mcp_unknown"
	TrustMCPTrusted        TrustLevel = "mcp_trusted"
	TrustSensitive         TrustLevel = "sensitive"
	TrustAgentGenerated    TrustLevel = "agent_generated"
)

// Flag constants for provenance labels.
const (
	FlagSecretRedacted  = "secret_redacted"
	FlagPromptInjection = "prompt_injection_detected"
	FlagHiddenUnicode   = "hidden_unicode_stripped"
	FlagHiddenText      = "hidden_text_removed"
	FlagCredentialLure  = "credential_lure_detected"
	FlagUnknownMCP      = "unknown_mcp"
	FlagMutationResult  = "mutation_result"
	FlagSanitized       = "sanitized"
)

// ProvenanceLabel is a metadata record for a single tool output.
// It is written to the local session JSONL and optionally synced to Supabase.
// Raw tool content is never included.
type ProvenanceLabel struct {
	ContextID         string     `json:"context_id"`
	SessionID         string     `json:"session_id"`
	Client            string     `json:"client"`
	SourceTool        string     `json:"source_tool"`
	MCPServer         string     `json:"mcp_server,omitempty"`
	OriginDomain      string     `json:"origin_domain,omitempty"`
	TrustLevel        TrustLevel `json:"trust_level"`
	Flags             []string   `json:"flags,omitempty"`
	RedactionsCount   int        `json:"redactions_count,omitempty"`
	SanitizationCount int        `json:"sanitization_count,omitempty"`
	RiskScore         int        `json:"risk_score,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
}

// TrustedMCPServers is the canonical allowlist of known, vetted MCP servers.
// Used by both provenance classification and the MCP risk classifier.
var TrustedMCPServers = map[string]bool{
	"github":      true,
	"slack":       true,
	"jira":        true,
	"atlassian":   true,
	"figma":       true,
	"confluence":  true,
	"clickup":     true,
	"notion":      true,
	"amplitude":   true,
	"fireflies":   true,
	"zapier":      true,
	"linear":      true,
	"asana":       true,
	"intercom":    true,
	"hubspot":     true,
	"salesforce":  true,
}

// nativeFetchTools are native tools that retrieve external content.
var nativeFetchTools = map[string]bool{
	"WebFetch":   true,
	"WebSearch":  true,
}

// nativeLocalTools are native tools that operate on local filesystem/shell.
var nativeLocalTools = map[string]bool{
	"Bash":         true,
	"Read":         true,
	"Write":        true,
	"Edit":         true,
	"Glob":         true,
	"Grep":         true,
	"NotebookRead": true,
	"NotebookEdit": true,
	"TodoRead":     true,
	"TodoWrite":    true,
	"LS":           true,
}

// Classify builds a ProvenanceLabel for a completed tool call.
// report is the sanitization report from the PostToolUse pass (may be zero-value).
func Classify(event intercept.InterceptEvent, report intercept.SanitizeReport) ProvenanceLabel {
	tool := event.Tool
	if tool == nil {
		return ProvenanceLabel{
			SessionID:  event.Session.ID,
			Client:     string(event.Host),
			TrustLevel: TrustAgentGenerated,
			CreatedAt:  time.Now(),
		}
	}

	label := ProvenanceLabel{
		SessionID: event.Session.ID,
		Client:    string(event.Host),
		SourceTool: tool.Name,
		MCPServer:  tool.MCPServer,
		CreatedAt:  time.Now(),
	}

	// Classify trust level.
	label.TrustLevel = classifyTrust(tool)

	// Extract origin domain from WebFetch input URL.
	if tool.Name == "WebFetch" {
		if url := extractURL(tool.Input); url != "" {
			label.OriginDomain = extractDomain(url)
		}
	}

	// Build flags from sanitization report.
	if report.SecretsRedacted > 0 {
		label.Flags = append(label.Flags, FlagSecretRedacted)
		label.RedactionsCount = report.SecretsRedacted
		label.TrustLevel = TrustSensitive // upgrade: secrets were found
	}
	if report.InjectionFound {
		label.Flags = append(label.Flags, FlagPromptInjection)
		label.SanitizationCount++
	}
	if report.HiddenUnicodeFound {
		label.Flags = append(label.Flags, FlagHiddenUnicode)
		label.SanitizationCount++
	}
	if report.SecretsRedacted > 0 || report.InjectionFound || report.HiddenUnicodeFound {
		label.Flags = append(label.Flags, FlagSanitized)
	}
	if label.TrustLevel == TrustMCPUnknown {
		label.Flags = append(label.Flags, FlagUnknownMCP)
	}

	// Mutation detection: if the tool name suggests it writes/sends data.
	if isMutatingTool(tool.Name) {
		label.Flags = append(label.Flags, FlagMutationResult)
	}

	return label
}

func classifyTrust(tool *intercept.Tool) TrustLevel {
	if !tool.IsMCP {
		if nativeFetchTools[tool.Name] {
			return TrustExternalUntrusted // native tool, but fetches external content
		}
		return TrustLocal
	}
	// MCP tool — check against trusted server allowlist.
	server := strings.ToLower(tool.MCPServer)
	if server == "" {
		server = extractServerFromName(tool.Name)
	}
	if TrustedMCPServers[server] {
		return TrustMCPTrusted
	}
	return TrustMCPUnknown
}

// extractServerFromName extracts the server hint from an MCP tool name
// like "mcp__github__get_pull_request" → "github".
func extractServerFromName(name string) string {
	name = strings.ToLower(name)
	if !strings.HasPrefix(name, "mcp__") {
		return ""
	}
	parts := strings.SplitN(name[5:], "__", 2)
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

// extractURL pulls a URL string from a tool's input (handles string or map).
func extractURL(input any) string {
	if input == nil {
		return ""
	}
	switch v := input.(type) {
	case string:
		return v
	case map[string]any:
		if url, ok := v["url"].(string); ok {
			return url
		}
	}
	return ""
}

// extractDomain extracts the hostname from a URL string without importing net/url.
func extractDomain(url string) string {
	// Strip scheme.
	for _, prefix := range []string{"https://", "http://"} {
		if strings.HasPrefix(url, prefix) {
			url = url[len(prefix):]
			break
		}
	}
	// Strip path.
	if i := strings.IndexByte(url, '/'); i >= 0 {
		url = url[:i]
	}
	// Strip port.
	if i := strings.LastIndexByte(url, ':'); i >= 0 {
		url = url[:i]
	}
	return url
}

var mutatingVerbs = []string{
	"write", "delete", "remove", "create", "update", "push", "publish",
	"send", "post", "email", "upload", "transfer", "share", "export",
	"deploy", "destroy", "drop", "truncate", "wipe", "purge",
}

func isMutatingTool(name string) bool {
	name = strings.ToLower(name)
	for _, v := range mutatingVerbs {
		if strings.Contains(name, v) {
			return true
		}
	}
	return false
}

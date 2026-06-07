package cmd

import (
	"fmt"
	"strings"

	"github.com/confire-dev/confire/config"
	"github.com/confire-dev/confire/intercept"
)

// sessionSummary formats the per-session security summary printed on SessionEnd.
func sessionSummary(sess *sessionStats, cfg config.Config) string {
	if !cfg.Notifications.Enabled || cfg.Notifications.Style == "off" {
		return ""
	}
	// Only print if there was security activity worth surfacing.
	hasActivity := sess.blockedCalls > 0 || sess.reviewedCalls > 0 ||
		sess.sanitizedCalls > 0 || sess.secretsRedacted > 0 ||
		sess.mcpUnknownCalls > 0 || sess.externalUntrustedCalls > 0
	if !hasActivity {
		return ""
	}

	lines := []string{colorCyan + "[confire] session summary" + colorReset}
	if sess.blockedCalls > 0 {
		lines = append(lines, fmt.Sprintf("   🚫 blocked:    %d", sess.blockedCalls))
	}
	if sess.reviewedCalls > 0 {
		lines = append(lines, fmt.Sprintf("   ⚠  reviewed:   %d", sess.reviewedCalls))
	}
	if sess.warnedCalls > 0 {
		lines = append(lines, fmt.Sprintf("   ⚠  warned:     %d", sess.warnedCalls))
	}
	if sess.sanitizedCalls > 0 {
		lines = append(lines, fmt.Sprintf("   ⚡ sanitized:  %d", sess.sanitizedCalls))
	}
	if sess.secretsRedacted > 0 {
		lines = append(lines, fmt.Sprintf("   🔑 secrets redacted: %d", sess.secretsRedacted))
	}
	if sess.externalUntrustedCalls > 0 || sess.mcpUnknownCalls > 0 {
		lines = append(lines, "   trust distribution:")
		if sess.externalUntrustedCalls > 0 {
			lines = append(lines, fmt.Sprintf("     external_untrusted: %d", sess.externalUntrustedCalls))
		}
		if sess.mcpUnknownCalls > 0 {
			lines = append(lines, fmt.Sprintf("     mcp_unknown:        %d", sess.mcpUnknownCalls))
		}
	}
	lines = append(lines, fmt.Sprintf("   total tool calls: %d", sess.totalCalls))
	return strings.Join(lines, "\n")
}

// notifyResult is kept for compatibility with PostToolSteer hosts.
// Security results are now logged directly in handleConn.
func notifyResult(_ intercept.InterceptResult, _ string, _ config.Config) (stderrLine, contextLine string) {
	return "", ""
}

func normalizeToolName(toolName string) string {
	lower := strings.ToLower(toolName)
	known := map[string]string{
		"figma":        "Figma",
		"github":       "GitHub",
		"atlassian":    "Jira",
		"confluence":   "Confluence",
		"clickup":      "ClickUp",
		"slack":        "Slack",
		"notion":       "Notion",
		"amplitude":    "Amplitude",
		"fireflies":    "Fireflies",
		"playwright":   "Browser",
		"zapier":       "Zapier",
		"google_drive": "Google Drive",
		"bash":         "Bash",
		"read":         "Read",
		"webfetch":     "WebFetch",
	}
	for key, display := range known {
		if strings.Contains(lower, key) {
			return display
		}
	}
	if len(toolName) > 0 {
		return strings.ToUpper(toolName[:1]) + toolName[1:]
	}
	return toolName
}

package cmd

import (
	"fmt"
	"strings"

	"github.com/confire-dev/confire/config"
	"github.com/confire-dev/confire/intercept"
)

// notifyResult formats a save notification based on the user's preferences.
//
// Returns two strings:
//   stderrLine:   a line to print to stderr (developer sees it; Claude doesn't)
//   contextLine:  a line for additionalContext (Claude sees it; only for huge saves)
//
// Thresholds (from config):
//   < min_saved_tokens:   silent
//   < big_save_tokens:    🔥 → stderr only
//   ≥ big_save_tokens:    🔥 → stderr + additionalContext to Claude
func notifyResult(
	result intercept.InterceptResult,
	toolName string,
	cfg config.Config,
) (stderrLine, contextLine string) {
	if !cfg.Notifications.Enabled || cfg.Notifications.Style == "off" {
		return "", ""
	}
	if result.Stats == nil {
		return "", ""
	}

	savedBytes  := result.Stats.BeforeBytes - result.Stats.AfterBytes
	savedTokens := savedBytes / 4 // rough estimate: 4 bytes ≈ 1 token

	if savedTokens < cfg.Notifications.MinSavedTokens {
		return "", ""
	}

	rawTok  := result.Stats.BeforeBytes / 4
	optTok  := result.Stats.AfterBytes / 4
	tool    := normalizeToolName(toolName)

	switch cfg.Notifications.Style {
	case "minimal":
		stderrLine = fmt.Sprintf("[confire] %s: ~%s → ~%s tokens",
			tool, tokenStr(rawTok), tokenStr(optTok))

	default: // "brand"
		if savedTokens >= cfg.Notifications.BigSaveTokens {
			stderrLine = fmt.Sprintf("🔥 Confire saved ~%s tokens on this %s response.",
				tokenStr(savedTokens), tool)
		} else {
			stderrLine = fmt.Sprintf("🔥 Confire optimized %s: ~%s → ~%s tokens.",
				tool, tokenStr(rawTok), tokenStr(optTok))
		}
		contextLine = stderrLine
	}

	return stderrLine, contextLine
}

// sessionSummary formats the per-session 🔥 summary printed on SessionEnd.
func sessionSummary(sess *sessionStats, cfg config.Config) string {
	if !cfg.Notifications.Enabled || cfg.Notifications.Style == "off" {
		return ""
	}
	if sess.optimizedCalls == 0 {
		return ""
	}

	savedTokens := int(sess.rawBytes-sess.optimizedBytes) / 4

	switch cfg.Notifications.Style {
	case "minimal":
		return fmt.Sprintf("[confire] %d calls optimized · ~%s tokens saved",
			sess.optimizedCalls, tokenStr(savedTokens))

	default: // brand
		lines := []string{
			"🔥 Confire session summary",
			fmt.Sprintf("   %d tool calls optimized", sess.optimizedCalls),
			fmt.Sprintf("   ~%s tokens saved", tokenStr(savedTokens)),
		}
		if sess.biggestWinTool != "" {
			lines = append(lines, fmt.Sprintf("   Biggest win: %s · ~%s → ~%s",
				sess.biggestWinTool,
				tokenStr(sess.biggestWinRaw/4),
				tokenStr(sess.biggestWinOpt/4),
			))
		}
		return strings.Join(lines, "\n")
	}
}

// ── Helpers ────────────────────────────────────────────────────────────────

func tokenStr(n int) string {
	if n >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	}
	if n >= 1_000 {
		return fmt.Sprintf("%.0fk", float64(n)/1_000)
	}
	return fmt.Sprintf("%d", n)
}

func normalizeToolName(toolName string) string {
	// "mcp__figma__get_design_context" → "Figma MCP"
	// "Bash" → "Bash"
	lower := strings.ToLower(toolName)
	known := map[string]string{
		"figma":        "Figma MCP",
		"github":       "GitHub PR",
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
		"read":         "File read",
		"webfetch":     "Web fetch",
	}
	for key, display := range known {
		if strings.Contains(lower, key) {
			return display
		}
	}
	// Capitalize first letter for unknown tools
	if len(toolName) > 0 {
		return strings.ToUpper(toolName[:1]) + toolName[1:]
	}
	return toolName
}

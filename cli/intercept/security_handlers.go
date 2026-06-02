package intercept

// ContentScanHandlers returns all handlers that scan tool *output* content
// for security threats (injection, secrets, hidden unicode, etc.).
//
// This is the single registration point for the security harness.
// To add a new content-scanning handler: append it here.
// The harness in security_harness_test.go picks it up automatically.
//
// Excluded by design:
//   - MCPRiskHandler  — scores tool *names*, not content (tool.pre)
//   - guardrail.Handler — policy firewall on tool invocation (tool.pre)
func ContentScanHandlers() []Handler {
	return []Handler{
		&MCPSanitizeHandler{},
	}
}

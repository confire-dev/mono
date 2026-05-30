package intercept

import "strings"

// GetServerHint returns a string that the optimizer registry uses to pick
// the right optimizer. For MCP tools it returns the server name ("figma");
// for native tools it returns the tool name ("Bash").
func (t *Tool) GetServerHint() string {
	if t == nil {
		return ""
	}
	if t.IsMCP && t.MCPServer != "" {
		return strings.ToLower(t.MCPServer)
	}
	return strings.ToLower(t.Name)
}

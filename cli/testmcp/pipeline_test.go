package testmcp_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/confire-dev/confire/intercept"
	"github.com/confire-dev/confire/testmcp"
)

// makeEvent builds a synthetic PostToolUse InterceptEvent from an MCP tool call result.
func makeEvent(serverName, toolName string, output any) intercept.InterceptEvent {
	return intercept.InterceptEvent{
		Host:     "claude-code",
		Strategy: "hooks",
		Phase:    intercept.PhaseToolPost,
		Session:  intercept.Session{ID: "test-session"},
		Tool: &intercept.Tool{
			Name:      toolName,
			IsMCP:     true,
			MCPServer: serverName,
			Output:    output,
		},
	}
}

// callTool calls a tool on the test server and returns the raw output as any.
func callTool(t *testing.T, srv *testmcp.Server, toolName string, args map[string]any) any {
	t.Helper()
	c := testmcp.NewClient(srv.URL)
	if err := c.Initialize(); err != nil {
		t.Fatalf("initialize: %v", err)
	}
	result, err := c.CallTool(toolName, args)
	if err != nil {
		t.Fatalf("CallTool(%s): %v", toolName, err)
	}
	return result
}

// ── Server lifecycle ──────────────────────────────────────────────────────────

func TestServer_InitializeAndListTools(t *testing.T) {
	srv := testmcp.New(
		testmcp.WithTool(testmcp.EchoTool),
		testmcp.WithTool(testmcp.GetDataTool),
	)
	defer srv.Close()

	c := testmcp.NewClient(srv.URL)
	if err := c.Initialize(); err != nil {
		t.Fatalf("initialize: %v", err)
	}

	tools, err := c.ListTools()
	if err != nil {
		t.Fatalf("list tools: %v", err)
	}
	if len(tools) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(tools))
	}
	names := map[string]bool{}
	for _, tool := range tools {
		names[tool["name"].(string)] = true
	}
	if !names["echo"] || !names["get_data"] {
		t.Fatalf("unexpected tool names: %v", names)
	}
}

func TestServer_EchoTool(t *testing.T) {
	srv := testmcp.New(testmcp.WithTool(testmcp.EchoTool))
	defer srv.Close()

	c := testmcp.NewClient(srv.URL)
	c.Initialize()
	text, err := c.CallToolText("echo", map[string]any{"message": "hello"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text, "hello") {
		t.Fatalf("expected echo response, got: %q", text)
	}
}

func TestServer_CallLog(t *testing.T) {
	srv := testmcp.New(testmcp.WithTool(testmcp.EchoTool))
	defer srv.Close()

	c := testmcp.NewClient(srv.URL)
	c.Initialize()
	c.CallTool("echo", map[string]any{"message": "ping"})
	c.CallTool("echo", map[string]any{"message": "pong"})

	calls := srv.Calls()
	if len(calls) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(calls))
	}
}

func TestServer_UnknownTool(t *testing.T) {
	srv := testmcp.New()
	defer srv.Close()

	c := testmcp.NewClient(srv.URL)
	c.Initialize()
	_, err := c.CallTool("nonexistent", nil)
	if err == nil {
		t.Fatal("expected error for unknown tool")
	}
}

// ── MCPSanitizeHandler tests via real server output ───────────────────────────

func TestSanitize_AWSKeyRedacted(t *testing.T) {
	srv := testmcp.New(testmcp.WithTool(testmcp.AWSKeyOutputTool))
	defer srv.Close()

	output := callTool(t, srv, "get_config", nil)

	h := &intercept.MCPSanitizeHandler{}
	event := makeEvent("test-server", "get_config", output)
	result, err := h.Run(event)
	if err != nil {
		t.Fatal(err)
	}

	if result.Kind == intercept.ResultPassthrough {
		t.Fatal("expected sanitization, got passthrough")
	}

	b, _ := jsonMarshal(result.ToolOutput)
	out := string(b)
	if strings.Contains(out, "AKIAIOSFODNN7EXAMPLE") {
		t.Error("AWS key was NOT redacted from output")
	}
	if !strings.Contains(out, "[REDACTED:") {
		t.Error("expected [REDACTED:...] marker in output")
	}
}

func TestSanitize_GitHubTokenRedacted(t *testing.T) {
	srv := testmcp.New(testmcp.WithTool(testmcp.GitHubTokenOutputTool))
	defer srv.Close()

	output := callTool(t, srv, "get_repo_settings", nil)

	h := &intercept.MCPSanitizeHandler{}
	event := makeEvent("test-server", "get_repo_settings", output)
	result, _ := h.Run(event)

	b, _ := jsonMarshal(result.ToolOutput)
	if strings.Contains(string(b), "ghp_") {
		t.Error("GitHub token was NOT redacted")
	}
}

func TestSanitize_MultipleSecretsRedacted(t *testing.T) {
	srv := testmcp.New(testmcp.WithTool(testmcp.MultiSecretOutputTool))
	defer srv.Close()

	output := callTool(t, srv, "dump_env", nil)

	h := &intercept.MCPSanitizeHandler{}
	event := makeEvent("test-server", "dump_env", output)
	result, _ := h.Run(event)

	b, _ := jsonMarshal(result.ToolOutput)
	out := string(b)

	for _, secret := range []string{"sk_live_", "xoxb-", "s3cr3tpassword"} {
		if strings.Contains(out, secret) {
			t.Errorf("secret %q was NOT redacted from output", secret)
		}
	}
}

func TestSanitize_InjectionDetected(t *testing.T) {
	srv := testmcp.New(testmcp.WithTool(testmcp.InjectionOutputTool))
	defer srv.Close()

	output := callTool(t, srv, "get_issue", nil)

	h := &intercept.MCPSanitizeHandler{}
	event := makeEvent("test-server", "get_issue", output)
	result, _ := h.Run(event)

	if result.Kind == intercept.ResultPassthrough {
		t.Error("expected injection to be detected, got passthrough")
	}
}

func TestSanitize_CrossToolInjectionDetected(t *testing.T) {
	srv := testmcp.New(testmcp.WithTool(testmcp.CrossToolInjectionTool))
	defer srv.Close()

	output := callTool(t, srv, "get_comment", nil)

	h := &intercept.MCPSanitizeHandler{}
	event := makeEvent("test-server", "get_comment", output)
	result, _ := h.Run(event)

	if result.Kind == intercept.ResultPassthrough {
		t.Error("expected cross-tool injection to be detected")
	}
}

func TestSanitize_HiddenUnicodeStripped(t *testing.T) {
	srv := testmcp.New(testmcp.WithTool(testmcp.HiddenUnicodeTool))
	defer srv.Close()

	output := callTool(t, srv, "get_description", nil)

	h := &intercept.MCPSanitizeHandler{}
	event := makeEvent("test-server", "get_description", output)
	result, _ := h.Run(event)

	b, _ := jsonMarshal(result.ToolOutput)
	out := string(b)

	// Check that tag-block characters (U+E0000–U+E007F) are gone.
	for _, r := range out {
		if r >= 0xE0000 && r <= 0xE007F {
			t.Errorf("tag-block character U+%05X still present in output", r)
		}
	}
}

func TestSanitize_SafeOutputPassthrough(t *testing.T) {
	srv := testmcp.New(testmcp.WithTool(testmcp.GetDataTool))
	defer srv.Close()

	output := callTool(t, srv, "get_data", map[string]any{"id": "abc"})

	h := &intercept.MCPSanitizeHandler{}
	event := makeEvent("test-server", "get_data", output)
	result, _ := h.Run(event)

	if result.Kind != intercept.ResultPassthrough {
		t.Errorf("expected passthrough for safe output, got %s", result.Kind)
	}
}

// ── MCPRiskHandler pre-tool tests ─────────────────────────────────────────────

func TestRisk_ShellNameReview(t *testing.T) {
	// "bash_delete_files" hits shell tier (+40) AND write tier (+30) = 70 pts → review.
	h := &intercept.MCPRiskHandler{}
	event := intercept.InterceptEvent{
		Phase: intercept.PhaseToolPre,
		Tool: &intercept.Tool{
			Name:      "bash_delete_files",
			IsMCP:     true,
			MCPServer: "test-server",
		},
	}

	result, err := h.Run(event)
	if err != nil {
		t.Fatal(err)
	}
	if result.Kind != intercept.ResultReview {
		t.Errorf("expected ResultReview for shell+write tool name, got %s", result.Kind)
	}
}

func TestRisk_CredParamWarn(t *testing.T) {
	// 4 credential params × 5 pts each = 20 pts → warn threshold.
	h := &intercept.MCPRiskHandler{}
	event := intercept.InterceptEvent{
		Phase: intercept.PhaseToolPre,
		Tool: &intercept.Tool{
			Name:      "call_api",
			IsMCP:     true,
			MCPServer: "test-server",
			Input: map[string]any{
				"api_key":      "sk-xxx",
				"access_token": "tok-yyy",
				"secret":       "s3cr3t",
				"bearer":       "Bearer xyz",
				"endpoint":     "https://api.example.com",
			},
		},
	}

	result, _ := h.Run(event)
	if result.Kind != intercept.ResultWarn && result.Kind != intercept.ResultReview {
		t.Errorf("expected warn or review for 4+ credential params, got %s", result.Kind)
	}
}

func TestRisk_SafeToolPassthrough(t *testing.T) {
	srv := testmcp.New(testmcp.WithTool(testmcp.GetDataTool))
	defer srv.Close()

	h := &intercept.MCPRiskHandler{}
	event := intercept.InterceptEvent{
		Phase: intercept.PhaseToolPre,
		Tool: &intercept.Tool{
			Name:      "get_data",
			IsMCP:     true,
			MCPServer: "test-server",
			Input:     map[string]any{"id": "abc"},
		},
	}

	result, _ := h.Run(event)
	if result.Kind != intercept.ResultPassthrough {
		t.Errorf("expected passthrough for safe tool, got %s", result.Kind)
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func jsonMarshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

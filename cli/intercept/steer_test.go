package intercept

import (
	"strings"
	"testing"
)

func TestFormatPostToolSteer_NativeSecretsAndInjection(t *testing.T) {
	got := FormatPostToolSteer(PostToolSteerInput{
		Host:                "cursor",
		ToolName:            "Shell",
		NativeUnreplaceable: true,
		Report: SanitizeReport{
			SecretsRedacted: 2,
			SecretTypes:     []string{"github_token", "jwt"},
			InjectionFound:    true,
		},
	})

	if !strings.HasPrefix(got, postToolSteerHeader) {
		t.Fatalf("expected standardized header, got: %q", got)
	}
	for _, want := range []string{
		"tool=Shell",
		"host=cursor",
		"mode=native_unreplaceable",
		"secrets_redacted=2",
		"secret_types=github_token,jwt",
		"injection_sanitized=true",
		"agent_instruction:",
		"Do not repeat secrets",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}
}

func TestFormatPostToolSteer_MCPReplaced(t *testing.T) {
	got := FormatPostToolSteer(PostToolSteerInput{
		Host:           "cursor",
		ToolName:       "MCP: figma/get_file",
		OutputReplaced: true,
		Stats: &Stats{
			BeforeBytes: 10000,
			AfterBytes:  500,
			Optimizer:   "worker/figma",
		},
	})

	if !strings.Contains(got, "mode=mcp_replaced") {
		t.Fatalf("expected mcp_replaced mode, got:\n%s", got)
	}
	if strings.Contains(got, "native_output_not_replaceable") {
		t.Fatal("MCP steer should not mention native unreplaceable")
	}
}

func TestFormatPostToolSteer_EmptyWhenNothingToSay(t *testing.T) {
	if got := FormatPostToolSteer(PostToolSteerInput{Host: "cursor", ToolName: "Read"}); got != "" {
		t.Fatalf("expected empty steer, got %q", got)
	}
}

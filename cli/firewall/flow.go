// Package firewall implements cross-tool flow detection (Phase 3).
// It detects dangerous action chains by inspecting a ring buffer of recent
// provenance labels accumulated during a session.
package firewall

import (
	"fmt"
	"strings"

	"github.com/confire-dev/confire/intercept"
	"github.com/confire-dev/confire/policy"
	"github.com/confire-dev/confire/provenance"
)

// CheckFlowRules evaluates the 5 cross-tool chain rules against recent labels.
// It returns a non-passthrough result if a dangerous chain is detected.
// Called in the PreToolUse path before the per-tool guardrail runs.
func CheckFlowRules(
	recentLabels []provenance.ProvenanceLabel,
	event intercept.InterceptEvent,
	mode policy.Mode,
) intercept.InterceptResult {
	if event.Tool == nil || len(recentLabels) == 0 {
		return intercept.InterceptResult{Kind: intercept.ResultPassthrough}
	}

	for _, check := range flowRules {
		if result, triggered := check(recentLabels, event, mode); triggered {
			return result
		}
	}
	return intercept.InterceptResult{Kind: intercept.ResultPassthrough}
}

type flowRule func(
	recent []provenance.ProvenanceLabel,
	event intercept.InterceptEvent,
	mode policy.Mode,
) (intercept.InterceptResult, bool)

var flowRules = []flowRule{
	ruleUntrustedThenSecretRead,
	ruleSecretReadThenExternalSend,
	ruleInjectionThenShell,
	ruleUnknownMCPThenMutatingMCP,
	ruleCredentialLureThenMessage,
}

// ruleUntrustedThenSecretRead: recent untrusted output → current tool reads a secret file.
func ruleUntrustedThenSecretRead(
	recent []provenance.ProvenanceLabel,
	event intercept.InterceptEvent,
	_ policy.Mode,
) (intercept.InterceptResult, bool) {
	if !hasRecentTrust(recent, provenance.TrustMCPUnknown, provenance.TrustExternalUntrusted) {
		return intercept.InterceptResult{}, false
	}
	if !isSecretFileAccess(event.Tool) {
		return intercept.InterceptResult{}, false
	}
	return intercept.InterceptResult{
		Kind: intercept.ResultReview,
		Reason: flowReview(
			"Untrusted source before secret file access",
			event.Tool,
			"An untrusted external source recently provided context. Reading a credential file immediately after may indicate a prompt-injection attack guiding secret exfiltration.",
		),
	}, true
}

// ruleSecretReadThenExternalSend: recent sensitive/secret output → current tool sends data externally.
func ruleSecretReadThenExternalSend(
	recent []provenance.ProvenanceLabel,
	event intercept.InterceptEvent,
	mode policy.Mode,
) (intercept.InterceptResult, bool) {
	if !hasRecentFlag(recent, provenance.FlagSecretRedacted) && !hasRecentTrust(recent, provenance.TrustSensitive) {
		return intercept.InterceptResult{}, false
	}
	if !isExternalSend(event.Tool) {
		return intercept.InterceptResult{}, false
	}
	if mode == policy.ModeStrict {
		return intercept.InterceptResult{
			Kind: intercept.ResultBlock,
			Reason: flowBlock(
				"Secret output before external send",
				event.Tool,
				"A secret was recently found in tool output. Sending data externally immediately after raises an active credential exfiltration risk. Blocked in strict mode.",
			),
		}, true
	}
	return intercept.InterceptResult{
		Kind: intercept.ResultReview,
		Reason: flowReview(
			"Secret output before external send",
			event.Tool,
			"A secret was recently found in tool output. Sending data externally immediately after risks credential exfiltration.",
		),
	}, true
}

// ruleInjectionThenShell: recent injection-detected output → current tool is a shell command.
func ruleInjectionThenShell(
	recent []provenance.ProvenanceLabel,
	event intercept.InterceptEvent,
	_ policy.Mode,
) (intercept.InterceptResult, bool) {
	if !hasRecentFlag(recent, provenance.FlagPromptInjection) {
		return intercept.InterceptResult{}, false
	}
	if !isShellTool(event.Tool) {
		return intercept.InterceptResult{}, false
	}
	return intercept.InterceptResult{
		Kind: intercept.ResultReview,
		Reason: flowReview(
			"Injection detected before shell command",
			event.Tool,
			"A prompt injection attempt was detected in recent tool output. Running a shell command immediately after could execute attacker-controlled instructions.",
		),
	}, true
}

// ruleUnknownMCPThenMutatingMCP: recent unknown-MCP output → current tool is a mutating MCP call.
func ruleUnknownMCPThenMutatingMCP(
	recent []provenance.ProvenanceLabel,
	event intercept.InterceptEvent,
	_ policy.Mode,
) (intercept.InterceptResult, bool) {
	if !hasRecentTrust(recent, provenance.TrustMCPUnknown) {
		return intercept.InterceptResult{}, false
	}
	if !event.Tool.IsMCP || !isMutatingMCPTool(event.Tool.Name) {
		return intercept.InterceptResult{}, false
	}
	return intercept.InterceptResult{
		Kind: intercept.ResultReview,
		Reason: flowReview(
			"Unknown MCP source before mutating action",
			event.Tool,
			"An unrecognized MCP server recently returned data. Performing a mutating operation based on that data could propagate attacker-controlled changes.",
		),
	}, true
}

// ruleCredentialLureThenMessage: recent credential-lure flag → current tool sends a message.
func ruleCredentialLureThenMessage(
	recent []provenance.ProvenanceLabel,
	event intercept.InterceptEvent,
	_ policy.Mode,
) (intercept.InterceptResult, bool) {
	if !hasRecentFlag(recent, provenance.FlagCredentialLure) {
		return intercept.InterceptResult{}, false
	}
	if !isMessageSendTool(event.Tool) {
		return intercept.InterceptResult{}, false
	}
	return intercept.InterceptResult{
		Kind: intercept.ResultReview,
		Reason: flowReview(
			"Credential lure before message send",
			event.Tool,
			"A credential lure pattern was detected in recent output. Sending a message immediately after could spread the lure to other users.",
		),
	}, true
}

// ── Flow message formatting ───────────────────────────────────────────────────

func flowReview(ruleName string, tool *intercept.Tool, risk string) string {
	return fmt.Sprintf(
		"CONFIRE REVIEW REQUIRED\nRule:    %s\nTool:    %s\nCommand: %s\nRisk:    %s\nTo allow once, run:\n  confire bypass-next\nThen ask the agent to retry.",
		ruleName, tool.Name, flowInputSummary(tool, 120), risk,
	)
}

func flowBlock(ruleName string, tool *intercept.Tool, reason string) string {
	return fmt.Sprintf(
		"CONFIRE BLOCKED TOOL CALL\nRule:    %s\nTool:    %s\nCommand: %s\nReason:  %s\n\nThis action has been blocked.\nIf this is intentional, update your Confire policy or switch modes outside the agent session.",
		ruleName, tool.Name, flowInputSummary(tool, 120), reason,
	)
}

func flowInputSummary(tool *intercept.Tool, maxLen int) string {
	if tool == nil {
		return ""
	}
	var s string
	switch v := tool.Input.(type) {
	case string:
		s = v
	case map[string]any:
		if cmd, ok := v["command"].(string); ok {
			s = cmd
		} else if path, ok := v["file_path"].(string); ok {
			s = path
		} else {
			s = fmt.Sprintf("%v", v)
		}
	default:
		s = fmt.Sprintf("%v", tool.Input)
	}
	s = strings.TrimSpace(s)
	if len(s) > maxLen {
		s = s[:maxLen] + "…"
	}
	return s
}

// ── Predicate helpers ─────────────────────────────────────────────────────────

func hasRecentTrust(recent []provenance.ProvenanceLabel, levels ...provenance.TrustLevel) bool {
	for _, label := range recent {
		for _, level := range levels {
			if label.TrustLevel == level {
				return true
			}
		}
	}
	return false
}

func hasRecentFlag(recent []provenance.ProvenanceLabel, flag string) bool {
	for _, label := range recent {
		for _, f := range label.Flags {
			if f == flag {
				return true
			}
		}
	}
	return false
}

var secretFilePaths = []string{
	".env", "id_rsa", "id_ed25519", ".pem", ".key", ".p12", ".pfx",
	"credentials", "secrets.yaml", "secrets.json", ".aws",
}

func isSecretFileAccess(tool *intercept.Tool) bool {
	if tool.IsMCP {
		return false
	}
	name := strings.ToLower(tool.Name)
	if name != "read" && name != "bash" && name != "glob" && name != "grep" {
		return false
	}
	path := extractInputPath(tool)
	if path == "" {
		return false
	}
	path = strings.ToLower(path)
	for _, pattern := range secretFilePaths {
		if strings.Contains(path, pattern) {
			return true
		}
	}
	return false
}

var externalSendVerbs = []string{
	"send", "post", "email", "upload", "transfer", "share", "export",
	"publish", "push", "webhook", "notify", "message",
}

func isExternalSend(tool *intercept.Tool) bool {
	name := strings.ToLower(tool.Name)
	for _, v := range externalSendVerbs {
		if strings.Contains(name, v) {
			return true
		}
	}
	return false
}

func isShellTool(tool *intercept.Tool) bool {
	if tool.IsMCP {
		return false
	}
	name := strings.ToLower(tool.Name)
	return name == "bash" || strings.Contains(name, "exec") || strings.Contains(name, "shell")
}

var mutatingMCPVerbs = []string{
	"write", "delete", "remove", "create", "update", "push", "publish",
	"send", "post", "deploy", "destroy", "drop", "wipe",
}

func isMutatingMCPTool(name string) bool {
	name = strings.ToLower(name)
	for _, v := range mutatingMCPVerbs {
		if strings.Contains(name, v) {
			return true
		}
	}
	return false
}

var messageSendVerbs = []string{
	"send", "message", "email", "post", "notify", "share", "forward",
}

func isMessageSendTool(tool *intercept.Tool) bool {
	name := strings.ToLower(tool.Name)
	for _, v := range messageSendVerbs {
		if strings.Contains(name, v) {
			return true
		}
	}
	return false
}

func extractInputPath(tool *intercept.Tool) string {
	switch v := tool.Input.(type) {
	case string:
		return v
	case map[string]any:
		for _, key := range []string{"file_path", "path", "command", "query"} {
			if s, ok := v[key].(string); ok {
				return s
			}
		}
	}
	return ""
}

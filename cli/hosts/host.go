// Package hosts defines the Host interface and the central registry of
// all supported AI coding agents.
//
// A Host knows three things:
//   1. How to detect itself on the system.
//   2. Which integration strategies it supports.
//   3. How to install and verify a strategy.
//
// The rest of the codebase — setup, status, config — never hard-codes
// host-specific logic. It loops over hosts.Registry() and delegates.
//
// Author: Efe <efe@efebehar.dev>
package hosts

// Strategy is the mechanism confire uses to integrate with a host.
type Strategy string

const (
	// StrategyHooks uses the host's native lifecycle hook system.
	// Supported by: Claude Code, Cursor, VS Code, Windsurf, Codex, Cline, OpenCode, OpenClaw.
	// Capabilities vary per host — see capabilities.go.
	StrategyHooks Strategy = "hooks"

	// StrategyMCPProxy intercepts at the MCP transport layer.
	// Fallback for MCP-speaking hosts without a native hook API.
	StrategyMCPProxy Strategy = "mcp-proxy"
)

// InstallOptions carries the runtime data needed to install a strategy.
type InstallOptions struct {
	BinaryPath   string // absolute path to the confire binary
	WorkerURL    string // Worker URL for cloud optimization
	SettingsPath string // override settings file path; "" = host default (global)
}

// Host represents one AI coding agent that confire can integrate with.
type Host interface {
	// ID is a stable machine identifier ("claude-code", "cursor", …).
	ID() string

	// Label is the human-readable display name ("Claude Code", "Cursor", …).
	Label() string

	// Detect returns true if this host is installed on the current system.
	Detect() bool

	// Strategies returns the integration mechanisms this host supports,
	// in order of preference (best first).
	Strategies() []Strategy

	// Preferred returns the recommended strategy for this host.
	Preferred() Strategy

	// Install configures the given strategy for this host.
	// Idempotent: safe to call on an already-installed host.
	Install(s Strategy, opts InstallOptions) error

	// IsInstalled returns true if confire is already configured for s.
	IsInstalled(s Strategy) bool

	// Uninstall removes confire's configuration for s.
	Uninstall(s Strategy) error

	// ComingSoon returns true when integration is not yet available.
	// status and setup show these differently from uninstalled-but-supported hosts.
	ComingSoon() bool
}

// Registry returns all known hosts in display order.
// Claude Code is always first (it's the primary target).
func Registry() []Host {
	return []Host{
		&ClaudeCodeHost{},
		&CursorHost{},
		&VSCodeHost{},
		&ClineHost{},
		&WindsurfHost{},
		&CodexHost{},
		&OpenCodeHost{},
		&OpenClawHost{},
	}
}

// Detect returns only the hosts currently installed on this system.
func Detect() []Host {
	var out []Host
	for _, h := range Registry() {
		if h.Detect() {
			out = append(out, h)
		}
	}
	return out
}

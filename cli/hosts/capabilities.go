package hosts

// Capabilities describe what Confire should do for a host integration.
// Defaults live here; users can override per host in ~/.confire/config.json.
type Capabilities struct {
	// NativeOutputReplaceable: host hook API can replace native tool output (Claude Code only today).
	NativeOutputReplaceable bool
	// OptimizeNative: run local/cloud optimizers on native tools (Shell, Read, WebFetch).
	OptimizeNative bool
	// OptimizeMCP: run optimizers on MCP tool output (postToolUse replace where supported).
	OptimizeMCP bool
	// PostToolSteer: inject standardized [Confire post_tool steer] additional_context.
	PostToolSteer bool
}

var defaultCapabilities = map[string]Capabilities{
	"claude-code": {
		NativeOutputReplaceable: true,
		OptimizeNative:          true,
		OptimizeMCP:             true,
		PostToolSteer:           true,
	},
	"cursor": {
		NativeOutputReplaceable: false,
		OptimizeNative:          false,
		OptimizeMCP:             true,
		PostToolSteer:           true,
	},
	"vscode": {
		NativeOutputReplaceable: false,
		OptimizeNative:          false,
		OptimizeMCP:             true,
		PostToolSteer:           true,
	},
}

// DefaultCapabilities returns built-in capability defaults for a host ID.
func DefaultCapabilities(hostID string) Capabilities {
	if c, ok := defaultCapabilities[hostID]; ok {
		return c
	}
	// Unknown host: optimize when possible, no steer envelope.
	return Capabilities{
		NativeOutputReplaceable: true,
		OptimizeNative:          true,
		OptimizeMCP:             true,
		PostToolSteer:           false,
	}
}

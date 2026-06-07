package hosts

// Capabilities describe what Confire should do for a host integration.
// Defaults live here; users can override per host in ~/.confire/config.json.
type Capabilities struct {
	// NativeOutputReplaceable: host hook API can replace native tool output (Claude Code only today).
	NativeOutputReplaceable bool
	// PostToolSteer: inject standardized [Confire post_tool steer] additional_context.
	PostToolSteer bool
}

var defaultCapabilities = map[string]Capabilities{
	"claude-code": {
		NativeOutputReplaceable: true,
		PostToolSteer:           true,
	},
	"cursor": {
		NativeOutputReplaceable: false,
		PostToolSteer:           true,
	},
	"vscode": {
		NativeOutputReplaceable: false,
		PostToolSteer:           true,
	},
}

// DefaultCapabilities returns built-in capability defaults for a host ID.
func DefaultCapabilities(hostID string) Capabilities {
	if c, ok := defaultCapabilities[hostID]; ok {
		return c
	}
	// Unknown host: no steer envelope by default.
	return Capabilities{
		NativeOutputReplaceable: true,
		PostToolSteer:           false,
	}
}

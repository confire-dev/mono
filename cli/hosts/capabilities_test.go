package hosts

import "testing"

func TestDefaultCapabilities_CursorSkipsNativeOptimize(t *testing.T) {
	c := DefaultCapabilities("cursor")
	if c.OptimizeNative {
		t.Fatal("cursor should not optimize native tools by default")
	}
	if !c.OptimizeMCP || !c.PostToolSteer {
		t.Fatalf("cursor should optimize MCP and steer: %+v", c)
	}
}

func TestDefaultCapabilities_ClaudeCodeFullOptimize(t *testing.T) {
	c := DefaultCapabilities("claude-code")
	if !c.OptimizeNative || !c.OptimizeMCP {
		t.Fatalf("claude-code should optimize native and MCP: %+v", c)
	}
	if c.PostToolSteer {
		t.Fatal("claude-code should not use post-tool steer envelope")
	}
}

package hosts

import "testing"

func TestDefaultCapabilities_CursorNativeNotReplaceable(t *testing.T) {
	c := DefaultCapabilities("cursor")
	if c.NativeOutputReplaceable {
		t.Fatal("cursor native output should not be replaceable")
	}
	if !c.PostToolSteer {
		t.Fatalf("cursor should use post-tool steer: %+v", c)
	}
}

func TestDefaultCapabilities_ClaudeCodePostToolSteer(t *testing.T) {
	c := DefaultCapabilities("claude-code")
	if !c.NativeOutputReplaceable {
		t.Fatal("claude-code native output should be replaceable")
	}
	if !c.PostToolSteer {
		t.Fatal("claude-code should use post-tool steer envelope (enabled for security advisories)")
	}
}

package config

import "testing"

func TestCapabilitiesFor_CursorDefaults(t *testing.T) {
	caps := Config{}.CapabilitiesFor("cursor")
	if !caps.PostToolSteer {
		t.Fatalf("cursor should use post-tool steer by default: %+v", caps)
	}
	if caps.NativeOutputReplaceable {
		t.Fatal("cursor native output is not replaceable")
	}
}

func TestCapabilitiesFor_Override(t *testing.T) {
	steer := false
	cfg := Config{
		Hosts: map[string]HostSettings{
			"cursor": {PostToolSteer: &steer},
		},
	}
	caps := cfg.CapabilitiesFor("cursor")
	if caps.PostToolSteer {
		t.Fatal("expected config override to disable post-tool steer for cursor")
	}
}

package config

import "testing"

func TestCapabilitiesFor_CursorDefaults(t *testing.T) {
	caps := Config{}.CapabilitiesFor("cursor")
	if caps.OptimizeNative {
		t.Fatal("cursor should not optimize native tools by default")
	}
	if !caps.OptimizeMCP || !caps.PostToolSteer {
		t.Fatalf("unexpected cursor caps: %+v", caps)
	}
	if caps.NativeOutputReplaceable {
		t.Fatal("cursor native output is not replaceable")
	}
}

func TestCapabilitiesFor_Override(t *testing.T) {
	native := true
	cfg := Config{
		Hosts: map[string]HostSettings{
			"cursor": {OptimizeNative: &native},
		},
	}
	caps := cfg.CapabilitiesFor("cursor")
	if !caps.OptimizeNative {
		t.Fatal("expected config override to enable native optimize for cursor")
	}
}

package hosts

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBackupOnce_CreatesBackupOnFirstWrite(t *testing.T) {
	home := redirectHome(t)
	p := filepath.Join(home, ".cursor", "hooks.json")
	writeJSON(t, p, map[string]any{"version": 1, "hooks": map[string]any{}})

	original := readString(t, p)

	// First install — should create backup.
	installCursorHooksAt(p, "/usr/bin/confire")

	backup := p + ".confire-backup"
	if _, err := os.Stat(backup); err != nil {
		t.Fatalf("backup not created after first install: %v", err)
	}
	if readString(t, backup) != original {
		t.Error("backup content does not match original pre-install file")
	}
}

func TestBackupOnce_SecondWriteDoesNotOverwriteBackup(t *testing.T) {
	home := redirectHome(t)
	p := filepath.Join(home, ".cursor", "hooks.json")
	writeJSON(t, p, map[string]any{"version": 1, "hooks": map[string]any{}})

	original := readString(t, p)
	installCursorHooksAt(p, "/usr/bin/confire")

	// Uninstall modifies the file again — backup must still equal the original.
	uninstallCursorHooksAt(p)

	backup := p + ".confire-backup"
	if readString(t, backup) != original {
		t.Error("backup was overwritten by uninstall — must stay as the original pre-Confire state")
	}
}

func TestBackupOnce_NoBackupForNewFile(t *testing.T) {
	home := redirectHome(t)
	p := filepath.Join(home, ".cursor", "hooks.json")
	// Do NOT create the file — install on a fresh machine.
	os.MkdirAll(filepath.Dir(p), 0700)

	installCursorHooksAt(p, "/usr/bin/confire")

	backup := p + ".confire-backup"
	if _, err := os.Stat(backup); err == nil {
		t.Error("backup should not be created when no original file exists")
	}
}

func TestBackupOnce_AllClients(t *testing.T) {
	// Verify every JSON-based client creates a backup on first install.
	clients := []struct {
		name    string
		setup   func(home string) string
		install func(p string) error
	}{
		{
			"cursor",
			func(home string) string {
				p := filepath.Join(home, ".cursor", "hooks.json")
				writeJSON(t, p, map[string]any{"version": 1})
				return p
			},
			func(p string) error { return installCursorHooksAt(p, "/usr/bin/confire") },
		},
		{
			"windsurf",
			func(home string) string {
				p := filepath.Join(home, ".windsurf", "hooks.json")
				writeJSON(t, p, map[string]any{"PreToolUse": []any{}})
				return p
			},
			func(p string) error { return installWindsurfHooks("/usr/bin/confire", false) },
		},
		{
			"codex",
			func(home string) string {
				p := filepath.Join(home, ".codex", "hooks.json")
				writeJSON(t, p, map[string]any{"version": 1})
				return p
			},
			func(p string) error { return installCodexHooks("/usr/bin/confire", false) },
		},
	}

	for _, tc := range clients {
		t.Run(tc.name, func(t *testing.T) {
			home := redirectHome(t)
			p := tc.setup(home)
			original := readString(t, p)

			if err := tc.install(p); err != nil {
				t.Fatalf("install: %v", err)
			}

			backup := p + ".confire-backup"
			if _, err := os.Stat(backup); err != nil {
				t.Fatalf("backup not created: %v", err)
			}
			if readString(t, backup) != original {
				t.Error("backup does not match original")
			}
		})
	}
}

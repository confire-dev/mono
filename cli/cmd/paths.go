package cmd

import (
	"os"
	"path/filepath"

	"github.com/confire-dev/confire/config"
)

// localSettingsPath walks up from cwd to find the git root, then returns
// {root}/.claude/settings.json. Falls back to {cwd}/.claude/settings.json.
func localSettingsPath() string {
	cwd, _ := os.Getwd()
	dir := cwd
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return filepath.Join(dir, ".claude", "settings.json")
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return filepath.Join(cwd, ".claude", "settings.json")
}

// globalSettingsPath returns the user-wide Claude Code settings path.
func globalSettingsPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude", "settings.json")
}

func confireDir() string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".confire")
	os.MkdirAll(dir, 0700)
	return dir
}

func daemonSocketPath() string {
	return filepath.Join(confireDir(), "daemon.sock")
}

func daemonPIDPath() string {
	return filepath.Join(confireDir(), "daemon.pid")
}

func apiKeyEnv() string {
	return os.Getenv("CONFIRE_KEY")
}

// workerURLEnv resolves the Worker base URL in priority order:
//  1. --local flag → localhost:8787
//  2. CONFIRE_API env var
//  3. config.WorkerURL (set via `confire config set worker_url=...`)
//  4. production default
func workerURLEnv() string {
	if flagLocal {
		return "http://localhost:8787"
	}
	if u := os.Getenv("CONFIRE_API"); u != "" {
		return u
	}
	cfg := config.Load()
	if cfg.WorkerURL != "" {
		return cfg.WorkerURL
	}
	return "https://api.confire.dev"
}

// platformURL resolves the platform base URL:
//  1. --local flag → localhost:4321
//  2. CONFIRE_PLATFORM_URL env var
//  3. production default
func platformURL() string {
	if flagLocal {
		return "http://localhost:4321"
	}
	if u := os.Getenv("CONFIRE_PLATFORM_URL"); u != "" {
		return u
	}
	return "https://confire.dev"
}

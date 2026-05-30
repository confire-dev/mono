package cmd

import (
	"os"
	"path/filepath"
)

func confireDir() string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".confire")
	os.MkdirAll(dir, 0700)
	return dir
}

func daemonSocketPath() string {
	return filepath.Join(confireDir(), "daemon.sock")
}

func apiKeyEnv() string {
	return os.Getenv("CONFIRE_KEY")
}

func workerURLEnv() string {
	if u := os.Getenv("CONFIRE_API"); u != "" {
		return u
	}
	return "https://api.confire.dev"
}

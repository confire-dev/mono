package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update confire to the latest version",
	Long: `Checks get.confire.dev for a newer release, downloads the binary for the
current platform, replaces the running executable, and restarts the daemon.

confire-dev builds do not have a release channel — rebuild from source instead:
  cd cli && make build-dev-env`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runUpdate()
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

type latestManifest struct {
	Version     string            `json:"version"`
	PublishedAt string            `json:"publishedAt"`
	Assets      map[string]string `json:"assets"`
}

func runUpdate() error {
	if buildBinaryName != "confire" {
		fmt.Printf("%s is a dev-env build — no release channel available.\n", buildBinaryName)
		fmt.Println("Rebuild from source:  cd cli && make build-dev-env && make install-dev-env")
		return nil
	}

	// 1. Fetch the latest manifest.
	fmt.Print("Checking for updates... ")
	manifest, err := fetchUpdateManifest()
	if err != nil {
		return fmt.Errorf("could not fetch latest version: %w", err)
	}
	fmt.Printf("latest: %s\n", manifest.Version)

	if buildVersion == manifest.Version {
		fmt.Println("Already up to date.")
		return nil
	}
	if buildVersion == "dev" {
		fmt.Printf("Local build detected — will install %s.\n", manifest.Version)
	} else {
		fmt.Printf("Updating %s → %s\n", buildVersion, manifest.Version)
	}

	// 2. Find the asset URL for the current platform.
	assetKey := fmt.Sprintf("confire_%s_%s", runtime.GOOS, runtime.GOARCH)
	assetURL, ok := manifest.Assets[assetKey]
	if !ok {
		return fmt.Errorf("no binary available for %s/%s — download manually from %s", runtime.GOOS, runtime.GOARCH, manifest.Assets["confire_checksums.txt"])
	}

	// 3. Resolve the path of the running executable (follow symlinks).
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not determine executable path: %w", err)
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return fmt.Errorf("could not resolve executable path: %w", err)
	}

	// 4. Stop the daemon if it is running — we will restart it after the swap.
	wasRunning := isDaemonRunning()
	if wasRunning {
		fmt.Print("Stopping optimizer... ")
		if err := stopDaemon(); err != nil {
			fmt.Printf("warning: %v\n", err)
		} else {
			// Give the daemon a moment to exit cleanly.
			for i := 0; i < 15; i++ {
				time.Sleep(100 * time.Millisecond)
				if !isDaemonRunning() {
					break
				}
			}
			fmt.Println("stopped.")
		}
	}

	// 5. Download the new binary into a temp file in the same directory
	//    so the final os.Rename is always on the same filesystem.
	fmt.Printf("Downloading... ")
	dir := filepath.Dir(exe)
	tmp, err := os.CreateTemp(dir, ".confire-update-*")
	if err != nil {
		return fmt.Errorf("could not create temp file in %s (try sudo?): %w", dir, err)
	}
	tmpPath := tmp.Name()

	downloadErr := downloadBinary(tmp, assetURL)
	tmp.Close()
	if downloadErr != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("download failed: %w", downloadErr)
	}
	fmt.Println("done.")

	// 6. Make the new binary executable.
	if err := os.Chmod(tmpPath, 0755); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("could not chmod new binary: %w", err)
	}

	// 7. Atomic replace. os.Rename is atomic on POSIX when src and dst share
	//    the same filesystem, which is guaranteed because we created the temp
	//    file in the same directory.
	fmt.Printf("Installing to %s... ", exe)
	if err := os.Rename(tmpPath, exe); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("could not replace binary (try: sudo confire update): %w", err)
	}
	fmt.Println("done.")

	// 8. Restart the daemon with the new binary.
	if wasRunning {
		fmt.Print("Restarting optimizer... ")
		if err := launchDaemon(); err != nil {
			fmt.Printf("warning: could not restart: %v\n", err)
			fmt.Println("Run `confire start` to restart manually.")
		} else {
			for i := 0; i < 10; i++ {
				time.Sleep(100 * time.Millisecond)
				if isDaemonRunning() {
					break
				}
			}
			fmt.Println("restarted.")
		}
	}

	fmt.Printf("\n✓ confire updated to %s\n", manifest.Version)
	return nil
}

func fetchUpdateManifest() (*latestManifest, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(getURL() + "/latest.json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	var m latestManifest
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return nil, err
	}
	if m.Version == "" {
		return nil, fmt.Errorf("manifest missing version field")
	}
	return &m, nil
}

func downloadBinary(w io.Writer, url string) error {
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d from %s", resp.StatusCode, url)
	}
	_, err = io.Copy(w, resp.Body)
	return err
}

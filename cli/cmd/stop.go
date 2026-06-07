package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the Confire firewall",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runStop()
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}

func runStop() error {
	if !isDaemonRunning() {
		fmt.Println("Firewall is not running.")
		return nil
	}

	if err := stopDaemon(); err != nil {
		return fmt.Errorf("failed to stop firewall: %w", err)
	}

	// Wait for the socket to disappear.
	for i := 0; i < 20; i++ {
		time.Sleep(100 * time.Millisecond)
		if !isDaemonRunning() {
			fmt.Println("✓ Firewall stopped.")
			return nil
		}
	}
	fmt.Println("✓ Stop signal sent.")
	return nil
}

// stopDaemon sends SIGTERM to the daemon via its PID file.
func stopDaemon() error {
	data, err := os.ReadFile(daemonPIDPath())
	if err != nil {
		return fmt.Errorf("PID file not found — daemon may have been started outside of confire")
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return fmt.Errorf("invalid PID file: %w", err)
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("process not found: %w", err)
	}
	return proc.Signal(syscall.SIGTERM)
}

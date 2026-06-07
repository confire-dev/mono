package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the Confire firewall daemon in the background",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runStart()
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}

func runStart() error {
	if isDaemonRunning() {
		fmt.Printf("%s✓%s  Firewall daemon is already running.\n", green, reset)
		return nil
	}

	if err := launchDaemon(); err != nil {
		return fmt.Errorf("failed to start firewall daemon: %w", err)
	}

	// Give it a moment to bind the socket.
	for i := 0; i < 10; i++ {
		time.Sleep(100 * time.Millisecond)
		if isDaemonRunning() {
			fmt.Println("✓ Firewall daemon started.")
			return nil
		}
	}
	return fmt.Errorf("firewall daemon started but isn't responding — check logs")
}

// launchDaemon starts `confire daemon` as a detached background process.
func launchDaemon() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	cmd := exec.Command(exe, "daemon")
	// Detach from the terminal so the daemon survives after the shell exits.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil
	return cmd.Start()
}

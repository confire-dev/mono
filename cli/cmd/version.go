package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// Build metadata — set by ldflags in Makefile, never hardcoded.
var (
	buildVersion     = "dev"
	buildCommit      = "none"
	buildDate        = "unknown"
	buildBuiltBy     = "source"
	buildAuthor      = "Efe <efe@efebehar.dev>"
	buildBinaryName  = "confire"  // overridden to "confire-dev" for dev-env builds
	buildWorkerURL   = ""         // overridden for non-prod builds; empty = use hardcoded fallback
	buildPlatformURL = ""         // overridden for non-prod builds; empty = use hardcoded fallback
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Confire %s\n", buildVersion)
		fmt.Printf("Commit:    %s\n", buildCommit)
		fmt.Printf("Built:     %s\n", buildDate)
		fmt.Printf("Built by:  %s\n", buildBuiltBy)
		fmt.Printf("Go:        %s %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
		fmt.Printf("Copyright © 2026 Trana, Inc.\n")
		fmt.Printf("License:   Proprietary\n")
		fmt.Printf("Website:   https://confire.dev\n")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

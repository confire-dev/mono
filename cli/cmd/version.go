package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// Build metadata — set by ldflags in Makefile, never hardcoded.
var (
	buildVersion = "dev"
	buildCommit  = "none"
	buildDate    = "unknown"
	buildAuthor  = "Efe <efe@efebehar.dev>"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("confire %s\n", buildVersion)
		fmt.Printf("  commit : %s\n", buildCommit)
		fmt.Printf("  built  : %s\n", buildDate)
		fmt.Printf("  author : %s\n", buildAuthor)
		fmt.Printf("  go     : %s %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

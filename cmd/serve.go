package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmustier/economist-cli/internal/daemon"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run a background daemon for faster reads",
	Long: `Run a local daemon that keeps a headless browser warm.

The daemon listens on a local Unix socket and speeds up article reads.

Examples:
  economist serve
  economist serve &`,
	RunE: runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func runServe(cmd *cobra.Command, args []string) error {
	fmt.Println("Starting economist serve daemon...")
	return daemon.Serve()
}

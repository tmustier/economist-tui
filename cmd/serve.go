package cmd

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmustier/economist-tui/internal/daemon"
	appErrors "github.com/tmustier/economist-tui/internal/errors"
)

var (
	serveStatus bool
	serveStop   bool
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run a background daemon for faster reads",
	Long: `Run a local daemon that keeps a headless browser warm.

The daemon listens on a local Unix socket and speeds up article reads.

Examples:
  economist serve
  economist serve &
  economist serve --status
  economist serve --stop`,
	RunE: runServe,
}

func init() {
	serveCmd.Flags().BoolVar(&serveStatus, "status", false, "Show daemon status")
	serveCmd.Flags().BoolVar(&serveStop, "stop", false, "Stop the daemon")
	rootCmd.AddCommand(serveCmd)
}

func runServe(cmd *cobra.Command, args []string) error {
	if serveStatus && serveStop {
		return appErrors.NewUserError("--status and --stop are mutually exclusive")
	}

	if serveStatus {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		latency, running, err := daemon.Status(ctx)
		if err != nil {
			return err
		}
		if running {
			fmt.Printf("running (%s)\n", latency)
		} else {
			fmt.Println("not running")
		}
		return nil
	}

	if serveStop {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := daemon.Shutdown(ctx); err != nil {
			if errors.Is(err, daemon.ErrNotRunning) {
				fmt.Println("not running")
				return nil
			}
			return err
		}
		fmt.Println("stopped")
		return nil
	}

	fmt.Println("Starting economist serve daemon...")
	return daemon.Serve()
}

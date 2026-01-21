package cmd

import (
	"context"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmustier/economist-cli/internal/browse"
	"github.com/tmustier/economist-cli/internal/daemon"
	appErrors "github.com/tmustier/economist-cli/internal/errors"
	"github.com/tmustier/economist-cli/internal/logging"
	"github.com/tmustier/economist-cli/internal/ui"
)

var browseCmd = &cobra.Command{
	Use:   "browse [section]",
	Short: "Browse headlines interactively",
	Long: `Browse headlines in an interactive TUI.

Use ↑/↓ to navigate, Enter to read, b to go back, c to toggle columns, q to quit.

Examples:
  economist browse
  economist browse finance`,
	Args: cobra.MaximumNArgs(1),
	RunE: runBrowse,
}

func init() {
	rootCmd.AddCommand(browseCmd)
}

func runBrowse(cmd *cobra.Command, args []string) error {
	if !ui.IsTerminal(int(os.Stdin.Fd())) {
		return appErrors.NewUserError("browse requires an interactive terminal - use 'headlines --json' for scripts")
	}

	logging.Debugf(debugMode, "browse: ensure daemon")
	if err := daemon.EnsureBackground(); err != nil {
		logging.Debugf(debugMode, "browse: daemon start error: %v", err)
	}
	if debugMode {
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		if latency, running, err := daemon.Status(ctx); err == nil {
			logging.Debugf(debugMode, "browse: daemon status running=%t latency=%s", running, latency)
		} else {
			logging.Debugf(debugMode, "browse: daemon status error: %v", err)
		}
	}

	section := "leaders"
	if len(args) > 0 {
		section = args[0]
	}

	return browse.Run(section, browse.Options{Debug: debugMode, NoColor: noColor})
}

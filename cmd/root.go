package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	appErrors "github.com/tmustier/economist-tui/internal/errors"
)

var (
	debugMode bool
	noColor   bool
	version   = "0.1.0"
	commit    = ""
	date      = ""
)

var rootCmd = &cobra.Command{
	Use:           "economist",
	Short:         "Terminal UI to read The Economist",
	Long:          `A terminal UI and CLI to browse and read articles from The Economist.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		if appErrors.IsUserError(err) {
			fmt.Fprintln(os.Stderr, err)
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(1)
	}
}

func buildVersion() string {
	v := version
	if commit != "" {
		v += " (" + commit + ")"
	}
	if date != "" {
		v += " " + date
	}
	return v
}

func detectNoColor() bool {
	if os.Getenv("NO_COLOR") != "" {
		return true
	}
	return os.Getenv("TERM") == "dumb"
}

func init() {
	noColor = detectNoColor()
	rootCmd.PersistentFlags().BoolVar(&debugMode, "debug", false, "Enable debug output")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", noColor, "Disable color output")
	rootCmd.Version = buildVersion()
	rootCmd.SetVersionTemplate("{{.Version}}\n")

	rootCmd.AddCommand(headlinesCmd)
	rootCmd.AddCommand(readCmd)
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(sectionsCmd)
}

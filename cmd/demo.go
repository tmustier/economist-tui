package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/tmustier/economist-tui/internal/browse"
	"github.com/tmustier/economist-tui/internal/demo"
	appErrors "github.com/tmustier/economist-tui/internal/errors"
	"github.com/tmustier/economist-tui/internal/ui"
)

var demoCmd = &cobra.Command{
	Use:   "demo [section]",
	Short: "Browse demo content",
	Long: `Browse a safe demo feed with placeholder content.

No login is required and all content is locally generated.

Examples:
  economist demo
  economist demo the-americas`,
	Args: cobra.MaximumNArgs(1),
	RunE: runDemo,
}

func init() {
	rootCmd.AddCommand(demoCmd)
}

func runDemo(cmd *cobra.Command, args []string) error {
	if !ui.IsTerminal(int(os.Stdin.Fd())) {
		return appErrors.NewUserError("demo requires an interactive terminal")
	}

	section := demo.DefaultSection
	if len(args) > 0 {
		section = args[0]
	}

	source := demo.NewSource()
	opts := browse.Options{Debug: debugMode, NoColor: noColor, Source: source}
	return browse.Run(section, opts)
}

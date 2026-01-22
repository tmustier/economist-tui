package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tmustier/economist-tui/internal/article"
	"github.com/tmustier/economist-tui/internal/config"
	appErrors "github.com/tmustier/economist-tui/internal/errors"
	"github.com/tmustier/economist-tui/internal/fetch"
	"github.com/tmustier/economist-tui/internal/ui"
)

var (
	rawOutput bool
	wrapWidth int
	columns   int
)

var readCmd = &cobra.Command{
	Use:   "read [url|-]",
	Short: "Read an article",
	Long: `Fetch and display a full article in the terminal.

Requires login first: economist login

Examples:
  economist read https://www.economist.com/leaders/2026/01/15/some-article
  economist read <url> --raw
  echo "https://www.economist.com/..." | economist read -`,
	Args: cobra.RangeArgs(0, 1),
	RunE: runRead,
}

func init() {
	readCmd.Flags().BoolVar(&rawOutput, "raw", false, "Output raw markdown")
	readCmd.Flags().IntVar(&wrapWidth, "wrap", 0, "Wrap width for rendered output (0 = auto)")
	readCmd.Flags().IntVar(&columns, "columns", 1, "Number of columns for article body (1 or 2)")
}

func runRead(cmd *cobra.Command, args []string) error {
	url, err := resolveURL(args)
	if err != nil {
		return err
	}

	if columns < 1 || columns > 2 {
		return appErrors.NewUserError("columns must be 1 or 2")
	}

	if !config.IsLoggedIn() {
		fmt.Fprintln(os.Stderr, "⚠️  Not logged in. Run 'economist login' first.")
		fmt.Fprintln(os.Stderr, "   (Articles behind the paywall require authentication)")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "   Attempting to fetch anyway (may hit paywall)...")
		fmt.Fprintln(os.Stderr, "")
	}

	art, err := fetch.FetchArticle(url, fetch.Options{Debug: debugMode})
	if err != nil {
		return err
	}

	if debugMode && art.DebugHTMLPath != "" {
		fmt.Fprintf(os.Stderr, "Debug HTML saved to: %s\n", art.DebugHTMLPath)
	}

	return outputArticle(art)
}

func outputArticle(art *article.Article) error {
	opts := ui.ArticleRenderOptions{
		Raw:       rawOutput,
		NoColor:   noColor,
		WrapWidth: wrapWidth,
		TwoColumn: columns == 2,
	}
	if wrapWidth == 0 && columns == 1 && ui.IsTerminal(int(os.Stdout.Fd())) {
		termWidth := ui.TermWidth(int(os.Stdout.Fd()))
		opts.TermWidth = termWidth
		opts.WrapWidth = ui.ReaderContentWidth(termWidth)
		opts.Center = true
	}

	out, err := ui.RenderArticle(art, opts)
	if err != nil {
		return err
	}

	fmt.Print(out)
	return nil
}

func resolveURL(args []string) (string, error) {
	if len(args) == 1 && args[0] != "-" {
		return args[0], nil
	}

	if len(args) == 0 || (len(args) == 1 && args[0] == "-") {
		if !stdinHasData() {
			return "", appErrors.NewUserError("no URL provided - pass a URL or use stdin")
		}
		return readURLFromStdin()
	}

	return "", appErrors.NewUserError("invalid arguments")
}

func stdinHasData() bool {
	info, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice == 0
}

func readURLFromStdin() (string, error) {
	data, err := io.ReadAll(bufio.NewReader(os.Stdin))
	if err != nil {
		return "", err
	}
	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return "", appErrors.NewUserError("no URL found on stdin")
	}
	return fields[0], nil
}

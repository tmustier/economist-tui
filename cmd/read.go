package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/spf13/cobra"
	"github.com/tmustier/economist-cli/internal/article"
	"github.com/tmustier/economist-cli/internal/browser"
	"github.com/tmustier/economist-cli/internal/config"
	"github.com/tmustier/economist-cli/internal/daemon"
	appErrors "github.com/tmustier/economist-cli/internal/errors"
)

var (
	rawOutput bool
	wrapWidth int
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
	readCmd.Flags().IntVar(&wrapWidth, "wrap", 100, "Wrap width for rendered output (0 = no wrap)")
}

func runRead(cmd *cobra.Command, args []string) error {
	url, err := resolveURL(args)
	if err != nil {
		return err
	}

	if !config.IsLoggedIn() {
		fmt.Fprintln(os.Stderr, "⚠️  Not logged in. Run 'economist login' first.")
		fmt.Fprintln(os.Stderr, "   (Articles behind the paywall require authentication)")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "   Attempting to fetch anyway (may hit paywall)...")
		fmt.Fprintln(os.Stderr, "")
	}

	art, err := fetchArticle(url)
	if err != nil {
		return err
	}

	if debugMode && art.DebugHTMLPath != "" {
		fmt.Fprintf(os.Stderr, "Debug HTML saved to: %s\n", art.DebugHTMLPath)
	}

	return outputArticle(art)
}

func fetchArticle(url string) (*article.Article, error) {
	if !debugMode {
		ctx, cancel := context.WithTimeout(context.Background(), browser.FetchTimeout)
		defer cancel()
		if art, err := daemon.Fetch(ctx, url, false); err == nil {
			return validateArticle(art)
		} else if !errors.Is(err, daemon.ErrNotRunning) {
			return nil, err
		}
	}

	art, err := fetchArticleLocal(url)
	if err != nil {
		return nil, err
	}

	return validateArticle(art)
}

func fetchArticleLocal(url string) (*article.Article, error) {
	art, err := article.Fetch(url, article.FetchOptions{Debug: debugMode})
	if err != nil {
		if appErrors.IsPaywallError(err) {
			return nil, appErrors.NewUserError("paywall detected - run 'economist login' to read full articles")
		}
		return nil, err
	}

	return art, nil
}

func validateArticle(art *article.Article) (*article.Article, error) {
	if art.Content == "" {
		return nil, appErrors.NewUserError("no article content found - try 'economist login'")
	}
	return art, nil
}

func outputArticle(art *article.Article) error {
	md := art.ToMarkdown()

	if rawOutput || noColor {
		fmt.Println(md)
		return nil
	}

	opts := []glamour.TermRendererOption{glamour.WithAutoStyle()}
	if wrapWidth > 0 {
		opts = append(opts, glamour.WithWordWrap(wrapWidth))
	}

	renderer, err := glamour.NewTermRenderer(opts...)
	if err != nil {
		fmt.Println(md)
		return nil
	}

	out, err := renderer.Render(md)
	if err != nil {
		fmt.Println(md)
		return nil
	}

	fmt.Println(out)
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

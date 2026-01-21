package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	appErrors "github.com/tmustier/economist-cli/internal/errors"
	"github.com/tmustier/economist-cli/internal/rss"
	"golang.org/x/term"
)

var (
	headlinesLimit  int
	headlinesSearch string
	headlinesJSON   bool
	headlinesPlain  bool
)

var headlinesCmd = &cobra.Command{
	Use:   "headlines [section]",
	Short: "Show latest headlines from a section",
	Long: `Show latest headlines from The Economist RSS feeds.

Examples:
  economist headlines leaders
  economist headlines finance -n 5
  economist headlines business -s "AI"
  economist headlines finance --json`,
	Args: cobra.MaximumNArgs(1),
	RunE: runHeadlines,
}

func init() {
	headlinesCmd.Flags().IntVarP(&headlinesLimit, "number", "n", 10, "Number of headlines to show")
	headlinesCmd.Flags().StringVarP(&headlinesSearch, "search", "s", "", "Search headlines for a term")
	headlinesCmd.Flags().BoolVar(&headlinesJSON, "json", false, "Output JSON")
	headlinesCmd.Flags().BoolVar(&headlinesPlain, "plain", false, "Output plain text (title\turl)")
}

func runHeadlines(cmd *cobra.Command, args []string) error {
	section := "leaders"
	if len(args) > 0 {
		section = args[0]
	}

	if headlinesJSON && headlinesPlain {
		return appErrors.NewUserError("--json and --plain are mutually exclusive")
	}

	items, title, err := fetchHeadlines(section)
	if err != nil {
		return err
	}

	if headlinesJSON {
		return printHeadlinesJSON(items, section)
	}
	if headlinesPlain {
		printHeadlinesPlain(items)
		return nil
	}

	printHeadlines(items, title)
	return nil
}

func fetchHeadlines(section string) ([]rss.Item, string, error) {
	if headlinesSearch != "" {
		items, err := rss.Search(section, headlinesSearch)
		if err != nil {
			return nil, "", err
		}
		title := fmt.Sprintf("Search: \"%s\" in %s", headlinesSearch, section)
		return items, title, nil
	}

	feed, err := rss.FetchSection(section)
	if err != nil {
		return nil, "", err
	}
	title := strings.TrimSpace(feed.Channel.Title)
	return feed.Channel.Items, title, nil
}

func limitItems(items []rss.Item) []rss.Item {
	if headlinesLimit > 0 && len(items) > headlinesLimit {
		return items[:headlinesLimit]
	}
	return items
}

type headlineOutput struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Date        string `json:"date"`
	PubDate     string `json:"pub_date"`
	URL         string `json:"url"`
	Section     string `json:"section"`
}

func printHeadlinesJSON(items []rss.Item, section string) error {
	items = limitItems(items)
	out := make([]headlineOutput, 0, len(items))
	for _, item := range items {
		out = append(out, headlineOutput{
			Title:       item.CleanTitle(),
			Description: item.CleanDescription(),
			Date:        item.FormattedDate(),
			PubDate:     item.PubDate,
			URL:         item.Link,
			Section:     section,
		})
	}

	data, err := json.Marshal(out)
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(append(data, '\n'))
	return err
}

func printHeadlinesPlain(items []rss.Item) {
	items = limitItems(items)
	for _, item := range items {
		fmt.Printf("%s\t%s\n", item.CleanTitle(), item.Link)
	}
}

func getTermWidth() int {
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
		return w
	}
	return 100
}

func printHeadlines(items []rss.Item, title string) {
	width := getTermWidth()

	// Styles
	titleStyle := lipgloss.NewStyle().Bold(true)
	dimStyle := lipgloss.NewStyle().Faint(true)
	if noColor {
		titleStyle = lipgloss.NewStyle()
		dimStyle = lipgloss.NewStyle()
	}

	// Header
	fmt.Printf("%s\n\n", title)

	items = limitItems(items)

	for i, item := range items {
		num := fmt.Sprintf("%2d. ", i+1)
		headline := item.CleanTitle()
		date := item.FormattedDate()

		// Calculate padding for right-aligned date
		// Format: "NN. Title...                    Date"
		contentWidth := width - len(num) - len(date) - 2
		if contentWidth < 20 {
			contentWidth = 20
		}

		// Truncate title if too long
		displayTitle := headline
		if len(headline) > contentWidth {
			displayTitle = headline[:contentWidth-3] + "..."
		}

		// Build the line with right-aligned date
		padding := contentWidth - len(displayTitle)
		if padding < 1 {
			padding = 1
		}

		fmt.Printf("%s%s%s%s\n",
			num,
			titleStyle.Render(displayTitle),
			strings.Repeat(" ", padding),
			dimStyle.Render(date),
		)

		// Description (indented)
		if desc := item.CleanDescription(); desc != "" {
			descWidth := width - 4
			if len(desc) > descWidth {
				desc = desc[:descWidth-3] + "..."
			}
			fmt.Printf("    %s\n", desc)
		}

		// URL (dimmed, indented)
		fmt.Printf("    %s\n\n", dimStyle.Render(item.Link))
	}

	if len(items) == 0 {
		fmt.Println("No articles found.")
	}
}

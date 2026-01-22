package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	appErrors "github.com/tmustier/economist-cli/internal/errors"
	"github.com/tmustier/economist-cli/internal/rss"
	"github.com/tmustier/economist-cli/internal/ui"
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

func printHeadlines(items []rss.Item, title string) {
	termWidth := ui.TermWidth(int(os.Stdout.Fd()))
	if termWidth <= 0 {
		termWidth = ui.DefaultWidth
	}
	contentWidth := ui.ReaderContentWidth(termWidth)

	styles := ui.NewBrowseStyles(noColor)
	accentStyles := ui.NewStyles(ui.CurrentTheme(), noColor)

	// Header
	fmt.Printf("%s\n", styles.Header.Render(title))
	fmt.Printf("%s\n\n", ui.AccentRule(contentWidth, accentStyles))

	items = limitItems(items)
	if len(items) == 0 {
		fmt.Println("No articles found.")
		return
	}

	numWidth := len(fmt.Sprintf("%d", len(items)))
	prefixWidth := len(fmt.Sprintf("%*d. ", numWidth, len(items)))
	dateWidth := ui.DefaultDateWidth
	useCompactDates := contentWidth < prefixWidth+ui.MinTitleWidth+dateWidth
	if useCompactDates {
		dateWidth = len("02.01.06")
	}
	layout := ui.NewHeadlineLayout(contentWidth, prefixWidth, dateWidth)
	prefixPad := strings.Repeat(" ", prefixWidth)
	subtitleWidth := contentWidth - 4
	if subtitleWidth < 1 {
		subtitleWidth = 1
	}

	for i, item := range items {
		num := fmt.Sprintf("%*d. ", numWidth, i+1)
		headline := item.CleanTitle()
		date := item.FormattedDate()
		if useCompactDates {
			date = item.CompactDate()
		}
		dateColumn := fmt.Sprintf("%*s", layout.DateWidth, date)

		titleLines := ui.WrapLines(headline, layout.TitleWidth)
		if len(titleLines) == 0 {
			titleLines = []string{""}
		}
		for idx, line := range titleLines {
			if idx == 0 {
				paddedTitle := fmt.Sprintf("%-*s", layout.TitleWidth, line)
				fmt.Printf("%s%s%s\n",
					num,
					styles.Title.Render(paddedTitle),
					styles.Dim.Render(dateColumn),
				)
				continue
			}
			if line == "" {
				fmt.Printf("%s\n", prefixPad)
				continue
			}
			fmt.Printf("%s%s\n", prefixPad, styles.Title.Render(line))
		}

		if desc := item.CleanDescription(); desc != "" {
			for _, line := range ui.WrapLines(desc, subtitleWidth) {
				if line == "" {
					fmt.Printf("    \n")
					continue
				}
				fmt.Printf("    %s\n", styles.Subtitle.Render(line))
			}
		}

		// URL (dimmed, indented)
		fmt.Printf("    %s\n\n", styles.Dim.Render(item.Link))
	}
}

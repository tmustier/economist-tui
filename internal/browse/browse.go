package browse

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tmustier/economist-tui/internal/rss"
	"github.com/tmustier/economist-tui/internal/ui"
)

type Options struct {
	Debug   bool
	NoColor bool
}

func Run(section string, opts Options) error {
	sectionTitle, items, err := loadSection(section)
	if err != nil {
		return err
	}

	ui.InitTheme()
	m := NewModel(section, items, sectionTitle, opts)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

func loadSection(section string) (string, []rss.Item, error) {
	feed, err := rss.FetchSection(section)
	if err != nil {
		return "", nil, err
	}

	items := feed.Channel.Items
	if len(items) > 50 {
		items = items[:50]
	}

	sectionTitle := strings.TrimSpace(feed.Channel.Title)
	if sectionTitle == "" {
		sectionTitle = section
	}

	return sectionTitle, items, nil
}

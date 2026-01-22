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
	feed, err := rss.FetchSection(section)
	if err != nil {
		return err
	}

	items := feed.Channel.Items
	if len(items) > 50 {
		items = items[:50]
	}

	sectionTitle := strings.TrimSpace(feed.Channel.Title)
	if sectionTitle == "" {
		sectionTitle = section
	}

	ui.InitTheme()
	m := NewModel(items, sectionTitle, opts)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

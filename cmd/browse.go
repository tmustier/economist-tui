package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/tmustier/economist-cli/internal/article"
	appErrors "github.com/tmustier/economist-cli/internal/errors"
	"github.com/tmustier/economist-cli/internal/rss"
	"golang.org/x/term"
)

var browseCmd = &cobra.Command{
	Use:   "browse [section]",
	Short: "Browse headlines interactively",
	Long: `Browse headlines in an interactive TUI.

Use ↑/↓ or j/k to navigate, Enter to read, q to quit.

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
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return appErrors.NewUserError("browse requires an interactive terminal - use 'headlines --json' for scripts")
	}

	section := "leaders"
	if len(args) > 0 {
		section = args[0]
	}

	feed, err := rss.FetchSection(section)
	if err != nil {
		return err
	}

	items := feed.Channel.Items
	if len(items) > 20 {
		items = items[:20]
	}

	m := initialModel(items, section)
	p := tea.NewProgram(m, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	// If user selected an article, read it
	if fm, ok := finalModel.(model); ok && fm.selected != nil {
		fmt.Println() // Clear line after TUI
		return readArticle(fm.selected.Link)
	}

	return nil
}

func readArticle(url string) error {
	fmt.Printf("Loading article...\n\n")

	art, err := article.Fetch(url, article.FetchOptions{Debug: debugMode})
	if err != nil {
		if appErrors.IsPaywallError(err) {
			return appErrors.NewUserError("paywall detected - run 'economist login' to read full articles")
		}
		return err
	}

	if art.Content == "" {
		return appErrors.NewUserError("no article content found - try 'economist login'")
	}

	return outputArticle(art)
}

// Model

type model struct {
	items    []rss.Item
	section  string
	cursor   int
	selected *rss.Item
	width    int
	height   int
}

func initialModel(items []rss.Item, section string) model {
	w, h, _ := term.GetSize(int(os.Stdout.Fd()))
	if w == 0 {
		w = 80
	}
	if h == 0 {
		h = 24
	}
	return model{
		items:   items,
		section: section,
		width:   w,
		height:  h,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case "enter":
			if len(m.items) > 0 {
				m.selected = &m.items[m.cursor]
				return m, tea.Quit
			}
		case "home":
			m.cursor = 0
		case "end":
			m.cursor = len(m.items) - 1
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m model) View() string {
	// Styles
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	dimStyle := lipgloss.NewStyle().Faint(true)
	helpStyle := lipgloss.NewStyle().Faint(true)

	var b strings.Builder

	// Header
	header := titleStyle.Render(fmt.Sprintf("The Economist — %s", m.section))
	b.WriteString(header + "\n")
	b.WriteString(strings.Repeat("─", min(m.width, 60)) + "\n\n")

	// Calculate visible items based on terminal height
	visibleItems := m.height - 6 // Reserve space for header and footer
	if visibleItems < 5 {
		visibleItems = 5
	}
	itemHeight := 3 // lines per item
	maxVisible := visibleItems / itemHeight
	if maxVisible > len(m.items) {
		maxVisible = len(m.items)
	}

	// Calculate scroll offset
	start := 0
	if m.cursor >= maxVisible {
		start = m.cursor - maxVisible + 1
	}
	end := start + maxVisible
	if end > len(m.items) {
		end = len(m.items)
	}

	// Items
	for i := start; i < end; i++ {
		item := m.items[i]
		cursor := "  "
		style := lipgloss.NewStyle()
		if i == m.cursor {
			cursor = "▸ "
			style = selectedStyle
		}

		title := item.CleanTitle()
		maxTitleLen := m.width - 20
		if maxTitleLen < 30 {
			maxTitleLen = 30
		}
		if len(title) > maxTitleLen {
			title = title[:maxTitleLen-3] + "..."
		}

		b.WriteString(fmt.Sprintf("%s%s\n", cursor, style.Render(title)))

		desc := item.CleanDescription()
		if desc != "" {
			maxDescLen := m.width - 6
			if maxDescLen < 30 {
				maxDescLen = 30
			}
			if len(desc) > maxDescLen {
				desc = desc[:maxDescLen-3] + "..."
			}
			b.WriteString(fmt.Sprintf("    %s\n", dimStyle.Render(desc)))
		}
		b.WriteString(fmt.Sprintf("    %s\n", dimStyle.Render(item.FormattedDate())))
	}

	// Footer
	b.WriteString("\n")
	help := helpStyle.Render("↑/↓ navigate • enter read • q quit")
	b.WriteString(help)

	return b.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

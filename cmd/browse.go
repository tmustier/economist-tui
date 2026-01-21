package cmd

import (
	"fmt"
	"os"
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	appErrors "github.com/tmustier/economist-cli/internal/errors"
	"github.com/tmustier/economist-cli/internal/rss"
	"github.com/tmustier/economist-cli/internal/search"
	"github.com/tmustier/economist-cli/internal/ui"
)

var browseCmd = &cobra.Command{
	Use:   "browse [section]",
	Short: "Browse headlines interactively",
	Long: `Browse headlines in an interactive TUI.

Use ↑/↓ or j/k to navigate, Enter to read, / to search, q to quit.

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

	section := "leaders"
	if len(args) > 0 {
		section = args[0]
	}

	feed, err := rss.FetchSection(section)
	if err != nil {
		return err
	}

	items := feed.Channel.Items
	if len(items) > 50 {
		items = items[:50]
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

	art, err := fetchArticle(url)
	if err != nil {
		return err
	}

	return outputArticle(art)
}

// Model

type model struct {
	allItems      []rss.Item
	filteredItems []rss.Item
	section       string
	cursor        int
	selected      *rss.Item
	width         int
	height        int
	searching     bool
	searchQuery   string
}

func initialModel(items []rss.Item, section string) model {
	w, h := ui.TermSize(int(os.Stdout.Fd()))
	return model{
		allItems:      items,
		filteredItems: items,
		section:       section,
		width:         w,
		height:        h,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m *model) filterItems() {
	if m.searchQuery == "" {
		m.filteredItems = m.allItems
		return
	}

	var filtered []rss.Item
	for _, item := range m.allItems {
		text := item.CleanTitle() + " " + item.CleanDescription()
		if search.Match(text, m.searchQuery) {
			filtered = append(filtered, item)
		}
	}
	m.filteredItems = filtered

	// Reset cursor if out of bounds
	if m.cursor >= len(m.filteredItems) {
		m.cursor = max(0, len(m.filteredItems)-1)
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Search mode input handling
		if m.searching {
			switch msg.Type {
			case tea.KeyEsc:
				m.searching = false
				m.searchQuery = ""
				m.filterItems()
			case tea.KeyEnter:
				m.searching = false
				// Keep the filter active
			case tea.KeyBackspace:
				if len(m.searchQuery) > 0 {
					m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
					m.filterItems()
				}
			case tea.KeyRunes:
				for _, r := range msg.Runes {
					if unicode.IsPrint(r) {
						m.searchQuery += string(r)
					}
				}
				m.filterItems()
			case tea.KeySpace:
				m.searchQuery += " "
				m.filterItems()
			}
			return m, nil
		}

		// Normal mode
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc":
			if m.searchQuery != "" {
				m.searchQuery = ""
				m.filterItems()
			} else {
				return m, tea.Quit
			}
		case "/":
			m.searching = true
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.filteredItems)-1 {
				m.cursor++
			}
		case "enter":
			if len(m.filteredItems) > 0 && m.cursor < len(m.filteredItems) {
				m.selected = &m.filteredItems[m.cursor]
				return m, tea.Quit
			}
		case "home", "g":
			m.cursor = 0
		case "end", "G":
			m.cursor = max(0, len(m.filteredItems)-1)
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
	searchStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("220"))

	var b strings.Builder

	// Header
	header := titleStyle.Render(fmt.Sprintf("The Economist — %s", m.section))
	b.WriteString(header + "\n")
	b.WriteString(strings.Repeat("─", min(m.width, 60)) + "\n")

	// Search bar
	if m.searching {
		b.WriteString(searchStyle.Render("/ ") + m.searchQuery + "█\n")
	} else if m.searchQuery != "" {
		b.WriteString(dimStyle.Render(fmt.Sprintf("filter: %s", m.searchQuery)) + "\n")
	} else {
		b.WriteString("\n")
	}

	items := m.filteredItems

	if len(items) == 0 {
		b.WriteString("\n" + dimStyle.Render("  No matching articles") + "\n")
	} else {
		// Calculate visible items based on terminal height
		reservedLines := 7 // header + search + footer
		if m.searchQuery != "" || m.searching {
			reservedLines++
		}
		visibleItems := m.height - reservedLines
		if visibleItems < 5 {
			visibleItems = 5
		}
		itemHeight := 2 // lines per item (title + description)
		maxVisible := visibleItems / itemHeight
		if maxVisible > len(items) {
			maxVisible = len(items)
		}
		if maxVisible < 1 {
			maxVisible = 1
		}

		// Calculate scroll offset
		start := 0
		if m.cursor >= maxVisible {
			start = m.cursor - maxVisible + 1
		}
		end := start + maxVisible
		if end > len(items) {
			end = len(items)
		}

		layout := ui.NewHeadlineLayout(m.width, len("▸ "))

		// Items
		for i := start; i < end; i++ {
			item := items[i]
			cursor := "  "
			style := lipgloss.NewStyle()
			if i == m.cursor {
				cursor = "▸ "
				style = selectedStyle
			}

			title := item.CleanTitle()
			date := item.FormattedDate()

			paddedTitle := layout.PadTitle(title)

			b.WriteString(fmt.Sprintf("%s%s%s\n",
				cursor,
				style.Render(paddedTitle),
				dimStyle.Render(date),
			))

			desc := item.CleanDescription()
			if desc != "" {
				maxDescLen := m.width - 6
				if maxDescLen < ui.MinTitleWidth {
					maxDescLen = ui.MinTitleWidth
				}
				desc = ui.Truncate(desc, maxDescLen)
				b.WriteString(fmt.Sprintf("    %s\n", dimStyle.Render(desc)))
			}
		}

		// Scroll indicator
		if len(items) > maxVisible {
			b.WriteString(dimStyle.Render(fmt.Sprintf("\n  (%d/%d)", m.cursor+1, len(items))))
		}
	}

	// Footer
	b.WriteString("\n\n")
	if m.searching {
		help := helpStyle.Render("type to search • enter confirm • esc cancel")
		b.WriteString(help)
	} else {
		help := helpStyle.Render("↑/↓ navigate • enter read • / search • q quit")
		b.WriteString(help)
	}

	return b.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

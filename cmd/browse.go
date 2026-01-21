package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/tmustier/economist-cli/internal/daemon"
	appErrors "github.com/tmustier/economist-cli/internal/errors"
	"github.com/tmustier/economist-cli/internal/rss"
	"github.com/tmustier/economist-cli/internal/search"
	"github.com/tmustier/economist-cli/internal/ui"
)

var browseCmd = &cobra.Command{
	Use:   "browse [section]",
	Short: "Browse headlines interactively",
	Long: `Browse headlines in an interactive TUI.

Use ↑/↓ to navigate, Enter to read, type to search, q to quit.

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

	_ = daemon.EnsureBackground()

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

	sectionTitle := strings.TrimSpace(feed.Channel.Title)
	if sectionTitle == "" {
		sectionTitle = section
	}

	m := initialModel(items, sectionTitle)
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
	sectionTitle  string
	cursor        int
	selected      *rss.Item
	width         int
	height        int
	searchQuery   string
}

func initialModel(items []rss.Item, sectionTitle string) model {
	w, h := ui.TermSize(int(os.Stdout.Fd()))
	return model{
		allItems:      items,
		filteredItems: items,
		sectionTitle:  sectionTitle,
		width:         w,
		height:        h,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m *model) applySearch() {
	query := strings.TrimSpace(m.searchQuery)
	if query == "" {
		m.filteredItems = m.allItems
		return
	}

	if isDigits(query) {
		m.filteredItems = m.allItems
		idx, err := strconv.Atoi(query)
		if err == nil && idx > 0 && idx <= len(m.allItems) {
			m.cursor = idx - 1
		}
		return
	}

	var filtered []rss.Item
	for _, item := range m.allItems {
		text := item.CleanTitle() + " " + item.CleanDescription()
		if search.Match(text, query) {
			filtered = append(filtered, item)
		}
	}
	m.filteredItems = filtered

	// Reset cursor if out of bounds
	if m.cursor >= len(m.filteredItems) {
		m.cursor = max(0, len(m.filteredItems)-1)
	}
}

func isDigits(input string) bool {
	for _, r := range input {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return input != ""
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			if m.searchQuery == "" {
				return m, tea.Quit
			}
		}

		switch msg.Type {
		case tea.KeyEsc:
			if m.searchQuery != "" {
				m.searchQuery = ""
				m.applySearch()
			} else {
				return m, tea.Quit
			}
		case tea.KeyBackspace:
			if len(m.searchQuery) > 0 {
				m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
				m.applySearch()
			}
		case tea.KeyEnter:
			if len(m.filteredItems) > 0 && m.cursor < len(m.filteredItems) {
				m.selected = &m.filteredItems[m.cursor]
				return m, tea.Quit
			}
		case tea.KeyUp:
			if m.cursor > 0 {
				m.cursor--
			}
		case tea.KeyDown:
			if m.cursor < len(m.filteredItems)-1 {
				m.cursor++
			}
		case tea.KeyLeft:
			page := m.pageSize()
			m.cursor = max(0, m.cursor-page)
		case tea.KeyRight:
			page := m.pageSize()
			m.cursor = min(m.cursor+page, len(m.filteredItems)-1)
		case tea.KeyHome:
			m.cursor = 0
		case tea.KeyEnd:
			m.cursor = max(0, len(m.filteredItems)-1)
		case tea.KeySpace:
			if m.searchQuery != "" {
				m.searchQuery += " "
				m.applySearch()
			}
		case tea.KeyRunes:
			for _, r := range msg.Runes {
				if unicode.IsPrint(r) {
					m.searchQuery += string(r)
				}
			}
			m.applySearch()
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m model) View() string {
	styles := ui.NewBrowseStyles(noColor)

	var b strings.Builder

	// Header
	header := styles.Header.Render(m.sectionTitle)
	b.WriteString(header + "\n")
	b.WriteString(strings.Repeat("─", min(m.width, 60)) + "\n")

	// Search bar
	if m.searchQuery != "" {
		b.WriteString(styles.Search.Render(fmt.Sprintf("search: %s", m.searchQuery)) + "\n")
	} else {
		b.WriteString("\n")
	}

	items := m.filteredItems

	if len(items) == 0 {
		b.WriteString("\n" + styles.Dim.Render("  No matching articles") + "\n")
	} else {
		maxVisible := m.pageSize()
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

		numWidth := len(fmt.Sprintf("%d", len(m.allItems)))
		prefix := fmt.Sprintf("%*d. ", numWidth, len(m.allItems))
		layout := ui.NewHeadlineLayout(m.width, len(prefix))

		// Items
		for i := start; i < end; i++ {
			item := items[i]
			lineStyle := styles.Title
			dateStyle := styles.Dim
			if i == m.cursor {
				lineStyle = styles.Selected
				dateStyle = styles.Selected
			}

			num := fmt.Sprintf("%*d. ", numWidth, i+1)
			title := item.CleanTitle()
			date := item.FormattedDate()

			paddedTitle := layout.PadTitle(title)

			b.WriteString(fmt.Sprintf("%s%s%s\n",
				num,
				lineStyle.Render(paddedTitle),
				dateStyle.Render(date),
			))

			desc := item.CleanDescription()
			if desc != "" {
				maxDescLen := m.width - 4
				if maxDescLen < ui.MinTitleWidth {
					maxDescLen = ui.MinTitleWidth
				}
				desc = ui.Truncate(desc, maxDescLen)
				b.WriteString(fmt.Sprintf("    %s\n", styles.Dim.Render(desc)))
			}

			b.WriteString("\n")
		}

		// Scroll indicator
		if len(items) > maxVisible {
			b.WriteString(styles.Dim.Render(fmt.Sprintf("\n  (%d/%d)", m.cursor+1, len(items))))
		}
	}

	// Footer
	b.WriteString("\n\n")
	help := styles.Help.Render("↑/↓ navigate • ←/→ page • enter read • type to search • esc clear • q quit")
	b.WriteString(help)

	return b.String()
}

func (m model) pageSize() int {
	reservedLines := 7 // header + search + footer
	visibleItems := m.height - reservedLines
	if visibleItems < 5 {
		visibleItems = 5
	}
	itemHeight := 3 // title + description + spacer
	page := visibleItems / itemHeight
	if page < 1 {
		page = 1
	}
	return page
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

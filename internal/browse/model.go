package browse

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tmustier/economist-cli/internal/article"
	"github.com/tmustier/economist-cli/internal/fetch"
	"github.com/tmustier/economist-cli/internal/logging"
	"github.com/tmustier/economist-cli/internal/rss"
	"github.com/tmustier/economist-cli/internal/search"
	"github.com/tmustier/economist-cli/internal/ui"
)

type viewMode int

const (
	modeBrowse viewMode = iota
	modeArticle
)

type articleMsg struct {
	url           string
	article       *article.Article
	err           error
	fetchDuration time.Duration
}

type Model struct {
	allItems      []rss.Item
	filteredItems []rss.Item
	sectionTitle  string
	cursor        int
	width         int
	height        int
	searchQuery   string

	mode        viewMode
	loading     bool
	loadingItem *rss.Item
	pendingURL  string
	article      *article.Article
	articleBase  string
	articleLines []string
	articleErr   error
	scroll       int
	twoColumn    bool

	fetchDuration  time.Duration
	baseDuration   time.Duration
	reflowDuration time.Duration

	opts Options
}

func NewModel(items []rss.Item, sectionTitle string, opts Options) Model {
	w, h := ui.TermSize(int(os.Stdout.Fd()))
	return Model{
		allItems:      items,
		filteredItems: items,
		sectionTitle:  sectionTitle,
		width:         w,
		height:        h,
		mode:          modeBrowse,
		opts:          opts,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m *Model) applySearch() {
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
		m.cursor = ui.Max(0, len(m.filteredItems)-1)
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

func fetchArticleCmd(url string, debug bool) tea.Cmd {
	return func() tea.Msg {
		start := time.Now()
		art, err := fetch.FetchArticle(url, fetch.Options{Debug: debug})
		return articleMsg{url: url, article: art, err: err, fetchDuration: time.Since(start)}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case articleMsg:
		if m.mode != modeArticle || msg.url != m.pendingURL {
			return m, nil
		}
		m.loading = false
		m.pendingURL = ""
		m.scroll = 0
		m.fetchDuration = msg.fetchDuration
		if msg.err != nil {
			m.articleErr = msg.err
			m.article = nil
			m.articleBase = ""
			m.articleLines = []string{fmt.Sprintf("Error: %v", msg.err)}
			return m, nil
		}
		m.articleErr = nil
		m.article = msg.article
		m.articleBase = ""
		m.refreshArticleLines()
		return m, nil
	case tea.KeyMsg:
		if m.mode == modeArticle {
			return m.updateArticle(msg)
		}
		return m.updateBrowse(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.mode == modeArticle && m.article != nil {
			m.refreshArticleLines()
		}
	}
	return m, nil
}

func (m Model) updateBrowse(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
			item := m.filteredItems[m.cursor]
			m.mode = modeArticle
			m.loading = true
			m.loadingItem = &item
			m.pendingURL = item.Link
			m.articleErr = nil
			m.article = nil
			m.articleBase = ""
			m.articleLines = nil
			m.scroll = 0
			return m, fetchArticleCmd(item.Link, m.opts.Debug)
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
		m.cursor = ui.Max(0, m.cursor-page)
	case tea.KeyRight:
		page := m.pageSize()
		m.cursor = ui.Min(m.cursor+page, len(m.filteredItems)-1)
	case tea.KeyHome:
		m.cursor = 0
	case tea.KeyEnd:
		m.cursor = ui.Max(0, len(m.filteredItems)-1)
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
	return m, nil
}

func (m Model) updateArticle(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "b", "left":
		m.mode = modeBrowse
		m.loading = false
		m.loadingItem = nil
		m.pendingURL = ""
		return m, nil
	case "c":
		m.twoColumn = !m.twoColumn
		m.refreshArticleLines()
		return m, nil
	}

	switch msg.Type {
	case tea.KeyEsc:
		m.mode = modeBrowse
		m.loading = false
		m.loadingItem = nil
		m.pendingURL = ""
	case tea.KeyUp:
		m.scroll--
	case tea.KeyDown:
		m.scroll++
	case tea.KeyPgUp:
		m.scroll -= m.articlePageSize()
	case tea.KeyPgDown:
		m.scroll += m.articlePageSize()
	case tea.KeyHome:
		m.scroll = 0
	case tea.KeyEnd:
		m.scroll = m.maxArticleScroll()
	}

	m.clampArticleScroll()
	return m, nil
}

func (m *Model) refreshArticleLines() {
	if m.article == nil {
		return
	}

	opts := m.articleRenderOptions()

	if m.articleBase == "" {
		baseStart := time.Now()
		base, err := ui.RenderArticleBodyBase(ui.ArticleBodyMarkdown(m.article), opts)
		m.baseDuration = time.Since(baseStart)
		if err != nil {
			m.articleErr = err
			m.articleLines = []string{fmt.Sprintf("Error: %v", err)}
			return
		}
		m.articleBase = base
		logging.Debugf(m.opts.Debug, "browse: article base render %s", m.baseDuration)
	}

	styles := ui.NewArticleStyles(m.opts.NoColor)
	reflowStart := time.Now()
	body := ui.ReflowArticleBody(m.articleBase, styles, opts)
	m.reflowDuration = time.Since(reflowStart)
	logging.Debugf(m.opts.Debug, "browse: article reflow %s", m.reflowDuration)

	header := ui.RenderArticleHeader(m.article, styles, opts)
	indent := ui.ArticleIndent(opts)
	if indent > 0 {
		header = ui.IndentBlock(header, indent)
	}

	footer := ui.ArticleFooter(m.article, styles, opts)
	if indent > 0 {
		footer = ui.IndentBlock(footer, indent)
	}

	m.articleErr = nil
	m.articleLines = strings.Split(strings.TrimRight(header+body+footer, "\n"), "\n")
	m.clampArticleScroll()
}

func (m Model) articleRenderOptions() ui.ArticleRenderOptions {
	termWidth := m.width
	if termWidth <= 0 {
		termWidth = ui.DefaultWidth
	}

	wrapWidth := 0
	center := false
	if !m.twoColumn {
		wrapWidth = ui.ReaderContentWidth(termWidth)
		center = true
	}

	return ui.ArticleRenderOptions{
		NoColor:   m.opts.NoColor,
		PlainBody: true,
		WrapWidth: wrapWidth,
		TermWidth: termWidth,
		Center:    center,
		TwoColumn: m.twoColumn,
	}
}

func (m Model) pageSize() int {
	visibleItems := m.height - browseReservedLines
	if visibleItems < browseMinVisibleLines {
		visibleItems = browseMinVisibleLines
	}
	page := visibleItems / browseItemHeight
	if page < 1 {
		page = 1
	}
	return page
}

func (m Model) articleViewHeight() int {
	reserved := articleReservedLines
	if m.opts.Debug {
		reserved++
	}
	visible := m.height - reserved
	if visible < articleMinVisibleLines {
		visible = articleMinVisibleLines
	}
	return visible
}

func (m Model) articlePageSize() int {
	page := m.articleViewHeight() - 2
	if page < 1 {
		page = 1
	}
	return page
}

func (m Model) maxArticleScroll() int {
	maxScroll := len(m.articleLines) - m.articleViewHeight()
	if maxScroll < 0 {
		maxScroll = 0
	}
	return maxScroll
}

func (m *Model) clampArticleScroll() {
	maxScroll := m.maxArticleScroll()
	m.scroll = ui.Clamp(m.scroll, 0, maxScroll)
}

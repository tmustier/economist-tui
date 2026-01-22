package browse

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tmustier/economist-tui/internal/article"
	"github.com/tmustier/economist-tui/internal/logging"
	"github.com/tmustier/economist-tui/internal/rss"
	"github.com/tmustier/economist-tui/internal/search"
	"github.com/tmustier/economist-tui/internal/ui"
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

type sectionMsg struct {
	section string
	title   string
	items   []rss.Item
	err     error
}

type Model struct {
	allItems      []rss.Item
	filteredItems []rss.Item
	sectionTitle  string
	sections      []rss.SectionInfo
	sectionIndex  int

	pendingSection      string
	pendingSectionIndex int
	sectionLoading      bool
	sectionErr          error

	cursor      int
	browseStart int
	width       int
	height      int
	searchQuery string

	mode         viewMode
	loading      bool
	loadingItem  *rss.Item
	pendingURL   string
	article      *article.Article
	articleBase  string
	articleLines []string
	articleErr   error
	scroll       int
	twoColumn    bool

	fetchDuration  time.Duration
	baseDuration   time.Duration
	reflowDuration time.Duration

	source DataSource
	opts   Options
}

func NewModel(section string, items []rss.Item, sectionTitle string, opts Options, source DataSource) Model {
	if source == nil {
		source = rssSource{debug: opts.Debug}
	}
	w, h := ui.TermSize(int(os.Stdout.Fd()))
	sections := rss.SectionList()
	sectionIndex, sections := resolveSectionIndex(section, sections)
	return Model{
		allItems:            items,
		filteredItems:       items,
		sectionTitle:        sectionTitle,
		sections:            sections,
		sectionIndex:        sectionIndex,
		width:               w,
		height:              h,
		mode:                modeBrowse,
		source:              source,
		opts:                opts,
		pendingSectionIndex: -1,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func resolveSectionIndex(section string, sections []rss.SectionInfo) (int, []rss.SectionInfo) {
	trimmed := strings.TrimSpace(section)
	normalized := strings.ToLower(trimmed)
	if normalized == "" {
		normalized = trimmed
	}

	path := normalized
	if resolved, ok := rss.Sections[normalized]; ok {
		path = resolved
	}

	for i, info := range sections {
		if info.Path == path {
			return i, sections
		}
	}

	if trimmed == "" {
		trimmed = path
	}
	custom := rss.SectionInfo{Primary: trimmed, Path: path, Aliases: []string{trimmed}}
	sections = append([]rss.SectionInfo{custom}, sections...)
	return 0, sections
}

func (m *Model) applySearch() {
	query := strings.TrimSpace(m.searchQuery)
	if query == "" {
		m.filteredItems = m.allItems
		m.ensureBrowseWindow()
		return
	}

	if isDigits(query) {
		m.filteredItems = m.allItems
		idx, err := strconv.Atoi(query)
		if err == nil && idx > 0 && idx <= len(m.allItems) {
			m.cursor = idx - 1
		}
		m.ensureBrowseWindow()
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
	m.ensureBrowseWindow()
}

func isDigits(input string) bool {
	for _, r := range input {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return input != ""
}

func (m Model) fetchArticleCmd(url string) tea.Cmd {
	source := m.source
	if source == nil {
		source = rssSource{debug: m.opts.Debug}
	}
	return func() tea.Msg {
		start := time.Now()
		art, err := source.Article(url)
		return articleMsg{url: url, article: art, err: err, fetchDuration: time.Since(start)}
	}
}

func (m Model) fetchSectionCmd(section string) tea.Cmd {
	source := m.source
	if source == nil {
		source = rssSource{debug: m.opts.Debug}
	}
	return func() tea.Msg {
		title, items, err := loadSection(source, section)
		return sectionMsg{section: section, title: title, items: items, err: err}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case sectionMsg:
		if msg.section != m.pendingSection {
			return m, nil
		}
		pendingIndex := m.pendingSectionIndex
		m.sectionLoading = false
		m.pendingSection = ""
		m.pendingSectionIndex = -1
		if msg.err != nil {
			m.sectionErr = msg.err
			return m, nil
		}
		m.sectionErr = nil
		if pendingIndex >= 0 {
			m.sectionIndex = pendingIndex
		}
		m.sectionTitle = msg.title
		m.allItems = msg.items
		m.filteredItems = msg.items
		m.cursor = 0
		m.browseStart = 0
		m.applySearch()
		return m, nil
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
		if m.mode == modeBrowse {
			m.ensureBrowseWindow()
		}
	}
	return m, nil
}

func (m Model) updateBrowse(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "ctrl+d":
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
	case tea.KeyTab:
		return m.queueSectionChange(1)
	case tea.KeyShiftTab:
		return m.queueSectionChange(-1)
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
			return m, m.fetchArticleCmd(item.Link)
		}
	case tea.KeyUp:
		if m.cursor > 0 {
			layout := m.browseLayout(len(m.filteredItems))
			maxVisible := layout.maxVisible
			if m.cursor == m.browseStart && m.browseStart > 0 {
				m.cursor--
				m.browseStart--
			} else {
				m.cursor--
			}
			if maxVisible > 0 && m.cursor < m.browseStart {
				m.browseStart = m.cursor
			}
		}
	case tea.KeyDown:
		if m.cursor < len(m.filteredItems)-1 {
			layout := m.browseLayout(len(m.filteredItems))
			maxVisible := layout.maxVisible
			lastVisible := m.browseStart + maxVisible - 1
			if m.cursor == lastVisible {
				m.cursor++
				m.browseStart++
			} else {
				m.cursor++
			}
			maxStart := ui.Max(0, len(m.filteredItems)-maxVisible)
			if m.browseStart > maxStart {
				m.browseStart = maxStart
			}
		}
	case tea.KeyLeft:
		itemCount := len(m.filteredItems)
		if itemCount > 0 {
			layout := m.browseLayout(len(m.filteredItems))
			maxVisible := layout.maxVisible
			offset := m.cursor - m.browseStart
			if offset < 0 {
				offset = 0
			} else if offset >= maxVisible {
				offset = maxVisible - 1
			}
			m.browseStart -= maxVisible
			if m.browseStart < 0 {
				m.browseStart = 0
			}
			m.cursor = ui.Min(m.browseStart+offset, itemCount-1)
		}
	case tea.KeyRight:
		itemCount := len(m.filteredItems)
		if itemCount > 0 {
			layout := m.browseLayout(len(m.filteredItems))
			maxVisible := layout.maxVisible
			maxStart := ui.Max(0, itemCount-maxVisible)
			offset := m.cursor - m.browseStart
			if offset < 0 {
				offset = 0
			} else if offset >= maxVisible {
				offset = maxVisible - 1
			}
			m.browseStart += maxVisible
			if m.browseStart > maxStart {
				m.browseStart = maxStart
			}
			m.cursor = ui.Min(m.browseStart+offset, itemCount-1)
		}
	case tea.KeyHome:
		m.cursor = 0
		m.browseStart = 0
	case tea.KeyEnd:
		itemCount := len(m.filteredItems)
		if itemCount > 0 {
			layout := m.browseLayout(len(m.filteredItems))
			maxVisible := layout.maxVisible
			m.cursor = itemCount - 1
			m.browseStart = ui.Max(0, itemCount-maxVisible)
		}
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
	m.ensureBrowseWindow()
	return m, nil
}

func (m Model) updateArticle(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "ctrl+d", "q":
		return m, tea.Quit
	case "b", "enter":
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

func (m Model) queueSectionChange(delta int) (tea.Model, tea.Cmd) {
	if len(m.sections) == 0 {
		return m, nil
	}

	baseIndex := m.sectionIndex
	if m.sectionLoading && m.pendingSectionIndex >= 0 {
		baseIndex = m.pendingSectionIndex
	}

	nextIndex := baseIndex + delta
	if nextIndex < 0 {
		nextIndex = len(m.sections) - 1
	} else if nextIndex >= len(m.sections) {
		nextIndex = 0
	}

	nextSection := m.sections[nextIndex].Primary
	if m.pendingSection == nextSection && m.sectionLoading {
		return m, nil
	}

	m.pendingSection = nextSection
	m.pendingSectionIndex = nextIndex
	m.sectionLoading = true
	m.sectionErr = nil
	return m, m.fetchSectionCmd(nextSection)
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
	layout := ui.ResolveArticleLayoutWithContent(m.articleBase, opts)
	reflowStart := time.Now()
	body := ui.ReflowArticleBodyWithLayout(m.articleBase, styles, opts, layout)
	m.reflowDuration = time.Since(reflowStart)
	logging.Debugf(m.opts.Debug, "browse: article reflow %s", m.reflowDuration)

	header := ui.RenderArticleHeaderWithLayout(m.article, styles, layout, opts)
	indent := ui.ArticleIndentForLayout(layout)
	if indent > 0 {
		header = ui.IndentBlock(header, indent)
	}

	footer := ui.ArticleFooterWithLayout(m.article, styles, layout, opts)
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
	center := true
	if !m.twoColumn {
		wrapWidth = ui.ReaderContentWidth(termWidth)
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

func (m Model) pageSize(itemCount int) int {
	layout := m.browseLayout(itemCount)
	return layout.maxVisible
}

func (m Model) articleViewHeight() int {
	spec := articleLayoutSpec(m.opts.Debug)
	return spec.VisibleLines(m.height)
}

type browseLayout struct {
	maxVisible    int
	titleLines    int
	subtitleLines int
	showPosition  bool
}

func (m Model) browseLayout(itemCount int) browseLayout {
	helpLineCount := m.browseHelpLineCount()
	layout := m.browseLayoutForFooter(helpLineCount, true, itemCount)
	if itemCount <= layout.maxVisible {
		layout = m.browseLayoutForFooter(helpLineCount, false, itemCount)
	}
	return layout
}

func (m Model) browseLayoutForFooter(helpLineCount int, showPosition bool, itemCount int) browseLayout {
	spec := browseLayoutSpec(helpLineCount, showPosition)
	visibleLines := spec.VisibleLines(m.height)
	titleLines, subtitleLines := resolveBrowseItemLines(visibleLines)
	maxVisible := m.maxVisibleItems(itemCount, visibleLines, titleLines, subtitleLines)
	if maxVisible < 1 {
		maxVisible = 1
	}
	return browseLayout{
		maxVisible:    maxVisible,
		titleLines:    titleLines,
		subtitleLines: subtitleLines,
		showPosition:  showPosition,
	}
}

func (m Model) browseHelpLineCount() int {
	termWidth := m.width
	if termWidth <= 0 {
		termWidth = ui.DefaultWidth
	}
	contentWidth := ui.ReaderContentWidth(termWidth)
	return len(browseHelpLines(contentWidth))
}

func (m Model) maxVisibleItems(itemCount, visibleLines, titleLines, subtitleLines int) int {
	if itemCount <= 0 {
		return 0
	}
	if visibleLines <= 0 {
		return 1
	}
	items := m.filteredItems
	if itemCount < len(items) {
		items = items[:itemCount]
	}
	start := m.browseStart
	if start < 0 {
		start = 0
	}
	if start >= len(items) {
		start = ui.Max(0, len(items)-1)
	}
	titleWidth := m.browseTitleWidth()
	used := 0
	count := 0
	for i := start; i < len(items); i++ {
		height := browseItemHeight(items[i], titleWidth, titleLines, subtitleLines)
		if count > 0 && used+height > visibleLines {
			break
		}
		used += height
		count++
		if used >= visibleLines {
			break
		}
	}
	if count < 1 {
		count = 1
	}
	return count
}

func (m Model) browseTitleWidth() int {
	termWidth := m.width
	if termWidth <= 0 {
		termWidth = ui.DefaultWidth
	}
	contentWidth := ui.ReaderContentWidth(termWidth)
	numWidth := len(fmt.Sprintf("%d", len(m.allItems)))
	prefixWidth := len(fmt.Sprintf("%*d. ", numWidth, len(m.allItems)))
	dateLayout := ui.ResolveDateLayout(contentWidth, prefixWidth)
	titleWidth := contentWidth - prefixWidth - dateLayout.ColumnWidth
	minWidth := prefixWidth + dateLayout.ColumnWidth + ui.MinTitleWidth
	if titleWidth < ui.MinTitleWidth && contentWidth >= minWidth {
		titleWidth = ui.MinTitleWidth
	}
	if titleWidth < 1 {
		titleWidth = 1
	}
	return titleWidth
}

func browseItemHeight(item rss.Item, titleWidth, titleLines, subtitleLines int) int {
	title := item.CleanTitle()
	titleLineCount := len(ui.LimitLines(ui.WrapLines(title, titleWidth), titleLines, titleWidth))
	if titleLineCount == 0 {
		titleLineCount = 1
	}
	subtitle := item.CleanDescription()
	subtitleLineCount := len(ui.LimitLines(ui.WrapLines(subtitle, titleWidth), subtitleLines, titleWidth))
	return titleLineCount + subtitleLineCount + browseItemGapLines
}

func resolveBrowseItemLines(visibleLines int) (int, int) {
	titleLines := browseTitleLines
	subtitleLines := browseSubtitleLines
	itemHeight := titleLines + subtitleLines + 1

	for subtitleLines > 0 && visibleLines < itemHeight*2 {
		subtitleLines--
		itemHeight--
	}
	for titleLines > 1 && visibleLines < itemHeight*2 {
		titleLines--
		itemHeight--
	}

	return titleLines, subtitleLines
}

func (m *Model) ensureBrowseWindow() {
	itemCount := len(m.filteredItems)
	if itemCount == 0 {
		m.cursor = 0
		m.browseStart = 0
		return
	}

	layout := m.browseLayout(itemCount)
	maxVisible := layout.maxVisible
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor >= itemCount {
		m.cursor = itemCount - 1
	}

	maxStart := ui.Max(0, itemCount-maxVisible)
	m.browseStart = ui.Clamp(m.browseStart, 0, maxStart)

	if m.cursor < m.browseStart {
		m.browseStart = m.cursor
	} else if m.cursor >= m.browseStart+maxVisible {
		m.browseStart = m.cursor - maxVisible + 1
		m.browseStart = ui.Clamp(m.browseStart, 0, maxStart)
	}
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

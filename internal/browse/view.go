package browse

import (
	"fmt"
	"math"
	"strings"

	"github.com/tmustier/economist-tui/internal/ui"
)

func (m Model) View() string {
	if m.mode == modeArticle {
		content, footer := m.articleView()
		return ui.LayoutWithFooter(content, footer, m.height, browseFooterPadding)
	}
	content, footer := m.browseView()
	return ui.LayoutWithFooter(content, footer, m.height, browseFooterPadding)
}

func (m Model) browseView() (string, string) {
	styles := ui.NewBrowseStyles(m.opts.NoColor)
	accentStyles := ui.NewStyles(ui.CurrentTheme(), m.opts.NoColor)

	var b strings.Builder

	termWidth := m.width
	if termWidth <= 0 {
		termWidth = ui.DefaultWidth
	}
	contentWidth := ui.ReaderContentWidth(termWidth)
	indent := ui.ArticleIndent(ui.ArticleRenderOptions{TermWidth: termWidth, WrapWidth: contentWidth, Center: true})

	b.WriteString("\n")
	header := styles.Header.Render(m.sectionTitle)
	b.WriteString(header + "\n")
	b.WriteString(ui.AccentRule(contentWidth, accentStyles) + "\n")

	// Status line (loading/error)
	statusLine := ""
	if m.sectionLoading && m.pendingSection != "" {
		statusLine = styles.Dim.Render(fmt.Sprintf("loading %s…", m.pendingSection))
	} else if m.sectionErr != nil {
		statusLine = styles.Dim.Render(fmt.Sprintf("error: %v", m.sectionErr))
	}

	// Search bar with states: idle, active, no-match
	searchLine := m.renderSearchBar(styles, contentWidth)
	if statusLine != "" {
		searchLine = fmt.Sprintf("%s  %s", searchLine, statusLine)
	}
	b.WriteString(searchLine + "\n\n")

	items := m.filteredItems
	maxVisible := 0
	layout := browseLayout{}

	if len(items) == 0 {
		b.WriteString("\n" + styles.Dim.Render("  No matching articles") + "\n")
	} else {
		layout = m.browseLayout(len(items))
		maxVisible = layout.maxVisible
		if maxVisible > len(items) {
			maxVisible = len(items)
		}
		if maxVisible < 1 {
			maxVisible = 1
		}

		start := m.browseStart
		maxStart := ui.Max(0, len(items)-maxVisible)
		if start > maxStart {
			start = maxStart
		}
		if start < 0 {
			start = 0
		}
		end := start + maxVisible
		if end > len(items) {
			end = len(items)
		}

		numWidth := ui.Max(2, len(fmt.Sprintf("%d", len(m.allItems))))
		prefixWidth := len(fmt.Sprintf("%*d. ", numWidth, len(m.allItems)))
		dateLayout := ui.ResolveDateLayout(contentWidth, prefixWidth)

		listItems := make([]ui.ListItem, len(items))
		for i, item := range items {
			date := item.FormattedDate()
			if dateLayout.Compact {
				date = item.CompactDate()
			}
			listItems[i] = ui.ListItem{
				Title:    item.CleanTitle(),
				Subtitle: item.CleanDescription(),
				Right:    date,
			}
		}

		listOpts := ui.ListOptions{
			Width:            contentWidth,
			PrefixWidth:      prefixWidth,
			RightColumnWidth: dateLayout.ColumnWidth,
			TitleLines:       layout.titleLines,
			SubtitleLines:    layout.subtitleLines,
			ItemGapLines:     browseItemGapLines,
			SelectedIndex:    m.cursor,
			Start:            start,
			End:              end,
			Prefix: func(index int) string {
				return fmt.Sprintf("%*d. ", numWidth, index+1)
			},
		}
		listStyles := ui.ListStyles{
			Title:         styles.Title,
			Subtitle:      styles.Subtitle,
			Selected:      styles.Selected,
			Right:         styles.Dim,
			RightSelected: styles.Selected,
		}

		b.WriteString(ui.RenderList(listItems, listOpts, listStyles))

	}

	content := b.String()
	divider := ui.SectionRule(contentWidth, accentStyles)
	helpLines := browseHelpLines(contentWidth)
	footerLines := make([]string, 0, len(helpLines)+2)

	// Position tracker (above dots)
	if layout.showPosition && len(items) > 0 {
		footerLines = append(footerLines, styles.Dim.Render(fmt.Sprintf("← (%d/%d) →", m.cursor+1, len(items))))
	}

	// Section dots navigation
	if len(m.sections) > 1 {
		footerLines = append(footerLines, m.renderSectionDots(styles, contentWidth))
	}

	for _, line := range helpLines {
		footerLines = append(footerLines, styles.Help.Render(line))
	}

	// Center all footer lines
	for i, line := range footerLines {
		footerLines[i] = ui.CenterText(line, contentWidth)
	}

	footer := ui.BuildFooter(divider, footerLines...)
	if indent > 0 {
		content = ui.IndentBlock(content, indent)
		footer = ui.IndentBlock(footer, indent)
	}

	if termWidth > 0 {
		content = ui.PadBlockRight(content, termWidth)
		footer = ui.PadBlockRight(footer, termWidth)
	}

	return content, footer
}

func (m Model) articleView() (string, string) {
	styles := ui.NewBrowseStyles(m.opts.NoColor)
	opts := m.articleRenderOptions()
	indent := ui.ArticleIndent(opts)
	contentWidth := opts.WrapWidth
	if contentWidth <= 0 {
		contentWidth = opts.TermWidth
		if contentWidth <= 0 {
			contentWidth = ui.DefaultWidth
		}
	}
	ruleStyles := ui.NewStyles(ui.CurrentTheme(), m.opts.NoColor)
	divider := ui.SectionRule(contentWidth, ruleStyles)
	padWidth := opts.TermWidth
	if padWidth <= 0 {
		padWidth = contentWidth
	}

	var b strings.Builder
	if m.loading {
		content := m.loadingSkeletonView()
		centeredHelp := ui.CenterText(styles.Help.Render(articleLoadingHelp), contentWidth)
		footer := ui.BuildFooter(divider, centeredHelp)
		if indent > 0 {
			footer = ui.IndentBlock(footer, indent)
		}
		return ui.PadBlockRight(content, padWidth), ui.PadBlockRight(footer, padWidth)
	}

	if m.articleErr != nil {
		b.WriteString(styles.Dim.Render(fmt.Sprintf("%v", m.articleErr)))
		content := b.String()
		centeredHelp := ui.CenterText(styles.Help.Render(articleLoadingHelp), contentWidth)
		footer := ui.BuildFooter(divider, centeredHelp)
		if indent > 0 {
			content = ui.IndentBlock(content, indent)
			footer = ui.IndentBlock(footer, indent)
		}
		return ui.PadBlockRight(content, padWidth), ui.PadBlockRight(footer, padWidth)
	}

	if len(m.articleLines) == 0 {
		b.WriteString("No article loaded.")
		content := b.String()
		centeredHelp := ui.CenterText(styles.Help.Render(articleLoadingHelp), contentWidth)
		footer := ui.BuildFooter(divider, centeredHelp)
		if indent > 0 {
			content = ui.IndentBlock(content, indent)
			footer = ui.IndentBlock(footer, indent)
		}
		return ui.PadBlockRight(content, padWidth), ui.PadBlockRight(footer, padWidth)
	}

	start := ui.Min(m.scroll, m.maxArticleScroll())
	viewHeight := m.articleViewHeight()
	end := ui.Min(len(m.articleLines), start+viewHeight)
	content := strings.Join(m.articleLines[start:end], "\n")

	columnLabel := "off"
	if m.twoColumn {
		columnLabel = "on"
	}
	help := fmt.Sprintf(articleHelpFormat, columnLabel)

	showMore := end < len(m.articleLines)
	hintLine := ""
	if showMore {
		pct := 0
		if len(m.articleLines) > 0 {
			pct = int(math.Round(float64(end) / float64(len(m.articleLines)) * 100))
			if pct < 1 {
				pct = 1
			}
			if pct > 99 {
				pct = 99
			}
		}
		hintLine = styles.Dim.Render(fmt.Sprintf("%d%% · more ↓", pct))
	}

	lastLine := lastNonBlankLine(m.articleLines[start:end])
	if ui.IsRuleLine(lastLine) {
		divider = ""
	}

	// Center footer lines
	centeredHint := ui.CenterText(hintLine, contentWidth)
	centeredHelp := ui.CenterText(styles.Help.Render(help), contentWidth)

	footer := ui.BuildFooter(divider, centeredHint, centeredHelp)
	if m.opts.Debug {
		detail := fmt.Sprintf("fetch=%s base=%s reflow=%s", m.fetchDuration, m.baseDuration, m.reflowDuration)
		footer = strings.TrimRight(footer, "\n") + "\n" + ui.CenterText(styles.Dim.Render(detail), contentWidth)
	}

	if indent > 0 {
		footer = ui.IndentBlock(footer, indent)
	}

	return ui.PadBlockRight(content, padWidth), ui.PadBlockRight(footer, padWidth)
}

func (m Model) loadingSkeletonView() string {
	header := ui.SkeletonHeader{
		Section: m.sectionTitle,
	}

	if m.loadingItem != nil {
		header.Title = m.loadingItem.CleanTitle()
		header.Subtitle = m.loadingItem.CleanDescription()
		header.Date = m.loadingItem.FormattedDate()
	}

	return ui.RenderArticleSkeleton(header, m.articleRenderOptions(), m.articleViewHeight())
}

func lastNonBlankLine(lines []string) string {
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(ui.StripANSI(lines[i])) == "" {
			continue
		}
		return lines[i]
	}
	return ""
}

// renderSectionDots renders section position dots with tab navigation hints.
// Format: ⇧⇥  ○ ○ ● ○ ○  ⇥
func (m Model) renderSectionDots(styles ui.BrowseStyles, width int) string {
	const (
		tabIcon      = "⇥"
		shiftTabIcon = "⇧⇥"
		dotActive    = "●"
		dotInactive  = "○"
		maxDots      = 20 // Collapse to avoid overflow on narrow terminals
	)

	numSections := len(m.sections)
	showAllDots := numSections <= maxDots && numSections*2-1+10 <= width // dots + spacing + icons

	var content string
	if showAllDots {
		// Build dots: ○ ○ ● ○ ○
		var dots strings.Builder
		for i := 0; i < numSections; i++ {
			if i > 0 {
				dots.WriteString(" ")
			}
			if i == m.sectionIndex {
				dots.WriteString(dotActive)
			} else {
				dots.WriteString(dotInactive)
			}
		}
		content = fmt.Sprintf("%s  %s  %s", shiftTabIcon, dots.String(), tabIcon)
	} else {
		// Compact: ⇧⇥  3/12  ⇥
		content = fmt.Sprintf("%s  %d/%d  %s", shiftTabIcon, m.sectionIndex+1, numSections, tabIcon)
	}

	return styles.Dim.Render(content)
}

// renderSearchBar renders the search bar with appropriate state styling.
// States: idle (placeholder), active (typing), no-match (red)
func (m Model) renderSearchBar(styles ui.BrowseStyles, width int) string {
	const cursor = "│"
	const prefix = "/"

	if m.searchQuery == "" {
		// Idle state: show placeholder
		return styles.SearchIdle.Render("/ type to filter...")
	}

	// Active state with query
	totalItems := len(m.allItems)
	matchCount := len(m.filteredItems)

	if matchCount == 0 {
		// No match state
		text := fmt.Sprintf("%s %s%s  0 matches", prefix, m.searchQuery, cursor)
		return styles.SearchNoMatch.Render(text)
	}

	// Normal active state with results - unified background
	text := fmt.Sprintf("%s %s%s", prefix, m.searchQuery, cursor)
	countText := fmt.Sprintf("%d of %d", matchCount, totalItems)
	return styles.SearchActive.Render(text) + styles.SearchCount.Render(countText)
}

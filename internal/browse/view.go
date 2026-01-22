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

	statusLine := ""
	if m.sectionLoading && m.pendingSection != "" {
		statusLine = styles.Dim.Render(fmt.Sprintf("loading %s…", m.pendingSection))
	} else if m.sectionErr != nil {
		statusLine = styles.Dim.Render(fmt.Sprintf("error: %v", m.sectionErr))
	}

	if m.searchQuery != "" {
		line := styles.Search.Render(fmt.Sprintf("search: %s", m.searchQuery))
		if statusLine != "" {
			line = fmt.Sprintf("%s  %s", line, statusLine)
		}
		b.WriteString(line + "\n")
	} else if statusLine != "" {
		b.WriteString(statusLine + "\n")
	} else {
		b.WriteString("\n")
	}

	items := m.filteredItems
	maxVisible := 0

	if len(items) == 0 {
		b.WriteString("\n" + styles.Dim.Render("  No matching articles") + "\n")
	} else {
		maxVisible = m.pageSize(len(items))
		if maxVisible > len(items) {
			maxVisible = len(items)
		}
		if maxVisible < 1 {
			maxVisible = 1
		}

		start := 0
		if m.cursor >= maxVisible {
			start = m.cursor - maxVisible + 1
		}
		end := start + maxVisible
		if end > len(items) {
			end = len(items)
		}

		numWidth := len(fmt.Sprintf("%d", len(m.allItems)))
		prefixWidth := len(fmt.Sprintf("%*d. ", numWidth, len(m.allItems)))
		dateLayout := ui.ResolveDateLayout(contentWidth, prefixWidth)
		layout := ui.NewHeadlineLayout(contentWidth, prefixWidth, dateLayout.ColumnWidth)
		prefixPad := strings.Repeat(" ", prefixWidth)
		subtitleWidth := layout.TitleWidth

		for i := start; i < end; i++ {
			item := items[i]
			lineStyle := styles.Title
			dateStyle := styles.Dim
			subtitleStyle := styles.Subtitle
			if i == m.cursor {
				lineStyle = styles.Selected
				dateStyle = styles.Selected
			}

			num := fmt.Sprintf("%*d. ", numWidth, i+1)
			title := item.CleanTitle()
			date := item.FormattedDate()
			if dateLayout.Compact {
				date = item.CompactDate()
			}
			dateColumn := fmt.Sprintf("%*s", layout.DateWidth, date)

			titleLines := ui.LimitLines(ui.WrapLines(title, layout.TitleWidth), browseTitleLines, layout.TitleWidth)
			if len(titleLines) == 0 {
				titleLines = []string{""}
			}
			for lineIdx, line := range titleLines {
				if lineIdx == 0 {
					paddedTitle := fmt.Sprintf("%-*s", layout.TitleWidth, line)
					b.WriteString(fmt.Sprintf("%s%s%s\n",
						lineStyle.Render(num),
						lineStyle.Render(paddedTitle),
						dateStyle.Render(dateColumn),
					))
					continue
				}
				if line == "" {
					b.WriteString(prefixPad + "\n")
					continue
				}
				b.WriteString(fmt.Sprintf("%s%s\n", prefixPad, lineStyle.Render(line)))
			}

			desc := item.CleanDescription()
			subtitleLines := ui.LimitLines(ui.WrapLines(desc, subtitleWidth), browseSubtitleLines, subtitleWidth)
			for _, line := range subtitleLines {
				if line == "" {
					b.WriteString(prefixPad + "\n")
					continue
				}
				b.WriteString(fmt.Sprintf("%s%s\n", prefixPad, subtitleStyle.Render(line)))
			}

			b.WriteString("\n")
		}

	}

	content := b.String()
	divider := ui.SectionRule(contentWidth, accentStyles)
	helpLines := browseHelpLines(contentWidth)
	footerLines := make([]string, 0, len(helpLines)+1)
	if len(items) > maxVisible {
		footerLines = append(footerLines, styles.Dim.Render(fmt.Sprintf("(%d/%d)", m.cursor+1, len(items))))
	}
	for _, line := range helpLines {
		footerLines = append(footerLines, styles.Help.Render(line))
	}
	footer := buildFooter(divider, footerLines...)
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
		footer := styles.Help.Render(articleLoadingHelp)
		footer = buildFooter(divider, footer)
		if indent > 0 {
			footer = ui.IndentBlock(footer, indent)
		}
		return ui.PadBlockRight(content, padWidth), ui.PadBlockRight(footer, padWidth)
	}

	if m.articleErr != nil {
		b.WriteString(styles.Dim.Render(fmt.Sprintf("%v", m.articleErr)))
		content := b.String()
		footer := styles.Help.Render(articleLoadingHelp)
		footer = buildFooter(divider, footer)
		if indent > 0 {
			content = ui.IndentBlock(content, indent)
			footer = ui.IndentBlock(footer, indent)
		}
		return ui.PadBlockRight(content, padWidth), ui.PadBlockRight(footer, padWidth)
	}

	if len(m.articleLines) == 0 {
		b.WriteString("No article loaded.")
		content := b.String()
		footer := styles.Help.Render(articleLoadingHelp)
		footer = buildFooter(divider, footer)
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

	columnLabel := "1-col"
	if m.twoColumn {
		columnLabel = "2-col"
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

	footer := buildFooter(divider, hintLine, styles.Help.Render(help))
	if m.opts.Debug {
		detail := fmt.Sprintf("fetch=%s base=%s reflow=%s", m.fetchDuration, m.baseDuration, m.reflowDuration)
		footer = strings.TrimRight(footer, "\n") + "\n" + styles.Dim.Render(detail)
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

func buildFooter(divider string, lines ...string) string {
	footerLines := []string{"", divider}
	for _, line := range lines {
		if line == "" {
			continue
		}
		footerLines = append(footerLines, line)
	}
	return strings.Join(footerLines, "\n")
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

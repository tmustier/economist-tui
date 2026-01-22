package browse

import (
	"fmt"
	"math"
	"strings"

	"github.com/tmustier/economist-cli/internal/ui"
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

		if len(items) > maxVisible {
			b.WriteString(styles.Dim.Render(fmt.Sprintf("\n  (%d/%d)", m.cursor+1, len(items))))
		}
	}

	content := b.String()
	footer := styles.Help.Render(browseHelpText)
	if indent > 0 {
		content = ui.IndentBlock(content, indent)
		footer = ui.IndentBlock(footer, indent)
	}

	return content, footer
}

func (m Model) articleView() (string, string) {
	styles := ui.NewBrowseStyles(m.opts.NoColor)
	indent := ui.ArticleIndent(m.articleRenderOptions())

	var b strings.Builder
	if m.loading {
		b.WriteString("Loading article...")
		content := b.String()
		footer := styles.Help.Render(articleLoadingHelp)
		if indent > 0 {
			content = ui.IndentBlock(content, indent)
			footer = ui.IndentBlock(footer, indent)
		}
		return content, footer
	}

	if m.articleErr != nil {
		b.WriteString(styles.Dim.Render(fmt.Sprintf("%v", m.articleErr)))
		content := b.String()
		footer := styles.Help.Render(articleLoadingHelp)
		if indent > 0 {
			content = ui.IndentBlock(content, indent)
			footer = ui.IndentBlock(footer, indent)
		}
		return content, footer
	}

	if len(m.articleLines) == 0 {
		b.WriteString("No article loaded.")
		content := b.String()
		footer := styles.Help.Render(articleLoadingHelp)
		if indent > 0 {
			content = ui.IndentBlock(content, indent)
			footer = ui.IndentBlock(footer, indent)
		}
		return content, footer
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
	footerLines := []string{}
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
		hint := fmt.Sprintf("%d%% · more ↓", pct)
		footerLines = append(footerLines, styles.Dim.Render(hint))
	}
	footerLines = append(footerLines, styles.Help.Render(help))

	if m.opts.Debug {
		detail := fmt.Sprintf("fetch=%s base=%s reflow=%s", m.fetchDuration, m.baseDuration, m.reflowDuration)
		footerLines = append(footerLines, styles.Dim.Render(detail))
	}

	footer := strings.Join(footerLines, "\n")
	if indent > 0 {
		footer = ui.IndentBlock(footer, indent)
	}

	return content, footer
}

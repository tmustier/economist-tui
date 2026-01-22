package browse

import (
	"fmt"
	"math"
	"strings"

	"github.com/tmustier/economist-cli/internal/ui"
)

func (m Model) View() string {
	if m.mode == modeArticle {
		return padView(m.articleView(), m.height)
	}
	return padView(m.browseView(), m.height)
}

func (m Model) browseView() string {
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
		dateWidth := ui.DefaultDateWidth
		dateGap := ui.DefaultDateGap
		useCompactDates := contentWidth < prefixWidth+ui.MinTitleWidth+dateWidth+dateGap
		if useCompactDates {
			dateWidth = len("02.01.06")
		}
		layout := ui.NewHeadlineLayout(contentWidth, prefixWidth, dateWidth+dateGap)
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
			if useCompactDates {
				date = item.CompactDate()
			}
			dateColumn := fmt.Sprintf("%*s", layout.DateWidth, date)

			titleLines := limitLines(ui.WrapLines(title, layout.TitleWidth), browseTitleLines, layout.TitleWidth)
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
			subtitleLines := limitLines(ui.WrapLines(desc, subtitleWidth), browseSubtitleLines, subtitleWidth)
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

	b.WriteString("\n\n")
	b.WriteString(styles.Help.Render(browseHelpText))

	content := b.String()
	if indent > 0 {
		content = ui.IndentBlock(content, indent)
	}

	return content
}

func (m Model) articleView() string {
	styles := ui.NewBrowseStyles(m.opts.NoColor)

	var b strings.Builder
	if m.loading {
		b.WriteString("Loading article...\n\n")
		b.WriteString(styles.Help.Render(articleLoadingHelp))
		return b.String()
	}

	if m.articleErr != nil {
		b.WriteString(styles.Dim.Render(fmt.Sprintf("%v", m.articleErr)))
		b.WriteString("\n\n")
		b.WriteString(styles.Help.Render(articleLoadingHelp))
		return b.String()
	}

	if len(m.articleLines) == 0 {
		b.WriteString("No article loaded.\n\n")
		b.WriteString(styles.Help.Render(articleLoadingHelp))
		return b.String()
	}

	start := ui.Min(m.scroll, m.maxArticleScroll())
	viewHeight := m.articleViewHeight()
	end := ui.Min(len(m.articleLines), start+viewHeight)
	b.WriteString(strings.Join(m.articleLines[start:end], "\n"))
	b.WriteString("\n")

	columnLabel := "1-col"
	if m.twoColumn {
		columnLabel = "2-col"
	}
	help := fmt.Sprintf(articleHelpFormat, columnLabel)

	indent := ui.ArticleIndent(m.articleRenderOptions())

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
		hint := fmt.Sprintf("%d%% · more ↓", pct)
		hintLine = styles.Dim.Render(hint)
	}
	if indent > 0 {
		b.WriteString(ui.IndentBlock(hintLine, indent))
	} else {
		b.WriteString(hintLine)
	}
	b.WriteString("\n")

	if indent > 0 {
		helper := ui.IndentBlock(styles.Help.Render(help), indent)
		b.WriteString(helper)
	} else {
		b.WriteString(styles.Help.Render(help))
	}

	if m.opts.Debug {
		b.WriteString("\n")
		detail := fmt.Sprintf("fetch=%s base=%s reflow=%s", m.fetchDuration, m.baseDuration, m.reflowDuration)
		if indent > 0 {
			b.WriteString(ui.IndentBlock(styles.Dim.Render(detail), indent))
		} else {
			b.WriteString(styles.Dim.Render(detail))
		}
	}

	return b.String()
}

func padView(view string, height int) string {
	if height <= 0 {
		return view
	}
	lines := strings.Count(view, "\n") + 1
	if lines >= height {
		return view
	}
	return view + strings.Repeat("\n", height-lines)
}

func limitLines(lines []string, count, width int) []string {
	if count <= 0 {
		return nil
	}
	if len(lines) > count {
		trimmed := append([]string(nil), lines[:count]...)
		lastIdx := count - 1
		trimmed[lastIdx] = addEllipsis(trimmed[lastIdx], width)
		return trimmed
	}
	return lines
}

func addEllipsis(line string, width int) string {
	line = strings.TrimRight(line, " ")
	if width <= 0 {
		if line == "" {
			return "..."
		}
		return line + "..."
	}
	if width <= 3 {
		return strings.Repeat(".", width)
	}
	if line == "" {
		return "..."
	}
	if len(line)+3 <= width {
		return line + "..."
	}
	return ui.Truncate(line, width)
}

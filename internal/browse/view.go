package browse

import (
	"fmt"
	"strings"

	"github.com/tmustier/economist-cli/internal/ui"
)

func (m Model) View() string {
	if m.mode == modeArticle {
		return m.articleView()
	}
	return m.browseView()
}

func (m Model) browseView() string {
	styles := ui.NewBrowseStyles(m.opts.NoColor)

	var b strings.Builder

	header := styles.Header.Render(m.sectionTitle)
	b.WriteString(header + "\n")
	rule := strings.Repeat("─", ui.Min(m.width, 60))
	b.WriteString(styles.Rule.Render(rule) + "\n")

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
		prefix := fmt.Sprintf("%*d. ", numWidth, len(m.allItems))
		layout := ui.NewHeadlineLayout(m.width, len(prefix))

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

		if len(items) > maxVisible {
			b.WriteString(styles.Dim.Render(fmt.Sprintf("\n  (%d/%d)", m.cursor+1, len(items))))
		}
	}

	b.WriteString("\n\n")
	b.WriteString(styles.Help.Render(browseHelpText))

	return b.String()
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
	end := ui.Min(len(m.articleLines), start+m.articleViewHeight())
	b.WriteString(strings.Join(m.articleLines[start:end], "\n"))
	b.WriteString("\n\n")

	columnLabel := "1-col"
	if m.twoColumn {
		columnLabel = "2-col"
	}
	help := fmt.Sprintf(articleHelpFormat, columnLabel)
	b.WriteString(styles.Help.Render(help))

	return b.String()
}

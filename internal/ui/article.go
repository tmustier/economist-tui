package ui

import (
	"fmt"
	"strings"

	"github.com/muesli/reflow/wordwrap"
	"github.com/tmustier/economist-tui/internal/article"
)

func RenderArticleHeader(art *article.Article, styles ArticleStyles, opts ArticleRenderOptions) string {
	var sb strings.Builder
	layout := resolveArticleLayout(opts)
	wrapWidth := layout.WrapWidth

	sb.WriteString("\n")
	wroteOvertitle := writeWrapped(&sb, art.Overtitle, wrapWidth, func(line string) string {
		return renderOvertitle(line, styles)
	})
	if wroteOvertitle && (art.Title != "" || art.Subtitle != "" || art.DateLine != "") {
		sb.WriteString("\n")
	}
	writeWrapped(&sb, art.Title, wrapWidth, func(line string) string {
		return styles.Title.Render(line)
	})
	writeWrapped(&sb, art.Subtitle, wrapWidth, func(line string) string {
		return styles.Subtitle.Render(line)
	})
	writeWrapped(&sb, art.DateLine, wrapWidth, func(line string) string {
		return styles.Date.Render(line)
	})

	sb.WriteString("\n")
	accentStyles := NewStyles(CurrentTheme(), opts.NoColor)
	sb.WriteString(AccentRule(layout.ContentWidth, accentStyles))
	sb.WriteString("\n\n")
	return sb.String()
}

func ArticleBodyMarkdown(art *article.Article) string {
	var sb strings.Builder
	if art.Content != "" {
		sb.WriteString(art.Content)
	}
	return sb.String()
}

func ArticleFooter(art *article.Article, styles ArticleStyles, opts ArticleRenderOptions) string {
	var sb strings.Builder
	layout := resolveArticleLayout(opts)
	accentStyles := NewStyles(CurrentTheme(), opts.NoColor)
	sb.WriteString("\n\n")
	sb.WriteString(AccentRule(layout.ContentWidth, accentStyles))
	sb.WriteString("\n\n")
	sb.WriteString(styles.Body.Render(art.URL))
	sb.WriteString("\n")
	return sb.String()
}

func HighlightTrailingMarker(text string, styles ArticleStyles) string {
	idx := strings.LastIndex(text, "■")
	if idx == -1 {
		return text
	}

	marker := styles.Title.Render("■")
	return text[:idx] + marker + text[idx+len("■"):]
}

func writeWrapped(sb *strings.Builder, text string, width int, render func(string) string) bool {
	if text == "" {
		return false
	}
	for _, line := range wrapHeaderText(text, width) {
		sb.WriteString(render(line))
		sb.WriteString("\n")
	}
	return true
}

func wrapHeaderText(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}
	wrapped := wordwrap.String(text, width)
	return strings.Split(wrapped, "\n")
}

func renderOvertitle(text string, styles ArticleStyles) string {
	parts := strings.SplitN(text, "|", 2)
	if len(parts) < 2 {
		return styles.Overtitle.Render(text)
	}

	section := strings.TrimSpace(parts[0])
	rest := strings.TrimSpace(parts[1])
	if section == "" {
		return styles.Overtitle.Render(text)
	}

	if rest == "" {
		return styles.Section.Render(section)
	}

	return fmt.Sprintf("%s | %s", styles.Section.Render(section), styles.Overtitle.Render(rest))
}

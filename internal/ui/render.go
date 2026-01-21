package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/muesli/reflow/ansi"
	"github.com/muesli/reflow/wrap"
	"github.com/tmustier/economist-cli/internal/article"
)

const (
	columnGap      = 4
	minColumnWidth = 32
)

type ArticleRenderOptions struct {
	Raw       bool
	NoColor   bool
	WrapWidth int
	TermWidth int
	TwoColumn bool
}

func RenderArticle(art *article.Article, opts ArticleRenderOptions) (string, error) {
	if opts.Raw {
		return art.ToMarkdown(), nil
	}

	styles := NewArticleStyles(opts.NoColor)
	header := RenderArticleHeader(art, styles)
	markdown := ArticleBodyMarkdown(art)

	contentWidth := resolveContentWidth(opts)
	columnWidth, useColumns := resolveColumnWidth(contentWidth, opts.TwoColumn)

	wrapWidth := contentWidth
	if useColumns {
		wrapWidth = columnWidth
	}

	body, err := renderArticleBody(markdown, wrapWidth, opts)
	if err != nil {
		return "", err
	}

	if useColumns {
		body = columnize(body, columnWidth)
	}

	if !opts.NoColor {
		body = HighlightTrailingMarker(body, styles)
	}

	return header + body, nil
}

func resolveContentWidth(opts ArticleRenderOptions) int {
	if opts.WrapWidth > 0 {
		return opts.WrapWidth
	}
	if opts.TermWidth > 0 {
		return opts.TermWidth
	}
	return TermWidth(int(os.Stdout.Fd()))
}

func resolveColumnWidth(contentWidth int, enabled bool) (int, bool) {
	if !enabled {
		return 0, false
	}
	if contentWidth <= 0 {
		return 0, false
	}
	width := (contentWidth - columnGap) / 2
	if width < minColumnWidth {
		return 0, false
	}
	return width, true
}

func renderArticleBody(markdown string, width int, opts ArticleRenderOptions) (string, error) {
	if opts.NoColor {
		if width > 0 {
			return wrap.String(markdown, width), nil
		}
		return markdown, nil
	}

	optsList := []glamour.TermRendererOption{glamour.WithAutoStyle()}
	if width > 0 {
		optsList = append(optsList, glamour.WithWordWrap(width))
	}

	renderer, err := glamour.NewTermRenderer(optsList...)
	if err != nil {
		return "", err
	}

	out, err := renderer.Render(markdown)
	if err != nil {
		return "", err
	}

	return out, nil
}

func columnize(text string, columnWidth int) string {
	trimmed := strings.TrimRight(text, "\n")
	lines := strings.Split(trimmed, "\n")
	if len(lines) == 0 {
		return text
	}

	rows := (len(lines) + 1) / 2
	gap := strings.Repeat(" ", columnGap)
	var b strings.Builder
	for i := 0; i < rows; i++ {
		left := ""
		if i < len(lines) {
			left = lines[i]
		}
		left = padRightANSI(left, columnWidth)

		if i+rows < len(lines) {
			right := lines[i+rows]
			b.WriteString(left)
			b.WriteString(gap)
			b.WriteString(right)
		} else {
			b.WriteString(left)
		}

		if i < rows-1 {
			b.WriteString("\n")
		}
	}

	if strings.HasSuffix(text, "\n") {
		b.WriteString("\n")
	}

	return b.String()
}

func padRightANSI(text string, width int) string {
	pad := width - ansi.PrintableRuneWidth(text)
	if pad <= 0 {
		return text
	}
	return fmt.Sprintf("%s%s", text, strings.Repeat(" ", pad))
}

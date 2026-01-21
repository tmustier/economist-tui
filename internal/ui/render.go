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
	columnGap       = 4
	minColumnWidth  = 32
	baseWrapWidth   = 2000
	headerIndentMin = 0
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

	base, err := RenderArticleBodyBase(markdown, opts)
	if err != nil {
		return "", err
	}

	body := ReflowArticleBody(base, styles, opts)
	indent := DetectIndent(body)
	if indent > headerIndentMin {
		header = IndentBlock(header, indent)
	}

	footer := ArticleFooter(art, styles)
	if indent > headerIndentMin {
		footer = IndentBlock(footer, indent)
	}

	return header + body + footer, nil
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

func RenderArticleBodyBase(markdown string, opts ArticleRenderOptions) (string, error) {
	if opts.NoColor {
		return markdown, nil
	}

	optsList := []glamour.TermRendererOption{
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(baseWrapWidth),
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

func ReflowArticleBody(base string, styles ArticleStyles, opts ArticleRenderOptions) string {
	contentWidth := resolveContentWidth(opts)
	columnWidth, useColumns := resolveColumnWidth(contentWidth, opts.TwoColumn)

	wrapWidth := contentWidth
	if useColumns {
		wrapWidth = columnWidth
	}

	body := base
	if wrapWidth > 0 {
		body = wrap.String(body, wrapWidth)
	}

	if useColumns {
		body = columnize(body, columnWidth)
	}

	if !opts.NoColor {
		body = HighlightTrailingMarker(body, styles)
	}

	return body
}

func columnize(text string, columnWidth int) string {
	trimmed := strings.TrimRight(text, "\n")
	lines := strings.Split(trimmed, "\n")
	lines = trimLeadingBlankLines(lines)
	lines = trimTrailingBlankLines(lines)
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

func trimLeadingBlankLines(lines []string) []string {
	start := 0
	for start < len(lines) && isLineBlank(lines[start]) {
		start++
	}
	return lines[start:]
}

func trimTrailingBlankLines(lines []string) []string {
	end := len(lines)
	for end > 0 && isLineBlank(lines[end-1]) {
		end--
	}
	return lines[:end]
}

func DetectIndent(text string) int {
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		if isLineBlank(line) {
			continue
		}
		return leadingIndent(line)
	}
	return 0
}

func IndentBlock(text string, indent int) string {
	if indent <= 0 {
		return text
	}
	pad := strings.Repeat(" ", indent)
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if line == "" {
			continue
		}
		lines[i] = pad + line
	}
	return strings.Join(lines, "\n")
}

func isLineBlank(line string) bool {
	inANSI := false
	for _, r := range line {
		switch {
		case inANSI:
			if ansi.IsTerminator(r) {
				inANSI = false
			}
			continue
		case r == ansi.Marker:
			inANSI = true
			continue
		case r == ' ' || r == '\t':
			continue
		default:
			return false
		}
	}
	return true
}

func leadingIndent(line string) int {
	indent := 0
	inANSI := false
	for _, r := range line {
		switch {
		case inANSI:
			if ansi.IsTerminator(r) {
				inANSI = false
			}
			continue
		case r == ansi.Marker:
			inANSI = true
			continue
		case r == ' ':
			indent++
		case r == '\t':
			indent += 4
		default:
			return indent
		}
	}
	return indent
}

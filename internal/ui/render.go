package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/glamour"
	cansi "github.com/charmbracelet/x/ansi"
	"github.com/muesli/reflow/ansi"
	"github.com/muesli/termenv"
	"github.com/tmustier/economist-tui/internal/article"
)

const (
	columnGap      = 4
	minColumnWidth = 32
	maxColumnWidth = MaxReadableWidth
	minColumnLines = 18
	baseWrapWidth  = 2000
	bodyIndent     = 2
)

type ArticleRenderOptions struct {
	Raw       bool
	NoColor   bool
	PlainBody bool
	WrapWidth int
	TermWidth int
	MaxWidth  int
	Center    bool
	TwoColumn bool
}

type ArticleLayout struct {
	TermWidth       int
	ContentWidth    int
	Indent          int
	OuterPadding    int
	WrapWidth       int
	HeaderWrapWidth int
	ColumnWidth     int
	ColumnCount     int
	UseColumns      bool
}

func RenderArticle(art *article.Article, opts ArticleRenderOptions) (string, error) {
	if opts.Raw {
		return art.ToMarkdown(), nil
	}

	styles := NewArticleStyles(opts.NoColor)
	markdown := ArticleBodyMarkdown(art)

	base, err := RenderArticleBodyBase(markdown, opts)
	if err != nil {
		return "", err
	}

	layout := ResolveArticleLayoutWithContent(base, opts)
	header := RenderArticleHeaderWithLayout(art, styles, layout, opts)
	body := ReflowArticleBodyWithLayout(base, styles, opts, layout)
	indent := ArticleIndentForLayout(layout)
	if indent > 0 {
		header = IndentBlock(header, indent)
	}

	footer := ArticleFooterWithLayout(art, styles, layout, opts)
	if indent > 0 {
		footer = IndentBlock(footer, indent)
	}

	return header + body + footer, nil
}

func resolveTerminalWidth(opts ArticleRenderOptions) int {
	if opts.TermWidth > 0 {
		return opts.TermWidth
	}
	return TermWidth(int(os.Stdout.Fd()))
}

func resolveContentWidth(opts ArticleRenderOptions) int {
	if opts.WrapWidth > 0 {
		return opts.WrapWidth
	}
	termWidth := resolveTerminalWidth(opts)
	if opts.MaxWidth > 0 && termWidth > opts.MaxWidth {
		return opts.MaxWidth
	}
	return termWidth
}

func ArticleIndent(opts ArticleRenderOptions) int {
	return ArticleIndentForLayout(ResolveArticleLayout(opts))
}

func ArticleIndentForLayout(layout ArticleLayout) int {
	return layout.Indent + layout.OuterPadding
}

func resolveColumnLayout(contentWidth int, enabled bool) (int, int, bool) {
	if !enabled {
		return 0, 1, false
	}
	if contentWidth <= 0 {
		return 0, 1, false
	}
	minWidth := minColumnWidth*2 + columnGap
	if contentWidth < minWidth {
		return 0, 1, false
	}

	numerator := contentWidth + columnGap
	denominator := maxColumnWidth + columnGap
	columns := (numerator + denominator - 1) / denominator
	if columns < 2 {
		columns = 2
	}

	width := (contentWidth - columnGap*(columns-1)) / columns
	if width < minColumnWidth {
		columns = 2
		width = (contentWidth - columnGap) / 2
		if width < minColumnWidth {
			return 0, 1, false
		}
	}
	return width, columns, true
}

func ResolveArticleLayout(opts ArticleRenderOptions) ArticleLayout {
	return resolveArticleLayout(opts, "")
}

func ResolveArticleLayoutWithContent(base string, opts ArticleRenderOptions) ArticleLayout {
	return resolveArticleLayout(opts, base)
}

func resolveArticleLayout(opts ArticleRenderOptions, base string) ArticleLayout {
	termWidth := resolveTerminalWidth(opts)
	contentWidth := resolveContentWidth(opts)
	if contentWidth < 0 {
		contentWidth = 0
	}

	indent := 0
	availableWidth := contentWidth
	if contentWidth > bodyIndent {
		indent = bodyIndent
		availableWidth = contentWidth - indent
	}

	columnWidth, columnCount, useColumns := resolveColumnLayout(availableWidth, opts.TwoColumn)
	if !useColumns {
		columnCount = 1
	}
	wrapWidth := availableWidth
	if useColumns {
		wrapWidth = columnWidth
	}

	if useColumns && base != "" && columnCount > 2 {
		lineCount := countWrappedLines(base, wrapWidth)
		maxByHeight := lineCount / minColumnLines
		if maxByHeight < 2 {
			maxByHeight = 2
		}
		if columnCount > maxByHeight {
			columnCount = maxByHeight
		}
	}

	if useColumns {
		if opts.Center {
			columnBlockWidth := columnBlockWidth(columnCount)
			if columnBlockWidth > 0 && availableWidth > columnBlockWidth {
				availableWidth = columnBlockWidth
				contentWidth = availableWidth + indent
			}
		}
		columnWidth = (availableWidth - columnGap*(columnCount-1)) / columnCount
		if columnWidth < minColumnWidth {
			columnWidth = 0
			columnCount = 1
			useColumns = false
			wrapWidth = availableWidth
		} else {
			wrapWidth = columnWidth
		}
	}

	headerWrapWidth := wrapWidth
	if useColumns {
		headerWrapWidth = availableWidth
	}
	if headerWrapWidth < 0 {
		headerWrapWidth = 0
	}

	outerPadding := 0
	if opts.Center && termWidth > contentWidth {
		outerPadding = (termWidth - contentWidth) / 2
	}

	return ArticleLayout{
		TermWidth:       termWidth,
		ContentWidth:    contentWidth,
		Indent:          indent,
		OuterPadding:    outerPadding,
		WrapWidth:       wrapWidth,
		HeaderWrapWidth: headerWrapWidth,
		ColumnWidth:     columnWidth,
		ColumnCount:     columnCount,
		UseColumns:      useColumns,
	}
}

func columnBlockWidth(columnCount int) int {
	if columnCount < 1 {
		return 0
	}
	return columnCount*maxColumnWidth + columnGap*(columnCount-1)
}

func wrapBody(text string, width int) string {
	if width <= 0 {
		return text
	}
	return cansi.Wrap(text, width, "")
}

func countWrappedLines(text string, width int) int {
	wrapped := wrapBody(text, width)
	wrapped = normalizeParagraphSpacing(wrapped)
	trimmed := strings.TrimRight(wrapped, "\n")
	if trimmed == "" {
		return 0
	}
	return strings.Count(trimmed, "\n") + 1
}

func writeHeaderAccent(sb *strings.Builder, layout ArticleLayout, opts ArticleRenderOptions) {
	sb.WriteString("\n")
	accentStyles := NewStyles(CurrentTheme(), opts.NoColor)
	sb.WriteString(AccentRule(layout.ContentWidth, accentStyles))
	sb.WriteString("\n\n")
}

func RenderArticleBodyBase(markdown string, opts ArticleRenderOptions) (string, error) {
	if opts.NoColor || opts.PlainBody {
		return markdown, nil
	}

	styles := glamour.DarkStyleConfig
	bodyColor := BodyColorDarkANSI
	if !termenv.HasDarkBackground() {
		styles = glamour.LightStyleConfig
		bodyColor = BodyColorLightANSI
	}
	styles.Document.Margin = uintPtr(0)
	styles.Document.Color = &bodyColor

	optsList := []glamour.TermRendererOption{
		glamour.WithStyles(styles),
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
	layout := ResolveArticleLayoutWithContent(base, opts)
	return ReflowArticleBodyWithLayout(base, styles, opts, layout)
}

func ReflowArticleBodyWithLayout(base string, styles ArticleStyles, opts ArticleRenderOptions, layout ArticleLayout) string {
	innerIndent := layout.Indent
	outerPadding := layout.OuterPadding
	columnWidth := layout.ColumnWidth
	columnCount := layout.ColumnCount
	useColumns := layout.UseColumns
	wrapWidth := layout.WrapWidth

	body := base
	if wrapWidth > 0 {
		body = wrapBody(body, wrapWidth)
	}
	body = normalizeParagraphSpacing(body)

	if useColumns {
		body = columnize(body, columnWidth, columnCount)
	}

	if !opts.NoColor {
		body = HighlightTrailingMarker(body, styles)
	}

	totalIndent := innerIndent + outerPadding
	if totalIndent > 0 {
		body = IndentBlock(body, totalIndent)
	}

	if opts.PlainBody && !opts.NoColor {
		body = styles.Body.Render(body)
	}

	return body
}

func columnize(text string, columnWidth int, columnCount int) string {
	if columnCount <= 1 {
		return text
	}
	trimmed := strings.TrimRight(text, "\n")
	lines := strings.Split(trimmed, "\n")
	lines = trimLeadingBlankLines(lines)
	lines = trimTrailingBlankLines(lines)
	if len(lines) == 0 {
		return text
	}

	rows := (len(lines) + columnCount - 1) / columnCount
	gap := strings.Repeat(" ", columnGap)
	var b strings.Builder
	for row := 0; row < rows; row++ {
		for col := 0; col < columnCount; col++ {
			idx := row + col*rows
			if idx >= len(lines) {
				break
			}
			line := padRightANSI(lines[idx], columnWidth)
			b.WriteString(line)

			if col < columnCount-1 {
				nextIdx := row + (col+1)*rows
				if nextIdx < len(lines) {
					b.WriteString(gap)
				}
			}
		}

		if row < rows-1 {
			b.WriteString("\n")
		}
	}

	if strings.HasSuffix(text, "\n") {
		b.WriteString("\n")
	}

	return b.String()
}

func padRightANSI(text string, width int) string {
	if width <= 0 {
		return text
	}
	trimmed := cansi.Truncate(text, width, "")
	pad := width - ansi.PrintableRuneWidth(trimmed)
	if pad <= 0 {
		return trimmed
	}
	return fmt.Sprintf("%s%s", trimmed, strings.Repeat(" ", pad))
}

func normalizeParagraphSpacing(text string) string {
	lines := strings.Split(text, "\n")
	var out []string
	blank := 0
	for _, line := range lines {
		if isLineBlank(line) {
			blank++
			if blank > 1 {
				continue
			}
		} else {
			blank = 0
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
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

func uintPtr(v uint) *uint {
	return &v
}

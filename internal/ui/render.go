package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/muesli/reflow/ansi"
	"github.com/muesli/reflow/wordwrap"
	"github.com/muesli/termenv"
	"github.com/tmustier/economist-tui/internal/article"
)

const (
	columnGap      = 4
	minColumnWidth = 32
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
	UseColumns      bool
}

func RenderArticle(art *article.Article, opts ArticleRenderOptions) (string, error) {
	if opts.Raw {
		return art.ToMarkdown(), nil
	}

	styles := NewArticleStyles(opts.NoColor)
	header := RenderArticleHeader(art, styles, opts)
	markdown := ArticleBodyMarkdown(art)

	base, err := RenderArticleBodyBase(markdown, opts)
	if err != nil {
		return "", err
	}

	body := ReflowArticleBody(base, styles, opts)
	indent := ArticleIndent(opts)
	if indent > 0 {
		header = IndentBlock(header, indent)
	}

	footer := ArticleFooter(art, styles, opts)
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
	layout := resolveArticleLayout(opts)
	return layout.Indent + layout.OuterPadding
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

func resolveArticleLayout(opts ArticleRenderOptions) ArticleLayout {
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

	columnWidth, useColumns := resolveColumnWidth(availableWidth, opts.TwoColumn)
	wrapWidth := availableWidth
	if useColumns {
		wrapWidth = columnWidth
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
		UseColumns:      useColumns,
	}
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
	layout := resolveArticleLayout(opts)
	innerIndent := layout.Indent
	outerPadding := layout.OuterPadding
	columnWidth := layout.ColumnWidth
	useColumns := layout.UseColumns
	wrapWidth := layout.WrapWidth

	body := base
	if wrapWidth > 0 {
		body = wordwrap.String(body, wrapWidth)
	}
	body = normalizeParagraphSpacing(body)

	if useColumns {
		body = columnize(body, columnWidth)
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

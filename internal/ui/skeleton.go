package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// SkeletonBlock characters
const (
	SkeletonSolid = "█"
	SkeletonLight = "░"

	skeletonOvertitleWidth  = 16
	skeletonMinLineLength   = 10
	skeletonParagraphStride = 4
	skeletonBodyMinLines    = 8
	skeletonBodyMaxLines    = 32
)

var skeletonLineLengths = []float64{0.95, 0.88, 1.0, 0.75, 0.92, 0.85, 1.0, 0.70}

// SkeletonStyles holds styles for skeleton loading states.
type SkeletonStyles struct {
	Section  lipgloss.Style
	Solid    lipgloss.Style
	Light    lipgloss.Style
	Title    lipgloss.Style
	Subtitle lipgloss.Style
	Date     lipgloss.Style
}

// NewSkeletonStyles creates themed skeleton styles.
func NewSkeletonStyles(noColor bool) SkeletonStyles {
	articleStyles := NewArticleStyles(noColor)
	theme := CurrentTheme()

	solid := lipgloss.NewStyle().Foreground(theme.TextFaint)
	light := lipgloss.NewStyle().Foreground(theme.Border)
	if noColor {
		solid = lipgloss.NewStyle()
		light = lipgloss.NewStyle()
	}

	return SkeletonStyles{
		Section:  articleStyles.Section,
		Solid:    solid,
		Light:    light,
		Title:    articleStyles.Title,
		Subtitle: articleStyles.Subtitle,
		Date:     articleStyles.Date,
	}
}

// SkeletonHeader contains known values for the loading skeleton.
type SkeletonHeader struct {
	Section  string // e.g., "Leaders" - known from browse context
	Title    string // Headline from RSS
	Subtitle string // Description from RSS
	Date     string // Formatted date from RSS
}

// RenderSkeletonHeader renders the article header skeleton.
func RenderSkeletonHeader(h SkeletonHeader, styles SkeletonStyles, opts ArticleRenderOptions) string {
	var sb strings.Builder
	layout := resolveArticleLayout(opts)
	wrapWidth := layout.WrapWidth

	sb.WriteString("\n")

	// Overtitle: Section | ████████████████
	if h.Section != "" {
		overtitleSkeleton := strings.Repeat(SkeletonSolid, skeletonOvertitleWidth)
		line := styles.Section.Render(h.Section) + " | " + styles.Solid.Render(overtitleSkeleton)
		sb.WriteString(line)
		sb.WriteString("\n\n")
	}

	// Title (real from RSS)
	if h.Title != "" {
		for _, line := range WrapLines(h.Title, wrapWidth) {
			sb.WriteString(styles.Title.Render(line))
			sb.WriteString("\n")
		}
	}

	// Subtitle (real from RSS)
	if h.Subtitle != "" {
		for _, line := range WrapLines(h.Subtitle, wrapWidth) {
			sb.WriteString(styles.Subtitle.Render(line))
			sb.WriteString("\n")
		}
	}

	// Date (real from RSS)
	if h.Date != "" {
		sb.WriteString(styles.Date.Render(h.Date))
		sb.WriteString("\n")
	}

	// Accent rule
	sb.WriteString("\n")
	accentStyles := NewStyles(CurrentTheme(), opts.NoColor)
	sb.WriteString(AccentRule(layout.ContentWidth, accentStyles))
	sb.WriteString("\n\n")

	return sb.String()
}

// RenderSkeletonBody renders placeholder lines for the article body.
func RenderSkeletonBody(styles SkeletonStyles, opts ArticleRenderOptions, lineCount int) string {
	layout := resolveArticleLayout(opts)
	wrapWidth := layout.WrapWidth
	if wrapWidth <= 0 {
		wrapWidth = 60
	}

	var sb strings.Builder
	for i := 0; i < lineCount; i++ {
		// Add paragraph breaks every ~4 lines
		if i > 0 && i%skeletonParagraphStride == 0 {
			sb.WriteString("\n")
		}

		lengthFactor := skeletonLineLengths[i%len(skeletonLineLengths)]
		lineLen := int(float64(wrapWidth) * lengthFactor)
		if lineLen < skeletonMinLineLength {
			lineLen = skeletonMinLineLength
		}

		skeletonLine := strings.Repeat(SkeletonLight, lineLen)
		sb.WriteString(styles.Light.Render(skeletonLine))
		sb.WriteString("\n")
	}

	return sb.String()
}

// RenderArticleSkeleton renders a complete skeleton loading view.
// availableLines is the visible area height (0 = use default).
func RenderArticleSkeleton(h SkeletonHeader, opts ArticleRenderOptions, availableLines int) string {
	styles := NewSkeletonStyles(opts.NoColor)

	header := RenderSkeletonHeader(h, styles, opts)

	// Calculate body lines based on available space
	headerLines := lineCount(header)
	availableForBody := availableLines - headerLines
	if availableLines <= 0 {
		availableForBody = skeletonBodyMaxLines
	}
	if availableForBody < 0 {
		availableForBody = 0
	}

	// Body adds paragraph breaks every 4 lines, so actual lines = content + breaks
	// For N content lines with breaks every 4: total = N + (N/4)
	// So content lines = availableForBody * 4 / 5
	bodyLines := (availableForBody * 4) / 5
	if bodyLines < skeletonBodyMinLines {
		bodyLines = skeletonBodyMinLines
	}
	if bodyLines > skeletonBodyMaxLines {
		bodyLines = skeletonBodyMaxLines
	}

	body := RenderSkeletonBody(styles, opts, bodyLines)

	indent := ArticleIndent(opts)
	if indent > 0 {
		header = IndentBlock(header, indent)
		body = IndentBlock(body, indent)
	}

	return header + body
}

package browse

import "github.com/tmustier/economist-tui/internal/ui"

func browseLayoutSpec(helpLineCount int, showPosition, showSectionDots bool) ui.LayoutSpec {
	return ui.LayoutSpec{
		HeaderLines:     browseHeaderLines,
		FooterLines:     browseFooterLines(helpLineCount, showPosition, showSectionDots),
		FooterPadding:   browseFooterPadding,
		FooterGapLines:  browseFooterGapLines,
		MinVisibleLines: browseMinVisibleLines,
	}
}

func articleLayoutSpec(debug bool) ui.LayoutSpec {
	footerLines := articleFooterLines
	if debug {
		footerLines++
	}
	return ui.LayoutSpec{
		FooterLines:     footerLines,
		FooterPadding:   articleFooterPadding,
		FooterGapLines:  articleFooterGapLines,
		MinVisibleLines: articleMinVisibleLines,
	}
}

func browseFooterLines(helpLineCount int, showPosition, showSectionDots bool) int {
	lines := 2 + helpLineCount
	if showPosition {
		lines++
	}
	if showSectionDots {
		lines++
	}
	return lines
}

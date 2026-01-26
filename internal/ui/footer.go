package ui

import (
	"strings"

	"github.com/muesli/reflow/ansi"
)

// BuildFooter builds a footer block with a leading blank line and divider.
func BuildFooter(divider string, lines ...string) string {
	footerLines := []string{"", divider}
	for _, line := range lines {
		if line == "" {
			continue
		}
		footerLines = append(footerLines, line)
	}
	return strings.Join(footerLines, "\n")
}

// CenterText centers text within the given width.
// Accounts for ANSI escape sequences when calculating visible width.
func CenterText(text string, width int) string {
	visibleLen := ansi.PrintableRuneWidth(text)
	if visibleLen >= width {
		return text
	}
	padLeft := (width - visibleLen) / 2
	padRight := width - visibleLen - padLeft
	return strings.Repeat(" ", padLeft) + text + strings.Repeat(" ", padRight)
}

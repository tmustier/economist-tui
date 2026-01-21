package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	RuleLight  = "─"
	RuleHeavy  = "━"
	RuleDouble = "═"
	RuleDotted = "┄"
)

func DrawRule(width int, char string, style lipgloss.Style) string {
	if width <= 0 {
		return ""
	}
	return style.Render(strings.Repeat(char, width))
}

func AccentRule(width int, styles Styles) string {
	return DrawRule(width, RuleHeavy, styles.RuleAccent)
}

func SectionRule(width int, styles Styles) string {
	return DrawRule(width, RuleLight, styles.Rule)
}

func SectionBadge(section string, styles Styles) string {
	return styles.Overline.Render(strings.ToUpper(section))
}

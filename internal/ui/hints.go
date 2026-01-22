package ui

import "github.com/muesli/reflow/ansi"

// SelectHintLine returns the first option that fits the width.
func SelectHintLine(width int, options ...string) string {
	if len(options) == 0 {
		return ""
	}
	if width <= 0 {
		return options[0]
	}
	for _, option := range options {
		if ansi.PrintableRuneWidth(option) <= width {
			return option
		}
	}
	return options[len(options)-1]
}

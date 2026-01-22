package ui

import (
	"fmt"
	"strings"

	"github.com/muesli/reflow/wordwrap"
)

const (
	DefaultDateWidth = 13
	MinTitleWidth    = 30
)

type HeadlineLayout struct {
	Width       int
	PrefixWidth int
	TitleWidth  int
	DateWidth   int
}

func NewHeadlineLayout(width, prefixWidth, dateWidth int) HeadlineLayout {
	if dateWidth <= 0 {
		dateWidth = DefaultDateWidth
	}
	titleWidth := width - prefixWidth - dateWidth
	minWidth := prefixWidth + dateWidth + MinTitleWidth
	if titleWidth < MinTitleWidth && width >= minWidth {
		titleWidth = MinTitleWidth
	}
	if titleWidth < 1 {
		titleWidth = 1
	}

	return HeadlineLayout{
		Width:       width,
		PrefixWidth: prefixWidth,
		TitleWidth:  titleWidth,
		DateWidth:   dateWidth,
	}
}

func (l HeadlineLayout) PadTitle(title string) string {
	truncated := Truncate(title, l.TitleWidth)
	return fmt.Sprintf("%-*s", l.TitleWidth, truncated)
}

func WrapLines(text string, width int) []string {
	if text == "" {
		return nil
	}
	if width <= 0 {
		return []string{text}
	}
	wrapped := wordwrap.String(text, width)
	return strings.Split(wrapped, "\n")
}

func Truncate(text string, width int) string {
	if width <= 0 || len(text) <= width {
		return text
	}
	if width <= 3 {
		return text[:width]
	}
	return text[:width-3] + "..."
}

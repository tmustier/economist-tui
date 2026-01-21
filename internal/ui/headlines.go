package ui

import "fmt"

const (
	DefaultDateWidth = 14
	MinTitleWidth    = 30
)

type HeadlineLayout struct {
	Width       int
	PrefixWidth int
	TitleWidth  int
	DateWidth   int
}

func NewHeadlineLayout(width, prefixWidth int) HeadlineLayout {
	dateWidth := DefaultDateWidth
	titleWidth := width - prefixWidth - dateWidth
	if titleWidth < MinTitleWidth {
		titleWidth = MinTitleWidth
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

func Truncate(text string, width int) string {
	if width <= 0 || len(text) <= width {
		return text
	}
	if width <= 3 {
		return text[:width]
	}
	return text[:width-3] + "..."
}

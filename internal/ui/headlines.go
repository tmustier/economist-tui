package ui

import (
	"fmt"
	"strings"

	"github.com/muesli/reflow/wordwrap"
)

const (
	DefaultDateWidth = 13
	DefaultDateGap   = 2
	MinTitleWidth    = 30
)

type HeadlineLayout struct {
	Width       int
	PrefixWidth int
	TitleWidth  int
	DateWidth   int
}

type DateLayout struct {
	Width       int
	Gap         int
	ColumnWidth int
	Compact     bool
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

func ResolveDateLayout(contentWidth, prefixWidth int) DateLayout {
	dateWidth := DefaultDateWidth
	gap := DefaultDateGap
	compact := contentWidth < prefixWidth+MinTitleWidth+dateWidth+gap
	if compact {
		dateWidth = len("02.01.06")
	}
	return DateLayout{
		Width:       dateWidth,
		Gap:         gap,
		ColumnWidth: dateWidth + gap,
		Compact:     compact,
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

func LimitLines(lines []string, count, width int) []string {
	if count <= 0 {
		return nil
	}
	if len(lines) > count {
		trimmed := append([]string(nil), lines[:count]...)
		lastIdx := count - 1
		trimmed[lastIdx] = addEllipsis(trimmed[lastIdx], width)
		return trimmed
	}
	return lines
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

func addEllipsis(line string, width int) string {
	line = strings.TrimRight(line, " ")
	if width <= 0 {
		if line == "" {
			return "..."
		}
		return line + "..."
	}
	if width <= 3 {
		return strings.Repeat(".", width)
	}
	if line == "" {
		return "..."
	}
	if len(line)+3 <= width {
		return line + "..."
	}
	return Truncate(line, width)
}

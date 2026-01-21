package ui

import "golang.org/x/term"

const (
	DefaultWidth  = 100
	DefaultHeight = 24
)

func TermSize(fd int) (int, int) {
	w, h, err := term.GetSize(fd)
	if err != nil || w <= 0 {
		w = DefaultWidth
	}
	if h <= 0 {
		h = DefaultHeight
	}
	return w, h
}

func TermWidth(fd int) int {
	w, _ := TermSize(fd)
	return w
}

func IsTerminal(fd int) bool {
	return term.IsTerminal(fd)
}

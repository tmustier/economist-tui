package ui

import (
	"os"
	"strconv"

	"golang.org/x/term"
)

const (
	DefaultWidth  = 100
	DefaultHeight = 24
)

func TermSize(fd int) (int, int) {
	w, h, err := term.GetSize(fd)
	if err != nil || w <= 0 {
		w = envInt("COLUMNS", DefaultWidth)
	}
	if h <= 0 {
		h = envInt("LINES", DefaultHeight)
	}
	return w, h
}

func envInt(name string, fallback int) int {
	if val, ok := os.LookupEnv(name); ok {
		if n, err := strconv.Atoi(val); err == nil && n > 0 {
			return n
		}
	}
	return fallback
}

func TermWidth(fd int) int {
	w, _ := TermSize(fd)
	return w
}

func IsTerminal(fd int) bool {
	return term.IsTerminal(fd)
}

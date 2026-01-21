package ui

const (
	MinReadableWidth = 45
	IdealWidth       = 66
	MaxReadableWidth = 72
)

func ReaderContentWidth(terminalWidth int) int {
	if terminalWidth <= 0 {
		return MaxReadableWidth
	}

	available := terminalWidth - 8
	if available <= 0 {
		available = terminalWidth
	}

	if available > MaxReadableWidth {
		return MaxReadableWidth
	}

	if available < MinReadableWidth {
		if terminalWidth < MinReadableWidth {
			return terminalWidth
		}
		return MinReadableWidth
	}

	return available
}

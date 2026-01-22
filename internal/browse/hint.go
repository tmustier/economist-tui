package browse

import "github.com/tmustier/economist-tui/internal/ui"

type helpLineSpec struct {
	Options []string
}

var browseHelpLineSpecs = []helpLineSpec{
	{
		Options: []string{
			"↑/↓ navigate • ←/→ page • tab/shift+tab section",
			"↑/↓ navigate • ←/→ page • tab section",
			"↑/↓ move • ←/→ page • tab section",
			"↑/↓ move • ←/→ page • tab",
			"↑/↓ move • ←/→ page",
			"↑/↓ move • tab",
			"↑/↓",
		},
	},
	{
		Options: []string{
			"enter read • type to search • esc clear • q quit",
			"enter read • type search • esc clear • q quit",
			"enter read • search • esc clear • q quit",
			"enter read • search • q quit",
			"enter • search • q quit",
			"enter • search • q",
			"enter • q",
			"q",
		},
	},
}

func browseHelpLines(width int) []string {
	lines := make([]string, 0, len(browseHelpLineSpecs))
	for _, spec := range browseHelpLineSpecs {
		lines = append(lines, ui.SelectHintLine(width, spec.Options...))
	}
	return lines
}

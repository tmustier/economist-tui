package browse

import "github.com/tmustier/economist-tui/internal/ui"

type helpLineSpec struct {
	Options []string
}

var browseHelpLineSpecs = []helpLineSpec{
	{
		Options: []string{
			"↑/↓ navigate • ←/→ page • ⇧⇥/⇥ section",
			"↑/↓ move • ←/→ page • ⇧⇥/⇥ section",
			"↑/↓ • ←/→ • ⇧⇥/⇥",
			"↑/↓ • ⇧⇥/⇥",
			"↑/↓",
		},
	},
	{
		Options: []string{
			"↵ read • esc clear • q quit",
			"↵ read • esc • q quit",
			"↵ • esc • q",
			"↵ • q",
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

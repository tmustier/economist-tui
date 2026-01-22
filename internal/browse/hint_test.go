package browse

import (
	"testing"

	"github.com/muesli/reflow/ansi"
)

func TestBrowseHelpLinesPreferFullText(t *testing.T) {
	lines := browseHelpLines(120)
	if len(lines) != len(browseHelpLineSpecs) {
		t.Fatalf("expected %d help lines, got %d", len(browseHelpLineSpecs), len(lines))
	}
	if lines[0] != browseHelpLineSpecs[0].Options[0] {
		t.Fatalf("expected full line 1, got %q", lines[0])
	}
	if lines[1] != browseHelpLineSpecs[1].Options[0] {
		t.Fatalf("expected full line 2, got %q", lines[1])
	}
}

func TestBrowseHelpLinesFitWidth(t *testing.T) {
	widths := []int{80, 60, 45, 30, 20}
	for _, width := range widths {
		lines := browseHelpLines(width)
		if len(lines) != len(browseHelpLineSpecs) {
			t.Fatalf("expected %d help lines, got %d", len(browseHelpLineSpecs), len(lines))
		}
		for i, line := range lines {
			if ansi.PrintableRuneWidth(line) > width {
				t.Fatalf("line %d too wide for width %d: %q", i, width, line)
			}
		}
	}
}

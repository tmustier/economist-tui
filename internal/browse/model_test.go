package browse

import (
	"testing"

	"github.com/tmustier/economist-tui/internal/rss"
)

func TestResolveSectionIndexUsesPrimaryAlias(t *testing.T) {
	sections := rss.SectionList()
	index, updated := resolveSectionIndex("finance-and-economics", sections)

	if index < 0 || index >= len(updated) {
		t.Fatalf("expected index in range, got %d", index)
	}
	if updated[index].Primary != "finance" {
		t.Fatalf("expected updated index to point at finance, got %q", updated[index].Primary)
	}
}

func TestQueueSectionChangeUsesPendingIndex(t *testing.T) {
	m := Model{
		sections: []rss.SectionInfo{
			{Primary: "leaders"},
			{Primary: "briefing"},
			{Primary: "business"},
		},
		sectionIndex:        0,
		pendingSection:      "briefing",
		pendingSectionIndex: 1,
		sectionLoading:      true,
	}

	next, cmd := m.queueSectionChange(1)
	if cmd == nil {
		t.Fatalf("expected a command to be returned")
	}

	updated := next.(Model)
	if updated.pendingSection != "business" {
		t.Fatalf("expected pending section business, got %q", updated.pendingSection)
	}
	if updated.pendingSectionIndex != 2 {
		t.Fatalf("expected pending section index 2, got %d", updated.pendingSectionIndex)
	}
	if !updated.sectionLoading {
		t.Fatalf("expected sectionLoading to be true")
	}
}

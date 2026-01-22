package demo

import (
	"strings"
	"testing"
	"time"

	"github.com/tmustier/economist-tui/internal/rss"
)

func TestDemoSource(t *testing.T) {
	source := NewSource()
	title, items, err := source.Section("")
	if err != nil {
		t.Fatalf("section: %v", err)
	}
	if title == "" {
		t.Fatalf("expected title")
	}
	if len(items) == 0 {
		t.Fatalf("expected items")
	}

	if !itemsSortedByDate(items) {
		t.Fatalf("expected items sorted by date")
	}

	link := ""
	for _, item := range items {
		if strings.HasSuffix(item.Link, "/fair-exchange") {
			link = item.Link
			break
		}
	}
	if link == "" {
		t.Fatalf("expected fair exchange item")
	}

	art, err := source.Article(link)
	if err != nil {
		t.Fatalf("article: %v", err)
	}
	if art.Title == "" || art.Content == "" {
		t.Fatalf("expected article content")
	}
	if !strings.Contains(strings.ToLower(art.Content), "destroyers") {
		snippet := art.Content
		if len(snippet) > 80 {
			snippet = snippet[:80]
		}
		t.Fatalf("expected fixture content, got %q", snippet)
	}
	if !strings.HasSuffix(strings.TrimSpace(art.Content), "■") {
		t.Fatalf("expected trailing marker")
	}
}

func itemsSortedByDate(items []rss.Item) bool {
	if len(items) < 2 {
		return true
	}
	prev, err := time.Parse(time.RFC1123Z, items[0].PubDate)
	if err != nil {
		return false
	}
	for _, item := range items[1:] {
		current, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			return false
		}
		if current.After(prev) {
			return false
		}
		prev = current
	}
	return true
}

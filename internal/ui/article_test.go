package ui

import (
	"strings"
	"testing"

	"github.com/tmustier/economist-tui/internal/article"
)

func TestRenderArticleHeaderUsesFullWidthInColumns(t *testing.T) {
	title := strings.TrimSpace(strings.Repeat("word ", 14))
	art := &article.Article{Title: title}
	opts := ArticleRenderOptions{
		NoColor:   true,
		WrapWidth: 80,
		TwoColumn: true,
	}
	layout := resolveArticleLayout(opts)
	if !layout.UseColumns {
		t.Fatalf("expected columns to be enabled")
	}

	header := RenderArticleHeader(art, NewArticleStyles(true), opts)
	var titleLines []string
	for _, line := range strings.Split(header, "\n") {
		if strings.Contains(line, "word") {
			titleLines = append(titleLines, line)
		}
	}

	if len(titleLines) != 1 {
		t.Fatalf("expected single-line title, got %d lines: %q", len(titleLines), titleLines)
	}
}

func TestArticleFooterFormatting(t *testing.T) {
	styles := NewArticleStyles(false)
	art := &article.Article{URL: "https://example.com/test"}

	footer := ArticleFooter(art, styles, ArticleRenderOptions{WrapWidth: 60})
	if strings.Contains(footer, "🔗") {
		t.Fatalf("expected footer without emoji, got %q", footer)
	}
	if strings.Count(footer, art.URL) != 1 {
		t.Fatalf("expected single URL occurrence, got %q", footer)
	}
	if !strings.Contains(footer, styles.Body.Render(art.URL)) {
		t.Fatalf("expected styled URL, got %q", footer)
	}
}

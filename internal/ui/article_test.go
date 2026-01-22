package ui

import (
	"strings"
	"testing"

	"github.com/tmustier/economist-tui/internal/article"
)

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

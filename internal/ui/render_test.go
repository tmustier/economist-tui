package ui

import (
	"strings"
	"testing"

	"github.com/tmustier/economist-tui/internal/article"
)

func TestRenderArticleIndent(t *testing.T) {
	art := &article.Article{
		Overtitle: "Section",
		Title:     "Headline",
		Subtitle:  "Subhead",
		DateLine:  "Jan 1st 2024",
		Content:   "This is a paragraph that should wrap across multiple lines to verify indentation is consistent across wraps.",
		URL:       "https://example.com/test",
	}

	out, err := RenderArticle(art, ArticleRenderOptions{
		NoColor:   true,
		WrapWidth: 40,
	})
	if err != nil {
		t.Fatalf("render article: %v", err)
	}

	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	indent := strings.Repeat(" ", bodyIndent)
	for _, line := range lines {
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, indent) {
			t.Fatalf("expected indent %q, got %q", indent, line)
		}
	}
}

func TestReflowArticleBodyColumns(t *testing.T) {
	styles := NewArticleStyles(true)
	base := strings.Repeat("word ", 80)
	opts := ArticleRenderOptions{
		NoColor:   true,
		WrapWidth: 80,
		TwoColumn: true,
	}
	layout := resolveArticleLayout(opts)

	body := ReflowArticleBody(base, styles, opts)
	lines := strings.Split(strings.TrimRight(body, "\n"), "\n")
	indent := strings.Repeat(" ", layout.Indent)
	gap := strings.Repeat(" ", columnGap)

	for _, line := range lines {
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, indent) {
			t.Fatalf("expected indent %q, got %q", indent, line)
		}
		trimmed := strings.TrimPrefix(line, indent)
		if len(trimmed) > layout.ColumnWidth {
			if len(trimmed) < layout.ColumnWidth+columnGap {
				t.Fatalf("expected column gap, got %q", trimmed)
			}
			if trimmed[layout.ColumnWidth:layout.ColumnWidth+columnGap] != gap {
				t.Fatalf("expected column gap %q, got %q", gap, trimmed)
			}
		}
	}
}

func TestHighlightTrailingMarker(t *testing.T) {
	styles := NewArticleStyles(false)
	input := "Ends with marker ■"
	output := HighlightTrailingMarker(input, styles)

	styled := styles.Title.Render("■")
	if !strings.Contains(output, styled) {
		t.Fatalf("expected styled marker, got %q", output)
	}

	replaced := strings.Replace(input, "■", styled, 1)
	if output != replaced {
		t.Fatalf("expected marker replacement, got %q", output)
	}
}

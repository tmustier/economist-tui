package article

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	appErrors "github.com/tmustier/economist-tui/internal/errors"
)

func loadFixture(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join("testdata", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	return string(data)
}

func TestParseArticleExtractsFields(t *testing.T) {
	html := loadFixture(t, "basic.html")
	art, err := parseArticle(html, "https://example.com/test")
	if err != nil {
		t.Fatalf("parse article: %v", err)
	}

	if art.Overtitle != "Finance & economics" {
		t.Fatalf("expected overtitle, got %q", art.Overtitle)
	}
	if art.Title != "A test headline for the ages" {
		t.Fatalf("expected title, got %q", art.Title)
	}
	if art.Subtitle != "A subheadline that should be captured" {
		t.Fatalf("expected subtitle, got %q", art.Subtitle)
	}
	if art.DateLine != "Jan 1st 2024" {
		t.Fatalf("expected date line, got %q", art.DateLine)
	}
	if art.URL != "https://example.com/test" {
		t.Fatalf("expected url, got %q", art.URL)
	}

	if !strings.Contains(art.Content, "First paragraph") {
		t.Fatalf("expected first paragraph, got %q", art.Content)
	}
	if !strings.Contains(art.Content, "Second paragraph ends with a marker ■") {
		t.Fatalf("expected second paragraph, got %q", art.Content)
	}
	if strings.Contains(art.Content, "extra text that should be removed") {
		t.Fatalf("expected trailing text removed, got %q", art.Content)
	}
	if !strings.HasSuffix(strings.TrimSpace(art.Content), "■") {
		t.Fatalf("expected trailing marker, got %q", art.Content)
	}
}

func TestParseArticlePaywall(t *testing.T) {
	html := loadFixture(t, "paywall.html")
	_, err := parseArticle(html, "https://example.com/paywall")
	if err == nil {
		t.Fatalf("expected paywall error")
	}
	if !appErrors.IsPaywallError(err) {
		t.Fatalf("expected paywall error, got %v", err)
	}
}

package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/tmustier/economist-tui/internal/article"
)

func setTempHome(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	return tmp
}

func TestSaveLoadArticle(t *testing.T) {
	setTempHome(t)
	art := &article.Article{URL: "https://example.com/test", Title: "Title"}

	if err := SaveArticle(art); err != nil {
		t.Fatalf("save article: %v", err)
	}

	loaded, ok, err := LoadArticle(art.URL)
	if err != nil {
		t.Fatalf("load article: %v", err)
	}
	if !ok {
		t.Fatalf("expected cache hit")
	}
	if loaded.Title != art.Title {
		t.Fatalf("expected title %q, got %q", art.Title, loaded.Title)
	}
}

func TestLoadArticleExpires(t *testing.T) {
	setTempHome(t)
	url := "https://example.com/expired"
	entry := articleEntry{
		CachedAt: time.Now().Add(-2 * articleTTL),
		Article:  article.Article{URL: url},
	}

	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	path := articleCachePath(url)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatalf("write: %v", err)
	}

	_, ok, err := LoadArticle(url)
	if err != nil {
		t.Fatalf("load article: %v", err)
	}
	if ok {
		t.Fatalf("expected cache miss for expired entry")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected expired cache file removed")
	}
}

func TestPurgeExpired(t *testing.T) {
	setTempHome(t)
	freshURL := "https://example.com/fresh"
	expiredURL := "https://example.com/stale"

	writeEntry := func(url string, cachedAt time.Time) string {
		entry := articleEntry{CachedAt: cachedAt, Article: article.Article{URL: url}}
		data, err := json.Marshal(entry)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		path := articleCachePath(url)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(path, data, 0600); err != nil {
			t.Fatalf("write: %v", err)
		}
		return path
	}

	freshPath := writeEntry(freshURL, time.Now().Add(-10*time.Minute))
	expiredPath := writeEntry(expiredURL, time.Now().Add(-2*articleTTL))

	if err := PurgeExpired(); err != nil {
		t.Fatalf("purge: %v", err)
	}

	if _, err := os.Stat(expiredPath); !os.IsNotExist(err) {
		t.Fatalf("expected expired cache removed")
	}
	if _, err := os.Stat(freshPath); err != nil {
		t.Fatalf("expected fresh cache retained: %v", err)
	}
}

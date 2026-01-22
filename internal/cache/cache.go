package cache

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/tmustier/economist-tui/internal/article"
	"github.com/tmustier/economist-tui/internal/config"
)

const articleTTL = time.Hour

const cacheDirName = "cache"

type articleEntry struct {
	CachedAt time.Time       `json:"cached_at"`
	Article  article.Article `json:"article"`
}

func LoadArticle(url string) (*article.Article, bool, error) {
	path := articleCachePath(url)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, err
	}

	var entry articleEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, false, err
	}

	if time.Since(entry.CachedAt) > articleTTL {
		_ = os.Remove(path)
		return nil, false, nil
	}

	return &entry.Article, true, nil
}

func SaveArticle(art *article.Article) error {
	if err := os.MkdirAll(cacheDir(), 0755); err != nil {
		return err
	}

	entry := articleEntry{
		CachedAt: time.Now().UTC(),
		Article:  *art,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	path := articleCachePath(art.URL)
	return os.WriteFile(path, data, 0600)
}

func PurgeExpired() error {
	dir := cacheDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var cached articleEntry
		if err := json.Unmarshal(data, &cached); err != nil {
			_ = os.Remove(path)
			continue
		}
		if time.Since(cached.CachedAt) > articleTTL {
			_ = os.Remove(path)
		}
	}

	return nil
}

func cacheDir() string {
	return filepath.Join(config.ConfigDir(), cacheDirName)
}

func articleCachePath(url string) string {
	h := sha1.Sum([]byte(url))
	name := hex.EncodeToString(h[:]) + ".json"
	return filepath.Join(cacheDir(), name)
}

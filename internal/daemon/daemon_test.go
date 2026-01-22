package daemon

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	appErrors "github.com/tmustier/economist-tui/internal/errors"
)

func withTestDaemon(t *testing.T, handler http.Handler) {
	t.Helper()
	home, err := os.MkdirTemp("/tmp", "economist-daemon-")
	if err != nil {
		t.Fatalf("tempdir: %v", err)
	}
	t.Setenv("HOME", home)

	socketPath := SocketPath()
	if err := os.MkdirAll(filepath.Dir(socketPath), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	ln, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	srv := &http.Server{Handler: handler}
	go func() {
		_ = srv.Serve(ln)
	}()

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
		_ = ln.Close()
		_ = os.Remove(socketPath)
		_ = os.RemoveAll(home)
	})
}

func TestFetchMapsPayload(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/fetch", func(w http.ResponseWriter, r *http.Request) {
		resp := FetchResponse{Article: &ArticlePayload{
			Overtitle: "Section",
			Title:     "Title",
			Subtitle:  "Subtitle",
			DateLine:  "Jan 1st 2024",
			Content:   "Body",
			URL:       "https://example.com/test",
		}}
		_ = json.NewEncoder(w).Encode(resp)
	})
	withTestDaemon(t, mux)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	art, err := Fetch(ctx, "https://example.com/test", false)
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if art.Overtitle != "Section" || art.Title != "Title" || art.Subtitle != "Subtitle" {
		t.Fatalf("unexpected article fields: %#v", art)
	}
	if art.URL != "https://example.com/test" {
		t.Fatalf("expected url, got %q", art.URL)
	}
}

func TestFetchMapsErrors(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/fetch", func(w http.ResponseWriter, r *http.Request) {
		resp := FetchResponse{Error: "paywall", ErrorType: "paywall"}
		_ = json.NewEncoder(w).Encode(resp)
	})
	withTestDaemon(t, mux)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := Fetch(ctx, "https://example.com/paywall", false)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !appErrors.IsPaywallError(err) {
		t.Fatalf("expected paywall error, got %v", err)
	}
}

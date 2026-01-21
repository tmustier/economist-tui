package daemon

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/tmustier/economist-cli/internal/article"
	"github.com/tmustier/economist-cli/internal/config"
	appErrors "github.com/tmustier/economist-cli/internal/errors"
)

const (
	socketName = "serve.sock"
	logName    = "serve.log"
)

var ErrNotRunning = errors.New("economist serve not running")

func SocketPath() string {
	return filepath.Join(config.ConfigDir(), socketName)
}

func LogPath() string {
	return filepath.Join(config.ConfigDir(), logName)
}

type FetchRequest struct {
	URL   string `json:"url"`
	Debug bool   `json:"debug"`
}

type FetchResponse struct {
	Article   *ArticlePayload `json:"article,omitempty"`
	Error     string          `json:"error,omitempty"`
	ErrorType string          `json:"error_type,omitempty"`
}

type ArticlePayload struct {
	Title         string `json:"title"`
	Subtitle      string `json:"subtitle,omitempty"`
	DateLine      string `json:"date_line,omitempty"`
	Content       string `json:"content,omitempty"`
	URL           string `json:"url"`
	DebugHTMLPath string `json:"debug_html_path,omitempty"`
}

func IsRunning() bool {
	_, err := ping(context.Background())
	return err == nil
}

func Status(ctx context.Context) (time.Duration, bool, error) {
	dur, err := ping(ctx)
	if err == nil {
		return dur, true, nil
	}
	if errors.Is(err, ErrNotRunning) {
		return 0, false, nil
	}
	return 0, false, err
}

func Shutdown(ctx context.Context) error {
	client, err := newClient()
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://unix/shutdown", nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		if isConnRefused(err) {
			return ErrNotRunning
		}
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("daemon HTTP %d", resp.StatusCode)
	}

	return nil
}

func EnsureBackground() error {
	if IsRunning() {
		return nil
	}
	return StartBackground()
}

func StartBackground() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(config.ConfigDir(), 0755); err != nil {
		return err
	}

	logFile, err := os.OpenFile(LogPath(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		logFile, _ = os.OpenFile(os.DevNull, os.O_WRONLY, 0644)
	}

	cmd := exec.Command(exe, "serve")
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.Stdin = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	return cmd.Start()
}

func Fetch(ctx context.Context, url string, debug bool) (*article.Article, error) {
	client, err := newClient()
	if err != nil {
		return nil, err
	}

	reqBody, err := json.Marshal(FetchRequest{URL: url, Debug: debug})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://unix/fetch", bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) || errors.Is(err, syscall.ENOENT) || errors.Is(err, net.ErrClosed) {
			return nil, ErrNotRunning
		}
		if isConnRefused(err) {
			return nil, ErrNotRunning
		}
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("daemon HTTP %d", resp.StatusCode)
	}

	var payload FetchResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	if payload.Error != "" {
		switch payload.ErrorType {
		case "paywall":
			return nil, appErrors.PaywallError{}
		case "user":
			return nil, appErrors.NewUserError(payload.Error)
		default:
			return nil, fmt.Errorf(payload.Error)
		}
	}
	if payload.Article == nil {
		return nil, fmt.Errorf("daemon returned empty response")
	}

	art := &article.Article{
		Title:         payload.Article.Title,
		Subtitle:      payload.Article.Subtitle,
		DateLine:      payload.Article.DateLine,
		Content:       payload.Article.Content,
		URL:           payload.Article.URL,
		DebugHTMLPath: payload.Article.DebugHTMLPath,
	}
	return art, nil
}

func Serve() error {
	if err := os.MkdirAll(config.ConfigDir(), 0755); err != nil {
		return err
	}

	socketPath := SocketPath()
	if _, err := os.Stat(socketPath); err == nil {
		_ = os.Remove(socketPath)
	}

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = listener.Close()
		_ = os.Remove(socketPath)
	}()

	_ = os.Chmod(socketPath, 0600)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	var server *http.Server
	mux.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		if server != nil {
			go func() {
				_ = server.Shutdown(context.Background())
			}()
		}
	})

	var fetchMu sync.Mutex
	mux.HandleFunc("/fetch", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var req FetchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		fetchMu.Lock()
		defer fetchMu.Unlock()

		art, err := article.Fetch(req.URL, article.FetchOptions{Debug: req.Debug})
		resp := FetchResponse{}
		if err != nil {
			resp.Error = err.Error()
			if appErrors.IsPaywallError(err) {
				resp.ErrorType = "paywall"
			} else if appErrors.IsUserError(err) {
				resp.ErrorType = "user"
			}
		} else {
			resp.Article = &ArticlePayload{
				Title:         art.Title,
				Subtitle:      art.Subtitle,
				DateLine:      art.DateLine,
				Content:       art.Content,
				URL:           art.URL,
				DebugHTMLPath: art.DebugHTMLPath,
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})

	server = &http.Server{
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 5 * time.Minute,
	}

	return server.Serve(listener)
}

func ping(ctx context.Context) (time.Duration, error) {
	client, err := newClient()
	if err != nil {
		return 0, err
	}

	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://unix/health", nil)
	if err != nil {
		return 0, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("daemon HTTP %d", resp.StatusCode)
	}
	return time.Since(start), nil
}

func newClient() (*http.Client, error) {
	socketPath := SocketPath()
	if _, err := os.Stat(socketPath); err != nil {
		return nil, ErrNotRunning
	}

	transport := &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, "unix", socketPath)
		},
	}

	return &http.Client{
		Transport: transport,
		Timeout:   2 * time.Minute,
	}, nil
}

func isConnRefused(err error) bool {
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		if errors.Is(opErr.Err, syscall.ECONNREFUSED) {
			return true
		}
	}
	return false
}

package browser

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/tmustier/economist-tui/internal/config"
)

const (
	UserAgent    = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	LoginURL     = "https://www.economist.com/api/auth/login"
	LoginTimeout = 5 * time.Minute
	FetchTimeout = 45 * time.Second
)

var (
	sharedMu     sync.Mutex
	sharedCtx    context.Context
	sharedCancel context.CancelFunc
	sharedDebug  bool
)

func headlessExecAllocatorOptions(debug bool) []chromedp.ExecAllocatorOption {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("blink-settings", "imagesEnabled=false"),
		chromedp.UserAgent(UserAgent),
	)
	if !debug {
		opts = append(opts,
			chromedp.Flag("disable-logging", true),
			chromedp.Flag("log-level", "3"),
		)
	}
	return opts
}

func newHeadlessContext(ctx context.Context, debug bool) (context.Context, context.CancelFunc) {
	opts := headlessExecAllocatorOptions(debug)
	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, opts...)
	logf := func(string, ...interface{}) {}
	if debug {
		logf = log.Printf
	}
	browserCtx, browserCancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(logf))

	cancel := func() {
		browserCancel()
		allocCancel()
	}

	return browserCtx, cancel
}

// HeadlessContext creates a headless browser context for fetching pages.
func HeadlessContext(ctx context.Context, debug bool) (context.Context, context.CancelFunc) {
	return newHeadlessContext(ctx, debug)
}

// SharedHeadlessContext returns a shared headless browser context for this process.
func SharedHeadlessContext(debug bool) context.Context {
	sharedMu.Lock()
	defer sharedMu.Unlock()

	if sharedCtx == nil || sharedDebug != debug {
		if sharedCancel != nil {
			sharedCancel()
		}
		sharedCtx, sharedCancel = newHeadlessContext(context.Background(), debug)
		sharedDebug = debug
	}

	return sharedCtx
}

// CloseSharedHeadless closes the shared browser, if any.
func CloseSharedHeadless() {
	sharedMu.Lock()
	defer sharedMu.Unlock()

	if sharedCancel != nil {
		sharedCancel()
		sharedCancel = nil
		sharedCtx = nil
	}
}

// VisibleContext creates a visible browser context for interactive login.
func VisibleContext(ctx context.Context, userDataDir string) (context.Context, context.CancelFunc) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-gpu", false),
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("disable-extensions", false),
		chromedp.UserDataDir(userDataDir),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, opts...)
	browserCtx, browserCancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(func(string, ...interface{}) {}))

	cancel := func() {
		browserCancel()
		allocCancel()
	}

	return browserCtx, cancel
}

// InjectCookies sets cookies from config into the browser context.
func InjectCookies(ctx context.Context, cookies []config.Cookie) error {
	if len(cookies) == 0 {
		return nil
	}

	var params []*network.CookieParam
	for _, c := range cookies {
		params = append(params, &network.CookieParam{
			Name:   c.Name,
			Value:  c.Value,
			Domain: c.Domain,
			Path:   c.Path,
		})
	}

	return chromedp.Run(ctx, network.SetCookies(params))
}

// ExtractCookies gets Economist cookies from the browser context.
func ExtractCookies(ctx context.Context) ([]config.Cookie, error) {
	var cookies []config.Cookie

	var networkCookies []*network.Cookie
	if err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		networkCookies, err = network.GetCookies().Do(ctx)
		return err
	})); err != nil {
		return nil, err
	}

	for _, c := range networkCookies {
		if isEconomistDomain(c.Domain) {
			cookies = append(cookies, config.Cookie{
				Name:   c.Name,
				Value:  c.Value,
				Domain: c.Domain,
				Path:   c.Path,
			})
		}
	}

	return cookies, nil
}

func isEconomistDomain(domain string) bool {
	return domain == ".economist.com" || domain == "www.economist.com" || domain == "economist.com"
}

// IsAuthCookie checks if a cookie indicates successful authentication.
func IsAuthCookie(name string) bool {
	authCookies := []string{
		"ec_permissions",
		"ec_subscriber",
		"SPC",
		"Authorization",
		"economist_session",
		"user_id",
	}
	for _, ac := range authCookies {
		if name == ac {
			return true
		}
	}
	return false
}

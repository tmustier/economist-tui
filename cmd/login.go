package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/spf13/cobra"
	"github.com/tmustier/economist-tui/internal/browser"
	"github.com/tmustier/economist-tui/internal/config"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to The Economist (opens browser)",
	Long: `Opens a browser window for you to log in to The Economist.
Once logged in, cookies are saved locally for future use.

The browser uses a persistent profile, so you only need to login once.`,
	RunE: runLoginCmd,
}

func runLoginCmd(cmd *cobra.Command, args []string) error {
	fmt.Println("🌐 Opening browser for login...")
	fmt.Println("   Please log in to your Economist account.")
	fmt.Println("   The browser will close when login is detected.")
	fmt.Println("")

	if err := runLogin(); err != nil {
		return err
	}

	fmt.Println("✅ Login successful! Cookies saved.")
	fmt.Println("   You can now use 'economist read <url>' to read articles.")
	return nil
}

func runLogin() error {
	userDataDir := config.BrowserDataDir()
	if err := os.MkdirAll(userDataDir, 0755); err != nil {
		return fmt.Errorf("failed to create browser data dir: %w", err)
	}

	ctx, cancel := browser.VisibleContext(context.Background(), userDataDir)
	defer cancel()

	ctx, cancelTimeout := context.WithTimeout(ctx, browser.LoginTimeout)
	defer cancelTimeout()

	if err := chromedp.Run(ctx, chromedp.Navigate(browser.LoginURL)); err != nil {
		return fmt.Errorf("failed to open login page: %w", err)
	}

	fmt.Println("⏳ Waiting for login (checking every 3s)...")
	fmt.Println("   (Close the browser manually if you need to cancel)")

	return waitForLogin(ctx)
}

func waitForLogin(ctx context.Context) error {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				return fmt.Errorf("login timed out after %v", browser.LoginTimeout)
			}
			return ctx.Err()
		case <-ticker.C:
			if done, err := checkLoginStatus(ctx); err != nil {
				continue // Non-fatal, keep polling
			} else if done {
				return nil
			}
		}
	}
}

func checkLoginStatus(ctx context.Context) (bool, error) {
	cookies, err := browser.ExtractCookies(ctx)
	if err != nil {
		return false, err
	}

	// Check for auth cookies
	for _, c := range cookies {
		if browser.IsAuthCookie(c.Name) {
			return saveCookies(cookies)
		}
	}

	// Check if user navigated away from login page
	var currentURL string
	if err := chromedp.Run(ctx, chromedp.Location(&currentURL)); err != nil {
		return false, err
	}

	isOnLoginPage := strings.Contains(currentURL, "/auth/") || strings.Contains(currentURL, "/login")
	if len(cookies) > 5 && !isOnLoginPage {
		fmt.Printf("   Detected navigation to: %s\n", currentURL)
		return saveCookies(cookies)
	}

	return false, nil
}

func saveCookies(cookies []config.Cookie) (bool, error) {
	cfg := &config.Config{Cookies: cookies}
	if err := cfg.Save(); err != nil {
		return false, fmt.Errorf("failed to save cookies: %w", err)
	}
	return true, nil
}

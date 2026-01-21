package article

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/tmustier/economist-cli/internal/browser"
	"github.com/tmustier/economist-cli/internal/config"
	appErrors "github.com/tmustier/economist-cli/internal/errors"
)

// Content length thresholds
const (
	minParagraphLen      = 40  // Skip very short paragraphs
	minContentLen        = 500 // Minimum content to consider article "loaded"
	articleReadySelector = "article h1, h1.article__headline, [data-test-id='headline'], article"
	articleWaitTimeout   = 8 * time.Second
)

var blockedURLPatterns = []string{
	"*.png",
	"*.jpg",
	"*.jpeg",
	"*.gif",
	"*.webp",
	"*.svg",
	"*.woff",
	"*.woff2",
	"*.ttf",
	"*.otf",
	"*.mp4",
	"*.mp3",
	"*.m4a",
	"*.mov",
	"*.avi",
	"*.m3u8",
	"*doubleclick.net/*",
	"*googletagmanager.com/*",
	"*google-analytics.com/*",
	"*googlesyndication.com/*",
	"*adservice.google.com/*",
	"*adsystem.com/*",
	"*adsrvr.org/*",
	"*scorecardresearch.com/*",
	"*criteo.com/*",
	"*taboola.com/*",
	"*outbrain.com/*",
}

type Article struct {
	Title         string
	Subtitle      string
	DateLine      string
	Content       string
	URL           string
	DebugHTMLPath string
}

type FetchOptions struct {
	Debug bool
}

func Fetch(articleURL string, opts FetchOptions) (*Article, error) {
	start := time.Now()
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	baseCtx := browser.SharedHeadlessContext(opts.Debug)
	ctx, cancel := chromedp.NewContext(baseCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, browser.FetchTimeout)
	defer cancel()

	debugf(opts.Debug, "context ready in %s", time.Since(start))

	// Inject saved cookies (ignore errors - will just hit paywall)
	_ = browser.InjectCookies(ctx, cfg.Cookies)

	if err := configureNetwork(ctx, opts.Debug); err != nil {
		debugf(opts.Debug, "network blocking error: %v", err)
	}

	navStart := time.Now()
	debugf(opts.Debug, "navigate start")
	err = chromedp.Run(ctx,
		navigateNoWait(articleURL),
		debugStep(opts.Debug, "navigate issued"),
		chromedp.WaitReady("body", chromedp.ByQuery),
		debugStep(opts.Debug, "body ready"),
		waitForArticleSelector(articleWaitTimeout),
		debugStep(opts.Debug, "article selector checked"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load page: %w", err)
	}

	html, err := captureHTML(ctx, opts.Debug)
	if err != nil {
		return nil, fmt.Errorf("failed to capture html: %w", err)
	}
	debugf(opts.Debug, "page loaded in %s", time.Since(navStart))

	parseStart := time.Now()
	art, parseErr := parseArticle(html, articleURL)
	debugf(opts.Debug, "parsed in %s", time.Since(parseStart))

	if opts.Debug {
		if path, err := writeDebugHTML(html); err == nil {
			if art == nil {
				art = &Article{URL: articleURL}
			}
			art.DebugHTMLPath = path
		}
		debugf(opts.Debug, "total fetch time %s", time.Since(start))
	}
	if parseErr != nil {
		return art, parseErr
	}

	return art, nil
}

func parseArticle(html, articleURL string) (*Article, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	article := &Article{URL: articleURL}
	article.Title = findFirst(doc, "h1.article__headline", "[data-test-id='headline']", "article h1", "h1")
	article.Subtitle = findFirst(doc, ".article__description", "[data-test-id='subheadline']", ".article__subheadline")
	article.DateLine = strings.TrimSpace(doc.Find("time").First().Text())
	article.Content = extractContent(doc)

	if err := checkPaywall(html, article.Content); err != nil {
		return article, err
	}

	return article, nil
}

func findFirst(doc *goquery.Document, selectors ...string) string {
	for _, sel := range selectors {
		if text := strings.TrimSpace(doc.Find(sel).First().Text()); text != "" {
			return text
		}
	}
	return ""
}

func extractContent(doc *goquery.Document) string {
	var paragraphs []string

	// Primary selectors for article body
	doc.Find(".article__body-text p, [data-component='article-body'] p").Each(func(i int, s *goquery.Selection) {
		if isInsideRelatedSection(s) {
			return
		}
		if text := cleanParagraph(s); text != "" {
			paragraphs = append(paragraphs, text)
		}
	})

	// Fallback to broader selectors
	if len(paragraphs) == 0 {
		doc.Find("article p, main p").Each(func(i int, s *goquery.Selection) {
			if text := cleanParagraph(s); text != "" && !looksLikeTeaser(text) {
				paragraphs = append(paragraphs, text)
			}
		})
	}

	content := strings.Join(paragraphs, "\n\n")
	content = strings.TrimSpace(content)
	return trimTrailingMarker(content)
}

func isInsideRelatedSection(s *goquery.Selection) bool {
	return s.ParentsFiltered("[class*='related'], [class*='teaser'], [class*='promo']").Length() > 0
}

func cleanParagraph(s *goquery.Selection) string {
	text := strings.TrimSpace(s.Text())
	if len(text) < minParagraphLen || isBoilerplate(text) {
		return ""
	}
	return text
}

// looksLikeTeaser detects short promotional text that isn't article content.
func looksLikeTeaser(text string) bool {
	// Real article paragraphs are typically longer
	return len(text) < 80
}

var boilerplatePatterns = []string{
	"subscribe",
	"sign up",
	"newsletter",
	"keep reading",
	"this article appeared",
	"reuse this content",
	"more from",
	"advertisement",
	"listen to this story",
	"enjoy more audio",
}

func isBoilerplate(text string) bool {
	lower := strings.ToLower(text)
	for _, pattern := range boilerplatePatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}
	return false
}

func trimTrailingMarker(content string) string {
	const marker = "■"
	if !strings.Contains(content, marker) {
		return content
	}

	paragraphs := strings.Split(content, "\n\n")
	if len(paragraphs) == 0 {
		return content
	}

	start := len(paragraphs) - 3
	if start < 0 {
		start = 0
	}

	for i := len(paragraphs) - 1; i >= start; i-- {
		idx := strings.LastIndex(paragraphs[i], marker)
		if idx == -1 {
			continue
		}

		paragraphs[i] = strings.TrimSpace(paragraphs[i][:idx+len(marker)])
		return strings.TrimSpace(strings.Join(paragraphs[:i+1], "\n\n"))
	}

	return content
}

var paywallIndicators = []string{
	"Subscribe to read",
	"Keep reading with a subscription",
	"This article is for subscribers",
	"Sign in to continue",
}

func checkPaywall(html, content string) error {
	for _, indicator := range paywallIndicators {
		if strings.Contains(html, indicator) && len(content) < minContentLen {
			return appErrors.PaywallError{}
		}
	}
	return nil
}

func writeDebugHTML(html string) (string, error) {
	file, err := os.CreateTemp("", "economist-article-*.html")
	if err != nil {
		return "", err
	}
	defer file.Close()

	if _, err := file.WriteString(html); err != nil {
		return "", err
	}

	return file.Name(), nil
}

func waitForArticleSelector(timeout time.Duration) chromedp.ActionFunc {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		waitCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		_ = chromedp.Run(waitCtx, chromedp.WaitVisible(articleReadySelector, chromedp.ByQuery))
		return nil
	})
}

func navigateNoWait(url string) chromedp.ActionFunc {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		_, _, _, _, err := page.Navigate(url).Do(ctx)
		return err
	})
}

func configureNetwork(ctx context.Context, debug bool) error {
	if len(blockedURLPatterns) == 0 {
		return nil
	}
	if err := chromedp.Run(ctx, network.Enable()); err != nil {
		return err
	}
	if err := chromedp.Run(ctx, network.SetBlockedURLs(blockedURLPatterns)); err != nil {
		return err
	}
	debugf(debug, "network blocking enabled (%d patterns)", len(blockedURLPatterns))
	return nil
}

func captureHTML(ctx context.Context, debug bool) (string, error) {
	selectors := []string{"article", "main", "body", "html"}
	for _, sel := range selectors {
		start := time.Now()
		html, err := captureOuterHTML(ctx, sel, 3*time.Second)
		if err == nil && strings.TrimSpace(html) != "" {
			debugf(debug, "html captured from %s in %s", sel, time.Since(start))
			return html, nil
		}
		debugf(debug, "html capture failed (%s): %v", sel, err)
	}
	return "", fmt.Errorf("no html captured")
}

func captureOuterHTML(ctx context.Context, selector string, timeout time.Duration) (string, error) {
	var html string
	capCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	err := chromedp.Run(capCtx, chromedp.OuterHTML(selector, &html, chromedp.ByQuery))
	return html, err
}

func debugStep(enabled bool, message string) chromedp.ActionFunc {
	return chromedp.ActionFunc(func(context.Context) error {
		debugf(enabled, message)
		return nil
	})
}

func debugf(enabled bool, format string, args ...any) {
	if !enabled {
		return
	}
	ts := time.Now().Format(time.RFC3339Nano)
	fmt.Fprintf(os.Stderr, "debug %s "+format+"\n", append([]any{ts}, args...)...)
}

func (a *Article) ToMarkdown() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s\n\n", a.Title))

	if a.Subtitle != "" {
		sb.WriteString(fmt.Sprintf("*%s*\n\n", a.Subtitle))
	}

	if a.DateLine != "" {
		sb.WriteString(fmt.Sprintf("%s\n\n", a.DateLine))
	}

	sb.WriteString("---\n\n")
	sb.WriteString(a.Content)
	sb.WriteString("\n\n---\n")
	sb.WriteString(fmt.Sprintf("🔗 %s\n", a.URL))

	return sb.String()
}

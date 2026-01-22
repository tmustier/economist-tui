package rss

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/tmustier/economist-cli/internal/browser"
	"github.com/tmustier/economist-cli/internal/search"
)

const httpTimeout = 10 * time.Second

// Sections maps aliases to canonical RSS feed paths.
var Sections = map[string]string{
	"leaders":               "leaders",
	"briefing":              "briefing",
	"finance":               "finance-and-economics",
	"finance-and-economics": "finance-and-economics",
	"us":                    "united-states",
	"united-states":         "united-states",
	"britain":               "britain",
	"europe":                "europe",
	"middle-east":           "middle-east-and-africa",
	"asia":                  "asia",
	"china":                 "china",
	"americas":              "the-americas",
	"business":              "business",
	"science":               "science-and-technology",
	"tech":                  "science-and-technology",
	"culture":               "culture",
	"graphic":               "graphic-detail",
	"world-this-week":       "the-world-this-week",
}

type RSS struct {
	Channel Channel `xml:"channel"`
}

type Channel struct {
	Title   string `xml:"title"`
	Link    string `xml:"link"`
	PubDate string `xml:"pubDate"`
	Items   []Item `xml:"item"`
}

type Item struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Link        string `xml:"link"`
	GUID        string `xml:"guid"`
	PubDate     string `xml:"pubDate"`
}

func (i Item) CleanTitle() string {
	return strings.TrimSpace(i.Title)
}

func (i Item) CleanDescription() string {
	return strings.TrimSpace(i.Description)
}

func (i Item) FormattedDate() string {
	if t, ok := parsePubDate(i.PubDate); ok {
		return formatOrdinalDate(t)
	}
	return strings.TrimSpace(i.PubDate)
}

func (i Item) CompactDate() string {
	if t, ok := parsePubDate(i.PubDate); ok {
		return t.Format("02.01.06")
	}
	return strings.TrimSpace(i.PubDate)
}

func parsePubDate(pubDate string) (time.Time, bool) {
	formats := []string{
		time.RFC1123Z,
		"Mon, 02 Jan 2006 15:04:05 +0000",
	}
	for _, format := range formats {
		if t, err := time.Parse(format, pubDate); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

func formatOrdinalDate(t time.Time) string {
	day := t.Day()
	suffix := ordinalSuffix(day)
	return fmt.Sprintf("%s %d%s %d", t.Format("Jan"), day, suffix, t.Year())
}

func ordinalSuffix(day int) string {
	if day%100 >= 11 && day%100 <= 13 {
		return "th"
	}
	switch day % 10 {
	case 1:
		return "st"
	case 2:
		return "nd"
	case 3:
		return "rd"
	default:
		return "th"
	}
}

func FetchSection(section string) (*RSS, error) {
	sectionPath := resolveSection(section)
	url := fmt.Sprintf("https://www.economist.com/%s/rss.xml", sectionPath)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", browser.UserAgent)

	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d from %s", resp.StatusCode, url)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rss RSS
	if err := xml.Unmarshal(body, &rss); err != nil {
		return nil, err
	}

	return &rss, nil
}

func Search(section, query string) ([]Item, error) {
	rss, err := FetchSection(section)
	if err != nil {
		return nil, err
	}

	var results []Item

	for _, item := range rss.Channel.Items {
		if matchesQuery(item, query) {
			results = append(results, item)
		}
	}

	return results, nil
}

func resolveSection(section string) string {
	if path, ok := Sections[strings.ToLower(section)]; ok {
		return path
	}
	return section
}

func matchesQuery(item Item, query string) bool {
	text := item.CleanTitle() + " " + item.CleanDescription()
	return search.Match(text, query)
}

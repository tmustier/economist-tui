package demo

import (
	_ "embed"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/tmustier/economist-tui/internal/article"
	"github.com/tmustier/economist-tui/internal/rss"
)

//go:embed fixtures/fair-exchange.txt
var fairExchangeFixture string

//go:embed fixtures/german-europe.txt
var germanEuropeFixture string

//go:embed fixtures/june-the-sixth.txt
var juneTheSixthFixture string

//go:embed fixtures/atom-bomb.txt
var atomBombFixture string

//go:embed fixtures/burnt-fingers-public-pulse.txt
var burntFingersPublicPulseFixture string

//go:embed fixtures/electronic-abacus.txt
var electronicAbacusFixture string

//go:embed fixtures/president-to-the-rescue.txt
var presidentToTheRescueFixture string

//go:embed fixtures/freeing-electronics.txt
var freeingElectronicsFixture string

//go:embed fixtures/both-sides-moon.txt
var bothSidesMoonFixture string

//go:embed fixtures/young-mans-america.txt
var youngMansAmericaFixture string

const DefaultSection = "leaders"

const demoSectionTitle = "Leaders - demo"

var demoBaseDate = time.Date(2026, time.January, 22, 9, 0, 0, 0, time.UTC)
var demoArchiveDate = time.Date(1940, time.September, 7, 9, 0, 0, 0, time.UTC)
var demoDdayDate = time.Date(1944, time.June, 10, 9, 0, 0, 0, time.UTC)
var demoAtomBombDate = time.Date(1945, time.August, 11, 9, 0, 0, 0, time.UTC)
var demoUnivacDate = time.Date(1952, time.November, 22, 9, 0, 0, 0, time.UTC)
var demoElectronicAbacusDate = time.Date(1954, time.March, 13, 9, 0, 0, 0, time.UTC)
var demoPresidentRescueDate = time.Date(1954, time.October, 30, 9, 0, 0, 0, time.UTC)
var demoFreeingElectronicsDate = time.Date(1956, time.February, 4, 9, 0, 0, 0, time.UTC)
var demoSputnikDate = time.Date(1957, time.October, 12, 9, 0, 0, 0, time.UTC)
var demoKennedyElectionDate = time.Date(1960, time.November, 12, 9, 0, 0, 0, time.UTC)

type Source struct {
	sections map[string]sectionData
	articles map[string]*article.Article
}

type sectionData struct {
	title string
	items []rss.Item
}

type demoArticle struct {
	slug     string
	title    string
	subtitle string
	content  string
	date     time.Time
}

func NewSource() *Source {
	source := &Source{
		sections: make(map[string]sectionData),
		articles: make(map[string]*article.Article),
	}
	source.addLeaders()
	return source
}

func (s *Source) Section(section string) (string, []rss.Item, error) {
	if section == "" {
		section = DefaultSection
	}
	key := strings.ToLower(section)
	if data, ok := s.sections[key]; ok {
		return data.title, data.items, nil
	}
	if data, ok := s.sections[DefaultSection]; ok {
		return data.title, data.items, nil
	}
	return "", nil, fmt.Errorf("demo section not found")
}

func (s *Source) Article(url string) (*article.Article, error) {
	if art, ok := s.articles[url]; ok {
		copy := *art
		return &copy, nil
	}
	return nil, fmt.Errorf("demo article not found")
}

func (s *Source) addLeaders() {
	articles := []demoArticle{
		{
			slug:     "fair-exchange",
			title:    "Fair Exchange",
			subtitle: "Destroyers for bases, and a new alliance",
			content:  strings.TrimSpace(fairExchangeFixture),
			date:     demoArchiveDate,
		},
		{
			slug:     "german-europe",
			title:    "German Europe",
			subtitle: "Conquest and the limits of domination",
			content:  strings.TrimSpace(germanEuropeFixture),
			date:     demoArchiveDate,
		},
		{
			slug:     "june-the-sixth",
			title:    "June the Sixth",
			subtitle: "Dunkirk’s reversal and the duty of liberation",
			content:  strings.TrimSpace(juneTheSixthFixture),
			date:     demoDdayDate,
		},
		{
			slug:     "atom-bomb",
			title:    "The atom bomb",
			subtitle: "A triumph of science and the terror of war",
			content:  strings.TrimSpace(atomBombFixture),
			date:     demoAtomBombDate,
		},
		{
			slug:     "burnt-fingers",
			title:    "Burnt Fingers on the Public Pulse",
			subtitle: "Polls, Univac, and the election night surprise",
			content:  strings.TrimSpace(burntFingersPublicPulseFixture),
			date:     demoUnivacDate,
		},
		{
			slug:     "electronic-abacus",
			title:    "Electronic Abacus",
			subtitle: "Lyons, Leo, and the electronic office",
			content:  strings.TrimSpace(electronicAbacusFixture),
			date:     demoElectronicAbacusDate,
		},
		{
			slug:     "president-to-the-rescue",
			title:    "President to the Rescue?",
			subtitle: "Unemployment figures and mid-term politics",
			content:  strings.TrimSpace(presidentToTheRescueFixture),
			date:     demoPresidentRescueDate,
		},
		{
			slug:     "freeing-electronics",
			title:    "Freeing Electronics",
			subtitle: "AT&T, IBM, and the $40,000-a-month brains",
			content:  strings.TrimSpace(freeingElectronicsFixture),
			date:     demoFreeingElectronicsDate,
		},
		{
			slug:     "both-sides-moon",
			title:    "Both Sides of the “Moon”",
			subtitle: "The little ball in the sky and the man on the ground",
			content:  strings.TrimSpace(bothSidesMoonFixture),
			date:     demoSputnikDate,
		},
		{
			slug:     "young-mans-america",
			title:    "Young Man's America?",
			subtitle: "Kennedy's narrow victory and a new mandate",
			content:  strings.TrimSpace(youngMansAmericaFixture),
			date:     demoKennedyElectionDate,
		},
		{slug: "imagined-markets", title: "Imaginary markets and measured optimism", subtitle: "A fictional briefing on sentiment and supply"},
		{slug: "soft-landing", title: "Why the demo economy always lands softly", subtitle: "Illustrative data without real-world stakes"},
		{slug: "office-coffee", title: "The quiet revolution of office coffee", subtitle: "Productivity gains in the imaginary workplace"},
		{slug: "bureaucracy", title: "Small experiments in better bureaucracy", subtitle: "A fictional reform agenda for calmer Mondays"},
	}

	type datedItem struct {
		item      rss.Item
		published time.Time
	}

	datedItems := make([]datedItem, 0, len(articles))
	for i, entry := range articles {
		published := demoBaseDate.AddDate(0, 0, -i)
		if !entry.date.IsZero() {
			published = entry.date
		}
		url := fmt.Sprintf("https://example.com/demo/%s", entry.slug)
		datedItems = append(datedItems, datedItem{
			item: rss.Item{
				Title:       entry.title,
				Description: entry.subtitle,
				Link:        url,
				GUID:        url,
				PubDate:     published.Format(time.RFC1123Z),
			},
			published: published,
		})
		content := strings.TrimSpace(entry.content)
		if content == "" {
			content = buildContent(entry.title)
		}
		s.articles[url] = &article.Article{
			Overtitle: "Leaders | Demo",
			Title:     entry.title,
			Subtitle:  entry.subtitle,
			DateLine:  formatDateLine(published),
			Content:   content,
			URL:       url,
		}
	}

	sort.SliceStable(datedItems, func(i, j int) bool {
		return datedItems[i].published.After(datedItems[j].published)
	})

	items := make([]rss.Item, 0, len(datedItems))
	for _, entry := range datedItems {
		items = append(items, entry.item)
	}

	s.sections[DefaultSection] = sectionData{title: demoSectionTitle, items: items}
}

func buildContent(title string) string {
	paragraphs := []string{
		"This demo content is stored locally so screenshots and tests can run without network access.",
		fmt.Sprintf("The headline \"%s\" is a placeholder used to show how headlines wrap and how the reader renders long paragraphs.", title),
		"Use ↑/↓ to scroll, b to go back, and c to toggle columns. Resize the terminal to see the layout adapt.",
		"Demo mode keeps everything local so you can explore the TUI without a subscription.",
		"Paragraph lengths are intentionally varied to show line wrapping, spacing, and the feel of the reading experience.",
		"If you are taking screenshots, this page is designed to be safe for public sharing.",
		"End of the sample article ■",
	}

	return strings.Join(paragraphs, "\n\n")
}

func formatDateLine(t time.Time) string {
	day := t.Day()
	suffix := "th"
	if day%100 < 11 || day%100 > 13 {
		switch day % 10 {
		case 1:
			suffix = "st"
		case 2:
			suffix = "nd"
		case 3:
			suffix = "rd"
		}
	}
	return fmt.Sprintf("%s %d%s %d", t.Format("Jan"), day, suffix, t.Year())
}

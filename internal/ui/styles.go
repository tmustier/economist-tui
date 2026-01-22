package ui

import "github.com/charmbracelet/lipgloss"

const (
	BodyColorLightANSI = "234"
	BodyColorDarkANSI  = "252"
)

type BrowseStyles struct {
	Header   lipgloss.Style
	Rule     lipgloss.Style
	Title    lipgloss.Style
	Subtitle lipgloss.Style
	Selected lipgloss.Style
	Dim      lipgloss.Style
	Help     lipgloss.Style
	Search   lipgloss.Style
}

type Styles struct {
	App     lipgloss.Style
	Sidebar lipgloss.Style
	Main    lipgloss.Style
	HelpBar lipgloss.Style

	RuleAccent lipgloss.Style
	RuleHeavy  lipgloss.Style
	Rule       lipgloss.Style

	Headline lipgloss.Style
	Subhead  lipgloss.Style
	Body     lipgloss.Style
	Caption  lipgloss.Style
	Overline lipgloss.Style

	Selected lipgloss.Style
	Focused  lipgloss.Style
	Disabled lipgloss.Style
}

type ArticleStyles struct {
	Overtitle lipgloss.Style
	Section   lipgloss.Style
	Title     lipgloss.Style
	Subtitle  lipgloss.Style
	Date      lipgloss.Style
	Rule      lipgloss.Style
	Body      lipgloss.Style
}

func NewStyles(theme Theme, noColor bool) Styles {
	if noColor {
		return newNoColorStyles()
	}

	return Styles{
		App:     lipgloss.NewStyle().Padding(1, 2),
		Sidebar: lipgloss.NewStyle().Width(28).BorderStyle(lipgloss.NormalBorder()).BorderRight(true).BorderForeground(theme.Border).Padding(0, 1),
		Main:    lipgloss.NewStyle().Padding(0, 2),
		HelpBar: lipgloss.NewStyle().Foreground(theme.TextFaint).Padding(1, 0, 0, 0),

		RuleAccent: lipgloss.NewStyle().Foreground(theme.Brand).Bold(true),
		RuleHeavy:  lipgloss.NewStyle().Foreground(theme.Text),
		Rule:       lipgloss.NewStyle().Foreground(theme.Border),

		Headline: lipgloss.NewStyle().Bold(true).Foreground(theme.Text),
		Subhead:  lipgloss.NewStyle().Foreground(theme.TextMuted).Italic(true),
		Body:     lipgloss.NewStyle().Foreground(theme.TextMuted),
		Caption:  lipgloss.NewStyle().Foreground(theme.TextFaint),
		Overline: lipgloss.NewStyle().Foreground(theme.TextMuted).Bold(true),

		Selected: lipgloss.NewStyle().Foreground(theme.Selection).Bold(true),
		Focused:  lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(theme.Brand),
		Disabled: lipgloss.NewStyle().Foreground(theme.TextFaint),
	}
}

func newNoColorStyles() Styles {
	return Styles{
		App:     lipgloss.NewStyle().Padding(1, 2),
		Sidebar: lipgloss.NewStyle().Width(28).BorderStyle(lipgloss.NormalBorder()).BorderRight(true).Padding(0, 1),
		Main:    lipgloss.NewStyle().Padding(0, 2),
		HelpBar: lipgloss.NewStyle().Padding(1, 0, 0, 0),

		RuleAccent: lipgloss.NewStyle(),
		RuleHeavy:  lipgloss.NewStyle(),
		Rule:       lipgloss.NewStyle(),

		Headline: lipgloss.NewStyle().Bold(true),
		Subhead:  lipgloss.NewStyle(),
		Body:     lipgloss.NewStyle(),
		Caption:  lipgloss.NewStyle(),
		Overline: lipgloss.NewStyle().Bold(true),

		Selected: lipgloss.NewStyle().Bold(true),
		Focused:  lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()),
		Disabled: lipgloss.NewStyle(),
	}
}

func NewBrowseStyles(noColor bool) BrowseStyles {
	theme := CurrentTheme()
	body := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: BodyColorLightANSI, Dark: BodyColorDarkANSI})
	header := lipgloss.NewStyle().Bold(true).Foreground(theme.Brand)
	rule := lipgloss.NewStyle().Foreground(theme.Border)
	title := body.Copy().Bold(true)
	subtitle := lipgloss.NewStyle().Foreground(theme.TextMuted)
	selected := lipgloss.NewStyle().Bold(true).Foreground(theme.Brand)
	dim := lipgloss.NewStyle().Foreground(theme.TextFaint)
	help := lipgloss.NewStyle().Foreground(theme.TextFaint)
	search := lipgloss.NewStyle().Foreground(theme.TextFaint)

	if noColor {
		body = lipgloss.NewStyle()
		header = lipgloss.NewStyle().Bold(true)
		rule = lipgloss.NewStyle()
		title = body.Copy().Bold(true)
		subtitle = lipgloss.NewStyle()
		selected = lipgloss.NewStyle().Bold(true)
		dim = lipgloss.NewStyle()
		help = lipgloss.NewStyle()
		search = lipgloss.NewStyle()
	}

	return BrowseStyles{
		Header:   header,
		Rule:     rule,
		Title:    title,
		Subtitle: subtitle,
		Selected: selected,
		Dim:      dim,
		Help:     help,
		Search:   search,
	}
}

func NewArticleStyles(noColor bool) ArticleStyles {
	theme := CurrentTheme()
	body := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: BodyColorLightANSI, Dark: BodyColorDarkANSI})
	overtitle := body.Copy()
	section := lipgloss.NewStyle().Bold(true).Foreground(theme.Brand)
	title := lipgloss.NewStyle().Bold(true).Foreground(theme.Brand)
	subtitle := lipgloss.NewStyle().Foreground(theme.TextMuted)
	date := lipgloss.NewStyle().Foreground(theme.TextFaint).Faint(true)
	rule := lipgloss.NewStyle().Foreground(theme.Border)

	if noColor {
		overtitle = lipgloss.NewStyle()
		section = lipgloss.NewStyle()
		title = lipgloss.NewStyle().Bold(true)
		subtitle = lipgloss.NewStyle()
		date = lipgloss.NewStyle()
		rule = lipgloss.NewStyle()
		body = lipgloss.NewStyle()
	}

	return ArticleStyles{
		Overtitle: overtitle,
		Section:   section,
		Title:     title,
		Subtitle:  subtitle,
		Date:      date,
		Rule:      rule,
		Body:      body,
	}
}

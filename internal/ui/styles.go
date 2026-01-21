package ui

import "github.com/charmbracelet/lipgloss"

const HighlightColor = "124"

type BrowseStyles struct {
	Header   lipgloss.Style
	Rule     lipgloss.Style
	Title    lipgloss.Style
	Selected lipgloss.Style
	Dim      lipgloss.Style
	Help     lipgloss.Style
	Search   lipgloss.Style
}

type ArticleStyles struct {
	Overtitle lipgloss.Style
	Title     lipgloss.Style
	Subtitle  lipgloss.Style
	Date      lipgloss.Style
	Rule      lipgloss.Style
}

func NewBrowseStyles(noColor bool) BrowseStyles {
	header := lipgloss.NewStyle().Bold(true)
	rule := lipgloss.NewStyle().Foreground(lipgloss.Color(HighlightColor))
	title := lipgloss.NewStyle().Bold(true)
	selected := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(HighlightColor))
	dim := lipgloss.NewStyle().Faint(true)
	help := lipgloss.NewStyle().Faint(true)
	search := lipgloss.NewStyle().Faint(true)

	if noColor {
		header = lipgloss.NewStyle().Bold(true)
		rule = lipgloss.NewStyle()
		title = lipgloss.NewStyle().Bold(true)
		selected = lipgloss.NewStyle().Bold(true)
		dim = lipgloss.NewStyle()
		help = lipgloss.NewStyle()
		search = lipgloss.NewStyle()
	}

	return BrowseStyles{
		Header:   header,
		Rule:     rule,
		Title:    title,
		Selected: selected,
		Dim:      dim,
		Help:     help,
		Search:   search,
	}
}

func NewArticleStyles(noColor bool) ArticleStyles {
	overtitle := lipgloss.NewStyle().Faint(true)
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(HighlightColor))
	subtitle := lipgloss.NewStyle().Faint(true)
	date := lipgloss.NewStyle().Faint(true)
	rule := lipgloss.NewStyle().Faint(true)

	if noColor {
		overtitle = lipgloss.NewStyle()
		title = lipgloss.NewStyle().Bold(true)
		subtitle = lipgloss.NewStyle()
		date = lipgloss.NewStyle()
		rule = lipgloss.NewStyle()
	}

	return ArticleStyles{
		Overtitle: overtitle,
		Title:     title,
		Subtitle:  subtitle,
		Date:      date,
		Rule:      rule,
	}
}

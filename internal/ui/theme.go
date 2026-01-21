package ui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// Economist Design System colors.
var (
	// Brand
	EconomistRed   = lipgloss.Color("#E3120B")
	EconomistRed60 = lipgloss.Color("#F6423C")

	// Accent (Chicago blues)
	Chicago20 = lipgloss.Color("#141F52")
	Chicago45 = lipgloss.Color("#2E45B8")
	Chicago90 = lipgloss.Color("#D6DBF5")

	// Secondary
	HongKong45 = lipgloss.Color("#1DC9A4") // Teal/success
	Tokyo45    = lipgloss.Color("#C91D42") // Rose/error

	// Tertiary
	Singapore55 = lipgloss.Color("#F97A1F") // Orange
	NewYork55   = lipgloss.Color("#F9C31F") // Gold/warning

	// Greyscale (London)
	London5   = lipgloss.Color("#0D0D0D")
	London10  = lipgloss.Color("#1A1A1A")
	London20  = lipgloss.Color("#333333")
	London35  = lipgloss.Color("#595959")
	London70  = lipgloss.Color("#B3B3B3")
	London85  = lipgloss.Color("#D9D9D9")
	London95  = lipgloss.Color("#F2F2F2")
	London100 = lipgloss.Color("#FFFFFF")

	// Canvas
	LosAngeles95 = lipgloss.Color("#F5F4EF") // Warm paper
	Paris95      = lipgloss.Color("#EFF5F5") // Cool paper
)

// Theme defines semantic color assignments.
type Theme struct {
	Brand      lipgloss.Color
	Text       lipgloss.Color
	TextMuted  lipgloss.Color
	TextFaint  lipgloss.Color
	Border     lipgloss.Color
	BorderDim  lipgloss.Color
	Background lipgloss.Color
	Selection  lipgloss.Color
	Success    lipgloss.Color
	Warning    lipgloss.Color
	Error      lipgloss.Color
}

var DefaultTheme = Theme{
	Brand:      EconomistRed,
	Text:       London20,
	TextMuted:  London35,
	TextFaint:  London70,
	Border:     London85,
	BorderDim:  London95,
	Background: London100,
	Selection:  EconomistRed,
	Success:    HongKong45,
	Warning:    NewYork55,
	Error:      Tokyo45,
}

var DarkTheme = Theme{
	Brand:      EconomistRed,
	Text:       London95,
	TextMuted:  London70,
	TextFaint:  London35,
	Border:     London20,
	BorderDim:  London10,
	Background: London5,
	Selection:  EconomistRed,
	Success:    lipgloss.Color("#36E2BD"), // HongKong55
	Warning:    lipgloss.Color("#FBD051"), // NewYork65
	Error:      lipgloss.Color("#E2365B"), // Tokyo55
}

func CurrentTheme() Theme {
	if termenv.HasDarkBackground() {
		return DarkTheme
	}
	return DefaultTheme
}

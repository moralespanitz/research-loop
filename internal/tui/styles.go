package tui

import "github.com/charmbracelet/lipgloss"

// Palette
var (
	colorPrimary  = lipgloss.Color("#3b82f6") // blue
	colorSuccess  = lipgloss.Color("#22c55e") // green
	colorWarn     = lipgloss.Color("#f59e0b") // amber
	colorDanger   = lipgloss.Color("#ef4444") // red
	colorMuted    = lipgloss.Color("#6b7280") // gray-500
	colorSubtle   = lipgloss.Color("#374151") // gray-700
	colorText     = lipgloss.Color("#f3f4f6") // gray-100
	colorBg       = lipgloss.Color("#111827") // gray-900
	colorBorder   = lipgloss.Color("#1f2937") // gray-800
	colorHighlight = lipgloss.Color("#1e3a5f") // dark blue bg
)

// Base styles
var (
	bold   = lipgloss.NewStyle().Bold(true)
	muted  = lipgloss.NewStyle().Foreground(colorMuted)
	subtle = lipgloss.NewStyle().Foreground(colorSubtle)

	successText = lipgloss.NewStyle().Foreground(colorSuccess).Bold(true)
	warnText    = lipgloss.NewStyle().Foreground(colorWarn).Bold(true)
	dangerText  = lipgloss.NewStyle().Foreground(colorDanger).Bold(true)
	primaryText = lipgloss.NewStyle().Foreground(colorPrimary).Bold(true)
)

// Layout styles
var (
	// Outer app frame
	appStyle = lipgloss.NewStyle().
		Padding(1, 2)

	// Card / panel box
	cardStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBorder).
		Padding(1, 2)

	// Header bar across the top
	headerStyle = lipgloss.NewStyle().
		Background(colorBg).
		Foreground(colorText).
		Bold(true).
		Padding(0, 2).
		Width(80)

	// Section title inside a card
	sectionTitle = lipgloss.NewStyle().
		Foreground(colorMuted).
		Bold(true).
		PaddingBottom(1)

	// Selected menu item
	selectedItem = lipgloss.NewStyle().
		Foreground(colorPrimary).
		Bold(true).
		PaddingLeft(2)

	// Unselected menu item
	normalItem = lipgloss.NewStyle().
		Foreground(colorText).
		PaddingLeft(2)

	// Status badges
	badgeBlue = lipgloss.NewStyle().
		Background(colorPrimary).
		Foreground(lipgloss.Color("#fff")).
		Padding(0, 1).
		Bold(true)

	badgeGreen = lipgloss.NewStyle().
		Background(colorSuccess).
		Foreground(lipgloss.Color("#fff")).
		Padding(0, 1).
		Bold(true)

	badgeGray = lipgloss.NewStyle().
		Background(colorSubtle).
		Foreground(colorMuted).
		Padding(0, 1)

	// Key hint (e.g. "q quit")
	keyHint = lipgloss.NewStyle().
		Foreground(colorMuted)

	keyLabel = lipgloss.NewStyle().
		Foreground(colorPrimary).
		Bold(true)

	// Metric value (big number in dashboard)
	metricValue = lipgloss.NewStyle().
		Foreground(colorText).
		Bold(true)

	metricLabel = lipgloss.NewStyle().
		Foreground(colorMuted).
		PaddingTop(0)

	// Input field
	inputStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorPrimary).
		Padding(0, 1).
		Width(60)

	// Dim overlay text
	dimText = lipgloss.NewStyle().
		Foreground(colorSubtle)
)

// helpBar renders the bottom key-hint row.
func helpBar(hints ...string) string {
	var parts []string
	for i := 0; i+1 < len(hints); i += 2 {
		key := keyLabel.Render(hints[i])
		desc := keyHint.Render(" " + hints[i+1])
		parts = append(parts, key+desc)
	}
	return muted.Render("  " + joinWith("   ", parts...))
}

func joinWith(sep string, parts ...string) string {
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += sep
		}
		result += p
	}
	return result
}

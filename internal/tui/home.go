package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const logo = `
 ██████╗ ███████╗███████╗███████╗ █████╗ ██████╗  ██████╗██╗  ██╗
 ██╔══██╗██╔════╝██╔════╝██╔════╝██╔══██╗██╔══██╗██╔════╝██║  ██║
 ██████╔╝█████╗  ███████╗█████╗  ███████║██████╔╝██║     ███████║
 ██╔══██╗██╔══╝  ╚════██║██╔══╝  ██╔══██║██╔══██╗██║     ██╔══██║
 ██║  ██║███████╗███████║███████╗██║  ██║██║  ██║╚██████╗██║  ██║
 ╚═╝  ╚═╝╚══════╝╚══════╝╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝ ╚═════╝╚═╝  ╚═╝
                                                         L O O P`

const tagline = "The Researcher's Agent OS"

type homeMenuItem struct {
	label string
	desc  string
	next  screen
}

var homeMenuItems = []homeMenuItem{
	{"  Explore problem", "MIT grad student: mental models, debates, Carlini scoring", screenExplore},
	{"  Setup provider", "Connect Claude Code, Codex, Gemini, Ollama, and more", screenSetup},
	{"  Start new investigation", "Ingest a paper and extract a hypothesis", screenIngest},
	{"  Browse sessions", "View and resume existing research sessions", screenSessions},
	{"  Parallel discovery", "Run multiple research lanes with Carlini gates", screenDiscovery},
	{"  Live dashboard", "Monitor running experiments in real time", screenDashboard},
	{"  Quit", "Exit Research Loop", screenQuit},
}

type homeModel struct {
	cursor int
}

func newHomeModel() homeModel {
	return homeModel{cursor: 0}
}

func (m homeModel) Init() tea.Cmd { return nil }

func (m homeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(homeMenuItems)-1 {
				m.cursor++
			}
		case "enter", " ":
			item := homeMenuItems[m.cursor]
			if item.next == screenQuit {
				return m, tea.Quit
			}
			return m, navigateTo(item.next)
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m homeModel) View() string {
	// Logo
	logoStyled := lipgloss.NewStyle().
		Foreground(colorPrimary).
		Bold(true).
		Render(logo)

	tag := lipgloss.NewStyle().
		Foreground(colorMuted).
		Italic(true).
		PaddingLeft(4).
		Render(tagline)

	version := badgeGray.Render(" v0.1.0 ")
	header := lipgloss.JoinHorizontal(lipgloss.Bottom, tag, "  ", version)

	// Menu
	menuLines := ""
	for i, item := range homeMenuItems {
		var line string
		if i == m.cursor {
			arrow := primaryText.Render("▶")
			label := selectedItem.Render(item.label)
			desc := lipgloss.NewStyle().Foreground(colorMuted).PaddingLeft(4).Render(item.desc)
			line = fmt.Sprintf("%s%s\n%s", arrow, label, desc)
		} else {
			label := normalItem.Render(item.label)
			line = label
		}
		if i < len(homeMenuItems)-1 {
			menuLines += line + "\n\n"
		} else {
			menuLines += line
		}
	}

	menuCard := cardStyle.Render(
		sectionTitle.Render("NAVIGATION") + "\n\n" + menuLines,
	)

	// Bottom hint
	hint := helpBar("↑↓", "navigate", "enter", "select", "q", "quit")

	return appStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			logoStyled,
			header,
			"",
			menuCard,
			"",
			hint,
		),
	)
}

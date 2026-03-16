package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ─── Tick message ─────────────────────────────────────────────────────────────

type sessionsTickMsg struct{}

func sessionsTickCmd() tea.Cmd {
	return tea.Tick(10*time.Second, func(t time.Time) tea.Msg {
		return sessionsTickMsg{}
	})
}

// ─── Model ───────────────────────────────────────────────────────────────────

type sessionsModel struct {
	workspace string
	table     table.Model
	sessions  []sessionRow
}

type sessionRow struct {
	id      string
	title   string
	runs    int
	created string
}

func newSessionsModel(workspace string) sessionsModel {
	cols := []table.Column{
		{Title: "SESSION ID", Width: 36},
		{Title: "PAPER", Width: 32},
		{Title: "RUNS", Width: 6},
		{Title: "CREATED", Width: 10},
	}

	t := table.New(
		table.WithColumns(cols),
		table.WithFocused(true),
		table.WithHeight(12),
	)

	ts := table.DefaultStyles()
	ts.Header = ts.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colorBorder).
		BorderBottom(true).
		Bold(true).
		Foreground(colorMuted)
	ts.Selected = ts.Selected.
		Foreground(colorText).
		Background(colorHighlight).
		Bold(true)
	t.SetStyles(ts)

	m := sessionsModel{workspace: workspace, table: t}
	m.reload()
	return m
}

func (m *sessionsModel) reload() {
	dir := filepath.Join(m.workspace, ".research-loop", "sessions")
	entries, _ := os.ReadDir(dir)

	m.sessions = nil
	var rows []table.Row
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		id := e.Name()
		info, _ := e.Info()
		created := ""
		if info != nil {
			created = info.ModTime().Format("2006-01-02")
		}
		title := readSessionHeading(filepath.Join(dir, id, "hypothesis.md"))
		runs := countFileLines(filepath.Join(dir, id, "autoresearch.jsonl"))
		m.sessions = append(m.sessions, sessionRow{id, title, runs, created})
		rows = append(rows, table.Row{id, truncateStr(title, 30), fmt.Sprintf("%d", runs), created})
	}
	m.table.SetRows(rows)
}

// ─── Update ──────────────────────────────────────────────────────────────────

func (m sessionsModel) Init() tea.Cmd {
	return sessionsTickCmd()
}

func (m sessionsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return m, navigateTo(screenHome)
		case "enter":
			// Open dashboard for selected session
			return m, navigateTo(screenDashboard)
		case "n":
			return m, navigateTo(screenIngest)
		}
	case sessionsTickMsg:
		m.reload()
		return m, sessionsTickCmd()
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// ─── View ────────────────────────────────────────────────────────────────────

func (m sessionsModel) View() string {
	header := headerStyle.Render("🔬  Research Loop  /  Sessions")

	count := badgeBlue.Render(fmt.Sprintf(" %d ", len(m.sessions)))
	title := lipgloss.JoinHorizontal(lipgloss.Center,
		primaryText.Render("Research Sessions  "), count,
	)

	var body string
	if len(m.sessions) == 0 {
		body = cardStyle.Render(
			dimText.Render("No sessions yet.\n\nPress ") +
				keyLabel.Render("n") +
				dimText.Render(" to start a new investigation."),
		)
	} else {
		tableBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Render(m.table.View())
		body = tableBox
	}

	hint := helpBar("↑↓", "navigate", "enter", "open dashboard", "n", "new session", "esc", "home")

	return appStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			header, "", title, "", body, "", hint,
		),
	)
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func readSessionHeading(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return "(no hypothesis)"
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimPrefix(line, "# ")
		}
	}
	return "(untitled)"
}

func countFileLines(path string) int {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	return strings.Count(string(data), "\n")
}

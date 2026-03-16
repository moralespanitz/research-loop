package tui

import (
	"encoding/json"
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

type discoverTickMsg struct{}

func discoverTickCmd() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return discoverTickMsg{}
	})
}

// ─── Model ───────────────────────────────────────────────────────────────────

type discoveryModel struct {
	workspace   string
	table       table.Model
	discoveries []discoveryRow
}

type discoveryRow struct {
	topic    string
	lanes    int
	bestVal  float64
	bestLane string
	verdict  string
	date     string
}

func newDiscoveryModel(workspace string) discoveryModel {
	cols := []table.Column{
		{Title: "TOPIC", Width: 28},
		{Title: "LANES", Width: 6},
		{Title: "BEST", Width: 10},
		{Title: "BEST LANE", Width: 20},
		{Title: "VERDICT", Width: 12},
		{Title: "DATE", Width: 10},
	}

	t := table.New(
		table.WithColumns(cols),
		table.WithFocused(true),
		table.WithHeight(10),
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

	m := discoveryModel{workspace: workspace, table: t}
	m.reload()
	return m
}

func (m *discoveryModel) reload() {
	dir := filepath.Join(m.workspace, ".research-loop", "discoveries")
	entries, _ := os.ReadDir(dir)

	m.discoveries = nil
	var rows []table.Row
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}

		path := filepath.Join(dir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var lanes []interface{}
		if err := json.Unmarshal(data, &lanes); err != nil {
			continue
		}

		row := discoveryRow{
			topic:    strings.TrimSuffix(e.Name(), ".json"),
			lanes:    len(lanes),
			bestVal:  0,
			bestLane: "-",
			verdict:  "-",
		}

		for _, l := range lanes {
			lane, ok := l.(map[string]interface{})
			if !ok {
				continue
			}

			if best, ok := lane["best_metric"].(float64); ok && (row.bestVal == 0 || best < row.bestVal) {
				row.bestVal = best
			}
			if bestNode, ok := lane["best_node"].(string); ok && bestNode != "" {
				row.bestLane = bestNode
			}
			if v, ok := lane["verdict"].(string); ok && v != "" {
				row.verdict = v
			}
		}

		info, _ := e.Info()
		if info != nil {
			row.date = info.ModTime().Format("2006-01-02")
		}

		m.discoveries = append(m.discoveries, row)
		rows = append(rows, table.Row{
			truncateStr(row.topic, 26),
			fmt.Sprintf("%d", row.lanes),
			fmt.Sprintf("%.4f", row.bestVal),
			truncateStr(row.bestLane, 18),
			row.verdict,
			row.date,
		})
	}
	m.table.SetRows(rows)
}

// ─── Update ──────────────────────────────────────────────────────────────────

func (m discoveryModel) Init() tea.Cmd {
	return discoverTickCmd()
}

func (m discoveryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return m, navigateTo(screenHome)
		case "r":
			m.reload()
		}
	case discoverTickMsg:
		m.reload()
		return m, discoverTickCmd()
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// ─── View ───────────────────────────────────────────────────────────────────

func (m discoveryModel) View() string {
	header := headerStyle.Render("🔬  Research Loop  /  Discover")

	count := badgeBlue.Render(fmt.Sprintf(" %d ", len(m.discoveries)))
	title := lipgloss.JoinHorizontal(lipgloss.Center,
		primaryText.Render("Parallel Discovery Results  "), count,
	)

	var body string
	if len(m.discoveries) == 0 {
		body = cardStyle.Render(
			dimText.Render("No discoveries yet.\n\nRun from CLI: ") +
				keyLabel.Render("research-loop discover \"topic\"") +
				dimText.Render("\n\nEach discovery runs multiple lanes in parallel,\nwith Carlini decision gates to kill poor hypotheses early."),
		)
	} else {
		tableBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Render(m.table.View())
		body = tableBox
	}

	hint := helpBar("r", "refresh", "esc", "home")

	return appStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			header, "", title, "", body, "", hint,
		),
	)
}

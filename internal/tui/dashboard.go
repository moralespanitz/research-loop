package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ─── Tick ────────────────────────────────────────────────────────────────────

type dashTickMsg struct{}

func dashTickCmd() tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return dashTickMsg{}
	})
}

// ─── Run record from JSONL ────────────────────────────────────────────────────

type runRecord struct {
	Event      string  `json:"event"`
	Node       string  `json:"node"`
	Result     string  `json:"result"`
	Status     string  `json:"status"`
	Annotation string  `json:"annotation"`
	MetricVal  float64 `json:"metric_value"`
	Timestamp  string  `json:"timestamp"`
}

// ─── Model ───────────────────────────────────────────────────────────────────

type dashboardModel struct {
	workspace  string
	spinner    spinner.Model
	sessionID  string
	title      string
	runs       int
	bestMetric string
	lastRuns   []runRecord
	uptime     time.Time
	kg         string // first N lines of knowledge_graph.md
}

func newDashboardModel(workspace string) dashboardModel {
	sp := spinner.New()
	sp.Spinner = spinner.MiniDot
	sp.Style = lipgloss.NewStyle().Foreground(colorPrimary)

	m := dashboardModel{
		workspace: workspace,
		spinner:   sp,
		uptime:    time.Now(),
	}
	m.reload()
	return m
}

func (m *dashboardModel) reload() {
	dir := filepath.Join(m.workspace, ".research-loop", "sessions")
	entries, _ := os.ReadDir(dir)
	if len(entries) == 0 {
		return
	}

	// Find most recent session dir
	var latestID string
	for i := len(entries) - 1; i >= 0; i-- {
		if entries[i].IsDir() {
			latestID = entries[i].Name()
			break
		}
	}
	if latestID == "" {
		return
	}
	m.sessionID = latestID

	sessionDir := filepath.Join(dir, latestID)
	m.title = readSessionHeading(filepath.Join(sessionDir, "hypothesis.md"))

	// Parse JSONL for run history
	jsonlPath := filepath.Join(sessionDir, "autoresearch.jsonl")
	data, _ := os.ReadFile(jsonlPath)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	m.runs = 0
	m.lastRuns = nil
	for _, line := range lines {
		if line == "" {
			continue
		}
		m.runs++
		var rec runRecord
		if err := json.Unmarshal([]byte(line), &rec); err == nil {
			if rec.Status != "" {
				m.lastRuns = append(m.lastRuns, rec)
			}
		}
	}
	// Keep last 5 experiment runs
	if len(m.lastRuns) > 5 {
		m.lastRuns = m.lastRuns[len(m.lastRuns)-5:]
	}
	// Find best metric (last improvement)
	for i := len(m.lastRuns) - 1; i >= 0; i-- {
		if m.lastRuns[i].Status == "improvement" && m.lastRuns[i].Result != "" {
			m.bestMetric = m.lastRuns[i].Result
			break
		}
	}

	// Read first 20 lines of knowledge graph
	kgPath := filepath.Join(sessionDir, "knowledge_graph.md")
	kgData, _ := os.ReadFile(kgPath)
	kgLines := strings.Split(string(kgData), "\n")
	if len(kgLines) > 20 {
		kgLines = kgLines[:20]
	}
	m.kg = strings.Join(kgLines, "\n")
}

// ─── Update ──────────────────────────────────────────────────────────────────

func (m dashboardModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, dashTickCmd())
}

func (m dashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return m, navigateTo(screenHome)
		case "s":
			return m, navigateTo(screenSessions)
		case "n":
			return m, navigateTo(screenIngest)
		case "r":
			m.reload()
			return m, nil
		}
	case dashTickMsg:
		m.reload()
		return m, dashTickCmd()
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

// ─── View ────────────────────────────────────────────────────────────────────

func (m dashboardModel) View() string {
	header := headerStyle.Render("🔬  Research Loop  /  Dashboard")

	uptime := time.Since(m.uptime).Round(time.Second)

	if m.sessionID == "" {
		empty := cardStyle.Render(
			dimText.Render("No active session.\n\nPress ") +
				keyLabel.Render("n") +
				dimText.Render(" to start a new investigation."),
		)
		return appStyle.Render(lipgloss.JoinVertical(lipgloss.Left, header, "", empty))
	}

	// ── Stat row ──────────────────────────────────────────────────────────────
	statW := 18
	statStyle := cardStyle.Copy().Width(statW).Align(lipgloss.Center)

	statRuns := statStyle.Render(
		metricValue.Render(fmt.Sprintf("%d", m.runs)) + "\n" +
			metricLabel.Render("TOTAL RUNS"),
	)

	bestVal := m.bestMetric
	if bestVal == "" {
		bestVal = "—"
	}
	statBest := statStyle.Render(
		metricValue.Render(truncateStr(bestVal, 10)) + "\n" +
			metricLabel.Render("BEST METRIC"),
	)

	statUptime := statStyle.Render(
		metricValue.Render(uptime.String()) + "\n" +
			metricLabel.Render("UPTIME"),
	)

	statsRow := lipgloss.JoinHorizontal(lipgloss.Top, statRuns, "  ", statBest, "  ", statUptime)

	// ── Session info ──────────────────────────────────────────────────────────
	spin := m.spinner.View()
	sessionInfo := cardStyle.Render(
		spin + "  " + primaryText.Render(truncateStr(m.title, 56)) +
			"\n" + muted.Render("Session: "+m.sessionID),
	)

	// ── Recent runs ───────────────────────────────────────────────────────────
	var runsContent string
	if len(m.lastRuns) == 0 {
		runsContent = dimText.Render("No experiment runs yet.")
	} else {
		for i := len(m.lastRuns) - 1; i >= 0; i-- {
			r := m.lastRuns[i]
			statusIcon := statusIcon(r.Status)
			node := lipgloss.NewStyle().Foreground(colorText).Bold(true).Render(truncateStr(r.Node, 28))
			result := muted.Render(truncateStr(r.Result, 20))
			runsContent += fmt.Sprintf("%s  %-28s  %s\n", statusIcon, node, result)
		}
		runsContent = strings.TrimRight(runsContent, "\n")
	}
	runsCard := cardStyle.Render(
		sectionTitle.Render("RECENT RUNS") + "\n\n" + runsContent,
	)

	// ── Knowledge graph preview ───────────────────────────────────────────────
	kgCard := cardStyle.Render(
		sectionTitle.Render("KNOWLEDGE GRAPH (preview)") + "\n\n" +
			dimText.Render(m.kg),
	)

	hint := helpBar("r", "refresh", "n", "new session", "s", "sessions", "esc", "home")

	return appStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			header, "",
			sessionInfo, "",
			statsRow, "",
			runsCard, "",
			kgCard, "",
			hint,
		),
	)
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func statusIcon(status string) string {
	switch status {
	case "improvement":
		return successText.Render("✓")
	case "regression":
		return dangerText.Render("✗")
	case "crash_failed":
		return dangerText.Render("💥")
	case "checks_failed":
		return warnText.Render("⚠")
	default:
		return muted.Render("•")
	}
}

package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ─── Messages ───────────────────────────────────────────────────────────────

type exploreStartMsg struct {
	topic string
}

type exploreProgressMsg struct {
	phase   string
	message string
}

type exploreDoneMsg struct {
	summary string
}

// ─── Model ───────────────────────────────────────────────────────────────────

type exploreModel struct {
	workspace string
	input     textinput.Model
	phase     string // "input" | "exploring" | "done"
	progress  string
	summary   string
	err       error
}

func newExploreModel(workspace string) exploreModel {
	ti := textinput.New()
	ti.Placeholder = "e.g. machine learning efficiency, attention mechanisms..."
	ti.Focus()
	ti.Prompt = "Topic: "
	ti.Width = 50

	return exploreModel{
		workspace: workspace,
		input:     ti,
		phase:     "input",
	}
}

func (m exploreModel) Init() tea.Cmd { return nil }

func (m exploreModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			return m, navigateTo(screenHome)
		case "enter":
			if m.phase == "input" && m.input.Value() != "" {
				m.phase = "exploring"
				m.progress = "Starting exploration..."
				return m, func() tea.Msg {
					return exploreStartMsg{topic: m.input.Value()}
				}
			}
		}
	case exploreProgressMsg:
		m.progress = fmt.Sprintf("%s: %s", msg.phase, msg.message)
	case exploreDoneMsg:
		m.phase = "done"
		m.summary = msg.summary
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m exploreModel) View() string {
	header := headerStyle.Render("🔬  Research Loop  /  Explore")

	var content string

	switch m.phase {
	case "input":
		content = cardStyle.Render(
			sectionTitle.Render("EXPLORE A RESEARCH PROBLEM") + "\n\n" +
				"Using the MIT grad student methodology:\n" +
				"  1. Gather papers & GitHub repos\n" +
				"  2. Extract mental models\n" +
				"  3. Find field debates\n" +
				"  4. Generate diagnostic questions\n" +
				"  5. Carlini scoring\n\n" +
				m.input.View() + "\n\n" +
				dimText.Render("Press Enter to start exploration"),
		)
	case "exploring":
		content = cardStyle.Render(
			sectionTitle.Render("EXPLORING...") + "\n\n" +
				primaryText.Render(m.progress) + "\n\n" +
				dimText.Render("This uses the idea-selection skill (Carlini methodology)"),
		)
	case "done":
		content = m.summary
	}

	hint := helpBar("esc", "home")

	return appStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			header, "", content, "", hint,
		),
	)
}

// ─── Exploration loader ───────────────────────────────────────────────────

type Exploration struct {
	Topic string `json:"topic"`
	Score struct {
		Taste       float64 `json:"taste"`
		Uniqueness  float64 `json:"uniqueness"`
		Impact      float64 `json:"impact"`
		Feasibility float64 `json:"feasibility"`
		Overall     float64 `json:"overall"`
		Verdict     string  `json:"verdict"`
		Reasoning   string  `json:"reasoning"`
	} `json:"score"`
	CreatedAt string `json:"created_at"`
}

func loadLatestExploration(workspace string) (*Exploration, error) {
	dir := filepath.Join(workspace, ".research-loop", "explorations")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var latest os.DirEntry
	var latestTime time.Time
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		info, _ := e.Info()
		if info.ModTime().After(latestTime) {
			latestTime = info.ModTime()
			latest = e
		}
	}

	if latest == nil {
		return nil, fmt.Errorf("no explorations found")
	}

	data, err := os.ReadFile(filepath.Join(dir, latest.Name()))
	if err != nil {
		return nil, err
	}

	var exp Exploration
	if err := json.Unmarshal(data, &exp); err != nil {
		return nil, err
	}

	return &exp, nil
}

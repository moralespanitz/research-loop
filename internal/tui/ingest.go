package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/research-loop/research-loop/internal/config"
	"github.com/research-loop/research-loop/internal/hypothesis"
	"github.com/research-loop/research-loop/internal/ingestion"
	"github.com/research-loop/research-loop/internal/llm"
	"github.com/research-loop/research-loop/internal/persistence"
)

// ─── States ──────────────────────────────────────────────────────────────────

type ingestState int

const (
	ingestInput      ingestState = iota // user types URL
	ingestFetching                      // downloading paper
	ingestExtracting                    // LLM extracting hypothesis
	ingestDone                          // hypothesis displayed
	ingestError                         // something went wrong
)

// ─── Messages ────────────────────────────────────────────────────────────────

type paperFetchedMsg struct{ paper *ingestion.Paper }
type hypothesisExtractedMsg struct {
	h       *hypothesis.Hypothesis
	session *persistence.Session
}
type ingestErrMsg struct{ err error }

// ─── Model ───────────────────────────────────────────────────────────────────

type ingestModel struct {
	workspace string
	cfg       *config.Config
	state     ingestState
	input     textinput.Model
	spinner   spinner.Model
	paper     *ingestion.Paper
	hyp       *hypothesis.Hypothesis
	session   *persistence.Session
	err       error
	step      string // current step description
}

func newIngestModel(workspace string, cfg *config.Config) ingestModel {
	ti := textinput.New()
	ti.Placeholder = "https://arxiv.org/abs/2403.05821  or  2403.05821"
	ti.Focus()
	ti.Width = 58
	ti.CharLimit = 256

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(colorPrimary)

	return ingestModel{
		workspace: workspace,
		cfg:       cfg,
		state:     ingestInput,
		input:     ti,
		spinner:   sp,
	}
}

// ─── Commands ────────────────────────────────────────────────────────────────

func fetchPaper(workspace, input string) tea.Cmd {
	return func() tea.Msg {
		paperDir := workspace + "/.research-loop/library/papers"
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		var paper *ingestion.Paper
		var err error
		if strings.HasSuffix(strings.ToLower(input), ".pdf") {
			paper, err = ingestion.FetchLocalPDF(input)
		} else {
			paper, err = ingestion.FetchArXiv(ctx, input, paperDir)
		}
		if err != nil {
			return ingestErrMsg{err}
		}
		return paperFetchedMsg{paper}
	}
}

func extractHypothesis(workspace string, cfg *config.Config, paper *ingestion.Paper) tea.Cmd {
	return func() tea.Msg {
		client, err := llm.New(cfg.LLM)
		if err != nil {
			return ingestErrMsg{fmt.Errorf("LLM not configured: %w\n\nSet ANTHROPIC_API_KEY or edit .research-loop/config.toml", err)}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()

		h, err := hypothesis.Extract(ctx, client, paper)
		if err != nil {
			return ingestErrMsg{err}
		}

		session, err := persistence.NewSession(workspace, h.PaperTitle)
		if err != nil {
			return ingestErrMsg{err}
		}
		if err := session.WriteHypothesis(h); err != nil {
			return ingestErrMsg{err}
		}
		_ = session.WriteKnowledgeGraph(h.PaperTitle)
		_ = session.WriteLabNotebook(h.PaperTitle)
		_ = session.AppendJSONL(map[string]interface{}{
			"event": "session_initialized_via_tui",
			"paper": h.PaperTitle,
		})
		return hypothesisExtractedMsg{h: h, session: session}
	}
}

// ─── Update ──────────────────────────────────────────────────────────────────

func (m ingestModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m ingestModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.state == ingestInput || m.state == ingestDone || m.state == ingestError {
				return m, navigateTo(screenHome)
			}
		case "enter":
			if m.state == ingestInput {
				url := strings.TrimSpace(m.input.Value())
				if url == "" {
					return m, nil
				}
				m.state = ingestFetching
				m.step = "Downloading paper…"
				return m, tea.Batch(
					m.spinner.Tick,
					fetchPaper(m.workspace, url),
				)
			}
			if m.state == ingestDone {
				return m, navigateTo(screenSessions)
			}
			if m.state == ingestError {
				m.state = ingestInput
				m.err = nil
				m.input.Reset()
				m.input.Focus()
				return m, textinput.Blink
			}
		case "esc":
			return m, navigateTo(screenHome)
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case paperFetchedMsg:
		m.paper = msg.paper
		m.state = ingestExtracting
		m.step = "Extracting hypothesis via LLM…"
		return m, tea.Batch(
			m.spinner.Tick,
			extractHypothesis(m.workspace, m.cfg, m.paper),
		)

	case hypothesisExtractedMsg:
		m.hyp = msg.h
		m.session = msg.session
		m.state = ingestDone
		return m, nil

	case ingestErrMsg:
		m.err = msg.err
		m.state = ingestError
		return m, nil
	}

	// Delegate input events
	if m.state == ingestInput {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	return m, nil
}

// ─── View ────────────────────────────────────────────────────────────────────

func (m ingestModel) View() string {
	header := headerStyle.Render("🔬  Research Loop  /  New Investigation")

	var body string
	switch m.state {

	case ingestInput:
		title := primaryText.Render("Start a new investigation")
		sub := muted.Render("Enter an ArXiv URL, bare ID, or path to a local PDF")
		inputBox := inputStyle.Render(m.input.View())
		hint := helpBar("enter", "fetch paper", "esc", "back", "q", "home")
		body = lipgloss.JoinVertical(lipgloss.Left,
			title, sub, "", inputBox, "", hint,
		)

	case ingestFetching, ingestExtracting:
		spin := m.spinner.View()
		step := primaryText.Render(m.step)
		var substeps string
		if m.state == ingestExtracting && m.paper != nil {
			substeps = "\n" + muted.Render("  Paper: "+truncateStr(m.paper.Title, 60))
			if m.paper.FullText == "" {
				substeps += "\n" + warnText.Render("  ⚠ Full text unavailable — using abstract only")
			}
		}
		body = lipgloss.JoinVertical(lipgloss.Left,
			spin+" "+step+substeps,
		)

	case ingestDone:
		h := m.hyp
		checkmark := successText.Render("✓  Hypothesis extracted")

		fields := []struct{ label, value string }{
			{"Paper", truncateStr(h.PaperTitle, 64)},
			{"ArXiv", "https://arxiv.org/abs/" + h.ArXivID},
			{"Session", m.session.ID},
		}
		meta := ""
		for _, f := range fields {
			meta += muted.Render(fmt.Sprintf("  %-10s", f.label)) +
				lipgloss.NewStyle().Foreground(colorText).Render(f.value) + "\n"
		}

		claimBox := cardStyle.Copy().BorderForeground(colorSuccess).Render(
			sectionTitle.Render("CORE CLAIM") + "\n\n" +
				lipgloss.NewStyle().Foreground(colorText).Width(60).Render(h.CoreClaim),
		)

		expBox := cardStyle.Render(
			sectionTitle.Render("PROPOSED EXPERIMENT") + "\n\n" +
				lipgloss.NewStyle().Foreground(colorText).Width(60).Render(h.ProposedExperiment) +
				"\n\n" +
				muted.Render("Baseline  ") + lipgloss.NewStyle().Foreground(colorPrimary).Render(h.BaselineRepo) +
				"   " +
				muted.Render("Metric  ") + lipgloss.NewStyle().Foreground(colorPrimary).Render(h.Metric),
		)

		hint := helpBar("enter", "view sessions", "esc", "home")
		body = lipgloss.JoinVertical(lipgloss.Left,
			checkmark, "", meta, "", claimBox, "", expBox, "", hint,
		)

	case ingestError:
		errBox := cardStyle.Copy().BorderForeground(colorDanger).Render(
			dangerText.Render("✗  Error") + "\n\n" +
				lipgloss.NewStyle().Foreground(colorText).Width(60).Render(m.err.Error()),
		)
		hint := helpBar("enter", "try again", "esc", "home")
		body = lipgloss.JoinVertical(lipgloss.Left, errBox, "", hint)
	}

	return appStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left, header, "", body),
	)
}

func truncateStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

// Package tui implements the Research Loop terminal UI using Bubble Tea.
package tui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/research-loop/research-loop/internal/config"
)

// ─── Screen enum ─────────────────────────────────────────────────────────────

type screen int

const (
	screenHome screen = iota
	screenSetup
	screenIngest
	screenSessions
	screenDashboard
	screenDiscovery
	screenExplore
	screenQuit
)

// ─── Navigation message ───────────────────────────────────────────────────────

type navMsg struct{ to screen }

func navigateTo(s screen) tea.Cmd {
	return func() tea.Msg { return navMsg{to: s} }
}

// ─── Root model ───────────────────────────────────────────────────────────────

// rootModel is the top-level Bubble Tea model.
// It owns the active screen and routes messages down.
type rootModel struct {
	current   screen
	workspace string
	cfg       *config.Config

	// Sub-models (one per screen)
	home      homeModel
	setup     setupModel
	ingest    ingestModel
	sessions  sessionsModel
	dashboard dashboardModel
	discovery discoveryModel
	explore   exploreModel
}

func newRootModel(workspace string, cfg *config.Config) rootModel {
	return rootModel{
		current:   screenHome,
		workspace: workspace,
		cfg:       cfg,
		home:      newHomeModel(),
		setup:     newSetupModel(workspace),
		ingest:    newIngestModel(workspace, cfg),
		sessions:  newSessionsModel(workspace),
		dashboard: newDashboardModel(workspace),
		discovery: newDiscoveryModel(workspace),
		explore:   newExploreModel(workspace),
	}
}

func (m rootModel) Init() tea.Cmd {
	return tea.Batch(
		m.home.Init(),
		m.sessions.Init(),
		m.dashboard.Init(),
	)
}

func (m rootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Global quit
	if km, ok := msg.(tea.KeyMsg); ok {
		if km.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	// Navigation
	if nav, ok := msg.(navMsg); ok {
		m.current = nav.to
		switch nav.to {
		case screenSetup:
			m.setup = newSetupModel(m.workspace)
			return m, m.setup.Init()
		case screenIngest:
			m.ingest = newIngestModel(m.workspace, m.cfg)
			return m, m.ingest.Init()
		case screenSessions:
			m.sessions = newSessionsModel(m.workspace)
			return m, m.sessions.Init()
		case screenDashboard:
			m.dashboard = newDashboardModel(m.workspace)
			return m, m.dashboard.Init()
		case screenDiscovery:
			m.discovery = newDiscoveryModel(m.workspace)
			return m, m.discovery.Init()
		case screenExplore:
			m.explore = newExploreModel(m.workspace)
			return m, m.explore.Init()
		case screenHome:
			m.home = newHomeModel()
			return m, m.home.Init()
		case screenQuit:
			return m, tea.Quit
		}
	}

	// Delegate to active screen
	var cmd tea.Cmd
	switch m.current {
	case screenHome:
		updated, c := m.home.Update(msg)
		m.home = updated.(homeModel)
		cmd = c
	case screenSetup:
		updated, c := m.setup.Update(msg)
		m.setup = updated.(setupModel)
		cmd = c
	case screenIngest:
		updated, c := m.ingest.Update(msg)
		m.ingest = updated.(ingestModel)
		cmd = c
	case screenSessions:
		updated, c := m.sessions.Update(msg)
		m.sessions = updated.(sessionsModel)
		cmd = c
	case screenDashboard:
		updated, c := m.dashboard.Update(msg)
		m.dashboard = updated.(dashboardModel)
		cmd = c
	case screenDiscovery:
		updated, c := m.discovery.Update(msg)
		m.discovery = updated.(discoveryModel)
		cmd = c
	case screenExplore:
		updated, c := m.explore.Update(msg)
		m.explore = updated.(exploreModel)
		cmd = c
	}
	return m, cmd
}

func (m rootModel) View() string {
	switch m.current {
	case screenHome:
		return m.home.View()
	case screenSetup:
		return m.setup.View()
	case screenIngest:
		return m.ingest.View()
	case screenSessions:
		return m.sessions.View()
	case screenDashboard:
		return m.dashboard.View()
	case screenDiscovery:
		return m.discovery.View()
	case screenExplore:
		return m.explore.View()
	}
	return ""
}

// ─── Entry point ─────────────────────────────────────────────────────────────

// Run starts the Bubble Tea TUI.
func Run(workspaceRoot string) error {
	cfg, err := config.Load(workspaceRoot)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	model := newRootModel(workspaceRoot, cfg)
	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),       // full-screen takeover
		tea.WithMouseCellMotion(), // mouse support
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
		return err
	}
	return nil
}

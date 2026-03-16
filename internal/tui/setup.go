package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/research-loop/research-loop/internal/auth"
)

// ─── Setup states ────────────────────────────────────────────────────────────

type setupState int

const (
	setupSelectProvider setupState = iota // pick from list
	setupCLIProbing                       // running `claude` liveness probe
	setupKeyInput                         // paste/type API key
	setupLocalConfig                      // configure base URL for local providers
	setupVerifying                        // checking / saving credential
	setupDone                             // success
	setupFailed                           // error
)

// ─── Messages ────────────────────────────────────────────────────────────────

type setupVerifyMsg struct {
	ok  bool
	err error
}
type cliProbeMsg struct {
	cliPath string
	version string
	err     error
}

// ─── Model ───────────────────────────────────────────────────────────────────

type setupModel struct {
	workspace string
	state     setupState
	cursor    int
	provider  auth.Provider
	input     textinput.Model
	spinner   spinner.Model
	err       error
	cliPath   string // resolved path after successful CLI probe
}

func newSetupModel(workspace string) setupModel {
	ti := textinput.New()
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'
	ti.Width = 56
	ti.CharLimit = 512

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(colorPrimary)

	return setupModel{
		workspace: workspace,
		state:     setupSelectProvider,
		input:     ti,
		spinner:   sp,
	}
}

// ─── Commands ────────────────────────────────────────────────────────────────

// probeCLI checks the local `claude` binary exists and reports its version.
// This is the Paperclip approach: no OAuth, no browser — just verify the CLI works.
func probeCLI() tea.Cmd {
	return func() tea.Msg {
		result := auth.ClaudeProbe()
		return cliProbeMsg{cliPath: result.CLIPath, version: result.Version, err: result.Err}
	}
}

// saveCLIProvider stores the claude-code provider config (no token needed —
// the CLI handles auth itself via ~/.claude/ or ANTHROPIC_API_KEY).
func saveCLIProvider(workspace string, p auth.Provider, cliPath string) tea.Cmd {
	return func() tea.Msg {
		// Store the resolved CLI path as the credential value so the LLM
		// runner knows which binary to spawn.
		cred := auth.Credential{ProviderID: p.ID, Value: cliPath}
		if err := auth.Save(workspace, cred); err != nil {
			return setupVerifyMsg{ok: false, err: err}
		}
		if err := auth.SetActiveProvider(workspace, p.ID, cred); err != nil {
			return setupVerifyMsg{ok: false, err: err}
		}
		return setupVerifyMsg{ok: true}
	}
}

func saveAPIKey(workspace string, p auth.Provider, value string) tea.Cmd {
	return func() tea.Msg {
		cred := auth.Credential{ProviderID: p.ID, Value: value, BaseURL: p.BaseURL}
		if err := auth.Save(workspace, cred); err != nil {
			return setupVerifyMsg{ok: false, err: err}
		}
		if err := auth.SetActiveProvider(workspace, p.ID, cred); err != nil {
			return setupVerifyMsg{ok: false, err: err}
		}
		return setupVerifyMsg{ok: true}
	}
}

// ─── Update ──────────────────────────────────────────────────────────────────

func (m setupModel) Init() tea.Cmd { return nil }

func (m setupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch m.state {

		case setupSelectProvider:
			switch msg.String() {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(auth.AllProviders)-1 {
					m.cursor++
				}
			case "enter", " ":
				m.provider = auth.AllProviders[m.cursor]
				return m.transitionToAuth()
			case "esc", "q":
				return m, navigateTo(screenHome)
			}

		case setupCLIProbing:
			// User can only cancel while the probe is running
			switch msg.String() {
			case "esc", "q":
				m.state = setupSelectProvider
				return m, nil
			}

		case setupKeyInput, setupLocalConfig:
			switch msg.String() {
			case "enter":
				val := strings.TrimSpace(m.input.Value())
				if val == "" {
					return m, nil
				}
				if m.state == setupLocalConfig {
					m.provider.BaseURL = val
					val = "" // local providers don't need a key
				}
				m.state = setupVerifying
				return m, tea.Batch(
					m.spinner.Tick,
					saveAPIKey(m.workspace, m.provider, val),
				)
			case "esc":
				m.state = setupSelectProvider
				m.input.Reset()
				return m, nil
			}

		case setupDone:
			switch msg.String() {
			case "enter", "esc", "q":
				return m, navigateTo(screenHome)
			}

		case setupFailed:
			switch msg.String() {
			case "enter":
				m.state = setupSelectProvider
				m.err = nil
				m.input.Reset()
				return m, nil
			case "esc", "q":
				return m, navigateTo(screenHome)
			}
		}

	// ── CLI probe result ──────────────────────────────────────────────────────

	case cliProbeMsg:
		if msg.err != nil {
			m.err = msg.err
			m.state = setupFailed
			return m, nil
		}
		// Probe passed — save config (store "path::version" as value)
		m.cliPath = msg.cliPath
		value := msg.cliPath
		if msg.version != "" {
			value = msg.cliPath + "::" + msg.version
		}
		m.state = setupVerifying
		return m, tea.Batch(
			m.spinner.Tick,
			saveCLIProvider(m.workspace, m.provider, value),
		)

	// ── Save result ───────────────────────────────────────────────────────────

	case setupVerifyMsg:
		if msg.ok {
			m.state = setupDone
		} else {
			m.err = msg.err
			m.state = setupFailed
		}
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	// Delegate text input
	if m.state == setupKeyInput || m.state == setupLocalConfig {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m setupModel) transitionToAuth() (tea.Model, tea.Cmd) {
	switch m.provider.AuthType {
	case auth.AuthTypeCLI:
		// Paperclip approach: no browser, no token — just probe the local CLI.
		m.state = setupCLIProbing
		return m, tea.Batch(m.spinner.Tick, probeCLI())

	case auth.AuthTypeAPIKey:
		m.state = setupKeyInput
		m.input.Placeholder = fmt.Sprintf("Paste your %s…", m.provider.KeyLabel)
		m.input.Focus()
		return m, textinput.Blink

	case auth.AuthTypeLocal:
		m.state = setupLocalConfig
		m.input.EchoMode = textinput.EchoNormal
		m.input.Placeholder = m.provider.BaseURL
		m.input.SetValue(m.provider.BaseURL)
		m.input.Focus()
		return m, textinput.Blink
	}
	return m, nil
}

// ─── View ────────────────────────────────────────────────────────────────────

func (m setupModel) View() string {
	header := headerStyle.Render("🔬  Research Loop  /  Setup Provider")

	var body string
	switch m.state {
	case setupSelectProvider:
		body = m.viewSelectProvider()
	case setupCLIProbing:
		body = m.viewCLIProbing()
	case setupKeyInput:
		body = m.viewKeyInput()
	case setupLocalConfig:
		body = m.viewLocalConfig()
	case setupVerifying:
		body = m.spinner.View() + "  " + primaryText.Render("Saving configuration…")
	case setupDone:
		body = m.viewDone()
	case setupFailed:
		body = m.viewFailed()
	}

	return appStyle.Render(lipgloss.JoinVertical(lipgloss.Left, header, "", body))
}

func (m setupModel) viewSelectProvider() string {
	title := primaryText.Render("Choose your model provider")
	sub := muted.Render("Research Loop will use this to extract hypotheses and propose experiments")

	items := ""
	for i, p := range auth.AllProviders {
		var authBadge string
		switch p.AuthType {
		case auth.AuthTypeCLI:
			authBadge = badgeBlue.Render(" cli ")
		case auth.AuthTypeAPIKey:
			authBadge = badgeGray.Render(" api key ")
		case auth.AuthTypeLocal:
			authBadge = badgeGreen.Render(" local ")
		}

		var line string
		if i == m.cursor {
			arrow := primaryText.Render("▶")
			name := selectedItem.Render(p.Name)
			desc := lipgloss.NewStyle().Foreground(colorMuted).PaddingLeft(4).Render(p.Description)
			line = fmt.Sprintf("%s%s  %s\n%s", arrow, name, authBadge, desc)
		} else {
			name := normalItem.Render(p.Name)
			line = fmt.Sprintf("%s  %s", name, authBadge)
		}

		if i < len(auth.AllProviders)-1 {
			items += line + "\n\n"
		} else {
			items += line
		}
	}

	card := cardStyle.Render(sectionTitle.Render("PROVIDERS") + "\n\n" + items)
	hint := helpBar("↑↓", "navigate", "enter", "select", "esc", "back")
	return lipgloss.JoinVertical(lipgloss.Left, title, sub, "", card, "", hint)
}

func (m setupModel) viewCLIProbing() string {
	spin := m.spinner.View()
	title := primaryText.Render("Checking Claude Code CLI…")

	steps := cardStyle.Render(
		sectionTitle.Render("CLI CHECK") + "\n\n" +
			spin + "  " + primaryText.Render("Locating `claude` binary…") + "\n\n" +
			dimText.Render("  No browser, no OAuth — auth is handled by your") + "\n" +
			dimText.Render("  existing `claude` setup (~/.claude/ or ANTHROPIC_API_KEY).") + "\n\n" +
			muted.Render("  If `claude` is not found, install it first:") + "\n" +
			lipgloss.NewStyle().Foreground(colorPrimary).Render("  https://claude.ai/download"),
	)

	hint := helpBar("esc", "cancel")
	return lipgloss.JoinVertical(lipgloss.Left, title, "", steps, "", hint)
}

func (m setupModel) viewKeyInput() string {
	p := m.provider
	title := primaryText.Render("Enter your " + p.Name + " " + p.KeyLabel)

	var envHint string
	if p.KeyEnv != "" {
		envHint = "\n" + muted.Render("  Or set environment variable: ") + keyLabel.Render(p.KeyEnv)
	}

	inputBox := inputStyle.Render(m.input.View())
	hint := helpBar("enter", "confirm", "esc", "back")
	return lipgloss.JoinVertical(lipgloss.Left, title, envHint, "", inputBox, "", hint)
}

func (m setupModel) viewLocalConfig() string {
	title := primaryText.Render("Configure " + m.provider.Name)
	sub := muted.Render("Enter the base URL where " + m.provider.Name + " is running")
	inputBox := inputStyle.Render(m.input.View())
	hint := helpBar("enter", "confirm", "esc", "back")
	return lipgloss.JoinVertical(lipgloss.Left, title, sub, "", inputBox, "", hint)
}

func (m setupModel) viewDone() string {
	p := m.provider
	check := successText.Render("✓  " + p.Name + " connected")

	authMethod := "CLI (existing claude auth)"
	if p.AuthType == auth.AuthTypeAPIKey {
		authMethod = "API key"
	} else if p.AuthType == auth.AuthTypeLocal {
		authMethod = "Local (no auth)"
	}

	cliLine := ""
	if p.AuthType == auth.AuthTypeCLI && m.cliPath != "" {
		cliLine = "\n" + muted.Render(fmt.Sprintf("  %-14s", "CLI path")) +
			lipgloss.NewStyle().Foreground(colorText).Render(m.cliPath)
	}

	details := cardStyle.Copy().BorderForeground(colorSuccess).Render(
		sectionTitle.Render("CONFIGURED") + "\n\n" +
			muted.Render(fmt.Sprintf("  %-14s", "Provider")) + lipgloss.NewStyle().Foreground(colorText).Render(p.Name) + "\n" +
			muted.Render(fmt.Sprintf("  %-14s", "Model")) + lipgloss.NewStyle().Foreground(colorPrimary).Render(p.DefaultModel) + "\n" +
			muted.Render(fmt.Sprintf("  %-14s", "Auth")) + lipgloss.NewStyle().Foreground(colorText).Render(authMethod) +
			cliLine + "\n" +
			muted.Render(fmt.Sprintf("  %-14s", "Credential")) + successText.Render("saved to .research-loop/credentials.toml"),
	)

	next := muted.Render("You can now start a new investigation.")
	hint := helpBar("enter", "back to home")
	return lipgloss.JoinVertical(lipgloss.Left, check, "", details, "", next, "", hint)
}

func (m setupModel) viewFailed() string {
	errBox := cardStyle.Copy().BorderForeground(colorDanger).Render(
		dangerText.Render("✗  Setup failed") + "\n\n" +
			lipgloss.NewStyle().Foreground(colorText).Width(60).Render(m.err.Error()),
	)
	hint := helpBar("enter", "try again", "esc", "home")
	return lipgloss.JoinVertical(lipgloss.Left, errBox, "", hint)
}

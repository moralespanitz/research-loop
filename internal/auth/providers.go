// Package auth manages provider credentials for Research Loop.
// Credentials are stored in .research-loop/credentials.toml (chmod 600).
package auth

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Provider represents a supported LLM/agent provider.
type Provider struct {
	ID          string // internal key, e.g. "claude-code"
	Name        string // display name
	Description string
	AuthType    AuthType
	AuthURL     string // for browser-based OAuth
	KeyEnv      string // env var to check first
	KeyLabel    string // label shown in TUI ("API Key", "Token", etc.)
	BaseURL     string // for local providers
	DefaultModel string
}

type AuthType int

const (
	AuthTypeBrowser AuthType = iota // open browser → paste token back (unused for Claude)
	AuthTypeAPIKey                  // paste key directly
	AuthTypeLocal                   // no auth needed (Ollama, LM Studio)
	AuthTypeCLI                     // spawn local CLI — auth already handled by the CLI itself
)

// Credential holds a saved provider credential.
type Credential struct {
	ProviderID string
	Value      string // token or API key
	BaseURL    string // for local/custom providers
}

// AllProviders is the ordered list shown in the setup screen.
var AllProviders = []Provider{
	{
		ID:           "claude-code",
		Name:         "Claude Code",
		Description:  "Spawns the local `claude` CLI — uses your existing claude auth",
		AuthType:     AuthTypeCLI,
		KeyEnv:       "ANTHROPIC_API_KEY",
		KeyLabel:     "API Key",
		DefaultModel: "claude-sonnet-4-5",
	},
	{
		ID:           "openai",
		Name:         "OpenAI / Codex",
		Description:  "GPT-4/5 and Codex via OpenAI API key",
		AuthType:     AuthTypeAPIKey,
		KeyEnv:       "OPENAI_API_KEY",
		KeyLabel:     "API Key (sk-...)",
		DefaultModel: "gpt-4o",
	},
	{
		ID:           "opencode",
		Name:         "OpenCode",
		Description:  "Open-source coding agent (OpenAI-compatible endpoint)",
		AuthType:     AuthTypeAPIKey,
		KeyEnv:       "OPENCODE_API_KEY",
		KeyLabel:     "API Key",
		DefaultModel: "gpt-4o",
	},
	{
		ID:           "gemini",
		Name:         "Google Gemini",
		Description:  "Google's Gemini models via API key",
		AuthType:     AuthTypeAPIKey,
		KeyEnv:       "GEMINI_API_KEY",
		KeyLabel:     "API Key",
		DefaultModel: "gemini-2.0-flash",
	},
	{
		ID:           "ollama",
		Name:         "Ollama (local)",
		Description:  "Run models locally — no API key needed",
		AuthType:     AuthTypeLocal,
		BaseURL:      "http://localhost:11434",
		DefaultModel: "llama3",
	},
	{
		ID:           "lmstudio",
		Name:         "LM Studio (local)",
		Description:  "Local models via LM Studio OpenAI-compatible server",
		AuthType:     AuthTypeLocal,
		BaseURL:      "http://localhost:1234/v1",
		DefaultModel: "local-model",
	},
}

// ProviderByID returns a provider by its ID.
func ProviderByID(id string) (Provider, bool) {
	for _, p := range AllProviders {
		if p.ID == id {
			return p, true
		}
	}
	return Provider{}, false
}

// ─── CLI probe ────────────────────────────────────────────────────────────────

// ClaudeProbeResult describes the outcome of a live claude CLI check.
type ClaudeProbeResult struct {
	CLIPath string // resolved path to the claude binary
	Version string // e.g. "2.1.76 (Claude Code)"
	Err     error  // nil = CLI found and responsive
}

// ClaudeProbe checks whether the `claude` CLI is installed and can run.
// Uses the Paperclip approach: no OAuth, no browser — just verify the binary
// exists and responds. Two-step:
//  1. Resolve the binary path via PATH
//  2. Run `claude --version` (fast, no auth required) to confirm it works
//
// We intentionally do NOT run the full streaming probe in the TUI setup flow
// because that would require a live API call and could open a browser if the
// user hasn't authenticated yet. Auth is already handled by `~/.claude/` or
// ANTHROPIC_API_KEY — the CLI will use those automatically when it's invoked
// later by the research runner.
func ClaudeProbe() ClaudeProbeResult {
	path, err := exec.LookPath("claude")
	if err != nil {
		return ClaudeProbeResult{Err: fmt.Errorf("`claude` not found on PATH\n\nInstall Claude Code: https://claude.ai/download\nThen run `claude` once to authenticate.")}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	out, err := exec.CommandContext(ctx, path, "--version").Output()
	if err != nil {
		return ClaudeProbeResult{CLIPath: path, Err: fmt.Errorf("`claude --version` failed: %w\n\nMake sure Claude Code CLI is properly installed.", err)}
	}

	version := strings.TrimSpace(string(out))
	return ClaudeProbeResult{CLIPath: path, Version: version}
}

// ─── Credential storage ───────────────────────────────────────────────────────

// credentialsPath returns the path to the credentials file.
func credentialsPath(workspaceRoot string) string {
	return filepath.Join(workspaceRoot, ".research-loop", "credentials.toml")
}

// Save writes a credential to disk (chmod 600).
func Save(workspaceRoot string, cred Credential) error {
	path := credentialsPath(workspaceRoot)

	// Load existing
	creds := loadAll(path)

	// Upsert
	found := false
	for i, c := range creds {
		if c.ProviderID == cred.ProviderID {
			creds[i] = cred
			found = true
			break
		}
	}
	if !found {
		creds = append(creds, cred)
	}

	return write(path, creds)
}

// Load returns the credential for a provider, checking env vars first.
func Load(workspaceRoot, providerID string) (Credential, bool) {
	p, ok := ProviderByID(providerID)
	if !ok {
		return Credential{}, false
	}

	// Env var takes priority
	if p.KeyEnv != "" {
		if val := os.Getenv(p.KeyEnv); val != "" {
			return Credential{ProviderID: providerID, Value: val}, true
		}
	}

	// Fall back to stored credentials
	creds := loadAll(credentialsPath(workspaceRoot))
	for _, c := range creds {
		if c.ProviderID == providerID {
			return c, true
		}
	}
	return Credential{}, false
}

// ActiveProvider returns the configured provider ID from config.
func ActiveProvider(workspaceRoot string) string {
	data, _ := os.ReadFile(filepath.Join(workspaceRoot, ".research-loop", "config.toml"))
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "provider") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				return strings.Trim(strings.TrimSpace(parts[1]), `"`)
			}
		}
	}
	return ""
}

// SetActiveProvider updates the provider in config.toml.
func SetActiveProvider(workspaceRoot, providerID string, cred Credential) error {
	p, ok := ProviderByID(providerID)
	if !ok {
		return fmt.Errorf("unknown provider: %s", providerID)
	}

	cfgPath := filepath.Join(workspaceRoot, ".research-loop", "config.toml")
	data, _ := os.ReadFile(cfgPath)
	lines := strings.Split(string(data), "\n")

	newLines := []string{}
	inLLM := false
	updated := map[string]bool{}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "[llm]" {
			inLLM = true
			newLines = append(newLines, line)
			continue
		}
		if strings.HasPrefix(trimmed, "[") && trimmed != "[llm]" {
			// Flush any missing fields before leaving [llm]
			if inLLM {
				if !updated["provider"] {
					newLines = append(newLines, fmt.Sprintf(`provider     = "%s"`, providerID))
				}
				if !updated["model"] {
					newLines = append(newLines, fmt.Sprintf(`model        = "%s"`, p.DefaultModel))
				}
			}
			inLLM = false
			newLines = append(newLines, line)
			continue
		}
		if inLLM {
			if strings.HasPrefix(trimmed, "provider") {
				newLines = append(newLines, fmt.Sprintf(`provider     = "%s"`, providerID))
				updated["provider"] = true
				continue
			}
			if strings.HasPrefix(trimmed, "model") {
				newLines = append(newLines, fmt.Sprintf(`model        = "%s"`, p.DefaultModel))
				updated["model"] = true
				continue
			}
			if strings.HasPrefix(trimmed, "api_key_env") && cred.Value != "" {
				newLines = append(newLines, fmt.Sprintf(`api_key_env  = "%s"`, p.KeyEnv))
				updated["api_key_env"] = true
				continue
			}
			if strings.HasPrefix(trimmed, "base_url") && p.BaseURL != "" {
				newLines = append(newLines, fmt.Sprintf(`base_url     = "%s"`, p.BaseURL))
				updated["base_url"] = true
				continue
			}
		}
		newLines = append(newLines, line)
	}

	return os.WriteFile(cfgPath, []byte(strings.Join(newLines, "\n")), 0644)
}

// ─── Simple TOML-like credential store ───────────────────────────────────────

func loadAll(path string) []Credential {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var creds []Credential
	var current *Credential
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "[[credential]]") {
			if current != nil {
				creds = append(creds, *current)
			}
			current = &Credential{}
			continue
		}
		if current == nil {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		k := strings.TrimSpace(parts[0])
		v := strings.Trim(strings.TrimSpace(parts[1]), `"`)
		switch k {
		case "provider_id":
			current.ProviderID = v
		case "value":
			current.Value = v
		case "base_url":
			current.BaseURL = v
		}
	}
	if current != nil {
		creds = append(creds, *current)
	}
	return creds
}

func write(path string, creds []Credential) error {
	var sb strings.Builder
	sb.WriteString("# Research Loop credentials\n# DO NOT COMMIT THIS FILE\n\n")
	for _, c := range creds {
		sb.WriteString("[[credential]]\n")
		sb.WriteString(fmt.Sprintf("provider_id = %q\n", c.ProviderID))
		sb.WriteString(fmt.Sprintf("value       = %q\n", c.Value))
		if c.BaseURL != "" {
			sb.WriteString(fmt.Sprintf("base_url    = %q\n", c.BaseURL))
		}
		sb.WriteString("\n")
	}
	if err := os.WriteFile(path, []byte(sb.String()), 0600); err != nil {
		return err
	}
	return nil
}

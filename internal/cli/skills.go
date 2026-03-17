package cli

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	embedassets "github.com/research-loop/research-loop/internal/embed"
	"github.com/spf13/cobra"
)

func newSkillsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "skills",
		Short: "Install Research Loop Claude Code integration into the current project",
		Long: `Installs skills, slash commands, and hooks into .claude/ in the current directory.

All files are merged idempotently — existing CLAUDE.md content and settings.json
hooks are preserved. Running this command twice is safe.`,
		RunE: runSkills,
	}
}

func runSkills(_ *cobra.Command, _ []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	claudeDir := filepath.Join(cwd, ".claude")
	claudeMdPath := filepath.Join(cwd, "CLAUDE.md")

	fmt.Println()

	err = fs.WalkDir(embedassets.Assets, "claude", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}

		rel, _ := filepath.Rel("claude", path)

		data, readErr := embedassets.Assets.ReadFile(path)
		if readErr != nil {
			return readErr
		}

		// CLAUDE.md lives at the project root, not inside .claude/
		if rel == "CLAUDE.md" {
			return mergeCLAUDEMd(claudeMdPath, data)
		}

		dst := filepath.Join(claudeDir, rel)
		if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
			return err
		}

		if rel == "settings.json" {
			return mergeSettingsJSON(dst, data)
		}

		if err := os.WriteFile(dst, data, 0644); err != nil {
			return err
		}
		if strings.HasSuffix(dst, ".sh") {
			_ = os.Chmod(dst, 0755)
		}

		fmt.Printf("  ✓  .claude/%s\n", rel)
		return nil
	})
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("  Research Loop Claude Code integration installed.")
	fmt.Println("  Reload Claude Code (or open a new session) to activate.")
	fmt.Println()
	return nil
}

// mergeCLAUDEMd appends the research-loop block to CLAUDE.md if not already present.
const claudeMdSentinel = "<!-- research-loop:begin -->"

func mergeCLAUDEMd(path string, block []byte) error {
	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if strings.Contains(string(existing), claudeMdSentinel) {
		fmt.Printf("  ·  CLAUDE.md (research-loop block already present — skipped)\n")
		return nil
	}

	var content []byte
	if len(existing) > 0 {
		content = append(existing, '\n')
	}
	content = append(content, block...)

	if err := os.WriteFile(path, content, 0644); err != nil {
		return err
	}
	fmt.Printf("  ✓  CLAUDE.md\n")
	return nil
}

// ─── settings.json merge ──────────────────────────────────────────────────────

type settingsHook struct {
	Type    string `json:"type"`
	Command string `json:"command"`
	Async   bool   `json:"async,omitempty"`
}

type settingsHookGroup struct {
	Matcher string         `json:"matcher,omitempty"`
	Hooks   []settingsHook `json:"hooks"`
}

type settingsFile struct {
	Hooks map[string][]settingsHookGroup `json:"hooks"`
}

// mergeSettingsJSON merges hook entries from src into the existing settings.json.
// Hook commands already present (by exact match) are not duplicated.
func mergeSettingsJSON(path string, src []byte) error {
	var incoming settingsFile
	if err := json.Unmarshal(src, &incoming); err != nil {
		return fmt.Errorf("parsing embedded settings.json: %w", err)
	}

	var current settingsFile
	if raw, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(raw, &current)
	}
	if current.Hooks == nil {
		current.Hooks = make(map[string][]settingsHookGroup)
	}

	changed := false
	for event, incomingGroups := range incoming.Hooks {
		for _, ig := range incomingGroups {
			for _, ih := range ig.Hooks {
				if !hookCommandExists(current.Hooks[event], ih.Command) {
					current.Hooks[event] = appendHookToGroups(current.Hooks[event], ig.Matcher, ih)
					changed = true
				}
			}
		}
	}

	if !changed {
		fmt.Printf("  ·  .claude/settings.json (hooks already present — skipped)\n")
		return nil
	}

	out, err := json.MarshalIndent(current, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, append(out, '\n'), 0644); err != nil {
		return err
	}
	fmt.Printf("  ✓  .claude/settings.json\n")
	return nil
}

func hookCommandExists(groups []settingsHookGroup, cmd string) bool {
	for _, g := range groups {
		for _, h := range g.Hooks {
			if h.Command == cmd {
				return true
			}
		}
	}
	return false
}

func appendHookToGroups(groups []settingsHookGroup, matcher string, h settingsHook) []settingsHookGroup {
	for i, g := range groups {
		if g.Matcher == matcher {
			groups[i].Hooks = append(groups[i].Hooks, h)
			return groups
		}
	}
	return append(groups, settingsHookGroup{Matcher: matcher, Hooks: []settingsHook{h}})
}

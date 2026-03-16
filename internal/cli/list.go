package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/research-loop/research-loop/internal/config"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all research sessions in the current workspace",
		RunE: func(cmd *cobra.Command, args []string) error {
			root := config.WorkspaceRoot()
			sessionsDir := filepath.Join(root, ".research-loop", "sessions")

			entries, err := os.ReadDir(sessionsDir)
			if err != nil {
				if os.IsNotExist(err) {
					fmt.Println("No sessions yet. Run 'research-loop start <url>' to begin.")
					return nil
				}
				return fmt.Errorf("reading sessions: %w", err)
			}

			if len(entries) == 0 {
				fmt.Println("No sessions yet. Run 'research-loop start <url>' to begin.")
				return nil
			}

			fmt.Printf("%-50s  %s\n", "SESSION", "HYPOTHESIS")
			fmt.Printf("%-50s  %s\n", strings.Repeat("─", 50), strings.Repeat("─", 40))
			for _, e := range entries {
				if !e.IsDir() {
					continue
				}
				hypPath := filepath.Join(sessionsDir, e.Name(), "hypothesis.md")
				title := readFirstHeading(hypPath)
				fmt.Printf("%-50s  %s\n", e.Name(), title)
			}
			fmt.Println()
			return nil
		},
	}
}

func readFirstHeading(path string) string {
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

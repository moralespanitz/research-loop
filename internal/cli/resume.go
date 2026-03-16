package cli

import (
	"fmt"

	"github.com/research-loop/research-loop/internal/config"
	"github.com/research-loop/research-loop/internal/persistence"
	"github.com/spf13/cobra"
)

func newResumeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "resume <session-id>",
		Short: "Resume a paused research session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := config.WorkspaceRoot()
			session, err := persistence.LoadSession(root, args[0])
			if err != nil {
				return err
			}

			printSuccess("Session loaded")
			fmt.Printf("  Session : %s\n", session.ID)
			fmt.Printf("  Dir     : %s\n", session.Root)
			fmt.Println()
			fmt.Println("Files:")
			fmt.Printf("  hypothesis.md      → %s/hypothesis.md\n", session.Root)
			fmt.Printf("  knowledge_graph.md → %s/knowledge_graph.md\n", session.Root)
			fmt.Printf("  lab_notebook.md    → %s/lab_notebook.md\n", session.Root)
			fmt.Printf("  autoresearch.jsonl → %s/autoresearch.jsonl\n", session.Root)
			fmt.Println()
			fmt.Println("Experiment loop coming in v0.2 — for now, inspect your session files above.")
			fmt.Println()
			return nil
		},
	}
}

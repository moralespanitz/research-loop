package cli

import (
	"fmt"

	"github.com/research-loop/research-loop/internal/config"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize a Research Loop workspace in the current directory",
		Long: `Creates a .research-loop/ directory with a default config.toml.

Run this once per project before using 'research-loop start'.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			root := config.WorkspaceRoot()
			if err := config.Init(root); err != nil {
				return fmt.Errorf("initializing workspace: %w", err)
			}
			printSuccess("Workspace initialized")
			fmt.Printf("  Location : %s/.research-loop/\n\n", root)
			fmt.Println("Configure your LLM backend:")
			fmt.Printf("  edit %s/.research-loop/config.toml\n\n", root)
			fmt.Println("Or set your API key directly:")
			fmt.Println("  export ANTHROPIC_API_KEY=sk-...")
			fmt.Println()
			fmt.Println("Then start your first investigation:")
			fmt.Println("  research-loop start \"https://arxiv.org/abs/2403.05821\"")
			fmt.Println()
			return nil
		},
	}
}

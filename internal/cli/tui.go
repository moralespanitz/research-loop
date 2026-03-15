package cli

import (
	"github.com/research-loop/research-loop/internal/config"
	researchtui "github.com/research-loop/research-loop/internal/tui"
	"github.com/spf13/cobra"
)

func newTUICmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tui",
		Short: "Open the interactive terminal UI",
		Long: `Launch the full-screen Research Loop terminal interface.

Navigate with arrow keys. Start a new investigation, browse sessions,
and monitor live experiment dashboards — all from the terminal.

  ↑↓ / j k   navigate
  enter       select
  esc / q     go back / home
  ctrl+c      quit`,
		RunE: func(cmd *cobra.Command, args []string) error {
			root := config.WorkspaceRoot()
			return researchtui.Run(root)
		},
	}
}

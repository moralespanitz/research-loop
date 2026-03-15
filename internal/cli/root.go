// Package cli wires together all Research Loop commands.
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Execute runs the root command. Called from main().
func Execute() {
	root := newRootCmd()
	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "research-loop",
		Short: "The Researcher's Agent OS",
		Long: `Research Loop — an open-source Agent OS for scientific researchers.

Point it at a paper. Get a running experiment loop.

  research-loop init                         Initialize a workspace
  research-loop start <arxiv-url>            Ingest a paper and extract a hypothesis
  research-loop list                         List all sessions
  research-loop resume <session-id>          Resume a paused session
  research-loop export [--session <id>]      Export a .research bundle

Run 'research-loop <command> --help' for more information.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(
		newInitCmd(),
		newStartCmd(),
		newListCmd(),
		newResumeCmd(),
		newExportCmd(),
		newMCPCmd(),
		newDashboardCmd(),
		newTUICmd(),
	)

	return cmd
}

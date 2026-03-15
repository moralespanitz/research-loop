package cli

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/research-loop/research-loop/internal/config"
	"github.com/research-loop/research-loop/internal/server"
	"github.com/spf13/cobra"
)

func newDashboardCmd() *cobra.Command {
	var port int
	var open bool

	cmd := &cobra.Command{
		Use:   "dashboard",
		Short: "Start the Research Loop web dashboard",
		Long: `Start the dashboard at http://localhost:<port>.

The dashboard shows all your research sessions, live experiment metrics,
knowledge graphs, and lab notebooks. It connects to Claude Code via the
MCP server automatically.

Examples:
  research-loop dashboard
  research-loop dashboard --port 3000
  research-loop dashboard --open`,
		RunE: func(cmd *cobra.Command, args []string) error {
			root := config.WorkspaceRoot()

			// Ensure workspace is initialized
			if err := config.Init(root); err != nil {
				return fmt.Errorf("initializing workspace: %w", err)
			}

			srv, err := server.New(root, port)
			if err != nil {
				return fmt.Errorf("creating server: %w", err)
			}

			if open {
				go openBrowser(fmt.Sprintf("http://localhost:%d", port))
			}

			return srv.ListenAndServe()
		},
	}

	cmd.Flags().IntVarP(&port, "port", "p", 4321, "Port to listen on")
	cmd.Flags().BoolVarP(&open, "open", "o", false, "Open dashboard in browser automatically")
	return cmd
}

func openBrowser(url string) {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
		args = []string{url}
	case "linux":
		cmd = "xdg-open"
		args = []string{url}
	default:
		return
	}
	_ = exec.Command(cmd, args...).Start()
}

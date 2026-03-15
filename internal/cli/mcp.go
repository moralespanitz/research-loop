package cli

import (
	"fmt"

	researchmcp "github.com/research-loop/research-loop/internal/mcp"
	"github.com/research-loop/research-loop/internal/config"
	"github.com/spf13/cobra"
)

func newMCPCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "MCP server commands",
		Long:  `Commands for running Research Loop as a Model Context Protocol (MCP) server.`,
	}
	cmd.AddCommand(newMCPServeCmd())
	return cmd
}

func newMCPServeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start the Research Loop MCP server (stdio transport)",
		Long: `Start Research Loop as an MCP server using the stdio transport.

Connect to Claude Code with:

  claude mcp add research-loop -- /path/to/research-loop mcp serve

Or add a .mcp.json to your project root for automatic registration:

  {
    "mcpServers": {
      "research-loop": {
        "type": "stdio",
        "command": "/path/to/research-loop",
        "args": ["mcp", "serve"]
      }
    }
  }

Once connected, Claude Code gains access to these tools:
  research_ingest_paper    Ingest a paper from ArXiv URL → hypothesis.md
  research_session_status  List all sessions
  research_read_hypothesis Read the current hypothesis
  research_read_notebook   Read the lab notebook
  research_kg_query        Query the knowledge graph
  research_update_kg       Append a causal annotation
  research_export_bundle   Export a .research bundle

And these resources (readable with @ mentions in Claude Code):
  research://<session-id>/hypothesis.md
  research://<session-id>/knowledge_graph.md
  research://<session-id>/lab_notebook.md`,
		RunE: func(cmd *cobra.Command, args []string) error {
			root := config.WorkspaceRoot()
			srv, err := researchmcp.New(root)
			if err != nil {
				return fmt.Errorf("starting MCP server: %w", err)
			}
			return srv.Serve()
		},
	}
}

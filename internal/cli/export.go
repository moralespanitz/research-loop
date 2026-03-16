package cli

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/research-loop/research-loop/internal/config"
	"github.com/research-loop/research-loop/internal/persistence"
	"github.com/spf13/cobra"
)

func newExportCmd() *cobra.Command {
	var sessionID string
	var output string

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export a session as a portable .research bundle",
		Long: `Packages the full session into a ZIP-compatible .research archive.

The bundle contains: hypothesis.md, knowledge_graph.md, lab_notebook.md,
autoresearch.jsonl, autoresearch.md, and a JSON manifest.

Any researcher or agent can load this bundle with:
  research-loop resume ./bundle.research`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExport(sessionID, output)
		},
	}

	cmd.Flags().StringVarP(&sessionID, "session", "s", "", "Session ID to export (defaults to most recent)")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file path (defaults to <session-id>.research)")
	return cmd
}

func runExport(sessionID, output string) error {
	root := config.WorkspaceRoot()

	// Find session
	if sessionID == "" {
		id, err := mostRecentSession(root)
		if err != nil {
			return err
		}
		sessionID = id
	}

	session, err := persistence.LoadSession(root, sessionID)
	if err != nil {
		return err
	}

	if output == "" {
		output = sessionID + ".research"
	}

	// Create bundle
	f, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("creating bundle file: %w", err)
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	defer zw.Close()

	// Files to include
	files := []string{
		"hypothesis.md",
		"knowledge_graph.md",
		"lab_notebook.md",
		"autoresearch.jsonl",
		"autoresearch.md",
	}

	bundledFiles := []string{}
	for _, name := range files {
		path := filepath.Join(session.Root, name)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue // skip missing files
		}
		if err := addFileToZip(zw, path, name); err != nil {
			return fmt.Errorf("adding %s to bundle: %w", name, err)
		}
		bundledFiles = append(bundledFiles, name)
	}

	// Write manifest
	manifest := map[string]interface{}{
		"version":    "0.1.0",
		"session_id": session.ID,
		"created_at": time.Now().UTC().Format(time.RFC3339),
		"files":      bundledFiles,
		"format":     "research-loop-bundle",
	}
	mdata, _ := json.MarshalIndent(manifest, "", "  ")
	mw, err := zw.Create("manifest.json")
	if err != nil {
		return err
	}
	mw.Write(mdata)

	printSuccess("Bundle exported")
	fmt.Printf("  File    : %s\n", output)
	fmt.Printf("  Session : %s\n", session.ID)
	fmt.Printf("  Files   : %s\n", strings.Join(bundledFiles, ", "))
	fmt.Println()
	fmt.Println("Share it:")
	fmt.Printf("  research-loop resume ./%s\n", output)
	fmt.Println()
	return nil
}

func addFileToZip(zw *zip.Writer, srcPath, destName string) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := zw.Create(destName)
	if err != nil {
		return err
	}

	_, err = io.Copy(dst, src)
	return err
}

func mostRecentSession(workspaceRoot string) (string, error) {
	dir := filepath.Join(workspaceRoot, ".research-loop", "sessions")
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) == 0 {
		return "", fmt.Errorf("no sessions found; run 'research-loop start <url>' first")
	}
	// ReadDir returns entries sorted by name; last entry is most recent (slug ends with date)
	for i := len(entries) - 1; i >= 0; i-- {
		if entries[i].IsDir() {
			return entries[i].Name(), nil
		}
	}
	return "", fmt.Errorf("no session directories found")
}

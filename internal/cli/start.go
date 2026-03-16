package cli

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/research-loop/research-loop/internal/config"
	"github.com/research-loop/research-loop/internal/hypothesis"
	"github.com/research-loop/research-loop/internal/ingestion"
	"github.com/research-loop/research-loop/internal/llm"
	"github.com/research-loop/research-loop/internal/persistence"
	"github.com/spf13/cobra"
)

func newStartCmd() *cobra.Command {
	var approveAll bool

	cmd := &cobra.Command{
		Use:   "start <arxiv-url-or-local-pdf>",
		Short: "Start a new research investigation from a paper",
		Long: `Download and ingest a paper, extract a hypothesis, and initialize a new session.

Examples:
  research-loop start "https://arxiv.org/abs/2403.05821"
  research-loop start 2403.05821
  research-loop start ./my-paper.pdf`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStart(args[0], approveAll)
		},
	}

	cmd.Flags().BoolVarP(&approveAll, "yes", "y", false, "Auto-approve the extracted hypothesis without prompting")
	return cmd
}

func runStart(input string, approveAll bool) error {
	root := config.WorkspaceRoot()
	cfg, err := config.Load(root)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// ── Step 1: Download / locate the paper ──────────────────────────────────
	printStep("Fetching paper...")
	t0 := time.Now()

	paperDir := root + "/.research-loop/library/papers"
	if err := os.MkdirAll(paperDir, 0755); err != nil {
		return err
	}

	var paper *ingestion.Paper
	if strings.HasSuffix(strings.ToLower(input), ".pdf") || fileExists(input) {
		paper, err = ingestion.FetchLocalPDF(input)
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		paper, err = ingestion.FetchArXiv(ctx, input, paperDir)
	}
	if err != nil {
		return fmt.Errorf("fetching paper: %w", err)
	}

	printDone(t0)
	if paper.FullText == "" {
		printWarning("Could not extract full text — working from abstract only")
	}

	// ── Step 2: Extract hypothesis via LLM ───────────────────────────────────
	printStep("Extracting hypothesis...")
	t1 := time.Now()

	llmClient, err := llm.New(cfg.LLM)
	if err != nil {
		return fmt.Errorf("initializing LLM client: %w\n\nMake sure your API key is set:\n  export ANTHROPIC_API_KEY=sk-...", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	h, err := hypothesis.Extract(ctx, llmClient, paper)
	if err != nil {
		return fmt.Errorf("extracting hypothesis: %w", err)
	}

	printDone(t1)

	// ── Step 3: Present to researcher ────────────────────────────────────────
	printHypothesisBox(h)

	if !approveAll {
		if !confirm("Approve this hypothesis and initialize the session?") {
			fmt.Println("\nAborted. You can edit the paper or try again.")
			return nil
		}
	}

	// ── Step 4: Initialize session ───────────────────────────────────────────
	session, err := persistence.NewSession(root, h.PaperTitle)
	if err != nil {
		return fmt.Errorf("creating session: %w", err)
	}

	if err := session.WriteHypothesis(h); err != nil {
		return err
	}
	if err := session.WriteKnowledgeGraph(h.PaperTitle); err != nil {
		return err
	}
	if err := session.WriteLabNotebook(h.PaperTitle); err != nil {
		return err
	}

	// Log initialization event
	_ = session.AppendJSONL(map[string]interface{}{
		"event":      "session_initialized",
		"arxiv_id":   h.ArXivID,
		"paper":      h.PaperTitle,
		"model":      llmClient.ModelName(),
		"session_id": session.ID,
	})

	// ── Step 5: Summary ──────────────────────────────────────────────────────
	fmt.Println()
	printSuccess("Session initialized")
	fmt.Printf("  Session ID : %s\n", session.ID)
	fmt.Printf("  Directory  : %s\n", session.Root)
	fmt.Printf("  Hypothesis : %s\n", session.HypothesisPath())
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Review hypothesis.md and edit if needed")
	fmt.Println("  2. Set up your baseline repo and benchmark command in .research-loop/config.toml")
	fmt.Println("  3. Run: research-loop loop start")
	fmt.Println()

	return nil
}

// ─── Terminal helpers ─────────────────────────────────────────────────────────

func printStep(msg string) {
	fmt.Printf("  %s ", msg)
}

func printDone(t time.Time) {
	fmt.Printf("done (%.1fs)\n", time.Since(t).Seconds())
}

func printWarning(msg string) {
	fmt.Printf("  \033[33m⚠\033[0m  %s\n", msg)
}

func printSuccess(msg string) {
	fmt.Printf("\033[32m✓\033[0m  %s\n", msg)
}

func printHypothesisBox(h *hypothesis.Hypothesis) {
	width := 66
	line := strings.Repeat("─", width)

	fmt.Printf("\n┌%s┐\n", line)
	fmt.Printf("│ %-*s │\n", width-2, "Hypothesis extracted")
	fmt.Printf("│ %-*s │\n", width-2, "")
	fmt.Printf("│ %-*s │\n", width-2, truncate("Paper: "+h.PaperTitle, width-2))
	fmt.Printf("│ %-*s │\n", width-2, "")
	fmt.Printf("│ %-*s │\n", width-2, "Claim:")
	for _, line2 := range wordWrap(h.CoreClaim, width-4) {
		fmt.Printf("│   %-*s │\n", width-4, line2)
	}
	fmt.Printf("│ %-*s │\n", width-2, "")
	fmt.Printf("│ %-*s │\n", width-2, "Proposed experiment:")
	for _, line2 := range wordWrap(h.ProposedExperiment, width-4) {
		fmt.Printf("│   %-*s │\n", width-4, line2)
	}
	fmt.Printf("│ %-*s │\n", width-2, "")
	fmt.Printf("│ %-*s │\n", width-2, truncate("Baseline: "+h.BaselineRepo+"   Metric: "+h.Metric, width-2))
	fmt.Printf("└%s┘\n\n", line)
}

func confirm(prompt string) bool {
	fmt.Printf("%s [y/N] ", prompt)
	var response string
	fmt.Scanln(&response)
	return strings.ToLower(strings.TrimSpace(response)) == "y"
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func wordWrap(text string, width int) []string {
	var lines []string
	words := strings.Fields(text)
	var current strings.Builder

	for _, w := range words {
		if current.Len()+1+len(w) > width && current.Len() > 0 {
			lines = append(lines, current.String())
			current.Reset()
		}
		if current.Len() > 0 {
			current.WriteString(" ")
		}
		current.WriteString(w)
	}
	if current.Len() > 0 {
		lines = append(lines, current.String())
	}
	return lines
}

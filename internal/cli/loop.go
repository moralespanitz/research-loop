package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/research-loop/research-loop/internal/config"
	"github.com/research-loop/research-loop/internal/loop"
	"github.com/research-loop/research-loop/internal/persistence"
	"github.com/spf13/cobra"
)

func newLoopCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "loop",
		Short: "Manage the experiment loop",
	}
	cmd.AddCommand(newLoopStartCmd())
	cmd.AddCommand(newLoopStatusCmd())
	return cmd
}

// ─── loop start ──────────────────────────────────────────────────────────────

func newLoopStartCmd() *cobra.Command {
	var (
		sessionID string
		repoDir   string
		maxRuns   int
		resume    bool
	)

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the experiment loop for a session",
		Long: `Run the PROPOSE → MUTATE → BENCHMARK → ANNOTATE loop until stopped.

The loop:
  1. Reads hypothesis.md for the research goal
  2. Asks the Epistemic agent to propose the next mutation
  3. Applies the mutation to the baseline repo
  4. Runs the benchmark command and parses METRIC from stdout
  5. Asks the Epistemic agent to write a causal annotation
  6. Updates knowledge_graph.md and lab_notebook.md
  7. Repeats

Press Ctrl+C to stop gracefully.

Examples:
  research-loop loop start
  research-loop loop start --session my-session-id --max-runs 20
  research-loop loop start --repo ./nanoGPT --resume`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLoopStart(sessionID, repoDir, maxRuns, resume)
		},
	}

	cmd.Flags().StringVarP(&sessionID, "session", "s", "", "Session ID to run (default: most recent session)")
	cmd.Flags().StringVar(&repoDir, "repo", "", "Path to baseline repository (default: session directory)")
	cmd.Flags().IntVarP(&maxRuns, "max-runs", "n", 0, "Maximum number of experiment iterations (0 = unlimited)")
	cmd.Flags().BoolVarP(&resume, "resume", "r", false, "Resume from last checkpoint in autoresearch.jsonl")

	return cmd
}

func runLoopStart(sessionID, repoDir string, maxRuns int, resume bool) error {
	root := config.WorkspaceRoot()
	cfg, err := config.Load(root)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Resolve session
	session, err := resolveSession(root, sessionID)
	if err != nil {
		return err
	}

	// Validate config
	if cfg.Metric.BenchmarkCommand == "" {
		fmt.Fprintf(os.Stderr, "\n\033[33m⚠\033[0m  benchmark_command is not set in .research-loop/config.toml\n")
		fmt.Fprintf(os.Stderr, "   Add something like:\n\n")
		fmt.Fprintf(os.Stderr, "   [metric]\n   benchmark_command = \"python train.py --eval\"\n\n")
		fmt.Fprintf(os.Stderr, "   Your script must print a line like:  METRIC val_loss=3.21\n\n")
	}

	opts := loop.Options{
		MaxRuns: maxRuns,
		RepoDir: repoDir,
		Resume:  resume,
	}

	// Progress channel — print to stdout
	progressCh := make(chan loop.Progress, 32)
	go func() {
		for p := range progressCh {
			printLoopProgress(p)
		}
	}()

	// Graceful shutdown on Ctrl+C
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		fmt.Println("\n\nStopping loop gracefully…")
		cancel()
	}()

	fmt.Printf("\n\033[34m🔬\033[0m  Research Loop — session \033[1m%s\033[0m\n", session.ID)
	fmt.Printf("   Metric: %s (%s)   Max runs: %s\n\n",
		cfg.Metric.Name, cfg.Metric.Direction, maxRunsStr(maxRuns))

	runner, err := loop.New(session, cfg, opts, progressCh)
	if err != nil {
		return fmt.Errorf("initializing loop: %w", err)
	}

	runErr := runner.Run(ctx)
	close(progressCh)

	if runErr != nil && runErr != ctx.Err() {
		return runErr
	}
	fmt.Println("\n\033[32m✓\033[0m  Loop stopped.")
	return nil
}

// ─── loop status ─────────────────────────────────────────────────────────────

func newLoopStatusCmd() *cobra.Command {
	var sessionID string

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show the current state of an experiment loop",
		RunE: func(cmd *cobra.Command, args []string) error {
			root := config.WorkspaceRoot()
			session, err := resolveSession(root, sessionID)
			if err != nil {
				return err
			}

			jsonlPath := session.Root + "/autoresearch.jsonl"
			cp, err := loop.LoadLastCheckpoint(jsonlPath)
			if err != nil || cp == nil {
				fmt.Printf("Session %s: no loop started yet.\n", session.ID)
				fmt.Printf("Run: research-loop loop start --session %s\n", session.ID)
				return nil
			}

			fmt.Printf("\n\033[1mSession\033[0m  %s\n", session.ID)
			fmt.Printf("State    %s\n", cp.State)
			fmt.Printf("Run #    %d\n", cp.RunNumber)
			if cp.BaselineVal != 0 {
				fmt.Printf("Baseline %.4f\n", cp.BaselineVal)
			}
			if cp.BestVal != 0 {
				fmt.Printf("Best     %.4f  (%s)\n", cp.BestVal, cp.BestNode)
			}
			fmt.Println()
			return nil
		},
	}
	cmd.Flags().StringVarP(&sessionID, "session", "s", "", "Session ID")
	return cmd
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func resolveSession(root, id string) (*persistence.Session, error) {
	if id != "" {
		s, err := persistence.LoadSession(root, id)
		if err != nil {
			return nil, fmt.Errorf("session %q not found: %w", id, err)
		}
		return s, nil
	}

	// Find the most recently modified session directory
	sessDir := root + "/.research-loop/sessions"
	entries, err := os.ReadDir(sessDir)
	if err != nil || len(entries) == 0 {
		return nil, fmt.Errorf("no sessions found in %s\nRun: research-loop start <arxiv-url>", sessDir)
	}
	// ReadDir returns entries sorted by name; find last directory
	var latest string
	for i := len(entries) - 1; i >= 0; i-- {
		if entries[i].IsDir() {
			latest = entries[i].Name()
			break
		}
	}
	if latest == "" {
		return nil, fmt.Errorf("no session directories found")
	}
	return persistence.LoadSession(root, latest)
}

func printLoopProgress(p loop.Progress) {
	stateColor := "\033[34m" // blue
	switch p.State {
	case loop.StateBenchmark:
		stateColor = "\033[33m" // amber
	case loop.StateAnnotate:
		stateColor = "\033[35m" // magenta
	case loop.StateDone:
		stateColor = "\033[32m" // green
	case loop.StateFailed:
		stateColor = "\033[31m" // red
	}

	prefix := fmt.Sprintf("%s[%s]\033[0m", stateColor, p.State)
	if p.RunNumber > 0 {
		prefix = fmt.Sprintf("%s[%s #%d]\033[0m", stateColor, p.State, p.RunNumber)
	}

	fmt.Printf("  %s  %s\n", prefix, p.Message)

	if p.Record != nil {
		r := p.Record
		icon := "✓"
		if r.Status == loop.StatusRegression || r.Status == loop.StatusCrash {
			icon = "✗"
		}
		fmt.Printf("         %s  metric=%.4f  Δ%+.4f  node=%s\n",
			icon, r.MetricVal, r.Delta, r.Node)
		if r.Annotation != "" {
			fmt.Printf("         \033[2m%s\033[0m\n", truncate(r.Annotation, 80))
		}
	}
}

func maxRunsStr(n int) string {
	if n == 0 {
		return "∞"
	}
	return fmt.Sprintf("%d", n)
}

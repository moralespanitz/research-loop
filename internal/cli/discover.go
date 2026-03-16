package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/research-loop/research-loop/internal/config"
	"github.com/research-loop/research-loop/internal/discovery"
	"github.com/spf13/cobra"
)

func newDiscoverCmd() *cobra.Command {
	var (
		maxLanes       int
		maxRunsPerLane int
		repoDir        string
		gateThreshold  float64
	)

	cmd := &cobra.Command{
		Use:   "discover <topic>",
		Short: "Run parallel discovery lanes on a research topic",
		Long: `Launch multiple independent research lanes that explore different angles
of a topic simultaneously.

Each lane runs the full pipeline:
  LITERATURE → GAP_ANALYSIS → HYPOTHESIS → EXPERIMENT → BENCHMARK → REVIEW

At each transition, a Carlini decision gate evaluates whether the lane is
worth continuing. Lanes that fail are killed early — no wasted compute.

Examples:
  research-loop discover "attention mechanism efficiency"
  research-loop discover "optimizer improvements for transformer training" --lanes 6
  research-loop discover "kv cache compression" --repo ./autoresearch --runs-per-lane 20`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDiscover(args[0], maxLanes, maxRunsPerLane, repoDir, gateThreshold)
		},
	}

	cmd.Flags().IntVar(&maxLanes, "lanes", 4, "Number of parallel research lanes")
	cmd.Flags().IntVar(&maxRunsPerLane, "runs-per-lane", 10, "Max experiment iterations per lane")
	cmd.Flags().StringVar(&repoDir, "repo", "", "Path to autoresearch/baseline repository")
	cmd.Flags().Float64Var(&gateThreshold, "gate-threshold", 0.4, "Minimum Carlini gate score to continue (0.0-1.0)")

	return cmd
}

func runDiscover(topic string, maxLanes, maxRunsPerLane int, repoDir string, gateThreshold float64) error {
	root := config.WorkspaceRoot()
	cfg, err := config.Load(root)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	opts := discovery.Options{
		MaxLanes:       maxLanes,
		MaxRunsPerLane: maxRunsPerLane,
		RepoDir:        repoDir,
		GateThreshold:  gateThreshold,
	}

	progressCh := make(chan discovery.LaneProgress, 64)
	go func() {
		for p := range progressCh {
			printDiscoveryProgress(p)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		fmt.Println("\n\nStopping discovery gracefully…")
		cancel()
	}()

	fmt.Printf("\n\033[34m🔬\033[0m  Research Loop — Parallel Discovery\n")
	fmt.Printf("   Topic: \033[1m%s\033[0m\n", topic)
	fmt.Printf("   Lanes: %d   Max runs/lane: %d   Gate threshold: %.1f\n\n",
		maxLanes, maxRunsPerLane, gateThreshold)

	orch, err := discovery.New(topic, root, cfg, opts, progressCh)
	if err != nil {
		return fmt.Errorf("initializing orchestrator: %w", err)
	}

	runErr := orch.Run(ctx)
	close(progressCh)

	if runErr != nil && runErr != ctx.Err() {
		return runErr
	}

	// Print summary
	fmt.Println("\n\033[1m── Discovery Summary ──\033[0m\n")
	for _, lane := range orch.Lanes() {
		fmt.Printf("  %s\n", lane.Summary())
	}
	fmt.Println()

	return nil
}

func printDiscoveryProgress(p discovery.LaneProgress) {
	color := "\033[34m" // blue
	switch p.State {
	case discovery.StateGapAnalysis:
		color = "\033[33m" // amber
	case discovery.StateHypothesis:
		color = "\033[35m" // magenta
	case discovery.StateExperiment:
		color = "\033[36m" // cyan
	case discovery.StateLaneBench:
		color = "\033[33m" // amber
	case discovery.StateReview:
		color = "\033[35m" // magenta
	case discovery.StateLaneDone:
		color = "\033[32m" // green
	case discovery.StateLaneKilled:
		color = "\033[31m" // red
	}

	prefix := ""
	if p.LaneID != "" {
		prefix = fmt.Sprintf("%s[%-12s %s]\033[0m", color, p.State, p.Angle)
	} else {
		prefix = fmt.Sprintf("%s[ORCHESTRATOR]\033[0m", color)
	}

	if p.Killed {
		fmt.Printf("  \033[31m✗\033[0m %s  %s\n", prefix, p.Message)
	} else {
		fmt.Printf("  %s  %s\n", prefix, p.Message)
	}
}

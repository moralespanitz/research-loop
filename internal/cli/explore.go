package cli

import (
	"context"
	"fmt"

	"github.com/research-loop/research-loop/internal/config"
	"github.com/research-loop/research-loop/internal/explore"
	"github.com/spf13/cobra"
)

func newExploreCmd() *cobra.Command {
	var opts struct {
		topic     string
		maxPapers int
		maxRepos  int
		noLaunch  bool
	}

	cmd := &cobra.Command{
		Use:   "explore <topic>",
		Short: "Explore a research problem: papers, mental models, debates, scoring",
		Long: `
Explore a research topic using the MIT grad student methodology:

1. Gather papers & GitHub repos on the topic
2. Extract core mental models experts share
3. Find where the field fundamentally disagrees
4. Generate diagnostic questions
5. Score using Carlini criteria

Then optionally launch parallel discovery lanes.

Examples:
  research-loop explore "machine learning efficiency"
  research-loop explore "attention mechanisms" --max-papers 20`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExplore(cmd.Context(), args[0], opts.maxPapers, opts.maxRepos, opts.noLaunch)
		},
	}

	cmd.Flags().IntVar(&opts.maxPapers, "max-papers", 15, "Maximum papers to gather")
	cmd.Flags().IntVar(&opts.maxRepos, "max-repos", 10, "Maximum GitHub repos to find")
	cmd.Flags().BoolVar(&opts.noLaunch, "no-launch", false, "Don't suggest launching discovery")

	return cmd
}

func runExplore(ctx context.Context, topic string, maxPapers, maxRepos int, noLaunch bool) error {
	workspace := config.WorkspaceRoot()
	cfg, err := config.Load(workspace)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	fmt.Printf("🔍 Exploring: %s\n\n", topic)

	expOpts := explore.Options{
		Topic:     topic,
		MaxPapers: maxPapers,
		MaxRepos:  maxRepos,
	}

	engine, err := explore.New(workspace, cfg, expOpts)
	if err != nil {
		return fmt.Errorf("creating explore engine: %w", err)
	}

	if err := engine.Run(ctx); err != nil {
		return fmt.Errorf("exploration failed: %w", err)
	}

	fmt.Println("\n" + engine.Summary())

	if !noLaunch {
		fmt.Println("\n🚀 To launch parallel discovery lanes:")
		fmt.Printf("   research-loop discover \"%s\"\n", topic)
	}

	return nil
}

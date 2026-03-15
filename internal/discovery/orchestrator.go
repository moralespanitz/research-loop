package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/research-loop/research-loop/internal/config"
	"github.com/research-loop/research-loop/internal/llm"
)

// Options configures a discovery run.
type Options struct {
	// MaxLanes is the maximum number of parallel research lanes (default: 4).
	MaxLanes int

	// MaxRunsPerLane is the max experiment iterations per lane (default: 10).
	MaxRunsPerLane int

	// RepoDir is the path to the autoresearch/baseline repository.
	RepoDir string

	// GateThreshold is the minimum Carlini gate score to pass (default: 0.4).
	GateThreshold float64
}

// Orchestrator manages multiple parallel discovery lanes for a topic.
type Orchestrator struct {
	topic    string
	cfg      *config.Config
	opts     Options
	client   llm.Client
	progress chan<- LaneProgress
	root     string // workspace root

	mu    sync.RWMutex
	lanes []*Lane
}

// New creates an Orchestrator for the given topic.
func New(topic, workspaceRoot string, cfg *config.Config, opts Options, progress chan<- LaneProgress) (*Orchestrator, error) {
	client, err := llm.New(cfg.LLM)
	if err != nil {
		return nil, fmt.Errorf("creating LLM client: %w", err)
	}

	if opts.MaxLanes <= 0 {
		opts.MaxLanes = 4
	}
	if opts.MaxRunsPerLane <= 0 {
		opts.MaxRunsPerLane = 10
	}
	if opts.GateThreshold <= 0 {
		opts.GateThreshold = 0.4
	}

	return &Orchestrator{
		topic:    topic,
		cfg:      cfg,
		opts:     opts,
		client:   client,
		progress: progress,
		root:     workspaceRoot,
	}, nil
}

// Run starts the parallel discovery process. Blocks until all lanes complete.
func (o *Orchestrator) Run(ctx context.Context) error {
	// ── Step 1: Generate research angles ──────────────────────────────────
	o.emit(LaneProgress{Message: fmt.Sprintf("Generating research angles for: %s", o.topic)})

	angles, err := o.generateAngles(ctx)
	if err != nil {
		return fmt.Errorf("generating research angles: %w", err)
	}

	if len(angles) > o.opts.MaxLanes {
		angles = angles[:o.opts.MaxLanes]
	}

	// ── Step 2: Create lanes ──────────────────────────────────────────────
	for i, angle := range angles {
		lane := &Lane{
			ID:          fmt.Sprintf("lane-%d", i+1),
			Topic:       o.topic,
			Angle:       angle.Slug,
			Description: angle.Description,
			State:       StateLiterature,
			StartedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		o.lanes = append(o.lanes, lane)
		o.emit(LaneProgress{
			LaneID: lane.ID, Angle: lane.Angle, State: StateLiterature,
			Message: fmt.Sprintf("Lane %d: %s — %s", i+1, angle.Slug, angle.Description),
		})
	}

	// ── Step 3: Run lanes in parallel ─────────────────────────────────────
	var wg sync.WaitGroup
	for _, lane := range o.lanes {
		wg.Add(1)
		go func(l *Lane) {
			defer wg.Done()
			o.runLane(ctx, l)
		}(lane)
	}
	wg.Wait()

	// ── Step 4: Save results ──────────────────────────────────────────────
	return o.saveResults()
}

// Lanes returns a snapshot of all lanes for TUI display.
func (o *Orchestrator) Lanes() []*Lane {
	o.mu.RLock()
	defer o.mu.RUnlock()
	cp := make([]*Lane, len(o.lanes))
	copy(cp, o.lanes)
	return cp
}

// ─── Lane execution ──────────────────────────────────────────────────────────

func (o *Orchestrator) runLane(ctx context.Context, lane *Lane) {
	stages := []struct {
		from, to LaneState
		fn       func(context.Context, *Lane) error
	}{
		{StateLiterature, StateGapAnalysis, o.stageLiterature},
		{StateGapAnalysis, StateHypothesis, o.stageGapAnalysis},
		{StateHypothesis, StateExperiment, o.stageHypothesis},
		{StateExperiment, StateLaneBench, o.stageExperiment},
		{StateLaneBench, StateReview, o.stageBenchmark},
		{StateReview, StateLaneDone, o.stageReview},
	}

	for _, stage := range stages {
		if ctx.Err() != nil || !lane.IsAlive() {
			return
		}

		// Execute stage
		if err := stage.fn(ctx, lane); err != nil {
			lane.AddError(err)
			o.emit(LaneProgress{
				LaneID: lane.ID, Angle: lane.Angle, State: lane.State,
				Message: fmt.Sprintf("Error in %s: %v", stage.from, err),
			})
			// Non-fatal for most stages — continue
			if stage.from == StateExperiment {
				// Experiment errors are expected (crashes etc)
				continue
			}
		}

		// Carlini gate check
		if ctx.Err() != nil {
			return
		}
		gate, err := EvaluateGate(ctx, o.client, lane, stage.from, stage.to)
		if err != nil {
			lane.AddError(err)
		}

		if !gate.Pass || gate.Score < o.opts.GateThreshold {
			lane.Kill(fmt.Sprintf("Carlini gate rejected (%s→%s): score=%.2f — %s",
				stage.from, stage.to, gate.Score, gate.Reason))
			o.emit(LaneProgress{
				LaneID: lane.ID, Angle: lane.Angle, State: StateLaneKilled,
				Message: fmt.Sprintf("KILLED: %s", gate.Reason), Killed: true,
			})
			return
		}

		lane.Transition(stage.to)
		o.emit(LaneProgress{
			LaneID: lane.ID, Angle: lane.Angle, State: stage.to,
			Message: fmt.Sprintf("Gate passed (%.2f): %s", gate.Score, gate.Reason),
		})
	}

	lane.Transition(StateLaneDone)
	o.emit(LaneProgress{
		LaneID: lane.ID, Angle: lane.Angle, State: StateLaneDone,
		Message: fmt.Sprintf("Lane complete — best=%.4f (%s) verdict=%s",
			lane.BestMetric, lane.BestNode, lane.Verdict),
	})
}

// ─── Stage implementations ───────────────────────────────────────────────────

func (o *Orchestrator) stageLiterature(ctx context.Context, lane *Lane) error {
	o.emit(LaneProgress{LaneID: lane.ID, Angle: lane.Angle, State: StateLiterature,
		Message: "Searching ArXiv for related papers…"})

	prompt := fmt.Sprintf(`Search for the most important recent papers on: %s
Focus specifically on the angle: %s

List 5-8 key papers. For each paper provide:
PAPER_ID: <arxiv id, e.g. 2403.05821>
TITLE: <paper title>
AUTHORS: <first author et al.>
YEAR: <year>
ABSTRACT: <1-2 sentence summary of key contribution>
---`, lane.Topic, lane.Angle)

	raw, err := o.client.Complete(ctx, "You are a research librarian. List the most relevant recent ArXiv papers for the given research angle. Be specific — list real papers with real ArXiv IDs where possible.", []llm.Message{
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return err
	}

	lane.mu.Lock()
	lane.Papers = parsePapers(raw)
	lane.mu.Unlock()

	o.emit(LaneProgress{LaneID: lane.ID, Angle: lane.Angle, State: StateLiterature,
		Message: fmt.Sprintf("Found %d papers", len(lane.Papers))})
	return nil
}

func (o *Orchestrator) stageGapAnalysis(ctx context.Context, lane *Lane) error {
	o.emit(LaneProgress{LaneID: lane.ID, Angle: lane.Angle, State: StateGapAnalysis,
		Message: "Analyzing research gaps…"})

	var papersText strings.Builder
	for _, p := range lane.Papers {
		papersText.WriteString(fmt.Sprintf("- %s (%d): %s\n", p.Title, p.Year, p.Abstract))
	}

	prompt := fmt.Sprintf(`Given these papers on "%s" (angle: %s):

%s

Identify 3-5 research gaps. For each gap, score these dimensions (0.0-1.0):
- IMPORTANCE: How much would solving this advance the field?
- NOVELTY: How new/unexplored is this direction?
- FEASIBILITY: Can we test this with a single-GPU training experiment (autoresearch)?

Format:
GAP: <description>
IMPORTANCE: <0.0-1.0>
NOVELTY: <0.0-1.0>
FEASIBILITY: <0.0-1.0>
---`, lane.Topic, lane.Angle, papersText.String())

	raw, err := o.client.Complete(ctx, `You are the Epistemic Agent performing gap analysis using Carlini's criteria:
- "The single most important skill is good taste in what problems are worth solving"
- "Pick your ideas for impact. One excellent paper > 1000 mediocre ones"
- Focus on gaps that are TESTABLE with a single-GPU 5-minute experiment`, []llm.Message{
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return err
	}

	lane.mu.Lock()
	lane.Gaps = parseGaps(raw)
	lane.mu.Unlock()

	if len(lane.Gaps) > 0 {
		o.emit(LaneProgress{LaneID: lane.ID, Angle: lane.Angle, State: StateGapAnalysis,
			Message: fmt.Sprintf("Top gap (score=%.2f): %s", lane.Gaps[0].Score, truncateStr(lane.Gaps[0].Description, 60))})
	}
	return nil
}

func (o *Orchestrator) stageHypothesis(ctx context.Context, lane *Lane) error {
	o.emit(LaneProgress{LaneID: lane.ID, Angle: lane.Angle, State: StateHypothesis,
		Message: "Formalizing hypothesis…"})

	if len(lane.Gaps) == 0 {
		return fmt.Errorf("no gaps to formalize")
	}
	topGap := lane.Gaps[0]

	prompt := fmt.Sprintf(`Based on this research gap:
%s

Formalize it into a concrete, testable hypothesis for a single-GPU training experiment
using karpathy/autoresearch (train.py with val_bpb metric).

CLAIM: <1-2 sentences — the specific empirical claim to test>
EXPERIMENT: <2-3 sentences — what to change in train.py and what metric improvement to expect>`, topGap.Description)

	raw, err := o.client.Complete(ctx, `You are the Epistemic Agent formalizing a hypothesis.
Be concrete: name specific constants in train.py (DEPTH, MATRIX_LR, WINDOW_PATTERN, etc.)
and predict a direction of improvement.`, []llm.Message{
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return err
	}

	lane.mu.Lock()
	for _, line := range strings.Split(raw, "\n") {
		if strings.HasPrefix(line, "CLAIM:") {
			lane.Claim = strings.TrimSpace(strings.TrimPrefix(line, "CLAIM:"))
		}
		if strings.HasPrefix(line, "EXPERIMENT:") {
			lane.Experiment = strings.TrimSpace(strings.TrimPrefix(line, "EXPERIMENT:"))
		}
	}
	lane.mu.Unlock()

	o.emit(LaneProgress{LaneID: lane.ID, Angle: lane.Angle, State: StateHypothesis,
		Message: fmt.Sprintf("Claim: %s", truncateStr(lane.Claim, 70))})
	return nil
}

func (o *Orchestrator) stageExperiment(ctx context.Context, lane *Lane) error {
	o.emit(LaneProgress{LaneID: lane.ID, Angle: lane.Angle, State: StateExperiment,
		Message: fmt.Sprintf("Running up to %d experiment iterations…", o.opts.MaxRunsPerLane)})

	// For now, log that experiments would run here.
	// The actual autoresearch integration uses the loop.Runner — to be wired in.
	lane.mu.Lock()
	lane.Runs = []LaneRun{
		{RunNumber: 0, Node: "baseline", MetricVal: 0, Delta: 0, Status: "pending",
			Timestamp: time.Now().UTC().Format(time.RFC3339)},
	}
	lane.mu.Unlock()

	o.emit(LaneProgress{LaneID: lane.ID, Angle: lane.Angle, State: StateExperiment,
		Message: "Experiment stage ready — wire to loop.Runner for live execution"})
	return nil
}

func (o *Orchestrator) stageBenchmark(ctx context.Context, lane *Lane) error {
	o.emit(LaneProgress{LaneID: lane.ID, Angle: lane.Angle, State: StateLaneBench,
		Message: "Collecting benchmark results…"})
	// Results are already in lane.Runs from the experiment stage
	return nil
}

func (o *Orchestrator) stageReview(ctx context.Context, lane *Lane) error {
	o.emit(LaneProgress{LaneID: lane.ID, Angle: lane.Angle, State: StateReview,
		Message: "Writing final review…"})

	prompt := fmt.Sprintf(`## Lane Review: %s

Topic: %s
Claim: %s
Experiment: %s
Runs: %d
Best metric: %.4f (%s)

Write a 3-5 sentence review of this research lane:
1. Did the experiments support the claim?
2. What did we learn?
3. Is this direction worth pursuing further?

VERDICT: promising | inconclusive | dead_end
REVIEW: <your review>`, lane.Angle, lane.Topic, lane.Claim, lane.Experiment,
		len(lane.Runs), lane.BestMetric, lane.BestNode)

	raw, err := o.client.Complete(ctx, "You are a senior researcher reviewing experimental results. Be honest and direct.", []llm.Message{
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return err
	}

	lane.mu.Lock()
	for _, line := range strings.Split(raw, "\n") {
		if strings.HasPrefix(line, "VERDICT:") {
			lane.Verdict = strings.TrimSpace(strings.TrimPrefix(line, "VERDICT:"))
		}
		if strings.HasPrefix(line, "REVIEW:") {
			lane.Review = strings.TrimSpace(strings.TrimPrefix(line, "REVIEW:"))
		}
	}
	if lane.Review == "" {
		lane.Review = raw
	}
	lane.mu.Unlock()

	return nil
}

// ─── Angle generation ────────────────────────────────────────────────────────

type researchAngle struct {
	Slug        string
	Description string
}

func (o *Orchestrator) generateAngles(ctx context.Context) ([]researchAngle, error) {
	prompt := fmt.Sprintf(`Generate %d distinct research angles for the topic: "%s"

Each angle should be a specific, testable direction that could be explored
with a single-GPU training experiment (karpathy/autoresearch — modifying train.py).

For each angle:
SLUG: <snake_case short name, max 30 chars>
DESCRIPTION: <1-2 sentences explaining the specific research direction>
---`, o.opts.MaxLanes, o.topic)

	raw, err := o.client.Complete(ctx, `You are a research strategist. Generate diverse, non-overlapping research angles.
Apply Carlini's criteria: each angle should be important, novel, and feasible.
Prioritize angles where single-GPU experiments can yield meaningful results.`, []llm.Message{
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return nil, err
	}

	return parseAngles(raw), nil
}

// ─── Persistence ─────────────────────────────────────────────────────────────

func (o *Orchestrator) saveResults() error {
	dir := filepath.Join(o.root, ".research-loop", "discoveries")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	slug := strings.ReplaceAll(strings.ToLower(o.topic), " ", "_")
	if len(slug) > 40 {
		slug = slug[:40]
	}
	filename := fmt.Sprintf("%s_%s.json", slug, time.Now().Format("2006-01-02"))

	data, err := json.MarshalIndent(o.lanes, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(dir, filename)
	o.emit(LaneProgress{Message: fmt.Sprintf("Results saved to %s", path)})
	return os.WriteFile(path, data, 0644)
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func (o *Orchestrator) emit(p LaneProgress) {
	if o.progress != nil {
		select {
		case o.progress <- p:
		default:
		}
	}
}

func parseAngles(raw string) []researchAngle {
	var angles []researchAngle
	var current researchAngle
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "SLUG:") {
			if current.Slug != "" {
				angles = append(angles, current)
			}
			current = researchAngle{Slug: strings.TrimSpace(strings.TrimPrefix(line, "SLUG:"))}
		}
		if strings.HasPrefix(line, "DESCRIPTION:") {
			current.Description = strings.TrimSpace(strings.TrimPrefix(line, "DESCRIPTION:"))
		}
	}
	if current.Slug != "" {
		angles = append(angles, current)
	}
	return angles
}

func parsePapers(raw string) []PaperRef {
	var papers []PaperRef
	var current PaperRef
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "---" {
			if current.Title != "" {
				papers = append(papers, current)
			}
			current = PaperRef{}
			continue
		}
		if strings.HasPrefix(line, "PAPER_ID:") {
			current.ArXivID = strings.TrimSpace(strings.TrimPrefix(line, "PAPER_ID:"))
			current.URL = "https://arxiv.org/abs/" + current.ArXivID
		}
		if strings.HasPrefix(line, "TITLE:") {
			current.Title = strings.TrimSpace(strings.TrimPrefix(line, "TITLE:"))
		}
		if strings.HasPrefix(line, "AUTHORS:") {
			current.Authors = strings.TrimSpace(strings.TrimPrefix(line, "AUTHORS:"))
		}
		if strings.HasPrefix(line, "YEAR:") {
			fmt.Sscanf(strings.TrimPrefix(line, "YEAR:"), "%d", &current.Year)
		}
		if strings.HasPrefix(line, "ABSTRACT:") {
			current.Abstract = strings.TrimSpace(strings.TrimPrefix(line, "ABSTRACT:"))
		}
	}
	if current.Title != "" {
		papers = append(papers, current)
	}
	return papers
}

func parseGaps(raw string) []Gap {
	var gaps []Gap
	var current Gap
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "---" {
			if current.Description != "" {
				current.Score = (current.Importance*0.4 + current.Novelty*0.3 + current.Feasibility*0.3)
				gaps = append(gaps, current)
			}
			current = Gap{}
			continue
		}
		if strings.HasPrefix(line, "GAP:") {
			current.Description = strings.TrimSpace(strings.TrimPrefix(line, "GAP:"))
		}
		if strings.HasPrefix(line, "IMPORTANCE:") {
			fmt.Sscanf(strings.TrimPrefix(line, "IMPORTANCE:"), "%f", &current.Importance)
		}
		if strings.HasPrefix(line, "NOVELTY:") {
			fmt.Sscanf(strings.TrimPrefix(line, "NOVELTY:"), "%f", &current.Novelty)
		}
		if strings.HasPrefix(line, "FEASIBILITY:") {
			fmt.Sscanf(strings.TrimPrefix(line, "FEASIBILITY:"), "%f", &current.Feasibility)
		}
	}
	if current.Description != "" {
		current.Score = (current.Importance*0.4 + current.Novelty*0.3 + current.Feasibility*0.3)
		gaps = append(gaps, current)
	}

	// Sort by score descending
	for i := 0; i < len(gaps); i++ {
		for j := i + 1; j < len(gaps); j++ {
			if gaps[j].Score > gaps[i].Score {
				gaps[i], gaps[j] = gaps[j], gaps[i]
			}
		}
	}
	return gaps
}

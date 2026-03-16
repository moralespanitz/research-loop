package loop

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/research-loop/research-loop/internal/config"
	"github.com/research-loop/research-loop/internal/llm"
	"github.com/research-loop/research-loop/internal/persistence"
)

// Options configures the experiment loop.
type Options struct {
	// MaxRuns is the maximum number of experiment iterations (0 = unlimited).
	MaxRuns int

	// RepoDir is the path to the baseline repository to mutate.
	// If empty, the session directory itself is used.
	RepoDir string

	// Resume attempts to continue from the last checkpoint in autoresearch.jsonl.
	Resume bool
}

// Runner drives the experiment loop state machine for one session.
type Runner struct {
	session  *persistence.Session
	cfg      *config.Config
	opts     Options
	client   llm.Client           // Empirical + Epistemic shared client (can be split later)
	progress chan<- Progress       // sends live updates to TUI / CLI

	// runtime state
	state       LoopState
	runNumber   int
	baselineVal float64
	bestVal     float64
	bestNode    string
	lastRuns    []RunRecord
	direction   string // "lower" | "higher"
}

// New creates a Runner for the given session.
// progress may be nil if no live updates are needed.
func New(session *persistence.Session, cfg *config.Config, opts Options, progress chan<- Progress) (*Runner, error) {
	client, err := llm.New(cfg.LLM)
	if err != nil {
		return nil, fmt.Errorf("creating LLM client: %w", err)
	}

	dir := cfg.Metric.Direction
	if dir == "" {
		dir = "lower"
	}

	repoDir := opts.RepoDir
	if repoDir == "" {
		repoDir = session.Root
	}

	return &Runner{
		session:   session,
		cfg:       cfg,
		opts:      opts,
		client:    client,
		progress:  progress,
		state:     StateIdle,
		direction: dir,
		bestVal:   math.MaxFloat64,
	}, nil
}

// Run executes the experiment loop until ctx is cancelled, MaxRuns is reached,
// or a fatal error occurs. It is safe to call from a goroutine.
func (r *Runner) Run(ctx context.Context) error {
	jsonlPath := filepath.Join(r.session.Root, "autoresearch.jsonl")

	// Resume from checkpoint if requested
	if r.opts.Resume {
		if cp, err := LoadLastCheckpoint(jsonlPath); err == nil && cp != nil {
			r.runNumber = cp.RunNumber
			r.baselineVal = cp.BaselineVal
			r.bestVal = cp.BestVal
			r.bestNode = cp.BestNode
			r.emit(Progress{State: StateIdle, Message: fmt.Sprintf("Resuming from run #%d (best: %.4f %s)", r.runNumber, r.bestVal, r.bestNode)})
		}
	}

	r.transition(StateHypothesize)

	// ── HYPOTHESIZE: read hypothesis.md ──────────────────────────────────────
	hypothesisMD, err := os.ReadFile(r.session.HypothesisPath())
	if err != nil {
		return r.fail(fmt.Errorf("reading hypothesis.md: %w", err))
	}
	r.emit(Progress{State: StateHypothesize, Message: "Hypothesis loaded"})

	// ── Establish baseline if first run ──────────────────────────────────────
	if r.runNumber == 0 && r.cfg.Metric.BenchmarkCommand != "" {
		r.emit(Progress{State: StateBenchmark, Message: "Running baseline benchmark…"})
		res := RunBenchmark(r.opts.RepoDir, r.cfg.Metric.BenchmarkCommand, r.cfg.Metric.TimeoutSecs)
		if res.Err != nil {
			return r.fail(fmt.Errorf("baseline benchmark failed: %w", res.Err))
		}
		r.baselineVal = res.MetricVal
		r.bestVal = res.MetricVal
		r.emit(Progress{State: StateBenchmark, Message: fmt.Sprintf("Baseline: %s = %.4f", r.cfg.Metric.Name, r.baselineVal)})

		_ = SaveCheckpoint(jsonlPath, Checkpoint{
			Event:       "baseline",
			State:       StateBenchmark,
			BaselineVal: r.baselineVal,
			BestVal:     r.bestVal,
			SessionID:   r.session.ID,
		})
	}

	// ── Main loop ─────────────────────────────────────────────────────────────
	for {
		if ctx.Err() != nil {
			r.transition(StateDone)
			return ctx.Err()
		}
		if r.opts.MaxRuns > 0 && r.runNumber >= r.opts.MaxRuns {
			r.transition(StateDone)
			r.emit(Progress{State: StateDone, Message: fmt.Sprintf("Done — %d runs. Best: %.4f (%s)", r.runNumber, r.bestVal, r.bestNode)})
			return nil
		}

		r.runNumber++
		r.emit(Progress{State: StatePropose, RunNumber: r.runNumber, Message: fmt.Sprintf("Run #%d — proposing next mutation…", r.runNumber)})

		// ── PROPOSE ──────────────────────────────────────────────────────────
		r.transition(StatePropose)
		kgMD := r.readKG()
		proposal, err := Propose(ctx, r.client, string(hypothesisMD), kgMD, r.lastRuns, r.opts.RepoDir)
		if err != nil {
			// Non-fatal: log and retry next iteration
			r.emit(Progress{State: StatePropose, RunNumber: r.runNumber, Message: fmt.Sprintf("Propose failed: %v — retrying next iteration", err)})
			time.Sleep(5 * time.Second)
			r.runNumber--
			continue
		}
		r.emit(Progress{State: StatePropose, RunNumber: r.runNumber,
			Message: fmt.Sprintf("Proposed: %s — %s", proposal.Node, proposal.Description)})

		// ── MUTATE ───────────────────────────────────────────────────────────
		r.transition(StateMutate)
		r.emit(Progress{State: StateMutate, RunNumber: r.runNumber,
			Message: fmt.Sprintf("Applying mutation: %s", proposal.Node)})

		if err := ApplyMutation(r.opts.RepoDir, proposal); err != nil {
			r.emit(Progress{State: StateMutate, RunNumber: r.runNumber,
				Message: fmt.Sprintf("Mutation apply failed: %v — skipping", err)})
			_ = RevertMutation(r.opts.RepoDir)
			continue
		}

		// Save diff before running (so we have it even if benchmark crashes)
		diffDir := filepath.Join(r.session.Root, "checkpoints")
		_ = os.MkdirAll(diffDir, 0755)
		diffPath := filepath.Join(diffDir, fmt.Sprintf("run-%04d-%s.diff", r.runNumber, proposal.Node))
		_ = SaveDiff(r.opts.RepoDir, diffPath)

		// ── BENCHMARK ────────────────────────────────────────────────────────
		r.transition(StateBenchmark)
		r.emit(Progress{State: StateBenchmark, RunNumber: r.runNumber,
			Message: fmt.Sprintf("Benchmarking %s…", proposal.Node)})

		rec := RunRecord{
			Event:       "run_complete",
			RunNumber:   r.runNumber,
			State:       StateBenchmark,
			Node:        proposal.Node,
			Mutation:    proposal.Description,
			Proposal:    proposal,
			BaselineVal: r.baselineVal,
			DiffPath:    diffPath,
		}

		var benchErr error
		if r.cfg.Metric.BenchmarkCommand == "" {
			benchErr = fmt.Errorf("benchmark_command not set in config.toml")
		} else {
			res := RunBenchmark(r.opts.RepoDir, r.cfg.Metric.BenchmarkCommand, r.cfg.Metric.TimeoutSecs)
			rec.BenchOutput = lastNLines(res.BenchOutput, 40)
			rec.MetricRaw = res.MetricRaw
			rec.MetricVal = res.MetricVal
			rec.Result = res.MetricRaw
			benchErr = res.Err
		}

		_ = RevertMutation(r.opts.RepoDir) // always revert after run

		if benchErr != nil {
			rec.Status = StatusCrash
			rec.Annotation = fmt.Sprintf("Benchmark crashed: %v", benchErr)
			r.emit(Progress{State: StateBenchmark, RunNumber: r.runNumber,
				Message: fmt.Sprintf("Benchmark crashed: %v", benchErr), Record: &rec})
		} else {
			rec.Delta = rec.MetricVal - r.baselineVal
			if r.isImprovement(rec.MetricVal) {
				rec.Status = StatusImprovement
				r.bestVal = rec.MetricVal
				r.bestNode = rec.Node
			} else {
				rec.Status = StatusRegression
			}
		}

		// ── ANNOTATE ─────────────────────────────────────────────────────────
		r.transition(StateAnnotate)
		r.emit(Progress{State: StateAnnotate, RunNumber: r.runNumber,
			Message: "Writing causal annotation…"})

		annCtx, annCancel := context.WithTimeout(ctx, 60*time.Second)
		annotation, annErr := Annotate(annCtx, r.client, string(hypothesisMD), rec)
		annCancel()
		if annErr == nil {
			rec.Annotation = annotation
		}

		// Persist to JSONL, KG, notebook
		rec.Timestamp = time.Now().UTC().Format(time.RFC3339)
		_ = r.session.AppendJSONL(runRecordToMap(rec))
		_ = AppendKnowledgeGraph(filepath.Join(r.session.Root, "knowledge_graph.md"), rec)
		_ = AppendLabNotebook(filepath.Join(r.session.Root, "lab_notebook.md"), rec)

		// Update last-5 ring buffer
		r.lastRuns = append(r.lastRuns, rec)
		if len(r.lastRuns) > 5 {
			r.lastRuns = r.lastRuns[len(r.lastRuns)-5:]
		}

		_ = SaveCheckpoint(jsonlPath, Checkpoint{
			Event:       "run_complete",
			State:       StatePropose,
			RunNumber:   r.runNumber,
			BaselineVal: r.baselineVal,
			BestVal:     r.bestVal,
			BestNode:    r.bestNode,
			SessionID:   r.session.ID,
		})

		r.emit(Progress{State: StateAnnotate, RunNumber: r.runNumber,
			Message: fmt.Sprintf("Run #%d done — %s %.4f (Δ %+.4f)", r.runNumber, rec.Node, rec.MetricVal, rec.Delta),
			Record:  &rec})
	}
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func (r *Runner) isImprovement(val float64) bool {
	if r.direction == "higher" {
		return val > r.bestVal
	}
	return val < r.bestVal
}

func (r *Runner) transition(s LoopState) {
	r.state = s
}

func (r *Runner) emit(p Progress) {
	if r.progress != nil {
		select {
		case r.progress <- p:
		default:
		}
	}
}

func (r *Runner) fail(err error) error {
	r.state = StateFailed
	r.emit(Progress{State: StateFailed, Message: err.Error()})
	return err
}

func (r *Runner) readKG() string {
	data, err := os.ReadFile(filepath.Join(r.session.Root, "knowledge_graph.md"))
	if err != nil {
		return "(knowledge graph not found)"
	}
	// Limit to ~4000 chars to stay within context
	s := string(data)
	if len(s) > 4000 {
		s = s[:4000] + "\n…(truncated)"
	}
	return s
}

func runRecordToMap(r RunRecord) map[string]interface{} {
	return map[string]interface{}{
		"event":          r.Event,
		"run_number":     r.RunNumber,
		"state":          string(r.State),
		"node":           r.Node,
		"mutation":       r.Mutation,
		"result":         r.Result,
		"metric_value":   r.MetricVal,
		"metric_raw":     r.MetricRaw,
		"baseline_value": r.BaselineVal,
		"delta":          r.Delta,
		"status":         string(r.Status),
		"annotation":     r.Annotation,
		"diff_path":      r.DiffPath,
		"bench_output":   strings.TrimSpace(lastNLines(r.BenchOutput, 20)),
	}
}

// Package discovery implements the parallel discovery orchestrator.
//
// A "discovery run" takes a research topic and spawns multiple independent
// research lanes, each pursuing a different angle. Lanes run the full
// pipeline concurrently:
//
//	LITERATURE → GAP_ANALYSIS → HYPOTHESIS → EXPERIMENT → BENCHMARK → REVIEW
//
// At each stage transition, a Carlini decision gate evaluates whether the
// lane is worth continuing. Lanes that fail the gate are killed early.
package discovery

import (
	"fmt"
	"sync"
	"time"
)

// LaneState is the current phase of a discovery lane.
type LaneState string

const (
	StateLiterature  LaneState = "LITERATURE"
	StateGapAnalysis LaneState = "GAP_ANALYSIS"
	StateHypothesis  LaneState = "HYPOTHESIS"
	StateExperiment  LaneState = "EXPERIMENT"
	StateLaneBench   LaneState = "BENCHMARK"
	StateReview      LaneState = "REVIEW"
	StateLaneDone    LaneState = "DONE"
	StateLaneKilled  LaneState = "KILLED"
)

// Lane is one independent research thread pursuing a specific angle.
type Lane struct {
	mu sync.RWMutex

	ID          string    `json:"id"`
	Topic       string    `json:"topic"`        // parent discovery topic
	Angle       string    `json:"angle"`        // specific research angle, e.g. "grouped_query_attention"
	Description string    `json:"description"`  // human-readable description of what this lane is exploring
	State       LaneState `json:"state"`
	StartedAt   time.Time `json:"started_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	KilledAt    *time.Time `json:"killed_at,omitempty"`
	KillReason  string    `json:"kill_reason,omitempty"`

	// Literature stage output
	Papers []PaperRef `json:"papers,omitempty"`

	// Gap analysis output
	Gaps []Gap `json:"gaps,omitempty"`

	// Hypothesis stage output
	Claim      string `json:"claim,omitempty"`
	Experiment string `json:"experiment,omitempty"`

	// Benchmark results
	Runs       []LaneRun `json:"runs,omitempty"`
	BestMetric float64   `json:"best_metric,omitempty"`
	BestNode   string    `json:"best_node,omitempty"`

	// Review output
	Review     string `json:"review,omitempty"`
	Verdict    string `json:"verdict,omitempty"` // "promising" | "inconclusive" | "dead_end"

	// Error tracking
	Errors []string `json:"errors,omitempty"`
}

// PaperRef is a reference to an ArXiv paper found during literature search.
type PaperRef struct {
	ArXivID  string `json:"arxiv_id"`
	Title    string `json:"title"`
	Authors  string `json:"authors"`
	Abstract string `json:"abstract"`
	Year     int    `json:"year"`
	URL      string `json:"url"`
}

// Gap is a research gap identified during gap analysis.
type Gap struct {
	Description string  `json:"description"`
	Importance  float64 `json:"importance"`  // 0.0-1.0 Carlini taste score
	Novelty     float64 `json:"novelty"`     // 0.0-1.0 how new is this angle
	Feasibility float64 `json:"feasibility"` // 0.0-1.0 can we test this with autoresearch
	Score       float64 `json:"score"`       // weighted composite
}

// LaneRun is one experiment iteration within a lane.
type LaneRun struct {
	RunNumber int     `json:"run_number"`
	Node      string  `json:"node"`
	MetricVal float64 `json:"metric_value"`
	Delta     float64 `json:"delta"`
	Status    string  `json:"status"` // "improvement" | "regression" | "crash"
	Timestamp string  `json:"timestamp"`
}

// LaneProgress is sent on the progress channel after each state transition.
type LaneProgress struct {
	LaneID  string
	Angle   string
	State   LaneState
	Message string
	Killed  bool
}

// Transition moves the lane to a new state and updates the timestamp.
func (l *Lane) Transition(state LaneState) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.State = state
	l.UpdatedAt = time.Now()
}

// Kill stops the lane and records the reason (Carlini gate rejection).
func (l *Lane) Kill(reason string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	l.State = StateLaneKilled
	l.KilledAt = &now
	l.KillReason = reason
	l.UpdatedAt = now
}

// AddError records a non-fatal error.
func (l *Lane) AddError(err error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.Errors = append(l.Errors, fmt.Sprintf("[%s] %v", time.Now().Format("15:04:05"), err))
}

// IsAlive returns true if the lane hasn't been killed or completed.
func (l *Lane) IsAlive() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.State != StateLaneKilled && l.State != StateLaneDone
}

// Summary returns a one-line summary for TUI display.
func (l *Lane) Summary() string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if l.State == StateLaneKilled {
		return fmt.Sprintf("%-30s  KILLED  %s", l.Angle, l.KillReason)
	}
	runs := len(l.Runs)
	if runs > 0 {
		return fmt.Sprintf("%-30s  %-12s  runs=%d  best=%.4f", l.Angle, l.State, runs, l.BestMetric)
	}
	return fmt.Sprintf("%-30s  %-12s", l.Angle, l.State)
}

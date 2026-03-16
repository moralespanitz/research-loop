// Package loop implements the Research Loop experiment state machine.
//
// State transitions:
//
//	IDLE → HYPOTHESIZE → PROPOSE → MUTATE → BENCHMARK → ANNOTATE → PROPOSE (next iteration)
//	                                                               ↘ DONE (max runs reached)
package loop

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// LoopState is the current phase of the experiment loop.
type LoopState string

const (
	StateIdle       LoopState = "IDLE"
	StateHypothesize LoopState = "HYPOTHESIZE" // Epistemic agent reads hypothesis.md
	StatePropose    LoopState = "PROPOSE"      // Epistemic agent proposes next mutation
	StateMutate     LoopState = "MUTATE"       // Empirical agent applies code change
	StateBenchmark  LoopState = "BENCHMARK"    // Empirical agent runs benchmark command
	StateAnnotate   LoopState = "ANNOTATE"     // Epistemic agent writes KG annotation
	StateDone       LoopState = "DONE"
	StateFailed     LoopState = "FAILED"
)

// RunStatus is the outcome of a single benchmark run.
type RunStatus string

const (
	StatusImprovement RunStatus = "improvement"
	StatusRegression  RunStatus = "regression"
	StatusChecksFlailed RunStatus = "checks_failed"
	StatusCrash       RunStatus = "crash"
)

// RunRecord is one completed experiment iteration, written to autoresearch.jsonl.
type RunRecord struct {
	Event       string    `json:"event"`
	RunNumber   int       `json:"run_number"`
	State       LoopState `json:"state"`
	Node        string    `json:"node"`        // short name of the mutation tried
	Mutation    string    `json:"mutation"`    // description of the change
	Result      string    `json:"result"`      // metric value as string, e.g. "3.21"
	MetricVal   float64   `json:"metric_value"`
	MetricRaw   string    `json:"metric_raw"`  // raw parsed string before float conversion
	BaselineVal float64   `json:"baseline_value"`
	Delta       float64   `json:"delta"`       // result - baseline (negative = improvement for "lower")
	Status      RunStatus `json:"status"`
	Annotation  string    `json:"annotation"`  // causal note from Epistemic agent
	DiffPath    string    `json:"diff_path"`   // relative path to saved git diff
	BenchOutput string    `json:"bench_output,omitempty"` // last N lines of benchmark stdout
	Proposal    Proposal  `json:"proposal"`    // the mutation that was applied
	Timestamp   string    `json:"timestamp"`
}

// Checkpoint is written to autoresearch.jsonl at every state transition so the
// loop can be resumed after a crash, reboot, or context reset.
type Checkpoint struct {
	Event       string    `json:"event"`
	State       LoopState `json:"state"`
	RunNumber   int       `json:"run_number"`
	BaselineVal float64   `json:"baseline_value,omitempty"`
	BestVal     float64   `json:"best_value,omitempty"`
	BestNode    string    `json:"best_node,omitempty"`
	SessionID   string    `json:"session_id"`
	Timestamp   string    `json:"timestamp"`
}

// Progress is sent on the Progress channel after each state transition.
type Progress struct {
	State     LoopState
	RunNumber int
	Message   string
	Record    *RunRecord // non-nil when a run completes
}

// ─── Checkpoint persistence ───────────────────────────────────────────────────

// SaveCheckpoint appends a Checkpoint to autoresearch.jsonl.
func SaveCheckpoint(jsonlPath string, cp Checkpoint) error {
	cp.Timestamp = time.Now().UTC().Format(time.RFC3339)
	if cp.Event == "" {
		cp.Event = "checkpoint"
	}
	data, err := json.Marshal(cp)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(jsonlPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprintf(f, "%s\n", data)
	return err
}

// LoadLastCheckpoint scans autoresearch.jsonl and returns the most recent
// Checkpoint record, so the loop can resume mid-run after a crash.
func LoadLastCheckpoint(jsonlPath string) (*Checkpoint, error) {
	data, err := os.ReadFile(jsonlPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // no checkpoint yet
		}
		return nil, err
	}

	var last *Checkpoint
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if line == "" {
			continue
		}
		var cp Checkpoint
		if err := json.Unmarshal([]byte(line), &cp); err != nil {
			continue
		}
		if cp.Event == "checkpoint" || cp.Event == "run_complete" {
			cp2 := cp
			last = &cp2
		}
	}
	return last, nil
}

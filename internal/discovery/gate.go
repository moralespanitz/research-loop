package discovery

import (
	"context"
	"fmt"
	"strings"

	"github.com/research-loop/research-loop/internal/llm"
)

// CarliniGate evaluates whether a discovery lane should continue at each
// stage transition, applying the decision criteria from Nicholas Carlini's
// "How to win a best paper award" essay.
//
// Gates:
//  1. LITERATURE → GAP_ANALYSIS: "Is this an important problem? Would Carlini pursue it?"
//  2. GAP_ANALYSIS → HYPOTHESIS: "Is this gap novel enough? Is there comparative advantage?"
//  3. HYPOTHESIS → EXPERIMENT:   "Is this the maximal version? Would someone else do it?"
//  4. EXPERIMENT → BENCHMARK:    "Kill papers that are not working — is the idea de-risked?"
//  5. BENCHMARK → REVIEW:        "Did it actually improve? Is the magnitude worth the complexity?"

// GateResult is the output of a Carlini gate evaluation.
type GateResult struct {
	Pass   bool    `json:"pass"`
	Score  float64 `json:"score"`  // 0.0-1.0 confidence
	Reason string  `json:"reason"` // why it passed or was killed
}

const gateSystemPrompt = `You are a research quality gate applying Nicholas Carlini's criteria from
"How to win a best paper award." Your job is to evaluate whether a research
direction is worth pursuing.

Carlini's key decision criteria:
1. TASTE: "Good taste for problems is the single most important skill."
   Does this problem matter? Would solving it teach us something important?
2. KILL EARLY: "Start with the sub-problem most likely to fail."
   Is there a fatal flaw visible already? Don't waste time on doomed ideas.
3. IMPACT: "Pick your ideas for impact. One excellent paper > 1000 mediocre ones."
   Would this result matter if it succeeded?
4. UNIQUENESS: "Do something only you can do."
   Would someone else publish this at the same conference anyway?
5. COMPARATIVE ADVANTAGE: "Find your corner of the high-dimensional space."
   Does the approach leverage unique skills or rare combinations?
6. SIMPLICITY: "All else being equal, simpler is better."
   Is this the cleanest way to test this idea?

Be harsh. Most ideas should be killed. A gate that passes everything is useless.

Respond in this exact format:
PASS: true or false
SCORE: 0.0 to 1.0
REASON: 1-3 sentences explaining why`

// EvaluateGate runs the Carlini gate for a lane at a specific transition.
func EvaluateGate(ctx context.Context, client llm.Client, lane *Lane, from, to LaneState) (GateResult, error) {
	prompt := buildGatePrompt(lane, from, to)

	raw, err := client.Complete(ctx, gateSystemPrompt, []llm.Message{
		{Role: "user", Content: prompt},
	})
	if err != nil {
		// On LLM failure, default to passing (don't block on infra issues)
		return GateResult{Pass: true, Score: 0.5, Reason: fmt.Sprintf("gate skipped (LLM error: %v)", err)}, nil
	}

	return parseGateResult(raw), nil
}

func buildGatePrompt(lane *Lane, from, to LaneState) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("## Discovery Lane: %s\n", lane.Angle))
	sb.WriteString(fmt.Sprintf("Topic: %s\n", lane.Topic))
	sb.WriteString(fmt.Sprintf("Transition: %s → %s\n\n", from, to))

	if lane.Description != "" {
		sb.WriteString(fmt.Sprintf("Description: %s\n\n", lane.Description))
	}

	switch from {
	case StateLiterature:
		sb.WriteString(fmt.Sprintf("Papers found: %d\n", len(lane.Papers)))
		for i, p := range lane.Papers {
			if i >= 5 {
				sb.WriteString(fmt.Sprintf("... and %d more\n", len(lane.Papers)-5))
				break
			}
			sb.WriteString(fmt.Sprintf("- %s (%d) — %s\n", p.Title, p.Year, truncateStr(p.Abstract, 200)))
		}
		sb.WriteString("\nShould we proceed to gap analysis on these papers?")

	case StateGapAnalysis:
		sb.WriteString("Gaps identified:\n")
		for _, g := range lane.Gaps {
			sb.WriteString(fmt.Sprintf("- [importance=%.1f novelty=%.1f feasibility=%.1f] %s\n",
				g.Importance, g.Novelty, g.Feasibility, g.Description))
		}
		sb.WriteString("\nIs the top gap worth formalizing into a testable hypothesis?")

	case StateHypothesis:
		sb.WriteString(fmt.Sprintf("Claim: %s\n", lane.Claim))
		sb.WriteString(fmt.Sprintf("Experiment: %s\n", lane.Experiment))
		sb.WriteString("\nIs this hypothesis concrete enough and important enough to run experiments?")

	case StateExperiment:
		sb.WriteString(fmt.Sprintf("Runs completed: %d\n", len(lane.Runs)))
		if len(lane.Runs) > 0 {
			last := lane.Runs[len(lane.Runs)-1]
			sb.WriteString(fmt.Sprintf("Last run: %s metric=%.4f delta=%+.4f status=%s\n",
				last.Node, last.MetricVal, last.Delta, last.Status))
		}
		sb.WriteString(fmt.Sprintf("Best metric so far: %.4f (%s)\n", lane.BestMetric, lane.BestNode))
		sb.WriteString("\nShould we continue experimenting or move to benchmark review?")

	case StateLaneBench:
		sb.WriteString(fmt.Sprintf("Best metric: %.4f (%s)\n", lane.BestMetric, lane.BestNode))
		sb.WriteString(fmt.Sprintf("Total runs: %d\n", len(lane.Runs)))
		improvements := 0
		for _, r := range lane.Runs {
			if r.Status == "improvement" {
				improvements++
			}
		}
		sb.WriteString(fmt.Sprintf("Improvements: %d / %d\n", improvements, len(lane.Runs)))
		sb.WriteString("\nIs this result significant enough to write up in a review?")
	}

	return sb.String()
}

func parseGateResult(raw string) GateResult {
	r := GateResult{Pass: true, Score: 0.5, Reason: "could not parse gate response"}
	lines := strings.Split(raw, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "PASS:") {
			val := strings.TrimSpace(strings.TrimPrefix(line, "PASS:"))
			r.Pass = strings.EqualFold(val, "true")
		}
		if strings.HasPrefix(line, "SCORE:") {
			fmt.Sscanf(strings.TrimPrefix(line, "SCORE:"), "%f", &r.Score)
		}
		if strings.HasPrefix(line, "REASON:") {
			r.Reason = strings.TrimSpace(strings.TrimPrefix(line, "REASON:"))
		}
	}
	return r
}

func truncateStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

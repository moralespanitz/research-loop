// Package hypothesis extracts structured research hypotheses from papers using an LLM.
package hypothesis

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/research-loop/research-loop/internal/ingestion"
	"github.com/research-loop/research-loop/internal/llm"
)

// Hypothesis is the structured output of the Epistemic agent's first pass.
type Hypothesis struct {
	PaperTitle     string
	PaperAuthors   []string
	ArXivID        string
	CoreClaim      string
	KeyInsight     string
	MathFormulation string
	ProposedExperiment string
	BaselineRepo   string
	Metric         string
	GeneratedAt    time.Time
}

const systemPrompt = `You are the Epistemic Agent in Research Loop — a scientific hypothesis extractor.
Your job is to read a research paper and extract a precise, actionable research hypothesis
that can be tested by running code experiments.

Be specific, not vague. Use concrete terms, not hedged language.
Your output will be used to guide autonomous experiment agents.`

const extractionPromptTemplate = `Paper title: %s
Authors: %s
ArXiv ID: %s

%s

---

Extract the following from this paper. Be specific and concrete.

CORE_CLAIM: (1-2 sentences) The central empirical or theoretical claim of this paper. What does it assert is true?

KEY_INSIGHT: (1 sentence) The single most important novel idea — the "aha" that makes this paper work.

MATH_FORMULATION: (1-3 sentences or equations) The key mathematical relationship or formulation, if any. Use plain text or LaTeX inline math.

PROPOSED_EXPERIMENT: (2-4 sentences) A concrete experiment that would test the core claim on a standard ML baseline (e.g., nanoGPT on OpenWebText, or a simple PyTorch model). What would you modify, what would you measure?

BASELINE_REPO: (1 line) The most appropriate open-source baseline repository for this experiment (e.g., "karpathy/nanoGPT", "huggingface/transformers", or "custom"). If unclear, write "nanoGPT".

METRIC: (1 line) The primary metric to optimize (e.g., "val_loss", "val_bpb", "accuracy"). Include direction: "lower" or "higher".

Respond with exactly these labeled fields and nothing else.`

// Extract calls the LLM to produce a structured Hypothesis from a Paper.
func Extract(ctx context.Context, client llm.Client, paper *ingestion.Paper) (*Hypothesis, error) {
	// Use full text if available, otherwise abstract only
	textSection := buildTextSection(paper)

	prompt := fmt.Sprintf(extractionPromptTemplate,
		paper.Title,
		strings.Join(paper.Authors, ", "),
		paper.ID,
		textSection,
	)

	response, err := client.Complete(ctx, systemPrompt, []llm.Message{
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return nil, fmt.Errorf("LLM extraction failed: %w", err)
	}

	h, err := parseResponse(response)
	if err != nil {
		return nil, fmt.Errorf("parsing LLM response: %w", err)
	}

	h.PaperTitle = paper.Title
	h.PaperAuthors = paper.Authors
	h.ArXivID = paper.ID
	h.GeneratedAt = time.Now()

	return h, nil
}

// buildTextSection returns full text (truncated) or abstract-only with a notice.
func buildTextSection(paper *ingestion.Paper) string {
	if paper.FullText != "" {
		// ~12,000 chars ≈ 3,000 tokens — enough for most models, keeps cost reasonable
		truncated := ingestion.TruncateText(paper.FullText, 12000)
		return "FULL TEXT (truncated):\n" + truncated
	}
	if paper.Abstract != "" {
		return "ABSTRACT (full text unavailable):\n" + paper.Abstract +
			"\n\n[Note: working from abstract only — some fields may be less precise]"
	}
	return "[No text available — please provide the paper text manually]"
}

// parseResponse parses the labeled field output from the LLM.
func parseResponse(raw string) (*Hypothesis, error) {
	h := &Hypothesis{}
	lines := strings.Split(raw, "\n")

	var currentKey string
	var currentVal strings.Builder

	flush := func() {
		if currentKey == "" {
			return
		}
		val := strings.TrimSpace(currentVal.String())
		switch currentKey {
		case "CORE_CLAIM":
			h.CoreClaim = val
		case "KEY_INSIGHT":
			h.KeyInsight = val
		case "MATH_FORMULATION":
			h.MathFormulation = val
		case "PROPOSED_EXPERIMENT":
			h.ProposedExperiment = val
		case "BASELINE_REPO":
			h.BaselineRepo = val
		case "METRIC":
			h.Metric = val
		}
		currentVal.Reset()
	}

	for _, line := range lines {
		if idx := strings.Index(line, ":"); idx > 0 {
			key := strings.TrimSpace(line[:idx])
			if isKnownKey(key) {
				flush()
				currentKey = key
				currentVal.WriteString(strings.TrimSpace(line[idx+1:]))
				continue
			}
		}
		if currentKey != "" {
			currentVal.WriteString("\n")
			currentVal.WriteString(line)
		}
	}
	flush()

	if h.CoreClaim == "" {
		return nil, fmt.Errorf("LLM did not return a CORE_CLAIM field; raw response:\n%s", raw)
	}
	return h, nil
}

func isKnownKey(k string) bool {
	switch k {
	case "CORE_CLAIM", "KEY_INSIGHT", "MATH_FORMULATION",
		"PROPOSED_EXPERIMENT", "BASELINE_REPO", "METRIC":
		return true
	}
	return false
}

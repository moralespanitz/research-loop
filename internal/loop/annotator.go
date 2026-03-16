package loop

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/research-loop/research-loop/internal/llm"
)

const annotateSystemPrompt = `You are the Epistemic Agent in Research Loop.
Your job is to write a precise causal annotation after each experiment run.

Rules:
- Be concise (1-3 sentences max).
- Explain WHY the result happened, not just what it was.
- If it improved: what mechanism drove the improvement?
- If it regressed: what went wrong? What does this rule out?
- Use language a co-author could quote in a paper.`

const annotatePromptTemplate = `## Hypothesis

%s

## Mutation tried

Node: %s
Change: %s
File: %s

## Result

Metric: %s (baseline was %.4f)
Delta: %+.4f (%s)

## Benchmark output (last 30 lines)

%s

---

Write a 1-3 sentence causal annotation for the knowledge graph.
Explain WHY this result happened. Be specific and scientific.`

// Annotate asks the Epistemic agent to write a causal annotation for a completed run.
func Annotate(ctx context.Context, client llm.Client, hypothesisMD string, r RunRecord) (string, error) {
	direction := "improvement"
	if r.Status == StatusRegression || r.Status == StatusCrash {
		direction = "regression"
	}

	lastLines := lastNLines(r.BenchOutput, 30)

	prompt := fmt.Sprintf(annotatePromptTemplate,
		hypothesisMD,
		r.Node, r.Mutation, r.Proposal.FilePath,
		r.MetricRaw, r.BaselineVal,
		r.Delta, direction,
		lastLines,
	)

	annotation, err := client.Complete(ctx, annotateSystemPrompt, []llm.Message{
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return "", fmt.Errorf("Epistemic agent annotation failed: %w", err)
	}
	return strings.TrimSpace(annotation), nil
}

// AppendKnowledgeGraph writes a new node to knowledge_graph.md.
func AppendKnowledgeGraph(kgPath string, r RunRecord) error {
	data, err := os.ReadFile(kgPath)
	if err != nil {
		return err
	}

	content := string(data)

	// Build the new node entry
	statusEmoji := "✓"
	if r.Status == StatusRegression {
		statusEmoji = "✗"
	} else if r.Status == StatusCrash || r.Status == StatusChecksFlailed {
		statusEmoji = "⚠"
	}

	node := fmt.Sprintf(`
### Run #%d — %s %s

- **Mutation**: %s
- **File**: %s
- **Metric**: %s (baseline: %.4f, delta: %+.4f)
- **Status**: %s
- **Annotation**: %s
- **Timestamp**: %s

`,
		r.RunNumber, r.Node, statusEmoji,
		r.Mutation,
		r.Proposal.FilePath,
		r.MetricRaw, r.BaselineVal, r.Delta,
		string(r.Status),
		r.Annotation,
		time.Now().Format("2006-01-02 15:04"),
	)

	// Insert under "## Explored Paths" or append at end
	if idx := strings.Index(content, "## Explored Paths"); idx >= 0 {
		insertAt := idx + len("## Explored Paths")
		content = content[:insertAt] + "\n" + node + content[insertAt:]
	} else {
		content += node
	}

	// Update "## Status" section
	content = updateKGStatus(content, r)

	return os.WriteFile(kgPath, []byte(content), 0644)
}

// UpdateKGSummary rewrites the ## Status section with current totals.
func updateKGStatus(content string, r RunRecord) string {
	statusLine := fmt.Sprintf("\nLast run: #%d %s — metric %.4f (delta %+.4f)\n",
		r.RunNumber, r.Node, r.MetricVal, r.Delta)

	if idx := strings.Index(content, "## Status"); idx >= 0 {
		// Find end of Status section (next ## heading)
		after := content[idx:]
		end := strings.Index(after[3:], "\n## ")
		if end > 0 {
			sectionEnd := idx + 3 + end
			content = content[:idx] + "## Status\n" + statusLine + "\n" + content[sectionEnd:]
		}
	}
	return content
}

// AppendLabNotebook appends a run summary to lab_notebook.md.
func AppendLabNotebook(notebookPath string, r RunRecord) error {
	f, err := os.OpenFile(notebookPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	statusLine := fmt.Sprintf("**%s**", string(r.Status))
	entry := fmt.Sprintf(`
---

## Run #%d — %s (%s)

- **Date**: %s
- **Mutation**: %s (%s)
- **Metric**: %s  baseline %.4f  delta %+.4f
- **Verdict**: %s
- **Why**: %s

`,
		r.RunNumber, r.Node, time.Now().Format("2006-01-02 15:04"),
		time.Now().Format("2006-01-02"),
		r.Node, r.Mutation,
		r.MetricRaw, r.BaselineVal, r.Delta,
		statusLine,
		r.Annotation,
	)

	_, err = f.WriteString(entry)
	return err
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func lastNLines(s string, n int) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	if len(lines) <= n {
		return s
	}
	return strings.Join(lines[len(lines)-n:], "\n")
}

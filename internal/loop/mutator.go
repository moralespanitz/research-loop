package loop

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/research-loop/research-loop/internal/llm"
)

// metricRE matches the autoresearch/karpathy summary format: "val_bpb:          0.997900"
// Also matches "val_loss: 3.21", "accuracy: 0.923", etc.
var metricRE = regexp.MustCompile(`(?i)^(val_bpb|val_loss|val_ppl|accuracy|perplexity|[a-z_]+)\s*:\s*([\d.]+)`)

// metricAssignRE matches METRIC name=value (our own convention)
var metricAssignRE = regexp.MustCompile(`(?i)METRIC\s+\S+=\s*([\d.]+)`)

// valueRE is a bare float fallback
var valueRE = regexp.MustCompile(`^\s*([\d]+\.[\d]+)\s*$`)

// Proposal is what the Epistemic agent proposes to try next.
type Proposal struct {
	Node        string // short slug, e.g. "gqa_4_groups"
	Description string // what to change and why
	FilePath    string // which file to modify (relative to repo root)
	Diff        string // unified diff or full replacement content
}

// MutationResult is the outcome of applying a mutation and running the benchmark.
type MutationResult struct {
	Proposal    Proposal
	MetricVal   float64
	MetricRaw   string
	BenchOutput string
	Err         error
	TimedOut    bool
}

// ─── Propose ─────────────────────────────────────────────────────────────────

const proposeSystemPrompt = `You are the Epistemic Agent in Research Loop.
Your job is to propose the next concrete code mutation to test a research hypothesis.

Rules:
- Propose exactly ONE change per turn.
- Be specific: name the file, the line/function, the exact change.
- Avoid changes you've already tried (the knowledge graph lists them).
- Prefer small, interpretable changes over large rewrites.
- Name the mutation with a short slug (snake_case, max 40 chars).`

const proposePromptTemplate = `## Hypothesis

%s

## Knowledge Graph (what we've tried so far)

%s

## Previous runs (last 5)

%s

## Current train.py constants (if autoresearch)

%s

---

Propose the next mutation to try. Respond in this exact format:

NODE: <slug, e.g. increase_matrix_lr>
DESCRIPTION: <what to change and why, 1-3 sentences>
FILE: <relative path to the file to modify, e.g. train.py>
DIFF:
<a unified diff OR the full new content of the changed section — for autoresearch, prefer showing only the changed constant lines>
END_DIFF`

// Propose asks the Epistemic agent to generate the next mutation to try.
// repoDir is passed so we can read current train.py constants for autoresearch.
func Propose(ctx context.Context, client llm.Client, hypothesisMD, kgMD string, lastRuns []RunRecord, repoDir ...string) (Proposal, error) {
	lastRunsText := formatLastRuns(lastRuns)

	// Read current hyperparameter constants from train.py if present (autoresearch pattern)
	trainConstants := readTrainConstants(repoDir...)

	prompt := fmt.Sprintf(proposePromptTemplate, hypothesisMD, kgMD, lastRunsText, trainConstants)
	raw, err := client.Complete(ctx, proposeSystemPrompt, []llm.Message{
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return Proposal{}, fmt.Errorf("Epistemic agent propose failed: %w", err)
	}

	return parseProposal(raw)
}

func parseProposal(raw string) (Proposal, error) {
	p := Proposal{}
	lines := strings.Split(raw, "\n")
	inDiff := false
	var diffLines []string

	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "NODE:"):
			p.Node = strings.TrimSpace(strings.TrimPrefix(line, "NODE:"))
		case strings.HasPrefix(line, "DESCRIPTION:"):
			p.Description = strings.TrimSpace(strings.TrimPrefix(line, "DESCRIPTION:"))
		case strings.HasPrefix(line, "FILE:"):
			p.FilePath = strings.TrimSpace(strings.TrimPrefix(line, "FILE:"))
		case strings.TrimSpace(line) == "DIFF:":
			inDiff = true
		case strings.TrimSpace(line) == "END_DIFF":
			inDiff = false
		case inDiff:
			diffLines = append(diffLines, line)
		}
	}
	p.Diff = strings.Join(diffLines, "\n")

	if p.Node == "" {
		p.Node = fmt.Sprintf("mutation_%d", time.Now().Unix())
	}
	if p.FilePath == "" || p.Diff == "" {
		return Proposal{}, fmt.Errorf("Epistemic agent did not produce a valid FILE or DIFF\nRaw:\n%s", raw)
	}
	return p, nil
}

// ─── Mutate ──────────────────────────────────────────────────────────────────

// ApplyMutation writes the diff/patch to the target file inside repoDir.
// If the diff starts with "---", it's treated as a unified diff and applied
// via `git apply`. Otherwise it's treated as a full file replacement.
func ApplyMutation(repoDir string, p Proposal) error {
	target := filepath.Join(repoDir, p.FilePath)

	if strings.HasPrefix(strings.TrimSpace(p.Diff), "---") {
		// Unified diff — write to a temp file and apply with git
		tmp, err := os.CreateTemp("", "rl-patch-*.diff")
		if err != nil {
			return fmt.Errorf("creating temp diff file: %w", err)
		}
		defer os.Remove(tmp.Name())
		if _, err := tmp.WriteString(p.Diff); err != nil {
			return err
		}
		tmp.Close()

		cmd := exec.Command("git", "apply", "--whitespace=fix", tmp.Name())
		cmd.Dir = repoDir
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("git apply failed: %w\n%s", err, string(out))
		}
		return nil
	}

	// Full replacement — write the diff content as the new file content
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return fmt.Errorf("creating parent dirs: %w", err)
	}
	return os.WriteFile(target, []byte(p.Diff), 0644)
}

// RevertMutation undoes the last change in the repo with `git checkout .`
func RevertMutation(repoDir string) error {
	cmd := exec.Command("git", "checkout", ".")
	cmd.Dir = repoDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git checkout failed: %w\n%s", err, string(out))
	}
	return nil
}

// SaveDiff captures the current working-tree diff and writes it to diffPath.
func SaveDiff(repoDir, diffPath string) error {
	cmd := exec.Command("git", "diff")
	cmd.Dir = repoDir
	out, err := cmd.Output()
	if err != nil {
		return err
	}
	return os.WriteFile(diffPath, out, 0644)
}

// ─── Benchmark ───────────────────────────────────────────────────────────────

// RunBenchmark executes benchmarkCmd in repoDir with the given timeout.
// Supports two patterns:
//
//  1. Plain command: output is captured directly and parsed for metrics.
//  2. autoresearch pattern: if benchmarkCmd ends with "> run.log 2>&1", the
//     command is run via shell and the log file is read back for parsing.
//     This matches karpathy/autoresearch's: uv run train.py > run.log 2>&1
func RunBenchmark(repoDir, benchmarkCmd string, timeoutSecs int) MutationResult {
	if timeoutSecs <= 0 {
		timeoutSecs = 360 // 5 min training + 1 min startup
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSecs)*time.Second)
	defer cancel()

	if benchmarkCmd == "" {
		return MutationResult{Err: fmt.Errorf("benchmark_command is empty — set it in .research-loop/config.toml")}
	}

	var output string

	// Shell-redirect pattern — run via sh -c so redirection works
	isShellRedirect := strings.Contains(benchmarkCmd, ">") || strings.Contains(benchmarkCmd, "|")
	if isShellRedirect {
		cmd := exec.CommandContext(ctx, "sh", "-c", benchmarkCmd)
		cmd.Dir = repoDir
		var buf strings.Builder
		cmd.Stdout = &lineWriter{w: &buf}
		cmd.Stderr = &lineWriter{w: &buf}
		err := cmd.Run()

		if ctx.Err() == context.DeadlineExceeded {
			return MutationResult{BenchOutput: buf.String(), TimedOut: true,
				Err: fmt.Errorf("benchmark timed out after %ds", timeoutSecs)}
		}

		// Read back the log file if redirected (e.g. "> run.log 2>&1")
		logPath := extractLogPath(benchmarkCmd)
		if logPath != "" {
			if logData, readErr := os.ReadFile(filepath.Join(repoDir, logPath)); readErr == nil {
				output = string(logData)
			}
		}
		if output == "" {
			output = buf.String()
		}

		if err != nil {
			// Check for crash in log
			val, raw, parseErr := parseMetric(output)
			if parseErr != nil {
				return MutationResult{BenchOutput: output,
					Err: fmt.Errorf("benchmark command failed: %w", err)}
			}
			return MutationResult{MetricVal: val, MetricRaw: raw, BenchOutput: output}
		}
	} else {
		// Direct command
		parts := strings.Fields(benchmarkCmd)
		cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
		cmd.Dir = repoDir
		var buf strings.Builder
		cmd.Stdout = &lineWriter{w: &buf}
		cmd.Stderr = &lineWriter{w: &buf}
		err := cmd.Run()
		output = buf.String()

		if ctx.Err() == context.DeadlineExceeded {
			return MutationResult{BenchOutput: output, TimedOut: true,
				Err: fmt.Errorf("benchmark timed out after %ds", timeoutSecs)}
		}
		if err != nil {
			return MutationResult{BenchOutput: output,
				Err: fmt.Errorf("benchmark command failed: %w", err)}
		}
	}

	val, raw, parseErr := parseMetric(output)
	if parseErr != nil {
		return MutationResult{BenchOutput: output, Err: parseErr}
	}
	return MutationResult{MetricVal: val, MetricRaw: raw, BenchOutput: output}
}

// extractLogPath pulls the log filename from a shell redirect like "> run.log 2>&1"
func extractLogPath(cmd string) string {
	re := regexp.MustCompile(`>\s*([^\s&|]+)`)
	if m := re.FindStringSubmatch(cmd); m != nil {
		return m[1]
	}
	return ""
}

// parseMetric scans benchmark output for a metric value. Priority order:
//  1. autoresearch format: "val_bpb:          0.997900" (after the "---" summary block)
//  2. METRIC name=value (research-loop convention)
//  3. bare float fallback (last one wins)
func parseMetric(output string) (float64, string, error) {
	// Pass 1: look for the autoresearch "---" summary block
	// Lines after "---" like "val_bpb:  0.997900" are canonical
	inSummary := false
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "---" {
			inSummary = true
			continue
		}
		if inSummary {
			if m := metricRE.FindStringSubmatch(line); m != nil {
				val, err := strconv.ParseFloat(m[2], 64)
				if err == nil {
					return val, m[2], nil
				}
			}
		}
	}

	// Pass 2: METRIC name=value anywhere in output
	scanner2 := bufio.NewScanner(strings.NewReader(output))
	for scanner2.Scan() {
		line := scanner2.Text()
		if m := metricAssignRE.FindStringSubmatch(line); m != nil {
			val, err := strconv.ParseFloat(m[1], 64)
			if err == nil {
				return val, m[1], nil
			}
		}
	}

	// Pass 3: bare float on its own line (last one wins)
	var lastFloat string
	scanner3 := bufio.NewScanner(strings.NewReader(output))
	for scanner3.Scan() {
		if m := valueRE.FindStringSubmatch(scanner3.Text()); m != nil {
			lastFloat = m[1]
		}
	}
	if lastFloat != "" {
		val, err := strconv.ParseFloat(lastFloat, 64)
		if err == nil {
			return val, lastFloat, nil
		}
	}

	return 0, "", fmt.Errorf("no metric value found in benchmark output\n" +
		"For autoresearch: output should contain a '---' summary with 'val_bpb: 0.997'\n" +
		"For custom scripts: print a line like 'METRIC val_loss=3.21'")
}

// readTrainConstants extracts the hyperparameter constant block from train.py.
// Returns a short excerpt of the ALLCAPS constants section, or empty string.
func readTrainConstants(repoDir ...string) string {
	if len(repoDir) == 0 || repoDir[0] == "" {
		return "(no repo dir provided)"
	}
	data, err := os.ReadFile(filepath.Join(repoDir[0], "train.py"))
	if err != nil {
		return "(train.py not found)"
	}
	// Extract lines between "Hyperparameters" comment and the next "---" comment block
	lines := strings.Split(string(data), "\n")
	var out []string
	inHyper := false
	for _, line := range lines {
		if strings.Contains(line, "Hyperparameters") {
			inHyper = true
			continue
		}
		if inHyper {
			if strings.HasPrefix(strings.TrimSpace(line), "# ---") || strings.HasPrefix(strings.TrimSpace(line), "# Setup") {
				break
			}
			// Only emit lines that look like constants (UPPER_CASE = value)
			trimmed := strings.TrimSpace(line)
			if len(trimmed) > 0 && trimmed[0] != '#' && strings.Contains(trimmed, "=") {
				out = append(out, line)
			}
		}
	}
	if len(out) == 0 {
		return "(no constants found in train.py)"
	}
	return strings.Join(out, "\n")
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func formatLastRuns(runs []RunRecord) string {
	if len(runs) == 0 {
		return "(no runs yet)"
	}
	var sb strings.Builder
	for _, r := range runs {
		delta := ""
		if r.Delta < 0 {
			delta = fmt.Sprintf("Δ %.4f ✓ improvement", r.Delta)
		} else {
			delta = fmt.Sprintf("Δ +%.4f ✗ regression", r.Delta)
		}
		sb.WriteString(fmt.Sprintf("- Run #%d  %s  metric=%.4f  %s\n  %s\n",
			r.RunNumber, r.Node, r.MetricVal, delta, r.Annotation))
	}
	return sb.String()
}

// lineWriter is an io.Writer that buffers output.
type lineWriter struct{ w *strings.Builder }

func (lw *lineWriter) Write(p []byte) (int, error) {
	lw.w.Write(p)
	return len(p), nil
}

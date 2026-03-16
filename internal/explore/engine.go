package explore

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/research-loop/research-loop/internal/config"
	"github.com/research-loop/research-loop/internal/llm"
)

type Options struct {
	Topic     string
	MaxPapers int
	MaxRepos  int
}

type Engine struct {
	topic string
	cfg   *config.Config
	opts  Options
	cl    llm.Client

	papers    []Paper
	repos     []Repo
	models    []MentalModel
	debates   []Debate
	questions []DiagnosticQuestion
	score     CarliniScore

	root string
}

type Paper struct {
	Title    string `json:"title"`
	ArXivID  string `json:"arxiv_id"`
	URL      string `json:"url"`
	Year     int    `json:"year"`
	Abstract string `json:"abstract"`
}

type Repo struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Stars       int    `json:"stars"`
}

type MentalModel struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	WhyImportant string `json:"why_important"`
}

type Debate struct {
	Topic      string `json:"topic"`
	SideA      string `json:"side_a"`
	SideB      string `json:"side_b"`
	StrongestA string `json:"strongest_a"`
	StrongestB string `json:"strongest_b"`
}

type DiagnosticQuestion struct {
	Question   string `json:"question"`
	GoodAnswer string `json:"good_answer"`
	BadAnswer  string `json:"bad_answer"`
}

type CarliniScore struct {
	Taste       float64 `json:"taste"`
	Uniqueness  float64 `json:"uniqueness"`
	Impact      float64 `json:"impact"`
	Feasibility float64 `json:"feasibility"`
	Overall     float64 `json:"overall"`
	Verdict     string  `json:"verdict"` // "promising" | "marginal" | "skip"
	Reasoning   string  `json:"reasoning"`
}

func New(workspaceRoot string, cfg *config.Config, opts Options) (*Engine, error) {
	cl, err := llm.New(cfg.LLM)
	if err != nil {
		return nil, fmt.Errorf("creating LLM client: %w", err)
	}

	if opts.MaxPapers == 0 {
		opts.MaxPapers = 15
	}
	if opts.MaxRepos == 0 {
		opts.MaxRepos = 10
	}

	return &Engine{
		topic: opts.Topic,
		cfg:   cfg,
		opts:  opts,
		cl:    cl,
		root:  workspaceRoot,
	}, nil
}

func (e *Engine) Run(ctx context.Context) error {
	fmt.Println("📚 Phase 1: Gathering literature...")

	if err := e.gatherPapers(ctx); err != nil {
		return fmt.Errorf("gathering papers: %w", err)
	}
	fmt.Printf("   Found %d papers\n", len(e.papers))

	if err := e.gatherRepos(ctx); err != nil {
		return fmt.Errorf("gathering repos: %w", err)
	}
	fmt.Printf("   Found %d repos\n", len(e.repos))

	fmt.Println("\n🧠 Phase 2: Extracting mental models...")

	if err := e.extractMentalModels(ctx); err != nil {
		return fmt.Errorf("extracting mental models: %w", err)
	}
	fmt.Printf("   Extracted %d mental models\n", len(e.models))

	fmt.Println("\n💬 Phase 3: Finding field debates...")

	if err := e.findDebates(ctx); err != nil {
		return fmt.Errorf("finding debates: %w", err)
	}
	fmt.Printf("   Found %d debates\n", len(e.debates))

	fmt.Println("\n❓ Phase 4: Generating diagnostic questions...")

	if err := e.generateQuestions(ctx); err != nil {
		return fmt.Errorf("generating questions: %w", err)
	}
	fmt.Printf("   Generated %d questions\n", len(e.questions))

	fmt.Println("\n📊 Phase 5: Carlini scoring...")

	if err := e.scoreProblem(ctx); err != nil {
		return fmt.Errorf("scoring: %w", err)
	}

	return e.save()
}

func (e *Engine) gatherPapers(ctx context.Context) error {
	prompt := fmt.Sprintf(`Search for the most important recent papers on: %s

Find %d key papers. For each paper provide:
PAPER_ID: <arxiv id if available>
TITLE: <paper title>
YEAR: <year>
ABSTRACT: <2-3 sentence summary>
---`, e.topic, e.opts.MaxPapers)

	raw, err := e.cl.Complete(ctx, `You are a research librarian. Find the most relevant papers for this topic. Prioritize recent work (2020-2026) and papers with high impact.`, []llm.Message{
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return err
	}

	e.papers = parsePapers(raw)
	return nil
}

func (e *Engine) gatherRepos(ctx context.Context) error {
	prompt := fmt.Sprintf(`Find %d popular GitHub repositories related to: %s

For each repo provide:
REPO: <owner/repo>
DESCRIPTION: <1-2 sentences>
STARS: <approximate star count>
---`, e.opts.MaxRepos, e.topic)

	raw, err := e.cl.Complete(ctx, `You are a developer advocate. Find the most relevant and popular open-source repos. Prioritize repos with good documentation and active maintenance.`, []llm.Message{
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return err
	}

	e.repos = parseRepos(raw)
	return nil
}

func (e *Engine) extractMentalModels(ctx context.Context) error {
	var papersText strings.Builder
	for _, p := range e.papers {
		papersText.WriteString(fmt.Sprintf("- %s (%d): %s\n", p.Title, p.Year, p.Abstract))
	}

	prompt := fmt.Sprintf(`Based on these papers and repos on "%s":

%s

Extract the %d core mental models that every expert in this field shares.
These are the intuitions, frameworks, or mental shortcuts that take years to develop.

For each mental model provide:
MODEL: <name>
DESCRIPTION: <what it is>
WHY_IMPORTANT: <why experts need this>
---`, e.topic, papersText.String(), 5)

	raw, err := e.cl.Complete(ctx, `You are a senior researcher extracting the deep insights that experts carry. Do NOT list facts — list mental models that shape how experts think about problems.`, []llm.Message{
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return err
	}

	e.models = parseModels(raw)
	return nil
}

func (e *Engine) findDebates(ctx context.Context) error {
	var papersText strings.Builder
	for _, p := range e.papers {
		papersText.WriteString(fmt.Sprintf("- %s: %s\n", p.Title, p.Abstract))
	}

	prompt := fmt.Sprintf(`Given these papers on "%s":

%s

Identify the %d places where experts fundamentally disagree.
For each debate provide:
DEBATE: <what they disagree about>
SIDE_A: <one side's position>
SIDE_B: <other side's position>
ARGUMENT_A: <strongest argument for side A>
ARGUMENT_B: <strongest argument for side B>
---`, e.topic, papersText.String(), 3)

	raw, err := e.cl.Complete(ctx, `You are a field historian. Find the genuine intellectual tensions in this field — not minor differences, but fundamental disagreements about how to think about the problem.`, []llm.Message{
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return err
	}

	e.debates = parseDebates(raw)
	return nil
}

func (e *Engine) generateQuestions(ctx context.Context) error {
	var modelsText strings.Builder
	for _, m := range e.models {
		modelsText.WriteString(fmt.Sprintf("- %s: %s\n", m.Name, m.Description))
	}

	prompt := fmt.Sprintf(`Based on these mental models for "%s":

%s

Generate %d diagnostic questions that would expose whether someone deeply understands this subject versus someone who just memorized facts.

For each question provide:
QUESTION: <the question>
GOOD_ANSWER: <what an expert would say>
BAD_ANSWER: <what a memorizer would say>
---`, e.topic, modelsText.String(), 5)

	raw, err := e.cl.Complete(ctx, `You are a thesis examiner. Create questions that separate surface-level understanding from deep comprehension.`, []llm.Message{
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return err
	}

	e.questions = parseQuestions(raw)
	return nil
}

func (e *Engine) scoreProblem(ctx context.Context) error {
	var papersText strings.Builder
	for _, p := range e.papers[:5] {
		papersText.WriteString(fmt.Sprintf("- %s\n", p.Title))
	}

	prompt := fmt.Sprintf(`Evaluate this research problem using Carlini's criteria:

TOPIC: %s

Papers:
%s

Answer each dimension (0.0-1.0):

TASTE: How important is this problem? Would solving it meaningfully change the field?
UNIQUENESS: What is ONLY you can bring? (skills, timing, framing, cross-field)
IMPACT: Best case — what's the conclusion beyond "X% improvement"?
FEASIBILITY: Can this be tested with single-GPU experiments in weeks?

Then provide:
VERDICT: promising | marginal | skip
REASONING: 2-3 sentences on why`, e.topic, papersText.String())

	raw, err := e.cl.Complete(ctx, `You are the Epistemic Agent applying Carlini's methodology. Be ruthlessly honest — if this isn't worth pursuing, say so.`, []llm.Message{
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return err
	}

	e.score = parseScore(raw)

	e.score.Overall = (e.score.Taste*0.3 + e.score.Uniqueness*0.25 + e.score.Impact*0.3 + e.score.Feasibility*0.15)

	return nil
}

func (e *Engine) Summary() string {
	var sb strings.Builder

	sb.WriteString("═" + strings.Repeat("═", 60) + "\n")
	sb.WriteString(fmt.Sprintf("  EXPLORATION: %s\n", e.topic))
	sb.WriteString("═" + strings.Repeat("═", 60) + "\n\n")

	sb.WriteString("📊 CARLINI SCORE\n")
	sb.WriteString(fmt.Sprintf("   Taste:       %.2f\n", e.score.Taste))
	sb.WriteString(fmt.Sprintf("   Uniqueness:  %.2f\n", e.score.Uniqueness))
	sb.WriteString(fmt.Sprintf("   Impact:      %.2f\n", e.score.Impact))
	sb.WriteString(fmt.Sprintf("   Feasibility: %.2f\n", e.score.Feasibility))
	sb.WriteString(fmt.Sprintf("   ─────────────\n"))
	sb.WriteString(fmt.Sprintf("   OVERALL:     %.2f  [%s]\n\n", e.score.Overall, e.score.Verdict))
	sb.WriteString(fmt.Sprintf("   %s\n\n", e.score.Reasoning))

	sb.WriteString("🧠 MENTAL MODELS\n")
	for i, m := range e.models {
		sb.WriteString(fmt.Sprintf("   %d. %s\n      %s\n", i+1, m.Name, truncateStr(m.Description, 60)))
	}
	sb.WriteString("\n")

	sb.WriteString("💬 FIELD DEBATES\n")
	for i, d := range e.debates {
		sb.WriteString(fmt.Sprintf("   %d. %s\n", i+1, d.Topic))
		sb.WriteString(fmt.Sprintf("      A: %s\n", truncateStr(d.SideA, 50)))
		sb.WriteString(fmt.Sprintf("      B: %s\n", truncateStr(d.SideB, 50)))
	}
	sb.WriteString("\n")

	sb.WriteString("❓ DIAGNOSTIC QUESTIONS (sample)\n")
	for i, q := range e.questions[:3] {
		sb.WriteString(fmt.Sprintf("   %d. %s\n", i+1, truncateStr(q.Question, 60)))
	}

	return sb.String()
}

func (e *Engine) save() error {
	dir := filepath.Join(e.root, ".research-loop", "explorations")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	slug := strings.ReplaceAll(strings.ToLower(e.topic), " ", "_")
	if len(slug) > 40 {
		slug = slug[:40]
	}
	filename := fmt.Sprintf("%s_%s.json", slug, time.Now().Format("2006-01-02"))

	type Exploration struct {
		Topic     string               `json:"topic"`
		Papers    []Paper              `json:"papers"`
		Repos     []Repo               `json:"repos"`
		Models    []MentalModel        `json:"mental_models"`
		Debates   []Debate             `json:"debates"`
		Questions []DiagnosticQuestion `json:"questions"`
		Score     CarliniScore         `json:"score"`
		CreatedAt string               `json:"created_at"`
	}

	exp := Exploration{
		Topic:     e.topic,
		Papers:    e.papers,
		Repos:     e.repos,
		Models:    e.models,
		Debates:   e.debates,
		Questions: e.questions,
		Score:     e.score,
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	data, err := json.MarshalIndent(exp, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(dir, filename)
	return os.WriteFile(path, data, 0644)
}

// ─── Parsers ─────────────────────────────────────────────────────────────────

func parsePapers(raw string) []Paper {
	var papers []Paper
	var current Paper
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "---" {
			if current.Title != "" {
				papers = append(papers, current)
			}
			current = Paper{}
			continue
		}
		if strings.HasPrefix(line, "PAPER_ID:") {
			id := strings.TrimSpace(strings.TrimPrefix(line, "PAPER_ID:"))
			current.ArXivID = id
			current.URL = "https://arxiv.org/abs/" + id
		}
		if strings.HasPrefix(line, "TITLE:") {
			current.Title = strings.TrimSpace(strings.TrimPrefix(line, "TITLE:"))
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

func parseRepos(raw string) []Repo {
	var repos []Repo
	var current Repo
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "---" {
			if current.Name != "" {
				repos = append(repos, current)
			}
			current = Repo{}
			continue
		}
		if strings.HasPrefix(line, "REPO:") {
			current.Name = strings.TrimSpace(strings.TrimPrefix(line, "REPO:"))
		}
		if strings.HasPrefix(line, "DESCRIPTION:") {
			current.Description = strings.TrimSpace(strings.TrimPrefix(line, "DESCRIPTION:"))
		}
		if strings.HasPrefix(line, "STARS:") {
			fmt.Sscanf(strings.TrimPrefix(line, "STARS:"), "%d", &current.Stars)
		}
	}
	if current.Name != "" {
		repos = append(repos, current)
	}
	return repos
}

func parseModels(raw string) []MentalModel {
	var models []MentalModel
	var current MentalModel
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "---" {
			if current.Name != "" {
				models = append(models, current)
			}
			current = MentalModel{}
			continue
		}
		if strings.HasPrefix(line, "MODEL:") {
			current.Name = strings.TrimSpace(strings.TrimPrefix(line, "MODEL:"))
		}
		if strings.HasPrefix(line, "DESCRIPTION:") {
			current.Description = strings.TrimSpace(strings.TrimPrefix(line, "DESCRIPTION:"))
		}
		if strings.HasPrefix(line, "WHY_IMPORTANT:") {
			current.WhyImportant = strings.TrimSpace(strings.TrimPrefix(line, "WHY_IMPORTANT:"))
		}
	}
	if current.Name != "" {
		models = append(models, current)
	}
	return models
}

func parseDebates(raw string) []Debate {
	var debates []Debate
	var current Debate
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "---" {
			if current.Topic != "" {
				debates = append(debates, current)
			}
			current = Debate{}
			continue
		}
		if strings.HasPrefix(line, "DEBATE:") {
			current.Topic = strings.TrimSpace(strings.TrimPrefix(line, "DEBATE:"))
		}
		if strings.HasPrefix(line, "SIDE_A:") {
			current.SideA = strings.TrimSpace(strings.TrimPrefix(line, "SIDE_A:"))
		}
		if strings.HasPrefix(line, "SIDE_B:") {
			current.SideB = strings.TrimSpace(strings.TrimPrefix(line, "SIDE_B:"))
		}
		if strings.HasPrefix(line, "ARGUMENT_A:") {
			current.StrongestA = strings.TrimSpace(strings.TrimPrefix(line, "ARGUMENT_A:"))
		}
		if strings.HasPrefix(line, "ARGUMENT_B:") {
			current.StrongestB = strings.TrimSpace(strings.TrimPrefix(line, "ARGUMENT_B:"))
		}
	}
	if current.Topic != "" {
		debates = append(debates, current)
	}
	return debates
}

func parseQuestions(raw string) []DiagnosticQuestion {
	var questions []DiagnosticQuestion
	var current DiagnosticQuestion
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "---" {
			if current.Question != "" {
				questions = append(questions, current)
			}
			current = DiagnosticQuestion{}
			continue
		}
		if strings.HasPrefix(line, "QUESTION:") {
			current.Question = strings.TrimSpace(strings.TrimPrefix(line, "QUESTION:"))
		}
		if strings.HasPrefix(line, "GOOD_ANSWER:") {
			current.GoodAnswer = strings.TrimSpace(strings.TrimPrefix(line, "GOOD_ANSWER:"))
		}
		if strings.HasPrefix(line, "BAD_ANSWER:") {
			current.BadAnswer = strings.TrimSpace(strings.TrimPrefix(line, "BAD_ANSWER:"))
		}
	}
	if current.Question != "" {
		questions = append(questions, current)
	}
	return questions
}

func parseScore(raw string) CarliniScore {
	s := CarliniScore{}
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "TASTE:") {
			fmt.Sscanf(strings.TrimPrefix(line, "TASTE:"), "%f", &s.Taste)
		}
		if strings.HasPrefix(line, "UNIQUENESS:") {
			fmt.Sscanf(strings.TrimPrefix(line, "UNIQUENESS:"), "%f", &s.Uniqueness)
		}
		if strings.HasPrefix(line, "IMPACT:") {
			fmt.Sscanf(strings.TrimPrefix(line, "IMPACT:"), "%f", &s.Impact)
		}
		if strings.HasPrefix(line, "FEASIBILITY:") {
			fmt.Sscanf(strings.TrimPrefix(line, "FEASIBILITY:"), "%f", &s.Feasibility)
		}
		if strings.HasPrefix(line, "VERDICT:") {
			s.Verdict = strings.TrimSpace(strings.TrimPrefix(line, "VERDICT:"))
		}
		if strings.HasPrefix(line, "REASONING:") {
			s.Reasoning = strings.TrimSpace(strings.TrimPrefix(line, "REASONING:"))
		}
	}
	return s
}

func truncateStr(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

---
name: paper-pipeline
description: >
  End-to-end paper generation pipeline ported from AutoResearchClaw (Aiming
  Lab). 14 phases covering topic initiation through export/publish, with human-
  in-the-loop gates and quality gating at each handoff. Use this when the user
  wants a full paper pipeline run — topic to submission-ready manuscript.
  Delegates to researcher/reviewer/writer/verifier subagents for stage
  execution and to autonomous-iteration for experiment optimization loops.
---

<SUBAGENT-STOP>
If you were dispatched as a subagent to execute a specific pipeline stage,
skip this skill. Execute the task and return structured results immediately.
</SUBAGENT-STOP>

<HARD-GATE>
DO NOT skip phase gates. Every gate stage (Phases 2, 5, 7, 11, 13) requires
explicit human approval before proceeding. No approval = no progression.
</HARD-GATE>

# Paper Pipeline Skill

> One-shot paper generation from topic to export, ported from AutoResearchClaw
> (Aiming Lab). Each phase produces structured artifacts in the session
> directory. Quality gates at every handoff catch issues before they compound.
> The pipeline is NOT linear — phases 8-10 form a tight experiment loop that
> may iterate.

## Relation to Other Skills

| Skill | When to Use |
|-------|-------------|
| `paper-pipeline` (this) | Full end-to-end run — topic to submission-ready manuscript |
| `writing-papers` | Standalone paper drafting from existing results |
| `loop` | Hypothesis testing with experiment ranking |
| `autonomous-iteration` | Metric-driven experiment optimization |
| `experiment-sandbox` | Running experiments in sandboxed environments |
| `figure-agent` | Generating publication-quality figures |
| `review-prep` | Handling reviews post-submission |

## Pipeline Overview

```
Phase A — Topic & Scope (Phases 1-2)
  Phase 1: Topic Init     → SMART research goal
  Phase 2: Literature Coll → Paper search [GATE]

Phase B — Knowledge & Synthesis (Phases 3-6)
  Phase 3: Literature Screen  → Relevance + quality scoring
  Phase 4: Knowledge Extract  → Structured evidence cards
  Phase 5: Hypothesis Gen     → Falsifiable predictions [GATE]
  Phase 6: Synthesis          → Topic clusters + research gaps

Phase C — Experiment (Phases 7-10)
  Phase 7: Experiment Design  → Compute budget, ablations, baselines [GATE]
  Phase 8: Code Generation    → Runnable experiment code
  Phase 9: Execution          → Sandbox running (ref experiment-sandbox)
  Phase 10: Result Analysis   → Statistical interpretation

Phase D — Paper & Review (Phases 11-14)
  Phase 11: Paper Draft       → NeurIPS/ICML/ICLR quality [GATE]
  Phase 12: Peer Review       → Simulated review + inline annotations
  Phase 13: Revision          → Address reviewer feedback [GATE]
  Phase 14: Export/Publish    → Conference template formatting
```

## Gate Stages

| Phase | Gate | Fail Action |
|-------|------|-------------|
| 2 (Literature Coll) | Sufficient papers collected? (< 10 = fail) | Retry search with broader queries |
| 5 (Hypothesis Gen) | At least 2 falsifiable hypotheses? | Iterate synthesis |
| 7 (Experiment Design) | Budget + baselines + ablations defined? | Refine design |
| 11 (Paper Draft) | All sections min word count met? | Expand draft |
| 13 (Revision) | All reviewer points addressed? | Iterate revision |

## Phase 1 — Topic Init

Goal: Establish SMART research goal, scope, and success criteria.

**Artifact:** `sessions/<slug>/01-topic-init.md`

Steps:
1. Define the research topic as a single, precise sentence
2. Create SMART goal (Specific, Measurable, Achievable, Relevant, Time-bound)
3. Set scope boundaries — what is IN and OUT
4. Define success criteria with measurable thresholds
5. List constraints: compute budget, data availability, timeline
6. Document the target venue (NeurIPS, ICML, ICLR, etc.) and format requirements

**Quality gate:** Goal must be ONE sentence. If it takes a paragraph to describe,
the topic is too broad. Split into sub-projects.

**Delegation:** Dispatch the `researcher` subagent if domain context is needed.

## Phase 2 — Literature Collection [GATE]

Goal: Gather candidate papers from arxiv, Semantic Scholar, and web sources.

**Artifact:** `sessions/<slug>/02-literature-collection.md`

Steps:
1. Delegate to `researcher` subagent:
   - Search arxiv with keyword queries (at least 3 diverse query angles)
   - Search Semantic Scholar for related work
   - Search Google Scholar / web for preprints and blog posts
2. Collect at least 20 candidate papers
3. For each paper record: title, authors, year, source, URL, abstract, citation key
4. Create a search plan document with query strings and results per source

**GATE:** Show the candidate list to the user. Ask: "Do you want to adjust the
search direction before we screen?" Wait for approval before proceeding.

**Quality thresholds:**
- Minimum 20 candidates (fewer = search too narrow)
- At least 3 distinct sources
- Coverage of both classic and recent work (last 3 years)

## Phase 3 — Literature Screening

Goal: Filter candidates to a shortlist of high-quality, relevant papers.

**Artifact:** `sessions/<slug>/03-literature-screen.md`

Steps:
1. Score each candidate on two axes:
   - **Relevance (0-1):** How closely does this paper address the research question?
   - **Quality (0-1):** Venue reputation, citation count, methodology rigor
2. Apply configured quality threshold (default: 0.4 for relevance, 0.3 for quality)
3. For each kept paper, provide a one-sentence keep reason
4. Sort shortlist by combined score (relevance * 0.6 + quality * 0.4)
5. Record the screening methodology so it's reproducible

**Output:** Shortlist of 8-15 papers with scores and keep reasons.

**Anti-patterns:**
- Rejecting a paper solely because it is old — classic works are often foundational
- Keeping a paper solely because it is highly cited — check relevance first
- Removing papers that challenge the research direction — those are the most valuable

## Phase 4 — Knowledge Extraction

Goal: Extract structured evidence cards from the shortlisted papers.

**Artifact:** `sessions/<slug>/04-evidence-cards.md`

Steps:
1. For each paper in the shortlist, extract:
   - Central problem addressed
   - Method/approach used
   - Key experimental setup (datasets, metrics, baselines)
   - Main quantitative findings
   - Limitations (self-acknowledged or apparent)
   - Citation key for referencing
2. Format as structured evidence cards (one per paper)
3. Identify contradictions between papers — note these for the synthesis phase

**Output:** Structured evidence table with one row per paper, cross-referenced.

**Delegation:** This can be delegated to the `researcher` subagent for bulk card
extraction. Provide the shortlist file as input.

## Phase 5 — Hypothesis Generation [GATE]

Goal: Produce at least 2 falsifiable, testable hypotheses from the evidence.

**Artifact:** `sessions/<slug>/05-hypotheses.md`

Steps:
1. Review evidence cards and identify open questions and contradictions
2. For each hypothesis, define:
   - **Rationale:** What evidence or gap motivates this hypothesis?
   - **Measurable prediction:** What specific numerical outcome would confirm it?
   - **Failure condition:** What result would falsify it?
   - **Experiment sketch:** Brief high-level plan for testing
3. Order hypotheses by information value (which teaches most if confirmed OR falsified)

**GATE:** Show hypotheses to user. Ask: "Which hypothesis should we pursue first?
Do you want to add, remove, or reorder any?" Wait for confirmation.

**Quality check:** Each hypothesis must be falsifiable. If it cannot be proven
wrong by an experiment, it is not a scientific hypothesis. Reformulate.

## Phase 6 — Synthesis

Goal: Organize knowledge into topic clusters and identify research gaps.

**Artifact:** `sessions/<slug>/06-synthesis.md`

Steps:
1. Cluster evidence cards into 3-5 topic groups by theme/methodology
2. For each cluster, write a 2-3 paragraph synthesis:
   - What is known
   - Where there is agreement
   - Where there is disagreement or uncertainty
3. Identify 2-4 research gaps:
   - Gap = an unanswered question that your hypothesis can address
   - For each gap, state what is missing and why it matters
4. Prioritize gaps by feasibility and impact

**Output:** Cluster overview + prioritized gap list with recommended hypothesis.

**Reference:** The `idea-selection` skill can be loaded here if the gaps need
a formal evaluation matrix.

## Phase 7 — Experiment Design [GATE]

Goal: Design a concrete experiment plan with compute budget, ablations, and baselines.

**Artifact:** `sessions/<slug>/07-experiment-design.md`

Steps:
1. Select the winning hypothesis from Phase 5
2. Define experimental conditions:
   - Independent variables (what you change)
   - Dependent variables (what you measure)
   - Controlled variables (what stays fixed)
3. Select baselines — at least 2 competitive methods from the literature
4. Design ablation study — remove one component at a time
5. Choose evaluation metrics with justification
6. Estimate compute budget — time estimate per run, total wall-clock
7. List risks and mitigations (convergence failure, NaN, OOM)

**GATE:** Present the experiment plan to the user. Ask: "Does the budget look
right? Are baselines fair? Ready to generate code?" Wait for approval.

**Format:** YAML experiment plan that can be directly consumed by the code
generation phase and the `experiment-sandbox` skill.

## Phase 8 — Code Generation

Goal: Generate executable experiment code implementing real algorithms.

**Artifact:** `sessions/<slug>/08-code/` — directory with project files

Steps:
1. Generate Python code from the experiment design using `researcher` or direct LLM
2. Code MUST include:
   - Real algorithm implementations (not random number generators)
   - Real objective/loss functions with proper math
   - Deterministic seed setting
   - Convergence stopping criteria (not fixed iterations)
   - Runtime time guard (stop at 80% budget, save partial results)
   - `TIME_ESTIMATE: Xs` print before main loop
   - `results.json` structured output
3. Validate code with AST syntax check and security scan (no subprocess/os.system/eval)
4. If validation fails → auto-repair with `iterative_repair` sub-prompt pattern
5. Write files to the experiment directory with `filename:xxx.py` format

**Quality checks:**
- Code must run end-to-end without errors
- All metrics from the experiment plan must be present in the output
- Seeds must be deterministic for reproducibility

**Reference:** The `experiment-sandbox` skill for sandbox configuration and
the `autonomous-iteration` skill for optimization loops.

## Phase 9 — Execution

Goal: Run experiments in sandboxed environment and collect results.

**Artifact:** `sessions/<slug>/09-results/` — metrics + logs

Steps:
1. Select sandbox mode based on experiment requirements:
   - `local` (venv) — for quick, low-resource experiments
   - `docker` — for isolated, reproducible environments
   - `ssh_remote` — for GPU compute on remote servers
   - `colab` — for Google Colab workflows
2. Create experiment workspace from generated code
3. Run the experiment with compute budget enforcement
4. Collect stdout/stderr, metrics, and `results.json`
5. Log the execution timestamp, environment, and seed

**Quality checks:**
- Verify `results.json` contains all declared metrics
- Check metric values are real (not NaN, not Inf)
- Validate against the time guard — did the guard fire? Was partial data saved?

**Reference:** Load the `experiment-sandbox` skill for detailed sandbox setup
and execution procedures.

## Phase 10 — Result Analysis

Goal: Produce statistical interpretation of experimental results.

**Artifact:** `sessions/<slug>/10-analysis.md`

Steps:
1. Load results from Phase 9 — metrics, logs, `results.json`
2. Compute summary statistics for each condition:
   - Mean, std, min, max across seeds
   - Statistical significance (if multiple trials)
3. Compare against baselines — is the proposed method better?
4. Analyze ablation results — which components matter most?
5. Identify surprising results — anything that contradicts the hypothesis
6. Produce a structured analysis report with:
   - Metrics Summary (with real values)
   - Comparative Findings
   - Statistical Checks
   - Limitations of the analysis
   - Conclusion

**Quality check:** Every number in the report must trace to an actual experiment
output. No approximations, no rounding without disclosure.

## Phase 11 — Paper Draft [GATE]

Goal: Write a full-length conference-quality paper draft.

**Artifact:** `sessions/<slug>/11-draft.md`

Steps:
1. Delegate to the `writer` subagent for draft generation
2. The draft MUST include all standard sections:
   - Title (10-15 words, informative)
   - Abstract (150-250 words with numbers)
   - Introduction (800-1000 words)
   - Related Work (600-800 words, 3-4 thematic groups)
   - Method (1000-1500 words, with equations)
   - Experiments (800-1200 words, setup + baselines)
   - Results (600-800 words, tables + analysis)
   - Discussion (400-600 words)
   - Limitations (200-300 words)
   - Conclusion (200-300 words)
3. EVERY metric value must exactly match Phase 10 analysis — no fabrication
4. Include figure placeholders referencing `figure-agent` output
5. Target venue template (NeurIPS, ICML, ICLR — see writing-papers skill)

**GATE:** Present the draft to the user. Ask for initial feedback before the
peer review phase. Do NOT skip this — first impressions matter.

**Quality check:** Total word count must be 5000-6500 words in main body. If
any section is below minimum, expand with substantive content — not filler.

## Phase 12 — Peer Review

Goal: Simulate peer review with at least 2 reviewer perspectives.

**Artifact:** `sessions/<slug>/12-reviews.md`

Steps:
1. Delegate to the `reviewer` subagent for review generation
2. Must include at least 2 reviewer perspectives (Reviewer A, Reviewer B)
3. Each review must include:
   - Strengths
   - Weaknesses
   - Actionable revision requests
4. Review must check:
   - **Methodology-evidence consistency:** Do paper claims match experiment evidence?
   - **Topic adherence:** Does the paper stay on topic?
   - **Trial counts:** Do reported trials match actual runs?
   - **Length compliance:** Are sections above minimum word counts?
   - **Novelty:** Is the contribution significant?
   - **Baselines:** Are baselines competitive and fairly tuned?
5. Assign scores 1-10 per the rubric:
   - 1-3 Reject (fundamental flaws)
   - 4-5 Borderline (significant weaknesses)
   - 6-7 Weak Accept (solid but not exciting)
   - 8-9 Accept (strong contribution)
   - 10 Strong Accept (exceptional)

**Output:** Structured review with inline annotations and actionable revision
requests.

## Phase 13 — Revision [GATE]

Goal: Address all reviewer feedback while maintaining or increasing word count.

**Artifact:** `sessions/<slug>/13-revision.md`

Steps:
1. Load the draft and all reviews
2. For each reviewer comment:
   - Mark as ADDRESSED or DEFERRED with justification
   - Make the corresponding change in the draft
3. NEVER shorten existing sections — only expand, improve, add
4. Maintain all section minimum word counts
5. After revision, run the `verifier` subagent to check quality:
   - Verify all numbers still match experimental data
   - Verify all reviewer points are addressed
   - Verify word counts per section

**GATE:** Present the revised draft with a change-log. Ask: "Ready for export?
Any lingering concerns?" Wait for approval.

**Quality check:** The revised paper must be longer than or equal to the draft.
If the revision shortened the paper, that is a failure.

## Phase 14 — Export/Publish

Goal: Format the final paper for submission to a conference.

**Artifact:** `sessions/<slug>/14-export/` — final formatted artifacts

Steps:
1. Load the venue template (NeurIPS, ICML, ICLR — see writing-papers skill)
2. Format the paper according to venue guidelines:
   - Page limits (9 pages + references for NeurIPS/ICML/ICLR)
   - Font size, margins, line spacing
   - Section ordering and headers
   - Citation format
3. Generate figures via the `figure-agent` skill
4. Export to final markdown and, if available, LaTeX/PDF
5. Final quality check against venue requirements
6. Write a submission checklist

**Output:** Final paper file + figures + submission checklist.

## Pipeline Configuration

The pipeline can be configured via inline YAML:

```yaml
pipeline:
  skip_phases: []            # Phases to skip (e.g., [3, 6])
  quality_threshold: 0.4     # Min relevance score for literature screen
  target_venue: "NeurIPS"    # One of: NeurIPS, ICML, ICLR
  experiment:
    mode: "sandbox"          # local | docker | ssh_remote | colab
    time_budget_sec: 300
    metric_key: "primary_metric"
    metric_direction: "minimize"
  figure_agent:
    enabled: true
    min_figures: 3
    max_figures: 10
```

## Anti-patterns

- **Starting code generation without experiment design.** Code without a plan
  produces uninterpretable results. Finish Phase 7 first.
- **Skipping literature screening.** Every kept paper must earn its place.
  "It looks relevant" is not a keep reason.
- **Fabricating results.** Every number in the paper must trace to an actual
  experiment output. The verifier checks this.
- **Writing the conclusion before results.** The conclusion reflects on what
  was actually found, not what was expected.
- **Ignoring surprising results.** The most valuable findings are often the
  ones that contradict the hypothesis. Investigate before discarding.

## Verification Checklist

- [ ] Phase 1: Topic is one sentence. SMART goal defined.
- [ ] Phase 2: >= 20 candidates from >= 3 sources. User approved.
- [ ] Phase 3: 8-15 papers in shortlist with scores and keep reasons.
- [ ] Phase 4: Structured evidence cards for all shortlisted papers.
- [ ] Phase 5: >= 2 falsifiable hypotheses. User approved.
- [ ] Phase 6: Topic clusters + prioritized gaps.
- [ ] Phase 7: Experiment plan with budget, baselines, ablations. User approved.
- [ ] Phase 8: Runnable code with real algorithms, validated and secured.
- [ ] Phase 9: Results collected with metrics from plan.
- [ ] Phase 10: Analysis with real numbers, not approximations.
- [ ] Phase 11: Full draft meeting section word counts. User approved.
- [ ] Phase 12: >= 2 reviewer perspectives with actionable feedback.
- [ ] Phase 13: All reviewer feedback addressed. Paper not shortened.
- [ ] Phase 14: Formatted for target venue. Figures generated.

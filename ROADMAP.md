# Roadmap

Research Loop ships in continuous iterations. This document is honest about what works today and what is actively being built.

---

## What ships today (v0.1)

The full conversational research workflow inside Claude Code — no commands required.

**Skills (Claude Code layer)**
- [x] `research-loop` — entry point, conversational advisor, automatic skill routing
- [x] `learn` — MIT grad student methodology + Socratic reverse learning
- [x] `explore` — 4 parallel search agents, lab notebook persistence
- [x] `idea-selection` — conversational Carlini gate (taste, uniqueness, impact, feasibility)
- [x] `discover` — 4 parallel hypothesis lanes with gates between stages
- [x] `loop` — PROPOSE → MUTATE → BENCHMARK → ANNOTATE experiment cycle
- [x] `execution` — result annotation, causal reasoning, continue/pivot/kill
- [x] `bootstrap` — mandatory activation layer, skill-loading protocol, Red Flags, HARD-GATE enforcement
- [x] `deep-research` — multi-phase deep research with subagent dispatch, provenance tracking, integrity verification
- [x] `literature-review` — structured literature review with parallel search, evidence tables, quality scoring
- [x] `replication` — reproduce published results with environment selection and integrity checks
- [x] `autonomous-iteration` — Karpathy-style metric-driven optimization loop (plan/debug/fix/security/ship/scenario/predict/learn/reason/probe)
- [x] `paper-pipeline` — 14-phase end-to-end paper generation pipeline from topic to submission-ready manuscript
- [x] `experiment-sandbox` — sandboxed experiment execution (local, Docker, SSH, Colab) with harness templates
- [x] `figure-agent` — publication-quality figure generation with style-matched conference templates
- [x] `writing-papers` — standalone paper drafting from existing results (Carlini methodology)
- [x] `review-prep` — submission preparation, rejection handling, resubmission evaluation
- [x] `getting-started` — bootstrap entry point, skill catalog, mandatory workflow primer
- [x] `autoresearch` — autonomous nanochat/GPT training experiments (karpathy/autoresearch)

**Skill infrastructure**
- [x] `<SUBAGENT-STOP>` — subagents skip advisor flow, execute tasks directly
- [x] `<HARD-GATE>` — non-negotiable rules enforced in explore, idea-selection, learn
- [x] SessionStart hook — active session state injected automatically on every open
- [x] Descriptions are triggering-condition-only (superpowers pattern)
- [x] Subagent prompts are task-shaped (no more interception bug)
- [x] Embedded skill copies — all skills mirrored in `internal/embed/claude/skills/` for distribution
- [x] Embedded agent definitions — all agents mirrored in `internal/embed/claude/agents/`
- [x] Embedded CLAUDE.md — full routing table, Red Flags, Priority Rules, HARD-GATE in embed
- [x] 21 skills registered in routing tables — no naming conflicts, non-overlapping descriptions
- [x] bootstrap/SKILL.md — complete skill inventory with SUBAGENT-STOP, EXTREMELY-IMPORTANT, Red Flags, Priority, HARD-GATE

**Go binary**
- [x] `research-loop init` — LLM backend configuration
- [x] `research-loop start <arxiv-url>` — paper ingestion → `hypothesis.md`
- [x] `research-loop loop` — experiment loop state machine with JSONL persistence
- [x] `research-loop discover` — parallel discovery orchestrator
- [x] `research-loop explore` — field exploration engine
- [x] `research-loop list / resume / export` — session management
- [x] `research-loop mcp serve` — MCP bridge server
- [x] Knowledge graph (`knowledge_graph.md` Markdown DAG)
- [x] `.research` bundle export/resume

**Repository**
- [x] `CONTRIBUTING.md` — skill writing guide, Fellow manifest format, PR process
- [x] `SECURITY.md` — local-first data policy, threat model, responsible disclosure
- [x] `LICENSE` — MIT
- [x] `CLAUDE.md` — workspace system prompt with clean skill routing table

---

## Phase 2 — Fellows + Full TUI (v0.2)

Target: Q2 2026

Fellows are autonomous scheduled agents that work for you on a cadence — ingesting papers while you sleep, running experiment loops overnight, monitoring ArXiv feeds, drafting paper sections from results.

**Fellows system**
- [ ] `FELLOW.toml` manifest format — name, description, schedule, tools, guardrails
- [ ] Fellow lifecycle: activate / pause / resume / schedule via CLI
- [ ] `research-loop fellow list` — show active fellows and their last run
- [ ] Built-in fellows:
  - [ ] **Ingestor** — scheduled PDF ingestion from ArXiv
  - [ ] **Experimenter** — overnight experiment loops with approval gates
  - [ ] **Librarian** — ArXiv RSS feed monitoring and auto-ingestion
  - [ ] **Scribe** — draft paper sections from accumulated results
  - [ ] **Reviewer** — methodology checking, p-hacking detection
  - [ ] **Replicator** — attempt to reproduce published results
  - [ ] **Deep Learner** — 5-phase corpus pipeline (see below)

**Deep Learner Fellow**
- [ ] Phase 1: Corpus assembly — accept textbooks, papers, transcripts, blog posts
- [ ] Phase 2: Mental model extraction → `mental_models.md`
- [ ] Phase 3: Intellectual landscape mapping → `landscape.md`
- [ ] Phase 4: Deep understanding question generation → `deep_questions.md`
- [ ] Phase 5: Socratic tutoring loop evaluated against full corpus → `learning_notebook.md`
- [ ] Auto-trigger when 3+ papers on same topic accumulate in library

**Full TUI (4 panes)**
- [ ] Dashboard pane (F2): live experiment metrics, color-coded results, cost tracker
- [ ] Library pane (F3): paper browser with full-text search, annotations, citation graph
- [ ] Writer pane (F4): Markdown editor with Vim keybindings, LaTeX preview, Scribe integration

**Paper ingestion pipeline**
- [ ] Full PDF text extraction with mathematical notation support
- [ ] Fallback to abstract-only with notification on parse failure
- [ ] BibTeX and DOI support
- [ ] Structured extraction: title, authors, abstract, core claim, math, baselines

**Core improvements**
- [ ] Semantic similarity check to skip duplicate hypotheses
- [ ] Hardware auto-detection (CUDA / Apple MPS / CPU)
- [ ] Cost tracking: real-time LLM API spend + GPU-hours
- [ ] Budget cap per session with auto-pause
- [ ] Backpressure correctness checks (`research.checks.sh`)
- [ ] Multi-paper sessions: chain papers for cross-hypothesis synthesis

---

## Phase 3 — Collaboration & Distribution (v0.3)

Target: Q3 2026

- [ ] MCP bridge improvements — full tool/resource coverage for Claude Code, OpenCode, Cursor
- [ ] Bundle registry — `research-loop publish` / `research-loop search`
- [ ] Collaborator bundle mode (read/write/run) vs. read-only reviewer mode
- [ ] LaTeX/PDF export from Writer pane (via pandoc)
- [ ] Knowledge graph Mermaid export
- [ ] Cross-session insight synthesis
- [ ] Community Fellow registry

---

## Phase 4 — Production (v1.0)

Target: Q4 2026

- [ ] Replay mode for `.research` bundles
- [ ] Homebrew formula + `curl | sh` installer
- [ ] Documentation site
- [ ] Cold start < 500ms, TUI @ 60fps
- [ ] SQLite FTS5 index for large paper libraries

---

## What we are not building

- A GUI or web dashboard — terminal-first, always
- A cloud service or SaaS product
- A fork of Claude Code or OpenCode
- Multi-node distributed training support
- Automatic institutional login for paywalled papers

---

## How to influence the roadmap

Open an issue with the label `roadmap` and describe:
1. The user problem you are trying to solve
2. How you currently work around it
3. What success would look like

We prioritize based on researcher impact and implementation feasibility.

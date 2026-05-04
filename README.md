# Research Loop

Research Loop is a **zero-dependency Claude Code plugin** that gives your coding agent a complete scientific research workflow. It merges the best patterns from [Superpowers](https://github.com/obra/superpowers) (auto-triggering skill UX), [Feynman](https://github.com/getcompanion-ai/feynman) (research subagents), [Autoresearch](https://github.com/uditgoenka/autoresearch) (autonomous iteration), and [AutoResearchClaw](https://github.com/aiming-lab/AutoResearchClaw) (paper pipeline) — all into one seamless agent experience.

21 composable skills. 4 specialist subagents. Zero dependencies. Clone and go.

## How it works

It starts from the moment you open your coding agent and mention anything research-related. As soon as it sees that you're exploring a topic, it *doesn't* just start searching and dumping information. Instead, it steps back and asks you what you're really trying to figure out.

Once it's teased the research framing out of the conversation, it explores in parallel — papers, repos, debates, open problems — and shows you the landscape in a synthesis short enough to actually read and act on.

After you've picked a direction, your agent finds the gaps, runs them through a Carlini gate (one question at a time, waiting for your answers), and surfaces the ideas worth pursuing. Then it spins up parallel hypothesis lanes, applies gates between them, and kills the weak ones early.

Next up, once you say "go", it launches a *subagent-driven experiment loop* — proposing code mutations, running benchmarks, annotating results causally, and building a living knowledge graph that remembers what was tried and why it failed.

When you want to understand something deeply, just say "explain X." The `learn` skill walks you through how experts actually think about a topic — not a summary, but the underlying reasoning structures: core mental models one at a time, the places where the field genuinely disagrees, questions that expose whether you understand or just recognize, and finally a reverse test where you explain it back. Every gap gets logged to the lab notebook.

Need a paper? Say "write a paper on X" and the `paper-pipeline` skill drives a 14-phase pipeline from topic framing all the way to conference-formatted export.

Need to optimize something? Say "optimize this metric" and the `autonomous-iteration` skill runs a Karpathy-style Modify → Verify → Keep/Discard loop with automatic rollback.

There's a bunch more to it, but that's the core of the system. And because the skills trigger automatically, you don't need to do anything special. Your agent just has Research Loop.

## Sponsorship

If Research Loop has helped you do work that matters and you are so inclined, consider [sponsoring the project](https://github.com/sponsors/moralespanitz).

Thanks!

— Alexander

## Installation

Research Loop is a **zero-dependency Claude Code plugin** — just like [Superpowers](https://github.com/obra/superpowers), it works by dropping skill files into your project. No binaries, no builds, no dependencies.

### Plugin install (recommended — no build required)

```bash
git clone https://github.com/moralespanitz/research-loop
cd research-loop
```

That's it. Open a new Claude Code session in this directory and the skills auto-activate.

### Plugin install into an existing project

```bash
git clone https://github.com/moralespanitz/research-loop /tmp/research-loop
cp -r /tmp/research-loop/.claude /tmp/research-loop/CLAUDE.md /your/project/
```

### Install from the Claude Code marketplace

```
/plugin install research-loop
```

### Go binary (optional — for CLI users)

A Go binary is available for users who want CLI tools (`research-loop init`, `research-loop start <arxiv-url>`, etc.):

```bash
go install github.com/moralespanitz/research-loop/cmd/research-loop@latest
research-loop init
```

### Verify installation

Open a new Claude Code session in the workspace and say anything research-related — "I want to explore transformer memory systems" or "explain policy compression." The `bootstrap` skill loads automatically, then routes to the right sub-skill. You should see the agent announce which skill it's using.

## Skills

Research Loop ships 21 skills organized by workflow layer. Skills auto-trigger — you just talk naturally and the right skill activates.

### Activation Layer

- **bootstrap** — Loads on every conversation start. Establishes the skill-loading protocol, Red Flags table to prevent rationalization, and priority rules. The foundation that makes the entire system work.

### Core Research Workflow

1. **research-loop** — Entry point. Activates when you mention research, a topic, papers, or experiments. Asks one question to confirm framing. Routes to the right sub-skill.
2. **status** — "Where are we?", "what do we have?", "what should I do next." Renders the full decision tree and updates the knowledge graph.
3. **learn** — "Explain X", "what is Y", "I don't understand Z." Builds the reasoning structures experts carry: mental models → field debates → diagnostic questions → reverse test.
4. **explore** — Parallel field mapping. 4 agents simultaneously search papers, repos, debates, and open problems. Saves everything to the lab notebook.
5. **idea-selection** — Conversational Carlini gate: taste, uniqueness, impact, feasibility. One question at a time, scored and saved.
6. **discover** — 4 parallel hypothesis lanes (incremental, cross-field, assumption challenge, systems/efficiency). Gates between stages. Kills weak lanes early.
7. **plan** — Breaks work into concrete actions with file paths, verification steps, and time estimates.
8. **loop** — PROPOSE → MUTATE → BENCHMARK → ANNOTATE experiment cycle. Living knowledge graph.
9. **execution** — Result annotation, causal reasoning, continue/pivot/kill decisions.

### Deep Research & Literature

10. **deep-research** — 7-phase research protocol: Plan → Scale Decision → Gather Evidence (parallel researcher subagents) → Draft → Cite → Review → Deliver with provenance sidecar. Integrity commandments: no fabricated sources, URL-or-it-didn't-happen.
11. **literature-review** — Parallel search across 3-4 angles, evidence table with A/B/C quality scoring, synthesis into topic clusters and research gaps, verifier dispatch for citation anchoring.
12. **replication** — Protocol for reproducing published results: extract → plan → environment → execute with integrity checks → expected vs. observed comparison.

### Autonomous Optimization

13. **autonomous-iteration** — Karpathy-style Modify → Verify → Keep/Discard → Repeat loop. One change per iteration, atomic commits, automatic git revert on failure, mechanical verification only, safety guardrails, interactive setup gate. Supports 11 specialized sub-commands:
    - **debug** — Scientific-method bug hunting
    - **fix** — Autonomous error repair (test/type/lint/build)
    - **security** — STRIDE + OWASP security audit with red-team personas
    - **ship** — 8-phase universal shipping workflow
    - **scenario** — 12-dimension scenario exploration
    - **predict** — Multi-persona swarm prediction
    - **learn** — Autonomous codebase documentation
    - **reason** — Adversarial refinement with blind judge panel
    - **probe** — Adversarial requirement interrogation
    - **plan** — Interactive wizard for building optimization config

### Paper Pipeline

14. **paper-pipeline** — End-to-end 14-phase pipeline: Topic Init → Literature Collection → Screening → Knowledge Extraction → Hypothesis Generation → Synthesis → Experiment Design → Code Generation → Execution → Result Analysis → Paper Draft → Peer Review → Revision → Export/Publish. Human-in-the-loop gates at key decision points.
15. **experiment-sandbox** — Four sandbox modes: local (venv), Docker (containerized), SSH remote, Colab. Code validation pipeline, compute budget enforcement, metric collection, deterministic reproducibility.
16. **figure-agent** — Decision agent for figure type selection (code plots vs. architecture diagrams). Matplotlib/Seaborn generation with conference style presets. Iterative improvement loop with critic feedback.

### Supporting

17. **getting-started** — First-run orientation. Teaches the skill system, routing table, and how to interact with Research Loop.
18. **review-prep** — Pre-submission checklist, methodology checks, reproducibility verification.
19. **writing-papers** — Standalone paper drafting with conference template awareness (NeurIPS, ICML, ICLR).
20. **autoresearch** — Karpathy's original nanochat/GPT autonano experiment protocol (train.py optimization).

## Research Subagents

Research Loop ships 4 specialist subagents for delegated research tasks:

- **researcher** — Evidence gatherer. Integrity commandments, numbered evidence table, source quality tiers (A/B/C/Reject), context hygiene.
- **reviewer** — Structured peer review. FATAL/MAJOR/MINOR classification, inline annotations quoting exact passages.
- **writer** — Evidence-only drafting. No fabrication, open questions section, claim sweep and provenance check before finishing.
- **verifier** — Citation verification. URL checking, source anchoring, orphan detection, removal of unsupported claims.

Subagents are invoked automatically by skills like `deep-research`, `literature-review`, and `paper-pipeline`. You don't call them directly — the workflow dispatches them.

## Session Persistence

Every session accumulates to a lab notebook:

```
.research-loop/sessions/<slug>/
  lab_notebook.md     # everything: framing, papers, gaps, scores, results
  knowledge_graph.md  # living DAG of hypotheses tried and why they failed
  autoresearch.jsonl  # machine-readable experiment history
```

Sessions are resumable. Bundles are portable. Any agent can resume from `lab_notebook.md` alone.

## Usage

Research Loop has no commands to memorize. You talk to Claude Code naturally and the right skill activates automatically.

### Explore a new field

> "I want to research policy compression and dopamine in transformers"

The agent confirms the framing, then spawns 4 parallel searches (papers, repos, debates, open problems) and shows you a synthesis — not a dump. Everything is saved to the lab notebook.

### Run deep research

> "Do a deep research on mechanistic interpretability of mixture of experts"

The agent creates a structured plan, spawns parallel researcher subagents, gathers evidence across papers and web, drafts a report with inline citations, runs a verifier pass, and delivers a final brief with provenance sidecar.

### Write a paper

> "Write a paper on sparse attention mechanisms for long-context transformers"

The `paper-pipeline` skill activates and walks you through 14 phases: from topic framing, literature collection and hypothesis generation, through experiment design, code generation and execution, to paper drafting, peer review, revision, and export in conference format.

### Optimize a metric autonomously

> "Optimize the inference latency of my transformer. Metric: ms per token. Direction: lower."

The `autonomous-iteration` skill activates. It asks for setup details (scope, verify command, guard command), then runs a Karpathy-style loop: modify → benchmark → keep or revert → repeat. Atomic commits, automatic rollback, mechanical verification.

### Debug something systematically

> "The API returns 500 on POST /users intermittently"

The `autonomous-iteration` debug workflow activates. It runs a scientific-method investigation: hypothesize → test → prove/disprove → log → repeat until root cause is found.

### Learn a concept deeply

> "Explain rate-distortion theory"
> "What is the information bottleneck?"
> "I don't understand fast weight programmers"

The `learn` skill activates. You get 5 core mental models one at a time, 3 field debates with both sides steel-manned, 5 diagnostic questions that test real understanding vs. memorization, and finally a Socratic reverse test where you explain it back. Every gap you reveal is logged.

### Find a research gap

> "What hasn't been tried in this space?"
> "Is this idea worth pursuing?"

The Carlini gate runs as a conversation — 4 questions (taste, uniqueness, impact, feasibility), one at a time, with scores saved to the lab notebook. Honest verdict at the end.

### Test multiple angles in parallel

> "Let's explore this from different angles"
> "What are the different ways to approach this hypothesis?"

4 hypothesis lanes run simultaneously — incremental improvement, cross-field transfer, assumption challenge, systems/efficiency. Weak lanes are killed early. You pick the one worth pursuing.

### Run an experiment loop

> "Let's start running experiments"
> "I want to test this hypothesis against karpathy/autoresearch"

The loop proposes a code mutation, applies it, runs the benchmark, and annotates the result causally. Repeat. The knowledge graph grows. You can interrupt at any time and resume exactly where you left off.

### Resume a session

> "Where did we leave off?"
> "Resume my research on dopamine and fast weights"

The lab notebook has everything. The agent reads it and picks up the thread — no re-exploration, no repeated work.

### At any point, ask anything

> "What does Carlini gate mean?"
> "Explain what a hypothesis lane is"
> "I don't understand the knowledge graph"

The `learn` skill activates mid-flow and teaches the concept, then hands control back to whatever you were doing.

## Philosophy

- **Researcher in control** — the agent proposes, you approve, the agent executes
- **One thing at a time** — never dump; always present as choices
- **Parallel by default** — 4 agents simultaneously, not sequentially
- **Persist everything** — lab notebook accumulates every decision and finding
- **Learn, don't just search** — understanding deeply is part of the research process
- **Skills auto-trigger** — no commands to memorize, the agent routes based on what you say

## References

- **Carlini gate** — [Nicholas Carlini, "How to win a best paper award"](https://nicholas.carlini.com/writing/2026/how-to-win-a-best-paper-award.html). The four axes (taste, uniqueness, impact, feasibility) come directly from his framework.
- **Autoresearch** — [Andrej Karpathy's autoresearch](https://github.com/karpathy/autoresearch). The Modify → Verify → Keep/Discard loop concept.
- **Superpowers** — [obra/superpowers](https://github.com/obra/superpowers) by Jesse Vincent. The skill architecture (`<SUBAGENT-STOP>`, `<HARD-GATE>`, Red Flags, bootstrap pattern).
- **Feynman** — [getcompanion-ai/feynman](https://github.com/getcompanion-ai/feynman). Research subagent system with integrity commandments.
- **Autoresearch (extended)** — [uditgoenka/autoresearch](https://github.com/uditgoenka/autoresearch). 11-command autonomous iteration framework.
- **AutoResearchClaw** — [aiming-lab/AutoResearchClaw](https://github.com/aiming-lab/AutoResearchClaw). End-to-end paper pipeline with experiment sandbox.
- **Superpowers skill architecture** — The description convention, priority rules, and Red Flags framework are adapted from Superpowers.

## What's coming

See [ROADMAP.md](ROADMAP.md) for the full plan. Highlights include:

- **Fellows** — autonomous scheduled agents (Ingestor, Experimenter, Librarian, Scribe, Reviewer, Replicator, Deep Learner)
- **Full 4-pane TUI** — dashboard, library, writer panes
- **PDF ingestion pipeline** — full text extraction with mathematical notation
- **Bundle registry** — `research-loop publish` / `research-loop search`
- **Multi-paper sessions** — chain papers for cross-hypothesis synthesis
- **Cost tracking** — real-time LLM API spend + GPU-hours

## Contributing

Skills live directly in this repository. To contribute:

1. Fork the repository
2. Create a branch for your skill or Fellow
3. Follow the guide in `CONTRIBUTING.md`
4. Submit a PR

See `CONTRIBUTING.md` for the complete guide, including skill writing rules, Fellow manifest format, and commit conventions.

## Updating

Pull the latest skills:

```bash
git pull origin main
```

## License

MIT — see `LICENSE` for details.

## Support

- **Issues**: https://github.com/moralespanitz/research-loop/issues
- **Discussions**: https://github.com/moralespanitz/research-loop/discussions
- **Security**: see `SECURITY.md`

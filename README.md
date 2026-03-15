# Research Loop

Research Loop is a complete scientific research workflow for your coding agents, built on top of a set of composable "skills" and some initial instructions that make sure your agent uses them.

## How it works

It starts from the moment you open your coding agent and mention anything research-related. As soon as it sees that you're exploring a topic, it *doesn't* just start searching and dumping information. Instead, it steps back and asks you what you're really trying to figure out.

Once it's teased the research framing out of the conversation, it explores in parallel — papers, repos, debates, open problems — and shows you the landscape in a synthesis short enough to actually read and act on.

After you've picked a direction, your agent finds the gaps, runs them through a Carlini gate (one question at a time, waiting for your answers), and surfaces the ideas worth pursuing. Then it spins up parallel hypothesis lanes, applies gates between them, and kills the weak ones early.

Next up, once you say "go", it launches a *subagent-driven experiment loop* — proposing code mutations, running benchmarks, annotating results causally, and building a living knowledge graph that remembers what was tried and why it failed.

When you want to understand something deeply, just say "explain X." The `learn` skill activates the full MIT grad student methodology: 5 core mental models (one at a time), 3 field debates (both sides steel-manned), 5 diagnostic questions that expose memorization vs. real understanding, and a Socratic reverse test where you teach it back. Every gap you reveal gets logged to the lab notebook.

There's a bunch more to it, but that's the core of the system. And because the skills trigger automatically, you don't need to do anything special. Your agent just has Research Loop.

## Sponsorship

If Research Loop has helped you do work that matters and you are so inclined, consider [sponsoring the project](https://github.com/sponsors/moralespanitz).

Thanks!

— Alexander

## Installation

### Claude Code (recommended)

Clone the repo into your research workspace:

```bash
git clone https://github.com/moralespanitz/research-loop
cd research-loop
```

The `.claude/` directory is picked up automatically by Claude Code. Open a new session and it's active.

### Go binary

```bash
go install github.com/moralespanitz/research-loop/cmd/research-loop@latest
research-loop init
```

Or with the install script:

```bash
curl -fsSL https://raw.githubusercontent.com/moralespanitz/research-loop/main/install.sh | sh
```

### Verify installation

Open a new Claude Code session in the workspace and say anything research-related — "I want to explore transformer memory systems" or "explain policy compression." The agent should automatically load the relevant skill without you typing any command.

## The Basic Workflow

1. **research-loop** — Activates when you mention research, a topic, papers, or experiments. Asks one question to confirm framing. Entry point for everything.

2. **learn** — Activates when you say "explain", "what is", "I don't understand", or ask about any term. Runs the MIT grad student methodology: mental models → debates → diagnostic questions → Socratic reverse test. Compresses field mastery from semesters into days.

3. **explore** — Activates when you want to map a field or find papers. Spawns 4 parallel search agents simultaneously (papers, repos, debates, open problems). Saves everything to the lab notebook. Presents a 3-sentence synthesis, not a data dump.

4. **idea-selection** — Activates when you want to find gaps or evaluate whether an idea is worth pursuing. Runs the conversational Carlini gate: 4 questions (taste, uniqueness, impact, feasibility), one at a time, scored and saved.

5. **discover** — Activates when you want to test multiple angles. Runs 4 parallel hypothesis lanes (incremental, cross-field transfer, assumption challenge, systems/efficiency). Applies Carlini gates between stages. Kills weak lanes early.

6. **loop** — Activates when you have a hypothesis and want to run experiments. Drives the PROPOSE → MUTATE → BENCHMARK → ANNOTATE cycle. Builds a living knowledge graph of what was tried and why it worked or failed.

7. **execution** — Activates when experiments complete. Annotates results causally, updates the lab notebook, and helps you decide: continue, pivot, or kill.

**The agent checks for relevant skills before any response.** Mandatory workflows, not suggestions.

## What's Inside

### Skills (shipping today)

**Learning**
- **learn** — MIT grad student methodology + Socratic reverse learning. Mental models, field debates, diagnostic questions, gap tracking.

**Exploration**
- **research-loop** — Entry point. Conversational advisor. Routes to the right skill based on what you say.
- **explore** — Parallel field mapping. 4 agents simultaneously. Saves full results to lab notebook.

**Idea Development**
- **idea-selection** — Conversational Carlini gate. Taste, uniqueness, impact, feasibility. One question at a time.
- **discover** — Parallel hypothesis lanes. 4 angles. Gates between stages. Kills weak lanes early.

**Experiments**
- **loop** — PROPOSE → MUTATE → BENCHMARK → ANNOTATE cycle. Living knowledge graph.
- **execution** — Result annotation, causal reasoning, continue/pivot/kill decisions.

### Session persistence

Every session accumulates to a lab notebook:

```
.research-loop/sessions/<slug>/
  lab_notebook.md     # everything: framing, papers, gaps, scores, results
  knowledge_graph.md  # living DAG of hypotheses tried and why they failed
  autoresearch.jsonl  # machine-readable experiment history
```

Sessions are resumable. Bundles are portable. Any agent can resume from `lab_notebook.md` alone.

### Go CLI (shipping today)

```bash
research-loop init                     # configure LLM backend
research-loop start <arxiv-url>        # ingest paper, extract hypothesis
research-loop loop start               # start experiment loop
research-loop list                     # list all sessions
research-loop resume <session-id>      # resume a paused session
research-loop export                   # export .research bundle
research-loop mcp serve                # start MCP bridge server
```

## Usage

Research Loop has no commands to memorize. You talk to Claude Code naturally and the right skill activates automatically.

### Explore a new field

> "I want to research policy compression and dopamine in transformers"

The agent confirms the framing, then spawns 4 parallel searches (papers, repos, debates, open problems) and shows you a synthesis — not a dump. Everything is saved to the lab notebook.

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

---

## Philosophy

- **Researcher in control** — the agent proposes, you approve, the agent executes
- **One thing at a time** — never dump; always present as choices
- **Parallel by default** — 4 agents simultaneously, not sequentially
- **Persist everything** — lab notebook accumulates every decision and finding
- **Learn, don't just search** — understanding deeply is part of the research process

## What's coming

Fellows (autonomous scheduled agents), the full 4-pane TUI, PDF ingestion pipeline, MCP bridge improvements, and the bundle registry are in active development. See [ROADMAP.md](ROADMAP.md) for the full plan.

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

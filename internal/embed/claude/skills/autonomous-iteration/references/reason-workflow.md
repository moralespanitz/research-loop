# Reason Workflow — $research-loop reason

Isolated multi-agent adversarial refinement for subjective domains. Generates,
critiques, synthesizes, and judges outputs through repeated rounds until
convergence — producing a lineage of evolving candidates.

**Core idea:** Generate-A → Critic attacks A → Generate-B from task+A+critique
→ Synthesize-AB → Blind judge panel picks winner → winner becomes new A →
repeat until convergence. Every agent is cold-start fresh with no shared
session — prevents sycophancy.

## Trigger

- User invokes `$research-loop reason`
- User says "reason through this", "debate and converge", "iterative argument"
- User wants subjective quality improvement with documented rationale
- Chained from another tool via `--chain reason`

## Loop Support

```
# Unlimited — keep refining until convergence
$research-loop reason

# Bounded — exactly N refinement rounds
$research-loop reason
Iterations: 10

# With task
$research-loop reason
Task: Should we use event sourcing for our order management system?
Domain: software
```

## PREREQUISITE: Interactive Setup

**CRITICAL — BLOCKING PREREQUISITE:** If invoked without task, domain, and
mode all provided, gather context before proceeding.

Adaptive questions (3-5 based on input):

| # | Header | Question | Options |
|---|--------|----------|---------|
| 1 | `Task` | "What should be reasoned about?" | Free text |
| 2 | `Domain` | "What domain?" | "Software", "Product", "Business", "Security", "Research", "Content" |
| 3 | `Mode` | "What refinement mode?" | "Convergent", "Creative", "Debate" |
| 4 | `Judges` | "How many judges?" | "3", "5", "7" |
| 5 | `Chain` | "Chain to another tool?" | "debug", "plan", "fix", "scenario", "No chain" |

Skip setup when Task + Domain + Mode all provided.

## Architecture

```
$research-loop reason
  ├── Phase 1: Setup — Interactive gate + config validation
  ├── Phase 2: Generate-A — Author-A produces first candidate
  ├── Phase 3: Critic — Adversarial attack on A (forced weaknesses)
  ├── Phase 4: Generate-B — Author-B sees task+A+critique, produces B
  ├── Phase 5: Synthesize-AB — Synthesizer produces AB from task+A+B
  ├── Phase 6: Judge Panel — N blind judges pick winner
  ├── Phase 7: Convergence Check — stop if incumbent wins N consecutive
  └── Phase 8: Handoff — write lineage files, optional --chain
```

## Phase 1: Setup — Configuration

- `--iterations N`: bounded mode (overrides convergence)
- `--judges N`: judge count (3-7, odd preferred)
- `--convergence N`: consecutive wins to stop (2-5, default: 3)
- `--mode`: convergent (default), creative, debate
  - convergent: stop when incumbent wins N consecutive rounds
  - creative: never auto-stop, generate diverse candidates
  - debate: no synthesis, judges evaluate A vs B directly
- `--domain`: shapes judge persona expertise
- `--chain`: validate targets

## Phase 2: Generate-A — First Candidate

Author-A receives ONLY the task (cold-start, no history):

```
TASK: [user's question/claim]
INSTRUCTION: Produce your best answer. Be thorough and specific.
```

**Output:** Candidate-A with reasoning documented.

## Phase 3: Critic — Adversarial Attack

Fresh agent (no prior context) receives ONLY Candidate-A:

```
TASK: [original task]
CANDIDATE-A: [full text of A]
INSTRUCTION: Attack this as a strawman. Find minimum 3 weaknesses.
Be ruthless but specific. Point to logical flaws, missing evidence,
unstated assumptions, and alternatives not considered.
```

**Output:** Critique with ≥3 specific weaknesses.

## Phase 4: Generate-B — Contested Alternative

Fresh agent sees: Task + Candidate-A + Critique:

```
TASK: [original task]
CANDIDATE-A: [full text of A]
CRITIQUE: [critique text]
INSTRUCTION: Produce an alternative candidate that addresses the
weaknesses in A. You can adopt parts of A, but must offer meaningful
differences. Do not simply improve A — consider different approaches.
```

**Output:** Candidate-B

## Phase 5: Synthesize-AB — Combined Candidate (convergent mode only)

Fresh agent sees: Task + Candidate-A + Candidate-B (no critique, no judge
history):

```
TASK: [original task]
CANDIDATE-A: [full text of A]
CANDIDATE-B: [full text of B]
INSTRUCTION: Produce a synthesis that captures the best elements of
both. This is not an average — pick the strongest arguments from each
and combine them into a superior whole.
```

**Output:** Candidate-AB

## Phase 6: Judge Panel — Blind Evaluation

N judges, each cold-start, receive randomized labels:

```
TASK: [original task]
CANDIDATE-X: [randomized — could be A, B, or AB]
CANDIDATE-Y: [another candidate]
CANDIDATE-Z: [another candidate]
INSTRUCTION: Rank these from best to worst. You MUST provide a clear
winner. Base your evaluation on: logical coherence, completeness,
practicality, evidence support, and originality. Explain your ranking.
```

**Labels are randomized per judge** (not X=A, Y=B, Z=AB — each judge gets
a different mapping). This prevents label bias.

**Winner:** Candidate with most first-place votes.

## Phase 7: Convergence Check

- If incumbent == winner → consecutive_wins++
- If incumbent != winner → reset consecutive_wins, new incumbent = winner
- If consecutive_wins >= convergence_threshold → CONVERGED → stop
- If oscillation detected (incumbent changes 5+ times without consecutive wins)
  → stop + flag "oscillation detected — no stable convergence"
- If bounded (`--iterations`): stop after N rounds regardless

## Phase 8: Handoff — Output & Chain

**Write lineage files:**
- `overview.md` — Setup, task, domain, configuration, rounds
- `lineage.md` — Round-by-round evolution: incumbent, challengers, winner
- `candidates.md` — All candidate texts with round and author labels
- `judge-transcripts.md` — Full judge reasoning with random labels
- `reason-results.tsv` — Tabular per-round results
- `reason-lineage.jsonl` — Machine-readable per-round records
- `handoff.json` — Winner text + convergence status for `--chain`

## Flags

| Flag | Purpose |
|------|---------|
| `--iterations N` | Bounded mode — run exactly N rounds |
| `--judges N` | Judge count (3-7, odd preferred) |
| `--convergence N` | Consecutive wins to converge (2-5) |
| `--mode <mode>` | convergent, creative, debate |
| `--domain <type>` | software, product, business, security, research, content |
| `--chain <targets>` | Chain converged output to other commands |
| `--judge-personas <list>` | Override default judge personas |
| `--no-synthesis` | Skip synthesis, judge A vs B only |

## Composite Metric

```
reason_score = quality_delta*30 + rounds_survived*5 + judge_consensus*20
              + critic_fatals_addressed*15 + convergence*10 + no_oscillation*5
```

Higher = more refined and stable convergence.

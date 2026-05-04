# Probe Workflow — $research-loop probe

Adversarial multi-persona requirement & assumption interrogation engine.
Probes user and codebase through N personas until net-new constraints per
round drop below a threshold (mechanical saturation), then emits the 5
research-loop primitives ready to feed any other command.

**Core idea:** Topic in → 8 personas interrogate → constraints harvested →
saturation reached → research-loop config out.

## Trigger

- User invokes `$research-loop probe`
- User says "interrogate requirements", "find hidden constraints"
- User says "what am I missing", "stress-test my goal"

## Loop Support

```
# Unlimited — keep probing until saturation
$research-loop probe

# Bounded — hard cap on rounds
$research-loop probe
Iterations: 15

# Focused with full flags
$research-loop probe --depth deep --personas 8
Topic: Event-driven order management system
```

## PREREQUISITE: Interactive Setup

**CRITICAL — BLOCKING PREREQUISITE:** If invoked without a topic description,
gather context before proceeding.

Adaptive questions (4-7 based on input fidelity):

| # | Header | Question | When to Ask |
|---|--------|----------|-------------|
| 1 | `Topic` | "What should be probed?" | If not provided |
| 2 | `Mode` | "Interactive or autonomous?" | If no `--mode` |
| 3 | `Depth` | "How deep?" | If no `--depth` |
| 4 | `Personas` | "How many personas? (3-8)" | If no `--personas` |
| 5 | `Saturation-Threshold` | "Stop when net-new drops below N?" | Always |
| 6 | `Scope` | "Files to scan for grounding?" | If no `--scope` |
| 7 | `Chain` | "Chain to another command?" | If no `--chain` |

Skip setup when Topic + mode + depth + scope all provided.

## Architecture

```
$research-loop probe
  ├── Phase 1:  Seed Capture       (parse topic / interactive setup)
  ├── Phase 2:  Persona Activation (pick N personas from 8 defaults)
  ├── Phase 3:  Codebase Grounding (scan --scope for prior art)
  ├── Phase 4:  Round Generation   (each persona drafts 1-2 questions)
  ├── Phase 5:  Question Synthesis (dedupe + batch ≤5 q/round)
  ├── Phase 6:  Answer Capture     (single batched prompt)
  ├── Phase 7:  Constraint Extraction (classify into 7 atom types)
  ├── Phase 8:  Cross-Check        (validate vs codebase + prior answers)
  ├── Phase 9:  Saturation Check   (net-new < threshold for K rounds)
  └── Phase 10: Synthesize & Handoff (probe-spec.md + research-loop config)
```

## Phase 1: Seed Capture

Parse the topic into structured atoms:
- Actors (who is involved?)
- Actions (what do they do?)
- Scope hints (what areas are relevant?)
- Constraints (what's already known?)

## Phase 2: Persona Activation

Select N personas from 8 defaults:

| # | Persona | Style |
|---|---------|-------|
| 1 | **Skeptic** | "Why is this needed? Prove it." |
| 2 | **Edge-Case Hunter** | "What about boundary conditions?" |
| 3 | **Scope Sentinel** | "This belongs to another concern." |
| 4 | **Ambiguity Detective** | "What does 'X' actually mean?" |
| 5 | **Contradiction Finder** | "These requirements conflict." |
| 6 | **Prior-Art Investigator** | "Hasn't this been tried before?" |
| 7 | **Success-Criteria Auditor** | "How do you measure success?" |
| 8 | **Constraint Excavator** | "What unspoken limits apply?" |

With `--adversarial`, rotate 3 most adversarial to the front (Skeptic,
Contradiction Finder, Edge-Case Hunter).

## Phase 3: Codebase Grounding

Scan the `--scope` glob to build a prior-art ledger:
- Existing features that relate to the topic
- Code patterns, conventions, and constraints
- Known limitations or tech stack boundaries
- Past decisions documented in code/comments

**This is mandatory — questions must be calibrated against real prior art.**

## Phase 4: Round Generation

Each persona drafts 1-2 candidate questions cold-start:

```
PERSONA: [name]
FOCUS: [persona's angle on the topic]
QUESTIONS:
  1. [specific question grounded in codebase context]
  2. [second question if persona has another angle]
```

## Phase 5: Question Synthesis

- Deduplicate: merge identical questions
- Drop already-answered questions from prior rounds
- Cap at ≤5 questions per round (prevents overload)
- Order by likely information gain

## Phase 6: Answer Capture

Single batched direct prompting call with 5 questions max:

```
INTERACTIVE mode: Ask user each question, capture answers
AUTONOMOUS mode: Self-answer using codebase inference
```

## Phase 7: Constraint Extraction

Classify each atomic constraint into 7 types:

| Type | Description | Example |
|------|-------------|---------|
| Requirement | Must-have | "Must support OAuth2" |
| Assumption | Implicit belief | "Users have stable internet" |
| Constraint | Limitation | "Budget < $10K/month" |
| Risk | Potential problem | "Third-party API may change" |
| Out-of-scope | Explicit exclusion | "Mobile app v2 only" |
| Ambiguity | Needs clarification | "What does 'real-time' mean?" |
| Contradiction | Conflicting needs | "Fast AND cheap processing" |

## Phase 8: Cross-Check

Validate new atoms against:
- Prior-art ledger (does codebase already address this?)
- Earlier round atoms (is this already captured?)
- Contradiction detection (does new info conflict with prior?)

## Phase 9: Saturation Check

```
FOR K=3 consecutive rounds:
  count = net_new_atoms_in_round
  IF count < threshold (default: 2):
    → SATURATED → stop probing
  ELSE:
    → continue to next round
```

**Stop conditions:**
- SATURATED (net-new < threshold for K rounds)
- BOUNDED (Iterations exhausted)
- USER_INTERRUPT (Ctrl+C, persists round atoms)
- SCOPE_LOCKED (all atoms classified out-of-scope for 2 rounds)

## Phase 10: Synthesize & Handoff

Emit output files:
- `probe-spec.md` — Full requirement specification from extracted atoms
- `constraints.tsv` — All extracted constraints with types and provenance
- `questions-asked.tsv` — Complete question-answer history per round
- `contradictions.md` — Detected contradictions and resolutions
- `hidden-assumptions.md` — Surfaced assumptions with risk assessment
- `autoresearch-config.yml` — Ready-to-use research-loop config
- `summary.md` — Executive summary: atoms found, saturation reached
- `handoff.json` — Machine-readable for `--chain` targets

## Flags

| Flag | Purpose |
|------|---------|
| `--depth <level>` | shallow (5), standard (15), deep (30) |
| `--personas N` | Active persona count (3-8, default 6) |
| `--saturation-threshold N` | Net-new atoms threshold (default 2) |
| `--scope <glob>` | Codebase glob for grounding |
| `--chain <targets>` | Comma-separated downstream commands |
| `--mode <mode>` | interactive or autonomous |
| `--adversarial` | Rotate adversarial personas to front |
| `--iterations N` | Hard cap on rounds |

## Composite Metric

```
probe_score = constraints_extracted*10 + contradictions_resolved*25
             + hidden_assumptions_surfaced*20 + ambiguities_clarified*15
             + (dimensions_covered/total)*30 + (saturated?100:0)
             + (config_complete?50:0)
```

Higher = more thorough requirement interrogation.

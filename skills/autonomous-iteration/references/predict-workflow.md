# Predict Workflow — $research-loop predict

Multi-persona swarm prediction that pre-analyzes code from multiple expert
perspectives. Simulates 3-5 personas that independently analyze, debate,
and reach consensus — producing ranked findings and hypotheses.

**Core idea:** Read code → Build knowledge files → Generate personas →
Independent analysis → Debate → Consensus → Report → Optional chain handoff.

## Trigger

- User invokes `$research-loop predict`
- User says "multi-perspective analysis", "swarm analysis"
- User wants pre-analysis before debugging, security, or shipping

## Loop Support

```
# Unlimited — keep refining until interrupted
$research-loop predict

# Bounded — exactly N persona debate rounds
$research-loop predict
Iterations: 3

# Focused scope with goal
$research-loop predict
Scope: src/api/**/*.ts
Goal: Security vulnerabilities and reliability gaps
```

## PREREQUISITE: Interactive Setup

**CRITICAL — BLOCKING PREREQUISITE:** If invoked without scope, goal, and
depth all provided, gather context before proceeding.

Adaptive questions (3-4 based on input):

| # | Header | Question | Options |
|---|--------|----------|---------|
| 1 | `Scope` | "Which files to analyze?" | Suggested globs + "Entire codebase" |
| 2 | `Goal` | "What should the swarm focus on?" | "Code quality", "Security", "Performance", "Architecture" |
| 3 | `Depth` | "How deep?" | "Shallow (3 people, 1 round)", "Standard (5, 2)", "Deep (8, 3)" |
| 4 | `Chain` | "Chain to another tool after?" | "Debug", "Security", "Fix", "Ship", "No chain" |

Skip setup when Scope + Goal + Depth all provided.

## Architecture

```
$research-loop predict
  ├── Phase 1: Setup — Interactive gate + config validation
  ├── Phase 2: Reconnaissance — Scan codebase, build knowledge files
  ├── Phase 3: Persona Generation — Create expert personas from context
  ├── Phase 4: Independent Analysis — Each persona analyzes independently
  ├── Phase 5: Debate — Structured cross-examination (1-3 rounds)
  ├── Phase 6: Consensus — Synthesizer aggregation + anti-herd check
  ├── Phase 7: Report — Generate findings, hypotheses, overview
  └── Phase 8: Handoff — Write handoff.json, optional chain
```

## Phase 1: Setup — Configuration

- Resolve `--scope` globs to actual file list
- Map `--depth` to persona/round count:
  - shallow → 3 personas, 1 round
  - standard → 5 personas, 2 rounds (default)
  - deep → 8 personas, 3 rounds
- Validate `--chain` target(s)
- If `--adversarial`, swap persona set

## Phase 2: Reconnaissance — Build Knowledge Files

Read in-scope source files and write structured knowledge files:

**codebase-analysis.md:** Files, imports, exports, entry points, data flows
**dependency-map.md:** Dependencies between modules, external packages
**component-clusters.md:** Logical groupings, patterns, conventions

## Phase 3: Persona Generation

Create expert personas from codebase context:

**Default persona set:**
- **Architect:** System design, patterns, tech debt, scalability
- **Security Analyst:** Vulnerabilities, auth, data exposure, injection
- **Performance Engineer:** Bottlenecks, N+1, memory, optimization
- **Reliability Engineer:** Error handling, retries, timeouts, resilience
- **Devil's Advocate:** What's wrong with this? Counter-arguments

**Adversarial set (with `--adversarial`):**
- Red Team, Blue Team, Insider Threat, Supply Chain Analyst, Judge

## Phase 4: Independent Analysis

Each persona analyzes independently using shared knowledge files:

```
Persona: [name]
Focus: [expertise area]
Findings:
  - [finding with file:line evidence and confidence]
Hypotheses:
  - [prediction about potential issue]
```

**Each finding must have:** file:line evidence, confidence score (Confirmed /
Likely / Possible), and impact description.

## Phase 5: Structured Debate

1-2 rounds of cross-examination:
- Each persona reviews others' findings
- Mandatory Devil's Advocate dissent
- Personas can upgrade/downgrade confidence based on evidence
- Track flip rate for anti-herd detection

## Phase 6: Consensus

Synthesizer aggregates findings:

- Merge duplicates, keep highest confidence
- Track minority opinions (anti-herd)
- Assign confidence: Confirmed > Likely > Possible
- Group by severity: Critical, High, Medium, Low, Info
- Check for groupthink (flip rate + entropy)

## Phase 7: Report

Write output files:
- `findings.md` — All findings with evidence and confidence
- `hypothesis-queue.md` — Untested hypotheses ranked by potential impact
- `overview.md` — Executive summary

## Phase 8: Handoff

Write `handoff.json` for `--chain` targets:
```json
{
  "findings": [...],
  "hypotheses": [...],
  "scope": "src/**/*.ts",
  "chain_to": "debug,fix"
}
```

## Flags

| Flag | Purpose |
|------|---------|
| `--chain <targets>` | Chain output to other commands |
| `--personas N` | Number of personas (3-8) |
| `--rounds N` | Debate rounds (1-3) |
| `--depth <level>` | Depth preset: shallow, standard, deep |
| `--adversarial` | Use adversarial persona set |
| `--budget <N>` | Max total findings (default: 40) |
| `--fail-on <severity>` | Non-zero exit for CI/CD gating |
| `--scope <glob>` | Limit analysis to specific files |

## Composite Metric

```
predict_score = findings_confirmed*15 + findings_probable*8
               + minority_preserved*3 + (personas/total)*20
               + (rounds/planned)*10 + anti_herd_passed*5
```

Higher = more thorough swarm analysis.

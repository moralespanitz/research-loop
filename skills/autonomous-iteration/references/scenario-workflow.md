# Scenario Workflow — $research-loop scenario

Scenario-driven use case generator that autonomously explores situations,
edge cases, failure modes, and derivative scenarios from a seed scenario.

**Core idea:** Seed scenario → Decompose into dimensions → Generate situations
→ Classify → Expand edge cases → Log → Repeat.

## Trigger

- User invokes `$research-loop scenario`
- User says "explore scenarios", "generate use cases", "what could go wrong"
- User says "stress test this feature", "edge cases for this"

## Loop Support

```
# Unlimited — keep generating until interrupted
$research-loop scenario

# Bounded — exactly N exploration iterations
$research-loop scenario
Iterations: 25

# Focused scope
$research-loop scenario
Scenario: User attempts checkout with multiple payment methods
Domain: software
```

## PREREQUISITE: Interactive Setup

**CRITICAL — BLOCKING PREREQUISITE:** If invoked without a scenario
description, gather context before proceeding.

Adaptive questions (4-8 based on input quality):

| # | Header | Question | When to Ask |
|---|--------|----------|-------------|
| 1 | `Scenario` | "Describe the scenario" | If not provided inline |
| 2 | `Domain` | "What domain?" | If not obvious |
| 3 | `Actors` | "Who are the key actors?" | If scenario doesn't mention |
| 4 | `Goal` | "What's your goal?" | If intent unclear |
| 5 | `Constraints` | "Any constraints?" | If no limits mentioned |
| 6 | `Depth` | "How deep?" | Always |
| 7 | `Output` | "What output format?" | If domain doesn't make it obvious |
| 8 | `Focus` | "Stress test first?" | If scenario is broad |

**Classification:**
- "checkout" → vague (1 word)
- "User resets password" → clear (actor + action)
- "Admin deploys with rollback" → clear + domain=software

## Architecture

```
$research-loop scenario
  ├── Phase 1: Seed — Capture, parse, and analyze the scenario
  ├── Phase 2: Decompose — Break into 12 exploration dimensions
  ├── Phase 3: Generate — Create ONE new situation
  ├── Phase 4: Classify — New? Duplicate? Out of scope?
  ├── Phase 5: Expand — Derive edge cases, what-ifs, failures
  ├── Phase 6: Log — Record to scenario-results.tsv
  └── Phase 7: Repeat — Next unexplored dimension
```

## Phase 1: Seed — Capture & Analyze Scenario

Parse scenario, identify:
- Actors (who participates?)
- Goals (what do they want?)
- Preconditions (what must be true?)
- Components (system parts involved)

**Output:** Structured seed in scenario seed document.

## Phase 2: Decompose — 12 Exploration Dimensions

| # | Dimension | Description |
|---|-----------|-------------|
| 1 | Happy Path | Everything works as expected |
| 2 | Error Path | Input validation, system errors |
| 3 | Edge Case | Boundary conditions, unusual inputs |
| 4 | Abuse | Intentional misuse, attack patterns |
| 5 | Scale | Load, concurrency, resource limits |
| 6 | Concurrent | Race conditions, locking, ordering |
| 7 | Temporal | Timeouts, delays, scheduling, expiry |
| 8 | Data Variation | Different data shapes, formats, sizes |
| 9 | Permission | Different roles, auth levels, tenants |
| 10 | Integration | External service failures, network issues |
| 11 | Recovery | Rollback, retry, compensation |
| 12 | State Transition | Invalid state transitions, partial updates |

## Phase 3: Generate — Create ONE Situation

Pick one dimension and generate a concrete situation:

```
DIMENSION: [name]
SITUATION: [concrete scenario]
TRIGGER: [what causes it]
FLOW: [step-by-step what happens]
EXPECTED OUTCOME: [what should happen]
```

## Phase 4: Classify — What Did We Find?

| Classification | Action |
|----------------|--------|
| New situation | Add to output, track dimension |
| Variant | Merge with existing, enrich |
| Duplicate | Skip, already recorded |
| Out of scope | Note boundary, move on |
| Low value | Skip (trivial / obvious) |

## Phase 5: Expand — Derive Edge Cases

From each kept situation, derive:
- What-ifs: "What if X is different?"
- Failure modes: "What if Y goes wrong?"
- Edge cases: "What about Z boundary?"
- Inverse: "What if the opposite happens?"

## Phase 6: Log — Record Results

**Append to scenario-results.tsv:**
```tsv
iteration	dimension	situation	classification	severity	edge_cases
1	error-path	invalid card number	new	HIGH	3
2	abuse	SQL injection in coupon code	new	CRITICAL	5
```

## Phase 7: Repeat

Pick next unexplored dimension or uncovered combinations.

**Every 10 iterations, print progress:**
```
=== Scenario Progress (iteration 20) ===
Dimensions covered: 9/12
Situations: 24 (18 new, 4 variants, 2 duplicates)
Edge cases: 67 derived
```

## Flags

| Flag | Purpose |
|------|---------|
| `--domain <type>` | software, product, business, security, marketing |
| `--depth <level>` | shallow (10), standard (25), deep (50+) |
| `--scope <glob>` | Limit to specific files/features |
| `--format <type>` | use-cases, user-stories, test-scenarios, threat-scenarios |
| `--focus <area>` | edge-cases, failures, security, scale |

## Composite Metric

```
scenario_score = scenarios_generated*10 + edge_cases_found*15
                + (dimensions_covered/12)*30 + unique_actors*5
```

Higher = more comprehensive exploration.

# Debug Workflow — $research-loop debug

Autonomous bug-hunting loop that applies the scientific method iteratively.
Doesn't stop at one bug — keeps investigating until the codebase is clean
or the user interrupts.

**Core idea:** Hypothesize → Test → Prove/Disprove → Log → Repeat.
Every finding needs code evidence. Every failed hypothesis teaches the next one.

## Trigger

- User invokes `$research-loop debug`
- User says "find all bugs", "debug this", "why is this failing"
- User reports a specific error and wants root cause analysis

## Loop Support

```
# Unlimited — keep hunting bugs until interrupted
$research-loop debug

# Bounded — exactly N investigation iterations
$research-loop debug
Iterations: 20

# Focused scope
$research-loop debug
Scope: src/api/**/*.ts
Symptom: API returns 500 on POST /users
```

## PREREQUISITE: Interactive Setup

**CRITICAL — BLOCKING PREREQUISITE:** If invoked without `--scope` or
`--symptom`, gather full context before proceeding.

**Single batched call — all 4 questions at once:**

| # | Header | Question | Options |
|---|--------|----------|---------|
| 1 | `Issue` | "What's the problem?" | "Hunt all bugs", "Specific error", "Failing tests", "CI/CD failure" |
| 2 | `Scope` | "Which files should I investigate?" | Suggested globs + "Entire codebase" |
| 3 | `Depth` | "How deep?" | "Quick scan (5)", "Standard (15)", "Deep (30+)", "Unlimited" |
| 4 | `After` | "When bugs are found?" | "Report only", "Find and fix", "Chain to another tool", "Ask me" |

If `--scope`, `--symptom`, `--fix`, or `--chain` flags are provided, skip setup.

## Architecture

```
$research-loop debug
  ├── Phase 1: Gather (symptoms + context)
  ├── Phase 2: Reconnaissance (scan codebase, map error surface)
  ├── Phase 3: Hypothesize (form falsifiable hypothesis)
  ├── Phase 4: Test (run experiment to prove/disprove)
  ├── Phase 5: Classify (bug found / disproven / inconclusive)
  ├── Phase 6: Log (record finding or elimination)
  └── Phase 7: Repeat (next hypothesis, next vector)
```

## Phase 1: Gather — Symptoms & Context

Collect everything known about the problem before investigating.

**If user provides symptoms:**
- Expected vs actual behavior
- Error messages, stack traces, log output
- When it started (commit, deploy, config change)
- Reproduction steps
- Environment (OS, runtime, versions)

**If no symptoms (autonomous bug hunting):**
- Run test suite, collect failures
- Run linter, collect errors
- Run type checker, collect issues
- Check build, collect warnings
- Scan for common anti-patterns

**Output:** `✓ Phase 1: Gathered — [N] symptoms, [M] error signals`

## Phase 2: Reconnaissance — Map the Error Surface

**Actions:**
1. Read files mentioned in stack traces / error messages
2. Trace call chains from error origin backward
3. Identify entry points (API routes, event handlers, CLI commands)
4. Map data flow through affected components
5. Check recent git changes (`git log --oneline -20 -- <path>`)
6. Identify external dependencies and integration points

**Error surface map:**
```
Entry Point → Data Flow → Failure Point → Side Effects
  POST /users → validate() → db.insert() → ← FAILS HERE
                                           → notification.send() ← cascading
```

**Output:** `✓ Phase 2: Recon — [N] files, [M] potential failure points`

## Phase 3: Hypothesize — Form Falsifiable Hypothesis

**A good hypothesis is:**
- Specific: "The JWT validation skips algorithm check on line 42"
- Testable: Can be proven/disproven with a concrete experiment
- Falsifiable: There exists evidence that would prove it wrong
- Prioritized: Most likely cause first

**Hypothesis formation strategy:**

| Priority | Strategy | When to Use |
|----------|----------|-------------|
| 1 | Error message literal | Stack trace points to exact line |
| 2 | Recent change | Bug started after specific commit |
| 3 | Data flow trace | Input → Transform → Output chain |
| 4 | Environment diff | Works locally, fails in CI/prod |
| 5 | Dependency issue | After upgrade/install |
| 6 | Race condition | Intermittent, timing-dependent |
| 7 | Edge case | Works for most inputs, fails for specific ones |

**Output:** `Hypothesis [N]: "[specific claim]" — testing...`

## Phase 4: Test — Run Experiment

**Experiment types:**

| Type | Method | Best For |
|------|--------|----------|
| Direct inspection | Read the code at suspected location | Logic errors, missing checks |
| Trace execution | Add logging, run, read output | Data flow issues |
| Minimal reproduction | Create smallest failing case | Complex interactions |
| Binary search | Comment out half the code, narrow | "Something in this file breaks" |
| Differential | Compare working vs broken | Regressions |
| Git bisect | Find exact commit that introduced bug | "It used to work" |
| Input variation | Change inputs systematically | Edge cases, boundary issues |

**Rules:**
- ONE experiment per iteration
- Record the exact command/action and its output
- If experiment is destructive, git stash first
- Timeout: >30 seconds → too complex, simplify

## Phase 5: Classify — What Did We Learn?

| Result | Action |
|--------|--------|
| Bug confirmed | Record with full evidence, severity, location |
| Hypothesis disproven | Log as eliminated, extract learnings |
| Inconclusive | Refine hypothesis, re-test |
| New lead discovered | Log discovery, add to hypothesis queue |

**Bug finding format:**
```
### [SEVERITY] Bug: [title]
- Location: `file:line`
- Hypothesis: [what we suspected]
- Evidence: [code snippet + experiment result]
- Reproduction: [exact steps to trigger]
- Impact: [what breaks, who's affected]
- Root cause: [WHY it happens]
- Suggested fix: [concrete code change]
```

**Severity classification:**
| Level | Criteria |
|-------|----------|
| CRITICAL | Data loss, security breach, system crash |
| HIGH | Feature broken, incorrect results, degradation >10x |
| MEDIUM | Edge case failure, degraded UX, workaround exists |
| LOW | Cosmetic, minor inconsistency, theoretical risk |

## Phase 6: Log — Record Everything

**Append to debug-results.tsv:**
```tsv
iteration	type	hypothesis	result	severity	location	description
1	hypothesis	JWT skips alg check	confirmed	CRITICAL	auth.ts:42	Algorithm confusion vulnerability
2	hypothesis	Rate limit missing	disproven	-	-	Rate limiter exists in middleware
```

**Every 5 iterations, print progress:**
```
=== Debug Progress (iteration 10) ===
Bugs found: 3 (1 Critical, 1 High, 1 Medium)
Hypotheses tested: 8 (3 confirmed, 4 disproven, 1 inconclusive)
Files investigated: 14 / 47 in scope
```

## Phase 7: Repeat — Next Investigation

1. Follow new leads discovered during previous experiments
2. Untested high-priority hypotheses
3. Uninvestigated files in the error surface
4. Deeper investigation of confirmed bugs
5. Pattern-based search

## Flags

| Flag | Purpose |
|------|---------|
| `--fix` | After finding bugs, switch to fix mode |
| `--scope <glob>` | Limit investigation to specific files |
| `--symptom "<text>"` | Pre-fill symptom |
| `--severity <level>` | Only report findings above threshold |
| `--chain <targets>` | Chain to downstream tool(s) |

## Composite Metric

```
debug_score = bugs_found * 15
            + hypotheses_tested * 3
            + (files_investigated / files_in_scope) * 40
            + (techniques_used / 7) * 10
```

Higher = more thorough. Incentivizes breadth AND depth.

## Investigation Techniques

- **Binary Search:** Comment out half the suspicious code, narrow down
- **Differential Debugging:** Compare working vs broken state
- **Minimal Reproduction:** Smallest possible case that reproduces the bug
- **Trace Execution:** Strategic logging at key data flow points
- **Pattern Search:** Found one bug? Search for the same pattern elsewhere
- **Working Backwards:** Trace from error backward to divergence point
- **The 5 Whys:** Ask "why" recursively until root cause is found

## What NOT to Do

| Anti-Pattern | Why It Fails |
|---|---|
| Fix before understanding | You fix symptoms, not causes |
| Change multiple things at once | Can't attribute improvement |
| Ignore disproven hypotheses | Repeat failed investigations |
| Assume instead of verify | Confirmation bias |
| Skip reproduction | Can't verify the fix |
| Debug in production | Never investigate with live data |
| Tunnel vision on one file | Bugs span boundaries |
| Trust error messages literally | Root cause is 2-3 layers deeper |
| Give up after 3 tries | Some bugs need 10+ hypotheses |

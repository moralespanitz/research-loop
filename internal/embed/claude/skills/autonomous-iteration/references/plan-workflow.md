# Plan Workflow — $research-loop plan

Convert a textual goal into a validated, ready-to-execute research-loop
configuration.

**Output:** Complete `$research-loop` invocation with Scope, Metric,
Direction, and Verify — all validated before launch.

## Trigger

- User invokes `$research-loop plan`
- User says "help me set up research-loop", "plan a research-loop run"

## Workflow

### Phase 1: Capture Goal

**CRITICAL — BLOCKING PREREQUISITE:** If no goal is provided inline, capture
it. Never proceed without a goal.

```
Question: "What do you want to improve? Describe your goal in plain language."
Options:
  - "Code quality" (tests, coverage, type safety, linting, bundle size)
  - "Performance" (response time, build speed, Lighthouse score)
  - "Content" (SEO score, readability, word count)
  - "Refactoring" (reduce LOC, eliminate patterns, simplify architecture)
```

If user provides goal text directly, skip to Phase 2.

### Phase 2: Analyze Context

1. Read codebase structure (package.json, project files, test config)
2. Identify domain: backend, frontend, ML, content, DevOps
3. Detect existing tooling: test runner, linter, bundler, benchmark scripts
4. Infer likely metric candidates from goal + tooling

### Phase 3: Define Scope

Present scope options based on codebase analysis:

```
Question: "Which files should research-loop be allowed to modify?"
Options:
  - "{inferred scope 1}" ({file count} files)
  - "{inferred scope 2}" ({file count} files)
  - "Entire project" (All source files — use with caution)
```

**Scope validation rules:**
- Scope must resolve to at least 1 file
- Warn if scope exceeds 50 files
- Warn if scope includes test files AND source files

### Phase 4: Define Metric

The metric must be **mechanical** — extractable from a command output.

```
Question: "What number tells you if things got better?"
Options:
  - "{metric 1} (Recommended)" — {extraction command}
  - "{metric 2}" — {extraction command}
  - "{metric 3}" — {extraction command}
```

**Metric validation rules:**

| Check | Pass | Fail |
|-------|------|------|
| Outputs a number | 87.3, 0.95, 42 | PASS, looks good, ✓ |
| Extractable by command | grep, awk, jq | Requires human judgment |
| Deterministic | Same input → same output | Random, flaky |
| Fast | < 30 seconds | > 2 minutes |

If metric fails validation, explain why and suggest alternatives.

### Phase 4.5: Define Guard (Optional)

```
Question: "Do you want a guard command? This prevents regressions."
Options: "npm test", "tsc --noEmit", "npm run build", "Skip — no guard"
```

### Phase 5: Define Direction

```
Question: "Higher or lower is better for 'metric'?"
Options: "Higher is better", "Lower is better"
```

### Phase 6: Define Verify & Dry-Run

Construct the shell command and **dry-run it**:

```
QUESTION: "I'll use this command to extract the metric — shall I dry-run it?"
SUGGESTED COMMAND: {constructed from metric + tooling}

DRY-RUN RESULT:
  ✓ Command ran successfully
  ✓ Output parsed as number: {value}
  → OR →
  ✗ Command failed: {error}
  → Offer to fix the command or choose a different metric
```

**Critical gates:**
- Metric MUST be mechanical
- Verify command MUST pass a dry run BEFORE accepting
- Scope MUST resolve to ≥1 file
- Guard command (if set) MUST pass a dry run

### Phase 7: Confirm & Launch

Present the complete config:

```
=== Research-Loop Configuration ===
Goal:        {goal}
Scope:       {scope} ({N} files)
Metric:      {metric}
Direction:   {direction}
Verify:      {command}
Guard:       {guard or "none"}

Options:
  [1] Launch — start autonomous loop (unlimited)
  [2] Launch with iteration limit — set N
  [3] Edit config — go back
  [4] Cancel — save config for later
```

After selection, either launch or save.

### Phase 8: Save Config

If user selects "Cancel" or "Edit," save the config for later use:
- Write `autoresearch-config.yml` with all fields
- User can re-run `$research-loop plan` to resume

## Interactive Setup Cheat Sheet

When invoked without inline goal:

```
BATCHED CALL (2 prompts, not one-at-a-time):

Prompt 1: "Goal" + options
Prompt 2: "Scope" + "Metric" + "Direction" + "Guard" + "Launch" (all in one)
```

Verify command dry-run happens between prompts 1 and 2.

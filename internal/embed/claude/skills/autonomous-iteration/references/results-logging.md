# Results Logging Protocol

Track every iteration in a structured log. Enables pattern recognition and
prevents repeating failed experiments.

## Setup & Initialization

Create the log automatically at Phase 0 (baseline):

```bash
# 1. Create log file with metric direction and header
echo "# metric_direction: higher_is_better" > research-loop-results.tsv
echo -e "iteration\tcommit\tmetric\tdelta\tguard\tguard-metric\tstatus\tdescription" >> research-loop-results.tsv

# 2. Add to .gitignore (log is local, not committed)
echo "research-loop-results.tsv" >> .gitignore

# 3. Run verify command to establish baseline metric
BASELINE=$(npx jest --coverage 2>&1 | grep 'All files' | awk '{print $4}')

# 4. Record baseline as iteration 0
COMMIT=$(git rev-parse --short HEAD)
echo -e "0\t${COMMIT}\t${BASELINE}\t0.0\tpass\t-\tbaseline\tinitial state" >> research-loop-results.tsv
```

## Logging Function

Called at Phase 7 of every iteration after the keep/discard/crash decision:

```bash
log_iteration() {
  local iteration=$1 commit=$2 metric=$3 delta=$4 guard=$5 guard_metric=$6 status=$7 description=$8
  echo -e "${iteration}\t${commit}\t${metric}\t${delta}\t${guard}\t${guard_metric}\t${status}\t${description}" \
    >> research-loop-results.tsv
}

# Usage:
log_iteration "1" "b2c3d4e" "87.1" "+1.9" "pass" "-" "keep" "add tests for auth middleware"
log_iteration "2" "-" "86.5" "-0.6" "-" "-" "discard" "refactor test helpers (broke 2 tests)"
log_iteration "3" "-" "0.0" "0.0" "-" "-" "crash" "add integration tests (DB connection failed)"
log_iteration "4" "-" "-" "-" "-" "-" "no-op" "attempted to modify read-only config"
log_iteration "5" "-" "-" "-" "-" "-" "metric-error" "verify output was 'PASS' — not a number"
```

## Reading & Using the Log

```bash
# Phase 1 (Review): Read recent entries for pattern recognition
tail -20 research-loop-results.tsv

# Count outcomes for progress tracking
KEEPS=$(grep -c 'keep' research-loop-results.tsv || echo 0)
DISCARDS=$(grep -c 'discard' research-loop-results.tsv || echo 0)
CRASHES=$(grep -c 'crash' research-loop-results.tsv || echo 0)

# Detect stuck state: >5 consecutive discards triggers recovery
LAST_5=$(tail -5 research-loop-results.tsv | awk -F'\t' '{print $6}')
# If all 5 are "discard" → trigger "When Stuck" protocol

# Pattern recognition: which file changes succeed?
grep 'keep' research-loop-results.tsv | awk -F'\t' '{print $7}'
```

## Log Format (TSV)

```
iteration	commit	metric	delta	guard	guard-metric	status	description
```

### Columns

| Column | Type | Description |
|--------|------|-------------|
| iteration | int | Sequential counter starting at 0 (baseline) |
| commit | string | Short git hash, "-" if reverted |
| metric | float | Measured value from verification |
| delta | float | Change from previous best |
| guard | enum | pass, fail, or "-" |
| guard-metric | float or "-" | Measured guard-metric value |
| status | enum | baseline, keep, keep (reworked), discard, crash, no-op, hook-blocked, metric-error |
| description | string | One-sentence description of what was tried |

### Example

```tsv
iteration	commit	metric	delta	guard	guard-metric	status	description
0	a1b2c3d	85.2	0.0	pass	-	baseline	initial state — test coverage 85.2%
1	b2c3d4e	87.1	+1.9	pass	-	keep	add tests for auth middleware edge cases
2	-	86.5	-0.6	-	-	discard	refactor test helpers (broke 2 tests)
3	-	0.0	0.0	-	-	crash	add integration tests (DB connection failed)
4	-	88.9	+1.8	fail	-	discard	inline hot-path functions (guard: 3 tests broke)
5	c3d4e5f	88.3	+1.2	pass	-	keep	add tests for error handling
```

## Log Management

- Create at setup (iteration 0 = baseline)
- Append after EVERY iteration (including crashes)
- Do NOT commit this file to git (add to .gitignore)
- Read last 10-20 entries at start of each iteration for context

## Summary Reporting

Every 10 iterations (or at loop completion in bounded mode), print:

```
=== Research-Loop Progress (iteration 20) ===
Baseline: 85.2% → Current best: 92.1% (+6.9%)
Keeps: 8 | Discards: 10 | Crashes: 2
Last 5: keep, discard, discard, keep, keep
```

## Metric Direction

Record direction in first line of results log:

```
# metric_direction: higher_is_better
```

- **Lower is better:** val_bpb, response time (ms), bundle size (KB), error count
- **Higher is better:** test coverage (%), lighthouse score, throughput (req/s)

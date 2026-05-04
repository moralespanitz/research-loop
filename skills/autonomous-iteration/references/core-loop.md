# Core Loop Protocol — Modify → Verify → Keep/Discard → Repeat

Detailed protocol for the autonomous iteration loop. SKILL.md has the
summary; this file has the full rules.

## Loop Modes

- **Unbounded (default):** Loop forever until manually interrupted (Ctrl+C)
- **Bounded:** Loop exactly N times when `Iterations: N` is set

## Phase 0: Precondition Checks (before loop starts)

**Must complete ALL checks before entering the loop. Fail fast if any fails.**

```bash
# 1. Verify git repo exists
git rev-parse --git-dir 2>/dev/null || echo "FAIL: not a git repo"

# 2. Check for dirty working tree
git status --porcelain

# 3. Check for stale lock files
ls .git/index.lock 2>/dev/null && echo "WARN: stale lock"

# 4. Check for detached HEAD
git symbolic-ref HEAD 2>/dev/null || echo "WARN: detached HEAD"

# 5. Check for git hooks
ls .git/hooks/pre-commit .git/hooks/commit-msg 2>/dev/null && echo "INFO: git hook detected"
```

If metric-valued guard is configured:

```bash
GUARD_BASELINE=$(<guard command>)
# Validate it's a valid number
# Record alongside the primary metric baseline in iteration 0
```

If any FAIL: Stop and inform user. Do not enter the loop.
If any WARN: Log the warning, proceed with caution.

## Phase 1: Review (situational awareness)

**Complete ALL 6 steps before each iteration:**

1. Read current state of in-scope files (full context)
2. Read last 10-20 entries from results log
3. Run: `git log --oneline -20` to see recent changes
4. Run: `git diff HEAD~1` (if last iteration was "keep")
5. Identify: what worked, what failed, what's untried
6. If bounded: check current_iteration vs max_iterations

Git IS the memory. After rollbacks, state may differ from expectations.

## Phase 2: Ideate (pick the next change)

**Priority order:**

1. Fix crashes/failures from previous iteration first
2. Exploit successes — run `git diff` on last kept commit, try variants
3. Explore new approaches — cross-reference results log AND git history
4. Combine near-misses — two changes that individually didn't help might work together
5. Simplify — remove code while maintaining metric. Simpler = better
6. Radical experiments — when incremental changes stall, try dramatically different

**Anti-patterns:**
- Don't repeat exact same change already discarded — CHECK git log first
- Don't make multiple unrelated changes at once
- Don't chase marginal gains with ugly complexity
- Don't ignore git history — it's the primary learning mechanism

## Phase 3: Modify (one atomic change)

- Make ONE focused change to in-scope files
- Describe in ONE sentence before making the change
- If description needs "and", split into separate iterations

### Multi-File Atomic Changes

One logical change may span multiple files if it serves a single purpose.

| One Change (OK) | Two Changes (Split) |
|-----------------|---------------------|
| Change port 3000→8080 in Dockerfile + compose + nginx | Change port AND add new service |
| Update Node 18→20 in Dockerfile + CI + package.json | Update Node AND switch to pnpm |

### Enforcing Atomicity

```bash
FILES_CHANGED=$(git diff --name-only | wc -l)
if [ "$FILES_CHANGED" -gt 5 ]; then
  echo "WARN: ${FILES_CHANGED} files changed — verify single intent"
fi
```

## Phase 4: Commit (before verification)

**Commit before running verification.** This enables clean rollback.

```bash
# Stage ONLY in-scope files
git add <file1> <file2> ...
# NEVER use git add -A — it stages ALL files

# Check if there's actually something to commit
git diff --cached --quiet
# → If exit code 0: skip commit, log as "no-op", go to next iteration

# Commit with descriptive message
git commit -m "experiment(<scope>): <one-sentence description>"
```

**Hook failure handling:** If pre-commit hook blocks:
1. Read the hook's error output
2. If fixable: fix, re-stage, retry — do NOT use `--no-verify`
3. If not fixable within 2 attempts: log as `hook-blocked`, revert

**Rollback strategy:**
```bash
# Preferred: git revert (safe, preserves history)
git revert HEAD --no-edit

# Alternative: git reset (if revert conflicts)
git revert --abort && git reset --hard HEAD~1
```

Prefer `git revert` over `git reset --hard` — revert preserves the experiment
in history so you can learn from it.

## Phase 5: Verify (mechanical only)

Run the agreed-upon verification command. Capture output.

**Timeout rule:** If verification exceeds 2x normal time, kill and treat as crash.

**Metric validation (MANDATORY after extraction):**

The extracted value MUST be a valid number before ANY decision logic runs.

```
extracted_value = <result of verify pipeline>
extracted_value = strip(extracted_value)

IF extracted_value does NOT match pattern: ^-?[0-9]+\.?[0-9]*$
    STATUS = "metric-error"
    LOG iteration as: status=metric-error
    safe_revert()
    PRINT "⚠ Metric extraction failed — got '{extracted_value}' instead of a number"
    PRINT "Raw verify output (last 5 lines):"
    PRINT <tail -5 of verify command output>

    IF previous_iteration.status == "metric-error":
        PRINT "✗ Two consecutive metric extraction failures — stopping."
        STOP
    CONTINUE to next iteration
```

### Verification Command Templates

| Language | Verify Command | Metric | Direction |
|----------|---------------|--------|-----------|
| Node.js | `npx jest --coverage 2>&1 \| grep 'All files' \| awk '{print $4}'` | Coverage % | higher |
| Python | `pytest --cov=src --cov-report=term 2>&1 \| grep TOTAL \| awk '{print $4}'` | Coverage % | higher |
| Rust | `cargo test 2>&1 \| grep -oP '\d+ passed' \| grep -oP '\d+'` | Tests passed | higher |
| Go | `go test -count=1 ./... 2>&1 \| grep -c '^ok'` | Packages passing | higher |
| Bundle | `npx esbuild src/index.ts --bundle --minify \| wc -c` | Bytes | lower |
| Latency | `wrk -t2 -c10 -d10s http://localhost:3000 \| grep 'Avg Lat' \| awk '{print $2}'` | ms | lower |

## Phase 6: Guard (regression prevention)

If a guard command is specified, run it after verification:

```bash
guard_result = run(guard_command)  # e.g., "npm test"
```

Guard prevents regressions. Fixing the metric shouldn't break other things.

## Phase 7: Decide — Keep, Revert, or Rework

| Condition | Action |
|-----------|--------|
| IMPROVED + guard passed | **KEEP** — commit stays, log "keep" |
| IMPROVED + guard failed | **REWORK** — revert, try different approach (max 2 attempts) |
| SAME/WORSE | **DISCARD** — revert immediately, log "discard" |
| Crash during fix | **RECOVER** — revert, try simpler approach (max 3 attempts) |

**Rework strategy (when guard fails):**
1. Read the guard failure — understand what regressed
2. Revert: `git revert HEAD --no-edit`
3. Understand why the fix broke something else
4. Find approach that fixes target WITHOUT breaking guard
5. If 2 rework attempts fail → skip, add to blocked.md, move on

## Phase 8: Log & Repeat

Append to results log (see `results-logging.md`), then go to Phase 1.

## Git as Memory — Detailed Protocol

At the start of EVERY iteration, the agent runs:

```bash
# Step 1: Read recent experiment history
git log --oneline -20

# Step 2: Inspect the last successful change
git diff HEAD~1

# Step 3: Check what was tried (avoid repeating failures)
git log --oneline -20 | grep "experiment"

# Step 4: Deep-dive a specific success
git show <hash> --stat
```

### Example: Memory in Action

```
# Agent reads git log and sees:
# a1b2c3d experiment(api): add response caching — KEPT (metric improved)
# d4e5f6g Revert "experiment(api): increase cache TTL to 60s" — REVERTED
# c3d4e5f experiment(api): add cache invalidation on write — KEPT
#
# Agent learns:
# ✓ Caching works (2 kept commits)
# ✗ Increasing TTL didn't help (reverted)
# → Next: try a different cache strategy, NOT longer TTL
```

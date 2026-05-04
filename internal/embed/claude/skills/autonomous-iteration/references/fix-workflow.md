# Fix Workflow — $research-loop fix

Autonomous fix loop that takes a broken state and iteratively repairs it
until everything passes. One fix per iteration. Atomic, committed, verified,
auto-reverted on failure.

**Core idea:** Detect → Prioritize → Fix ONE thing → Verify → Keep/Revert
→ Repeat until zero errors.

## Trigger

- User invokes `$research-loop fix`
- User says "fix all errors", "make tests pass", "fix the build"
- User has output from `$research-loop debug` and wants to fix findings

## Loop Support

```
# Unlimited — keep fixing until everything passes
$research-loop fix

# Bounded — exactly N fix iterations
$research-loop fix
Iterations: 30

# With explicit target
$research-loop fix
Target: make all tests pass
Scope: src/**/*.ts
Guard: npm run typecheck
```

## PREREQUISITE: Interactive Setup

**CRITICAL — BLOCKING PREREQUISITE:** If invoked without `--target`,
`--guard`, or `--scope`, auto-detect all failures first, then gather input.

**Single batched call — all 4 questions at once:**

| # | Header | Question | Options |
|---|--------|----------|---------|
| 1 | `Fix What` | "Found [N] failures. What should I fix?" | "Fix everything", "Only tests", "Only type errors", "Only lint" |
| 2 | `Guard` | "What must ALWAYS pass?" | "npm test", "tsc --noEmit", "npm run build", "Skip" |
| 3 | `Scope` | "Which files can I modify?" | Suggested globs from error locations |
| 4 | `Launch` | "Ready to fix?" | "Fix until zero", "Fix with limit", "Edit config", "Cancel" |

If `--target`, `--guard`, `--scope`, or `--from-debug` flags provided, skip.

## Architecture

```
$research-loop fix
  ├── Phase 1: Detect (what's broken?)
  ├── Phase 2: Prioritize (fix order)
  ├── Phase 3: Fix ONE thing (atomic change)
  ├── Phase 4: Commit (before verification)
  ├── Phase 5: Verify (did error count decrease?)
  ├── Phase 6: Guard (did anything else break?)
  ├── Phase 7: Decide (keep / revert / rework)
  └── Phase 8: Log & Repeat
```

## Phase 1: Detect — What's Broken?

Auto-detect the failure domain:

```
FUNCTION detectFailures(context):
  failures = []

  IF test runner detected:    result = run_tests()
  IF typescript detected:     result = run("tsc --noEmit")
  IF linter detected:         result = run_lint()
  IF build script detected:   result = run_build()
  IF debug/{latest}/findings.md exists: bugs = parse_findings()

  RETURN failures sorted by severity
```

**Output:** `✓ Phase 1: Detected — [N] test failures, [M] type errors`

## Phase 2: Prioritize — Fix Order

| Priority | Category | Why First |
|----------|----------|-----------|
| 1 | Build failures | Nothing works without compile |
| 2 | Critical/High bugs | Data loss, security |
| 3 | Type errors | Type safety prevents cascading bugs |
| 4 | Test failures | Tests verify correctness |
| 5 | Medium/Low bugs | From debug findings |
| 6 | Lint errors | Code quality |
| 7 | Warnings | Polish |

Within a category: cascading impact first, then simplicity, then file locality.

## Phase 3: Fix ONE Thing — Atomic Change

**Fix strategies by category:**

| Category | Strategy |
|----------|----------|
| Build failure | Read error, fix the exact line/import/config |
| Type error | Add proper types, fix signatures, handle null cases |
| Test failure | Read test + implementation, find mismatch, fix implementation |
| Lint error | Apply the rule — auto-fix where possible |
| Bug (from debug) | Apply the suggested fix from findings.md |

**Language-specific rules:**

| Language | Never Do | Correct Pattern |
|----------|----------|-----------------|
| TypeScript | `any`, `@ts-ignore` | Proper interfaces, generics, discriminated unions |
| Python | Bare `except:` | `except SpecificError:` |
| Go | Ignoring errors with `_` | Explicit error wrapping |
| Rust | `.unwrap()` in production | `Result<T, E>` propagation with `?` |

**Rules:**
- ONE fix per iteration
- Fix the IMPLEMENTATION, not the test
- Never add `@ts-ignore`, `eslint-disable` to suppress errors
- Prefer minimal changes — smallest diff that fixes the issue

## Phase 4: Commit — Before Verification

```bash
git add <modified-files>
git commit -m "fix: [what was fixed] — [file:line]"
```

## Phase 5: Verify — Did It Help?

```
previous_errors = error_count_before
current_errors = error_count_after
delta = previous_errors - current_errors
```

Expected: `delta > 0` (fewer errors than before)

## Phase 6: Guard — Did Anything Else Break?

Run guard command. Guard prevents regressions.

## Phase 7: Decide — Keep, Revert, or Rework

| Condition | Action |
|-----------|--------|
| delta > 0 AND guard passes | **KEEP** — commit stays |
| delta > 0 AND guard fails | **REWORK** — revert, try again (max 2) |
| delta == 0 | **DISCARD** — revert, no effect |
| delta < 0 | **DISCARD** — revert immediately |
| Crash | **RECOVER** — revert, try simpler (max 3) |

## Phase 8: Log & Repeat

**Append to fix-results.tsv:**
```tsv
iteration	category	target	delta	guard	status	description
0	-	-	-	pass	baseline	47 test failures, 12 type errors
1	type	auth.ts:42	-2	pass	fixed	add return type annotation
```

**Every 5 iterations, print progress:**
```
=== Fix Progress (iteration 15) ===
Baseline: 62 errors → Current: 23 errors (-39, -63%)
Keeps: 11 | Discards: 3 | Reworks: 1
```

**Completion detection:** When current_errors == 0, print "All Clear" and stop.

## Flags

| Flag | Purpose |
|------|---------|
| `--target <command>` | Explicit verify command |
| `--guard <command>` | Safety command |
| `--scope <glob>` | Limit fixes to specific files |
| `--category <type>` | Only fix: test, type, lint, build, bug |
| `--from-debug` | Read findings from latest debug/ session |
| `--chain <targets>` | Chain to downstream tool(s) |

## Composite Metric

```
fix_score = reduction_score + quality_score + guard_score + bonus_score

reduction_score = ((baseline - current) / baseline) * 60
guard_score = (guard_always_passed ? 25 : 0)
bonus_score = (zero_errors ? 10 : 0) + (no_discards ? 5 : 0)
```

- **100+** = perfect: all fixed, no regressions
- **80-99** = good: significant progress, guards held
- **60-79** = acceptable: meaningful reduction
- **<60** = needs work

## What NOT to Do

| Anti-Pattern | Why It's Wrong |
|---|---|
| Add `@ts-ignore` / `eslint-disable` | Hides the problem |
| Use `any` type | Defeats type safety |
| Delete or skip failing tests | Removes the safety net |
| `catch (e) {}` empty catch | Swallows errors |
| Comment out broken code | Never uncommented |
| Hardcode values to pass tests | Feature broken for real data |

# Learn Workflow — $research-loop learn

Autonomous codebase documentation engine. Scouts codebase structure, learns
patterns and architecture, generates/updates comprehensive documentation —
then validates and iteratively improves until docs match codebase reality.

**Core idea:** Scout → Generate → Validate → Fix → Repeat until docs are
accurate.

## Trigger

- User invokes `$research-loop learn`
- User says "learn this codebase", "generate docs", "document this project"
- User says "update docs", "check docs health"

## Loop Support

```
# Default — auto-detect mode and learn
$research-loop learn

# Specific mode
$research-loop learn --mode update

# Bounded iterations for validation-fix loop
$research-loop learn
Iterations: 5

# Scoped learning
$research-loop learn --scope src/api/**
```

## PREREQUISITE: Interactive Setup

**CRITICAL — BLOCKING PREREQUISITE:** If invoked without `--mode` or
sufficient inline context, gather config before proceeding.

Pre-scan project state for smart defaults:
1. Does `docs/` exist? How many .md files?
2. Project type indicators (package.json, Cargo.toml, etc.)
3. Staleness: last commit to docs/ vs main branch
4. Scale: total file count

**Single batched call — 4 questions:**

| # | Header | Question | Options |
|---|--------|----------|---------|
| 1 | `Mode` | "What operation?" | "Init (generate all)", "Update (refresh)", "Check (read-only)", "Summarize (quick)" |
| 2 | `Scope` | "Which parts?" | Detected top-level dirs + "Everything" |
| 3 | `Depth` | "How comprehensive?" | "Quick", "Standard", "Deep" |
| 4 | `Launch` | "Ready?" | "Launch", "Edit config", "Cancel" |

## 4 Modes

| Mode | Purpose | Loop? |
|------|---------|-------|
| `init` | Learn codebase from scratch, generate all docs | Yes — validate-fix cycle |
| `update` | Learn what changed, refresh existing docs | Yes — validate-fix cycle |
| `check` | Read-only health/staleness assessment | No — diagnostic only |
| `summarize` | Quick codebase summary with file inventory | Minimal — size check only |

## Architecture

```
$research-loop learn
  ├── Phase 1: Scout — Parallel codebase reconnaissance
  ├── Phase 2: Analyze — Structure detection + project type classification
  ├── Phase 3: Map — Dynamic doc discovery + gap analysis
  ├── Phase 4: Generate — Spawn docs-manager with structured prompt
  ├── Phase 5: Validate — Mechanical verification (refs, links, completeness)
  ├── Phase 6: Fix — Re-generate failed docs with feedback (LOOP)
  ├── Phase 7: Finalize — Size check, inventory, git diff summary
  └── Phase 8: Log — Record results to learn-results.tsv
```

## Phase 1: Scout — Parallel Reconnaissance

- Scan codebase: files/LOC per directory
- Exclusion list: .git, node_modules, __pycache__, dist, build, vendor
- Test directories are NOT excluded — metadata matters for testing-guide.md
- Scale awareness: >5000 files → increase parallelism, >10000 → warn user
- Monorepo detection: check workspaces in package.json, lerna.json, etc.

## Phase 2: Analyze — Structure Detection

- Classify project type (backend, frontend, library, CLI, ML, etc.)
- Detect tech stack (frameworks, languages, tools)
- Measure staleness: last doc update vs last code change
- Identify documentation gaps

## Phase 3: Map — Doc Discovery

- Dynamic discovery: `docs/*.md`
- Gap analysis: what docs exist vs. what the project needs
- Conditional selection: deployment-guide.md only if deploy config exists

## Phase 4: Generate — Documentation Creation

Generate structured docs using full scout context:

**Core docs (standard depth):**
- `README.md` — Project overview, setup, usage
- `system-architecture.md` — Architecture overview, component diagram
- `getting-started.md` — Development setup guide
- `api-reference.md` — API endpoints, data models (if API exists)
- `testing-guide.md` — Test patterns, fixtures, running tests

**Optional docs (deep depth):**
- `deployment-guide.md` — CI/CD, Docker, infrastructure
- `design-decisions.md` — Key architectural decisions and rationale
- `contributing.md` — Contribution workflow

## Phase 5: Validate — Mechanical Verification

Check each generated doc:
- Code references resolve to actual files
- Internal links work (no dead anchors)
- Sections are complete (not truncated)
- Size compliance: no doc exceeds 100KB
- No placeholder text ("TODO", "FIXME")

## Phase 6: Fix — Validation-Fix Loop

For each failed validation:
1. Re-generate with validation feedback
2. Re-validate
3. Max 3 retries per doc
4. Escalate to user if unresolved after 3 attempts

## Phase 7: Finalize

- Inventory check: all expected docs exist
- Git diff summary: show what changed
- Size compliance: check all docs
- Print summary of generated/updated docs

## Phase 8: Log — Record Results

**Append to learn-results.tsv:**
```tsv
iteration	doc	name	status	validation_pct	size_bytes
1	README.md	generated	100%	2450
2	system-architecture.md	generated	85%	8900
```

## Flags

| Flag | Purpose |
|------|---------|
| `--mode <mode>` | init, update, check, summarize |
| `--scope <glob>` | Limit learning to specific dirs |
| `--depth <level>` | quick, standard, deep |
| `--file <name>` | Selective update — target single doc |
| `--no-fix` | Skip validation-fix loop |
| `--format <fmt>` | markdown (default) |

## Composite Metric

```
learn_score = validation% * 0.5 + coverage% * 0.3 + size_compliance% * 0.2
```

Higher = more accurate and complete documentation.

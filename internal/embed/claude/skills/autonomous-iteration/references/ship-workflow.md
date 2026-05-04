# Ship Workflow — $research-loop ship

Universal shipping workflow that applies autonomous loop principles to the
last mile — taking anything from "done" to "deployed/published/delivered."
Works for code, content, marketing, sales, research, or design.

**Core idea:** Shipping has a universal pattern. Identify → Checklist →
Prepare → Dry-run → Ship → Verify → Log.

## Trigger

- User invokes `$research-loop ship`
- User says "ship it", "deploy this", "publish this", "launch this"
- User says "get this out the door", "push to prod", "go live"

## Loop Support

```
# Ship with automatic preparation
$research-loop ship

# Bounded — iterate N times before shipping
$research-loop ship
Iterations: 10

# Ship specific artifact
$research-loop ship
Target: src/features/auth/**
Destination: production
```

## PREREQUISITE: Interactive Setup

**CRITICAL — BLOCKING PREREQUISITE:** If invoked without `--type` or target,
scan for staged changes, open PRs, recent commits, then gather input.

**Single batched call — all 3 questions at once:**

| # | Header | Question | Options |
|---|--------|----------|---------|
| 1 | `What` | "What are you shipping?" | "Code PR", "Release/tag", "Deployment", "Blog post/docs" |
| 2 | `Mode` | "How should I ship it?" | "Full workflow", "Dry-run only", "Checklist only", "Auto-approve" |
| 3 | `Monitor` | "Post-ship monitoring?" | "None", "5 min", "10 min", "30 min" |

## Architecture

```
$research-loop ship
  ├── Phase 1: Identify (what are we shipping?)
  ├── Phase 2: Inventory (what's the current state?)
  ├── Phase 3: Checklist (domain-specific pre-ship gates)
  ├── Phase 4: Prepare (autonomous loop until checklist passes)
  ├── Phase 5: Dry-run (simulate the ship action)
  ├── Phase 6: Ship (execute the actual delivery)
  ├── Phase 7: Verify (post-ship health check)
  └── Phase 8: Log (record the shipment)
```

## Phase 1: Identify — What Are We Shipping?

Auto-detect shipment type from context:

| Detection Signal | Shipment Type |
|-----------------|---------------|
| Dockerfile/k8s/deploy configs | deployment |
| Open PR or branch changes | code-pr |
| *.md in content/ | content |
| Mentions email/campaign | marketing-email |
| Mentions deck/proposal | sales |
| Mentions paper/report | research |
| Mentions assets/mockup | design |

## Phase 2: Inventory — Current State

- Read the target artifact(s)
- Check git status (clean? staged? committed?)
- Check CI status if applicable
- Check dependency health
- Assess readiness gaps

## Phase 3: Checklist — Domain-Specific Pre-Ship Gates

**Examples by type:**

**Code PR checklist:**
- [ ] All tests pass
- [ ] No type errors (tsc --noEmit)
- [ ] No lint errors
- [ ] Build succeeds
- [ ] CHANGELOG updated
- [ ] PR description written
- [ ] Breaking changes documented
- [ ] Security review completed (if applicable)

**Deployment checklist:**
- [ ] CI pipeline green
- [ ] Migration dry-run passes
- [ ] Rollback plan exists
- [ ] Health check endpoints ready
- [ ] Monitoring alerts configured
- [ ] Feature flags set

**Content checklist:**
- [ ] Spelling/grammar checked
- [ ] Links verified
- [ ] Images loaded
- [ ] SEO metadata complete
- [ ] Call to action included
- [ ] Preview renders correctly

## Phase 4: Prepare — Iterative Fix Loop

Run autonomous loop for each failing checklist item:

```
LOOP:
  1. Pick next failing checklist item
  2. Fix ONE thing
  3. Commit
  4. Re-run checklist
  5. If fixed → keep, move to next item
  6. If not → revert, try another approach
  7. Repeat until checklist 100% passes
```

## Phase 5: Dry-Run

Simulate the ship action without side effects:

| Type | Dry-Run Action |
|------|----------------|
| code-pr | `gh pr create --dry-run` |
| deployment | `kubectl apply --dry-run=client -f deploy.yaml` |
| content | Render preview, check links |
| marketing-email | Send test to internal address |

## Phase 6: Ship — Execute Delivery

**Requires user confirmation** unless `--auto` flag is set.

| Type | Ship Action |
|------|-------------|
| code-pr | `gh pr create` or merge existing PR |
| code-release | Git tag + release |
| deployment | `kubectl apply`, push to deploy branch |
| content | Publish via CMS or commit |
| marketing-email | Send via ESP |
| sales | Send proposal, share deck |

## Phase 7: Verify — Post-Ship Health Check

- Confirm ship action succeeded (exit code 0, URL accessible)
- Run health check endpoint
- Check monitoring dashboard (if `--monitor`)
- Watch for immediate errors (first 60 seconds)

## Phase 8: Log — Record Shipment

**Append to ship-log.tsv:**
```tsv
timestamp	type	target	checklist_pct	dry_run	status	notes
20250101-1200	deployment	production	100%	pass	shipped	All checks green
```

## Flags

| Flag | Purpose |
|------|---------|
| `--dry-run` | Validate without shipping |
| `--auto` | Auto-approve if checklist passes |
| `--force` | Skip non-critical checklist items |
| `--rollback` | Undo last ship action |
| `--monitor N` | Post-ship monitoring for N minutes |
| `--type <type>` | Override auto-detection |
| `--checklist-only` | Stop after Phase 3 |

## Composite Metric

```
ship_score = (checklist_passing / checklist_total) * 80
           + (dry_run_passed ? 15 : 0)
           + (no_blockers ? 5 : 0)
```

Score of 100 = fully ready. Below 80 = not shippable.

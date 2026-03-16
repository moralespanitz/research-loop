---
name: status
description: Use when user asks "where are we", "show me the status", "what do we have", "what's pending", "what should I do next", or wants to see the research decision tree.
---

<SUBAGENT-STOP>
If you were dispatched as a subagent, skip this skill. Execute your task and return results.
</SUBAGENT-STOP>

# Status Skill — Decision Tree View

Read the active session and render a structured decision tree. Not a log. Not a summary. A tree that shows every branch taken, every branch open, and exactly where the researcher is right now.

## Step 1 — Read session files

```bash
# Find active session
ls -t .research-loop/sessions/

# Read all structured files
cat .research-loop/sessions/<slug>/lab_notebook.md
cat .research-loop/sessions/<slug>/knowledge_graph.md  # if exists
cat .research-loop/sessions/<slug>/hypothesis.md       # if exists
cat .research-loop/sessions/<slug>/insights.md         # if exists
```

## Step 2 — Render the decision tree

Show this exact structure. Fill in what exists, mark what's open:

```
RESEARCH TREE — [Topic]
Last updated: [date]
═══════════════════════════════════════════════════

STAGE 1: EXPLORATION ✓
  Papers found: [N]
  Repos found: [N]
  Debates mapped: [N]

STAGE 2: GAPS ✓
  ├── GAP 1: [name] — [one line]
  ├── GAP 2: [name] — [one line]
  ├── GAP 3: [name] — [one line]
  └── ... [N total]

STAGE 3: HYPOTHESES
  ├── [H-A] [name] — Carlini: [score] — [PURSUING / PARKED / KILLED]
  ├── [H-B] [name] — Carlini: [score] — [PURSUING / PARKED / KILLED]
  └── [H-C] [name] — Carlini: [score] — [PURSUING / PARKED / KILLED]

STAGE 4: EXPERIMENT LANES (for active hypothesis)
  Hypothesis: [H-X name]
  ├── Lane 1: [name] — Gate: [score] — [ACTIVE / KILLED]
  ├── Lane 2: [name] — Gate: [score] — [ACTIVE / KILLED]
  ├── Lane 3: [name] — Gate: [score] — [SELECTED ◄]
  └── Lane 4: [name] — Gate: [score] — [ACTIVE / KILLED]

STAGE 5: EXPERIMENTS
  Active lane: [Lane name]
  ├── Condition 1 (priority 1): [name] — [PENDING / RUNNING / DONE: metric]
  ├── Condition 2 (priority 2): [name] — [PENDING / RUNNING / DONE: metric]
  ├── Condition 3 (priority 3): [name] — [PENDING / RUNNING / DONE: metric]
  └── Condition 4 (priority 4): [name] — [PENDING / RUNNING / DONE: metric]
  Falsification test: [PENDING / DONE]
  Kill criterion: [condition]

STAGE 6: INSIGHTS LEDGER
  Insights logged: [N]
  ├── Insight 1: [one line — what was learned]
  ├── Insight 2: [one line — what was learned]
  └── ...
  Hypothesis status: [strengthened / weakened / unchanged / killed]

STAGE 7: KNOWLEDGE
  Mental models: [N] learned
  Field debates: [N] mapped
  Understanding tests: [passed/failed/pending]

═══════════════════════════════════════════════════
YOU ARE HERE: [exact stage and step]

NEXT DECISION:
  [The one specific thing to do or decide right now]

OPEN BRANCHES (available if current path stalls):
  → [H-B] could be pursued if [H-A] fails
  → Lane 1 is alive as a fallback
  → [GAP 2] unexplored
```

## Step 3 — State the next decision explicitly

After the tree, say one thing only:

> "You are at [stage]. The next decision is: [specific question or action]. Options: [A], [B], or [C]."

Do not explain the whole tree. The tree does that. Just name the decision point.

## Step 4 — Update knowledge_graph.md

After rendering, update `.research-loop/sessions/<slug>/knowledge_graph.md` using Bash (not the Write tool — use a heredoc or `tee` to avoid read-first constraints):

```bash
cat > .research-loop/sessions/<slug>/knowledge_graph.md << 'EOF'
[content]
EOF
```

Write the current tree in this format:

```markdown
# Knowledge Graph — [Topic]
Updated: [date]

## Decision Tree

### Gaps
| ID | Name | Status |
|----|------|--------|
| GAP-1 | [name] | open / pursuing / closed |
| GAP-2 | [name] | open / pursuing / closed |

### Hypotheses
| ID | Name | Carlini Score | Status |
|----|------|--------------|--------|
| H-A | [name] | [score] | pursuing / parked / killed |
| H-B | [name] | [score] | pursuing / parked / killed |

### Experiment Lanes (H-A)
| Lane | Name | Gate Score | Status |
|------|------|-----------|--------|
| L-1 | [name] | [score] | alive / killed |
| L-3 | [name] | [score] | selected |

### Experiments
| Condition | DA_write | DA_read | Status | Result |
|-----------|----------|---------|--------|--------|
| Baseline | 1.0 | 1.0 | pending | — |
| Write-ablated | 0.1 | 1.0 | pending | — |
| Read-ablated | 1.0 | 0.1 | pending | — |
| Both-ablated | 0.1 | 0.1 | pending | — |

### Current position
Stage: [N]
Next decision: [what]
```

## Rules

- **Never summarize the lab notebook** — the tree is the view, not a prose summary
- **Always show all branches** — including killed ones, so the researcher sees why decisions were made
- **Mark one thing as current** with ◄
- **State one next decision** — not a list of options, one specific question
- **Update knowledge_graph.md every time** this skill runs

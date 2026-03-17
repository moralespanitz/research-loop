---
name: status
description: Use when user asks "where are we", "show me the status", "what do we have", "what's pending", "what should I do next", or wants to see the research decision tree.
---

<SUBAGENT-STOP>
If you were dispatched as a subagent, skip this skill. Execute your task and return results.
</SUBAGENT-STOP>

# Status Skill — Decision Tree + Session Briefing

Two things happen every time this skill runs:
1. The decision tree is shown in the conversation
2. `SESSION.md` is written — a one-screen briefing any future session can read to instantly resume

## Step 1 — Read all session files

```bash
SLUG=$(ls -t .research-loop/sessions/ | head -1)
cat .research-loop/sessions/$SLUG/lab_notebook.md
cat .research-loop/sessions/$SLUG/knowledge_graph.md 2>/dev/null
cat .research-loop/sessions/$SLUG/hypothesis.md 2>/dev/null
cat .research-loop/sessions/$SLUG/insights.md 2>/dev/null
```

## Step 2 — Render the decision tree in the conversation

```
RESEARCH TREE — [Topic]
Last updated: [date]
═══════════════════════════════════════════════════

STAGE 1: EXPLORATION [✓ / pending]
  Papers: [N] | Repos: [N] | Debates: [N]

STAGE 2: GAPS [✓ / pending]
  ├── GAP-1: [name] — [one line] — [pursuing / open]
  ├── GAP-2: [name] — [one line] — [open]
  └── ...

STAGE 3: HYPOTHESES [✓ / pending]
  ├── H-A: [name] — Carlini: [score] — [PURSUING / PARKED / KILLED]
  ├── H-B: [name] — Carlini: [score] — [PARKED]
  └── H-C: [name] — Carlini: [score] — [PARKED]

STAGE 4: LANES [✓ / pending]
  ├── Lane 1: [name] — Gate: [score] — [alive / killed]
  ├── Lane 3: [name] — Gate: [score] — [SELECTED ◄]
  └── ...

STAGE 5: EXPERIMENTS [← YOU ARE HERE / done]
  ├── Condition 1: [name] — [PENDING / DONE: metric]
  ├── Condition 2: [name] — [PENDING / DONE: metric]
  └── ...
  Insights: [N] logged

STAGE 6: KNOWLEDGE [✓ / pending]
  Mental models: [N] | Debates: [N] | Tests: [passed/pending]

═══════════════════════════════════════════════════
YOU ARE HERE: Stage [N] — [exact step]
NEXT DECISION: [one specific thing]
OPEN FALLBACKS: [what's available if current path fails]
═══════════════════════════════════════════════════
```

## Step 3 — Write SESSION.md

Write `.research-loop/sessions/<slug>/SESSION.md` using Bash heredoc.
This file is the single source of truth for resuming — designed to be read in 30 seconds.

```bash
cat > .research-loop/sessions/$SLUG/SESSION.md << 'EOF'
# Session: [Topic]
Last updated: [date]

## What this research is about
[2-3 sentences. What question are we trying to answer? Why does it matter?]

## Where we are
Stage: [N of 6]
Current step: [exact description]

## What has been decided
- Hypothesis: [one sentence]
- Approach: [one sentence — which lane, which route]
- Experiment design: [one sentence — what conditions, what the key test is]

## What has been learned so far
[One bullet per insight logged. If none yet: "No experiments run yet."]
- Insight 1: [one line]
- Insight 2: [one line]

## The next action
[One specific thing. Not a list. The single next step.]

## Open branches (fallbacks if current path fails)
- [option 1]: [one line why it's viable]
- [option 2]: [one line why it's viable]

## Key references
- [paper 1]: [one line]
- [paper 2]: [one line]
- [repo]: [one line]

## Files in this session
- lab_notebook.md — full log of everything
- knowledge_graph.md — structured tables of all decisions
- insights.md — experiment results and what they mean
- hypothesis.md — formal hypothesis and falsification criteria (if written)
EOF
```

## Step 4 — Update knowledge_graph.md

```bash
cat > .research-loop/sessions/$SLUG/knowledge_graph.md << 'EOF'
# Knowledge Graph — [Topic]
Updated: [date]

## Gaps
| ID | Name | Status |
|----|------|--------|
| GAP-1 | [name] | pursuing / open / closed |

## Hypotheses
| ID | Name | Carlini | Status |
|----|------|---------|--------|
| H-A | [name] | [score] | pursuing / parked / killed |

## Lanes
| Lane | Name | Gate | Status |
|------|------|------|--------|
| L-3 | [name] | [score] | selected |

## Experiments
| Condition | Status | Result |
|-----------|--------|--------|
| Baseline | pending | — |

## Current position
Stage: [N]
Next: [what]
EOF
```

## Rules

- **SESSION.md is always written** — every time this skill runs, no exceptions
- **SESSION.md fits on one screen** — if it's longer than 40 lines, cut it
- **Never summarize the lab notebook** — SESSION.md is a briefing, not a log
- **One next action** — not a list, not a plan, one specific thing
- **Show all branches** — including killed ones, so decisions are traceable

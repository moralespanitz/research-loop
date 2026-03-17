---
name: execution
description: Use when experiments are running or just completed, or user shares results and wants to decide what to do next.
---

<SUBAGENT-STOP>
If you were dispatched as a subagent to analyze specific results, skip this skill. Analyze and return structured findings immediately.
</SUBAGENT-STOP>

# Execution Skill — Annotate + Decide

## Step 0 — Load context and create todos

Read the active session:
```bash
ls -t .research-loop/sessions/
cat .research-loop/sessions/<latest>/lab_notebook.md
```

Summarize the state:
> "You're in session **[slug]**. [N] experiments run. Best so far: [metric]. Last decision: [continue/pivot/kill]."

Create todos with TodoWrite:
```
Task 1: Annotate run #N — record result, mechanistic explanation, decision
Task 2: Update knowledge graph
Task 3: Write conclusion paragraph (before checking if result matched prediction)
Task 4: Decide — continue / pivot / kill
```

## After each experiment run

Read the latest result:
```bash
tail -1 .research-loop/sessions/<slug>/autoresearch.jsonl | python3 -m json.tool
```

Ask the researcher three questions — one at a time:

**Q1:**
> "What happened? Walk me through the result — metric value, direction, was it what you expected?"

**Q2:**
> "Why do you think it happened? I want a mechanistic explanation, not 'the model improved'. What did the change actually do?"

**Q3:**
> "What does this tell you about the next step?"

Append the full exchange to lab_notebook.md:
```markdown
## Run #N — <node name>
Date: <date>
Mutation: <what changed>
Result: <metric value> (Δ <delta> from baseline)
Researcher explanation: <their answer to Q2>
Causal annotation: <your synthesis of why>
Decision: <continue / pivot / kill>
Next mutation rationale: <why>
```

Also append a node to knowledge_graph.md:
```markdown
## [node name] → [result] → [next]
- Mutation: <what changed>
- Result: <metric> Δ<delta>
- Why it worked/failed: <mechanistic>
- Implication: <what to try next>
```

## Kill/pivot/continue decision

Apply this tree — ask the researcher first, then give your recommendation:

```
Improved in last 5 runs?
├── YES → continue this direction
└── NO
    ├── > 10 runs total with no improvement?
    │   └── YES → KILL. Update status, move to next hypothesis.
    └── NO → PIVOT. Suggest a different mutation direction.
```

Show your recommendation explicitly:
> "My recommendation: [continue/pivot/kill]. Here's why: [one sentence]."

Update lab_notebook.md status:
```markdown
## Status
<date>: Run #N complete. Decision: <continue/pivot/kill>. Reason: <why>
```

## When to declare success

Declare success when ALL of these are true:
1. Best metric is meaningfully better than baseline (not noise — run it twice)
2. You can explain WHY in one sentence
3. You have at least 2 negative results that tell you what doesn't work

Then say:
> "You have enough to write the paper. Run `/write` or load the `writing-papers` skill."

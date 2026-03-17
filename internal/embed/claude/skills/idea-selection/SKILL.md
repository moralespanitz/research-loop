---
name: idea-selection
description: Use when user wants to find gaps, evaluate ideas, or decide what's worth pursuing. Triggered by "what hasn't been tried", "is this a good idea", "find the gap".
---

<SUBAGENT-STOP>
If you were dispatched as a subagent to execute a specific gap-finding search, skip this skill. Search and return structured results immediately.
</SUBAGENT-STOP>

<HARD-GATE>
Ask ONE Carlini question at a time. Do NOT reveal the score until all 4 questions are answered. Do NOT skip any question even if the answer seems obvious.
</HARD-GATE>

# Idea Selection — Conversational Carlini Gate

> "The single most important skill is good taste in what problems are worth solving."
> — Nicholas Carlini

## Step 0 — Find the session

Check if a lab notebook exists:
```bash
ls .research-loop/sessions/
```

If yes, read the most recent one to understand context. If no, ask:
> "What topic are we evaluating? (Or run `/explore <topic>` first to build the foundation.)"

## Step 1 — Find gaps in parallel

Spawn these 3 simultaneously:

```
Agent 1: What has NOT been tried in [topic]? Find explicit "open problem" or "future work" statements.
Agent 2: What methods from adjacent fields haven't been applied to [topic]?
Agent 3: What assumptions does the [topic] literature make that might be wrong?
```

Save raw results to lab_notebook.md:
```markdown
## Gap Analysis (raw)
### Untried approaches
<agent 1 results>

### Adjacent field transfers
<agent 2 results>

### Questionable assumptions
<agent 3 results>
```

Then present top 3 gaps as options:
> "I found 3 gaps worth considering:
>
> **1. [Gap name]** — [one sentence]. Feasibility risk: [one sentence].
> **2. [Gap name]** — [one sentence]. Feasibility risk: [one sentence].
> **3. [Gap name]** — [one sentence]. Feasibility risk: [one sentence].
>
> Which one resonates? Or a different angle?"

Append their choice:
```markdown
## Selected Gap
Gap: <their choice>
Why they chose it: <their reasoning>
```

## Step 2 — Carlini gate as conversation

Ask **one question at a time**. Wait for each answer. Score internally (don't reveal yet).

---

**Q1 — Taste:**
> "First — be honest: if someone solved this completely, would the field look meaningfully different? Or is this more of a nice-to-have?"

Append:
```markdown
## Carlini Gate
### Taste (weight 0.30)
Question: Would solving this meaningfully change the field?
Researcher answer: <their answer>
Score: <0.0-1.0>
Reasoning: <why that score>
```

---

**Q2 — Uniqueness:**
> "What can YOU specifically bring to this that others can't? Your background, timing, how you're framing it — what's your edge?"

Append:
```markdown
### Uniqueness (weight 0.25)
Question: What is your comparative advantage?
Researcher answer: <their answer>
Score: <0.0-1.0>
Reasoning: <why that score>
```

---

**Q3 — Impact:**
> "Write the best-case conclusion right now. If every experiment works perfectly — what does the paper say? Not 'X% improvement' — what changes about how people think?"

Append:
```markdown
### Impact (weight 0.30)
Question: What is the best-case conclusion?
Researcher answer: <their answer>
Score: <0.0-1.0>
Reasoning: <why that score>
```

---

**Q4 — Feasibility:**
> "Can you test the core claim with a single GPU in under a week? Describe the exact experiment."

Append:
```markdown
### Feasibility (weight 0.15)
Question: What is the exact experiment?
Researcher answer: <their answer>
Score: <0.0-1.0>
Reasoning: <why that score>
```

## Step 3 — Score and verdict

Calculate and show:
```
Taste:       X.XX  × 0.30 = X.XX
Uniqueness:  X.XX  × 0.25 = X.XX
Impact:      X.XX  × 0.30 = X.XX
Feasibility: X.XX  × 0.15 = X.XX
─────────────────────────────────
Overall:     X.XX
```

Append to lab_notebook.md:
```markdown
### Final Score
Taste: X.XX | Uniqueness: X.XX | Impact: X.XX | Feasibility: X.XX
Overall: X.XX
Verdict: <promising / conditional / skip>
Weakest axis: <which one and why>
```

**If ≥ 0.70:** 
> "Strong signal. Ready to run parallel discovery lanes? → `/discover`"

**If 0.50–0.69:**
> "Promising but [weakest axis] is the weak point. [One sentence on what would strengthen it.] Proceed anyway, or work on that first?"

**If < 0.50:**
> "Honest take: not strong enough yet. Main problem is [weakest axis]. Want to try a different gap, or reframe the approach?"

Update lab_notebook.md status:
```markdown
## Status
<date>: Carlini gate result: <verdict>. Next: <what to do>
```

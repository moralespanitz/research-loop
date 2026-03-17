---
description: "Find gaps in the literature and score them — interactive Carlini gate"
argument-hint: "[topic or leave blank to continue current session]"
---

Load the `idea-selection` skill, then run an interactive gap analysis for: $ARGUMENTS

## How to run gap analysis

Spawn these in parallel:
```
Agent 1: What has NOT been tried in $ARGUMENTS? Find explicit statements of open problems.
Agent 2: What methods from adjacent fields have NOT been applied to $ARGUMENTS?
Agent 3: What assumptions does the $ARGUMENTS literature take for granted that might be wrong?
```

Launch all 3 at once.

## After results

Present the top 3 gaps as numbered options:

> "I found 3 gaps worth considering:
> 
> 1. **[Gap name]** — [one sentence]. Risk: [one sentence].
> 2. **[Gap name]** — [one sentence]. Risk: [one sentence].  
> 3. **[Gap name]** — [one sentence]. Risk: [one sentence].
>
> Which resonates with you? Or is there a different angle you had in mind?"

## After the user picks a gap

Run the Carlini gate as a **conversation**:

Ask each question one at a time and wait for the user's answer:

1. > "**Taste check**: If someone solved this completely, would the field look meaningfully different? Tell me your honest take."
   
   (Wait for answer. Score 0.0–1.0 internally.)

2. > "**Uniqueness check**: What can YOU specifically bring to this that others can't? Think about your background, timing, or how you're framing it."
   
   (Wait for answer. Score 0.0–1.0 internally.)

3. > "**Impact check**: Write the best-case conclusion right now — if all your experiments work perfectly, what does the paper say? Not 'X% improvement' — something that changes how people think."
   
   (Wait for answer. Score 0.0–1.0 internally.)

4. > "**Feasibility check**: Can you test the core claim with a single GPU in under a week? What's the experiment?"
   
   (Wait for answer. Score 0.0–1.0 internally.)

## After all 4 answers

Show the score:
```
Taste:       X.XX
Uniqueness:  X.XX  
Impact:      X.XX
Feasibility: X.XX
─────────────────
Overall:     X.XX  [verdict]
```

If overall ≥ 0.70: "Strong signal. Ready to run parallel discovery lanes? → /discover"
If overall 0.50–0.69: "Promising but [weakest axis] needs work. Want to strengthen it before proceeding?"
If overall < 0.50: "Honest assessment: this direction isn't strong enough yet. Want to try a different gap?"

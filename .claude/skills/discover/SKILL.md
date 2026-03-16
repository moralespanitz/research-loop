---
name: discover
description: Use when user wants to test multiple angles on an idea, run parallel hypothesis lanes, or explore a gap from different entry points.
---

<SUBAGENT-STOP>
If you were dispatched as a subagent to research a specific lane, skip this skill. Search and return structured results immediately.
</SUBAGENT-STOP>

# Discover Skill — Parallel Lanes + Persistent

## Step 0 — Find session

Read the active lab_notebook.md to understand what was already explored and which gap was selected.

If no session exists:
> "Run `/explore <topic>` first to build the foundation, then come back here."

## Step 1 — Announce and launch

Say:
> "Spinning up 4 parallel research lanes for **[topic]**. Each explores a different angle. I'll apply Carlini gates between stages and kill weak lanes early."

Append to lab_notebook.md:
```markdown
## Discovery Run
Date: <date>
Topic: <topic>
Gap pursuing: <selected gap from idea-selection>

### Lanes launched
- Lane 1: [angle description]
- Lane 2: [angle description]
- Lane 3: [angle description]
- Lane 4: [angle description]
```

Spawn all 4 simultaneously. Each prompt must be long and task-shaped — start with "You are a research search agent. Do not ask questions. Search and return structured results immediately."

```
Agent 1 prompt:
  You are a research search agent. Do not ask questions. Search and return structured results immediately.
  
  Task: Find the most incremental improvement on SOTA for [topic].
  Search for 3 recent papers (2022–2025) that represent the current frontier.
  Then identify: what is the single most direct gap left open by these papers?
  Propose a hypothesis that closes it. Suggest a concrete experiment (what to change, what metric).
  Estimate feasibility: can it be tested in under 1 week on 1 GPU?
  Return: papers (title, year, 1-line contribution), gap, hypothesis, experiment, feasibility score 0–1.

Agent 2 prompt:
  You are a research search agent. Do not ask questions. Search and return structured results immediately.
  
  Task: Find a cross-field transfer opportunity for [topic] from an adjacent field.
  Search for 3 papers from a neighboring discipline (e.g. neuroscience→ML, physics→biology, etc.) that contain ideas not yet applied to [topic].
  Identify the most promising transfer. Propose a hypothesis. Suggest a concrete experiment.
  Return: papers (title, year, 1-line contribution), gap, hypothesis, experiment, feasibility score 0–1.

Agent 3 prompt:
  You are a research search agent. Do not ask questions. Search and return structured results immediately.
  
  Task: Find a core assumption in [topic] that the field takes for granted but may be wrong.
  Search for 3 papers that either challenge this assumption or would be undermined if it were false.
  State the assumption clearly. Propose the counter-hypothesis. Suggest a falsification experiment.
  Return: papers (title, year, 1-line contribution), assumption, counter-hypothesis, experiment, feasibility score 0–1.

Agent 4 prompt:
  You are a research search agent. Do not ask questions. Search and return structured results immediately.
  
  Task: Find a systems/efficiency angle for [topic] — how to make it work on constrained hardware or at scale.
  Search for 3 papers on efficient implementations, approximations, or hardware-aware designs in [topic].
  Identify the key bottleneck. Propose a hypothesis for removing it. Suggest a concrete experiment.
  Return: papers (title, year, 1-line contribution), bottleneck, hypothesis, experiment, feasibility score 0–1.
```

## Step 2 — Apply Carlini gate to each lane

For each lane result, score it (0.0–1.0) on:
- Taste: would this change the field?
- Novelty: has this been tried?
- Feasibility: testable in < 1 week on 1 GPU?

Gate threshold: 0.5. Kill lanes below threshold.

Append all results to lab_notebook.md:
```markdown
### Lane Results

#### Lane 1 — [angle]
Claim: <hypothesis>
Experiment: <what to change and measure>
Gate score: <X.XX>
Status: <survived / killed>
Kill reason (if killed): <why>
Papers found: <list>

#### Lane 2 — [angle]
...
```

## Step 3 — Present surviving lanes as options

For each surviving lane, show a card:

> **Lane [N]: [angle name]** (score: X.XX)
> **Claim:** [one sentence — what you're testing]
> **Experiment:** [one sentence — what to change, what metric to watch]
> **Why it survived:** [one sentence]

Then:
> "Which lane do you want to pursue? You can also ask me to dig deeper into any before deciding."

Append their choice:
```markdown
### Selected Lane
Lane: <N>
Angle: <angle>
Claim: <full hypothesis>
Experiment: <full experiment description>
Researcher notes: <anything they said>
```

Update status:
```markdown
## Status
<date>: Discovery complete. Selected lane: <N>. Next: /loop to start experiments.
```

## Step 4 — Transition

> "Good choice. Now set up the experiment. Run `/loop` once your baseline repo is ready.
> 
> Your hypothesis is saved in the lab notebook. To resume later: `/resume`"

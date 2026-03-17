---
description: "Run parallel hypothesis lanes with Carlini gates — presents results as options"
argument-hint: "<topic or hypothesis>"
---

Load the `discover` skill, then run parallel discovery for: $ARGUMENTS

## How to run discovery

Tell the user upfront:
> "Spinning up 4 parallel research lanes for **$ARGUMENTS**. Each lane explores a different angle. I'll kill weak ones early using Carlini gates."

Spawn 4 Agent calls simultaneously, each exploring a different angle:
```
Agent 1: Literature + gap + hypothesis for angle: [most conservative/incremental angle]
Agent 2: Literature + gap + hypothesis for angle: [most novel/risky angle]  
Agent 3: Literature + gap + hypothesis for angle: [cross-field transfer angle]
Agent 4: Literature + gap + hypothesis for angle: [systems/efficiency angle]
```

## After lanes complete

For each surviving lane (not killed by gate), present as a card:

> **Lane [N]: [angle name]**
> Claim: [one sentence]
> Experiment: [one sentence — what to change, what metric to watch]
> Gate score: X.XX
> Why it survived: [one sentence]

Then ask:
> "Which lane do you want to pursue? You can also ask me to go deeper on any of them before deciding."

## If user picks a lane

Say:
> "Good choice. To start experiments, run `/loop` and point me at your baseline repo."

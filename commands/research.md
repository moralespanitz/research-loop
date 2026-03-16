---
description: "Full research pipeline: explore → score → discover → loop"
argument-hint: "<topic>"
---

Invoke the `research-loop` skill and run the FULL research pipeline for: $ARGUMENTS

Follow these steps in order:
1. Invoke `research-loop:explore` — gather papers, mental models, debates, score
2. If Carlini score ≥ 0.5: invoke `research-loop:discover` — parallel lanes with gates
3. If a lane survives as "promising": invoke `research-loop:loop` — run experiments
4. After experiments: invoke `research-loop:execution` — annotate and decide next step

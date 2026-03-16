---
description: "Resume an active research session — reloads full context from lab notebook"
argument-hint: "[session-slug or leave blank for latest]"
---

Load the `research-loop` skill, then resume the research session: $ARGUMENTS

## How to resume

1. Find the session:
```bash
ls -t .research-loop/sessions/
```

If $ARGUMENTS is blank, use the most recent directory.

2. Read the full lab notebook:
```bash
cat .research-loop/sessions/<slug>/lab_notebook.md
```

3. Summarize where we left off:
> "Resuming **[session name]**.
>
> **Topic:** [topic]
> **Last action:** [last status line]
> **Carlini score:** [score if done]
> **Selected hypothesis:** [claim if chosen]
> **Experiments run:** [N]
> **Current best metric:** [value if experiments started]
>
> What do you want to do next?"

Give 3 options based on current status:
- If status = "exploring" → "A) Continue exploring B) Run gap analysis C) Jump to Carlini gate"
- If status = "gap analysis done" → "A) Run discover B) Refine the gap C) Start over with new angle"
- If status = "hypothesis selected" → "A) Start experiments B) Review the hypothesis C) Run more discovery lanes"
- If status = "experiments running" → "A) Annotate latest result B) View knowledge graph C) Declare done"

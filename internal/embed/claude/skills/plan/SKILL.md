---
name: plan
description: Use when user has a selected hypothesis or route and needs a concrete execution plan broken into tasks. Triggered by "what are the next steps", "how do I start", "plan this out".
---

<SUBAGENT-STOP>
If you were dispatched as a subagent, skip this skill. Execute your task and return results.
</SUBAGENT-STOP>

<HARD-GATE>
Do NOT produce a vague timeline or milestone list. Every task must be specific enough that someone with no context could execute it. If a task takes more than 1 day, break it down further.
</HARD-GATE>

# Plan Skill — Concrete Execution Plan

The difference between a plan that works and one that doesn't is specificity. "Set up the codebase" is not a task. "Fork IDSIA/recurrent-fwp, rename to dopamine-read-gate, run `python train.py` to confirm baseline, record output" is a task.

## Step 1 — Read the session

Read the lab notebook to understand exactly where the researcher is:
```bash
cat .research-loop/sessions/<slug>/lab_notebook.md
```

Summarize back to the researcher:
> "You've selected [hypothesis/route]. You have [what's done]. The next phase is [what's needed]."

Ask one clarifying question if anything is ambiguous. Then proceed.

## Step 2 — Write the plan

Break the work into tasks. Each task must have:
- A specific action (verb + object)
- The exact file, repo, command, or person involved
- How to verify it's done (what does success look like?)
- Time estimate (be honest — hours, not days)

Format:
```markdown
## Execution Plan — [hypothesis name]
Date: <date>

### Phase: [name]

**Task 1: [action]**
- Do: [exact steps]
- Verify: [how you know it worked]
- Time: [estimate]

**Task 2: [action]**
- Do: [exact steps]
- Verify: [how you know it worked]
- Time: [estimate]

...
```

## Step 3 — Create todos

Use TodoWrite to create one todo per task. Mark the first one as `in_progress`.

Show the researcher the plan and ask:
> "Does this cover everything? Any task that's unclear or needs breaking down further?"

Wait for approval. Then say:
> "Starting Task 1. Tell me when you're ready."

## Step 4 — Track progress

As each task completes:
1. Mark it complete in TodoWrite
2. Append a one-line note to lab_notebook.md: `[date]: Task [N] complete — [what was done]`
3. Surface the next task explicitly: > "Task [N] done. Next: [Task N+1 description]. Ready?"

## What makes a bad plan

- Milestones instead of tasks ("Week 1: build the model")
- Tasks without verification steps ("set up environment")
- Tasks that depend on unknowns without surfacing them ("get lab data" — do you have the contact? have you emailed?)
- No time estimates

If you catch yourself writing any of these, stop and break it down further.

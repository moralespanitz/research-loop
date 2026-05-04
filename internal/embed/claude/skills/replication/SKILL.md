---
name: replication
description: Plan and execute a structured replication workflow for a paper, claim, or benchmark with environment selection and integrity checks.
---

<SUBAGENT-STOP>
If you were dispatched as a subagent to execute a specific replication task, skip this skill. Do the task and return structured results immediately.
</SUBAGENT-STOP>

<HARD-GATE>
Do NOT install packages, run training, or execute experiments without confirming the execution environment first.
</HARD-GATE>

# Replication Skill — Reproduce Published Results

## Workflow Overview

```
EXTRACT → PLAN → ENVIRONMENT → EXECUTE → LOG → REPORT
```

Derive a short slug from the paper/claim name (lowercase, hyphens, no filler words, ≤5 words). Use this slug for all files.

---

## Step 1: Extract

Use the researcher subagent (`.claude/agents/researcher.md`) to pull implementation details from the target paper and any linked code repositories.

The researcher should extract:
- Paper citation and URL
- Linked code/data repositories (GitHub, Hugging Face, etc.)
- Key claims, reported metrics, and experimental setup
- Dataset references and preprocessing steps
- Training hyperparameters (learning rate, batch size, optimizer, scheduler, seed)
- Evaluation protocol (metrics, splits, baselines)
- Environment requirements (framework versions, CUDA, OS)

Save extraction to `.research-loop/sessions/<slug>/extraction.md`.

If `CHANGELOG.md` (in `.research-loop/`) exists and this is a continuation, read the most recent relevant entries before planning.

---

## Step 2: Plan

Create `.research-loop/sessions/<slug>/plan.md` with:
- **Replication target** — paper citation, claim, or benchmark to reproduce
- **Expected outcomes** — exact metrics, figures, or behaviors to compare against
- **Code to implement** — what needs to be written, adapted, or run
- **Datasets required** — name, source URL, expected size
- **Environment requirements** — framework versions, hardware needs
- **Success criteria** — explicit test oracles: "replication succeeds if metric X is within Y% of reported value"
- **Task ledger** — implementation steps ordered by dependency
- **Verification log** — what checks will be performed and how
- **Risk factors** — paper details that are underspecified, ambiguous, or known-hard to reproduce

Be explicit about what is verified, what is inferred, what is still missing, and which checks or test oracles will be used to decide whether the replication succeeded.

---

## Step 3: Environment

Before running anything, ask the user where to execute:

| Environment | When to use | Setup |
|-------------|-------------|-------|
| **Local** | Simple experiments in the current working directory | No setup needed |
| **Virtual environment** | Isolated Python environment needed | `python -m venv .venv && source .venv/bin/activate` |
| **Docker** | Full isolation, reproducible environment | Write a `Dockerfile` and `docker build/run` |
| **Modal** | Serverless GPU for burst jobs | `pip install modal && modal setup`; write a Modal-decorated script |
| **RunPod** | Long-running GPU experiments with SSH | `runpodctl` CLI + `RUNPOD_API_KEY`; provision pod, transfer files, execute |
| **Plan only** | No execution — produce the plan without running | — |

Do not proceed without user confirmation of the environment.

---

## Step 4: Execute

Implement and run the replication steps in the chosen environment.

For each step:
1. Write the code/script to `.research-loop/sessions/<slug>/scripts/step-N-<name>.py` or equivalent
2. Run it in the chosen environment
3. Save raw outputs to `.research-loop/sessions/<slug>/outputs/step-N-output.txt`
4. Log observations, discrepancies, and unexpected results

### Integrity checks (mandatory)
- **Run actual code.** Do not describe what code would do — run it and capture real output.
- **Compare metrics directly.** Expected metric vs. actual metric, side by side.
- **Note discrepancies.** If the paper reports 94.2% and you observe 92.1%, do not round or soften the gap. Report it exactly.
- **Distinguish setup differences.** If you had to use a different framework version, smaller batch size, or fewer GPUs, say so explicitly.
- **Dead end detection.** If a step fails (dependency conflict, missing dataset, underspecified method), record the failure and continue with the next independent step. Do not silently skip.

Do not call the outcome replicated unless the planned checks actually passed.

---

## Step 5: Log

For multi-step or resumable replication work, append concise entries to `.research-loop/CHANGELOG.md` after:
- Meaningful progress
- Failed attempts
- Major verification outcomes
- Before stopping mid-workflow

Each entry must record: the active objective, what changed, what was checked, and the next step.

---

## Step 6: Report

Write the final replication report to `.research-loop/sessions/<slug>/report.md`.

Include:
- **Target** — paper/claim being replicated
- **Environment** — exact hardware, software versions, execution context
- **Method** — what you implemented and how
- **Results** — expected vs. observed, side by side
  - Use a table for metrics comparison
- **Discrepancies** — list every difference between reported and observed results
- **Verification status** — SUCCESSFUL / PARTIAL / FAILED / BLOCKED
- **Root cause analysis** — if partial or failed, what factors likely explain the gap
- **Sources** — paper URL, repository URL, dataset URL, any additional references

End with a `Sources` section containing direct URLs for all primary references.

---

## What NOT to do
- Do not install packages or run experiments without environment confirmation
- Do not claim replication succeeded when checks failed
- Do not round away discrepancies
- Do not silently skip underspecified steps — log them as BLOCKED
- Do not produce a report without running actual code (unless user chose "Plan only")

---

## Subagent Reference

| Agent | When dispatched |
|-------|-----------------|
| researcher | Step 1 — extract implementation details from paper and linked code |

The researcher agent definition lives in `.claude/agents/researcher.md`.

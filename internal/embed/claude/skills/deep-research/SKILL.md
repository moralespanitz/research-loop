---
name: deep-research
description: Run a thorough, multi-phase deep research investigation on a topic with subagent dispatch, provenance tracking, and integrity verification.
---

<SUBAGENT-STOP>
If you were dispatched as a subagent to execute a specific research task, skip this skill. Do the task and return structured results immediately.
</SUBAGENT-STOP>

<HARD-GATE>
Do NOT run searches, fetch sources, call subagents, or produce final artifacts until the user has confirmed the research plan. Write the plan first, then ask for confirmation.
</HARD-GATE>

# Deep Research Skill — Multi-Phase Investigation

This is an execution request, not a request to explain or implement the workflow instructions. Execute the workflow. Do not answer by describing the protocol or restating the instructions.

## Required Artifacts

Derive a short slug from the topic: lowercase, hyphenated, no filler words, at most 5 words.

Every run must leave these files on disk:
- `.research-loop/sessions/<slug>/plan.md`
- `.research-loop/sessions/<slug>/draft.md`
- `.research-loop/sessions/<slug>/draft-cited.md`
- `.research-loop/sessions/<slug>/report.md`
- `.research-loop/sessions/<slug>/provenance.md`

After the user approves the plan, if any capability fails, continue in degraded mode and still write a blocked or partial final output and provenance sidecar. Never end with chat-only output after plan approval. Use `Verification: BLOCKED` when verification could not be completed.

---

## Step 1: Plan

Create `.research-loop/sessions/<slug>/plan.md` immediately. The plan must include:
- Key questions
- Evidence needed
- Scale decision (direct vs. subagent dispatch)
- Task ledger
- Verification log
- Decision log

Make the scale decision before assigning owners in the plan. If the topic is a narrow "what is X" explainer, the plan must use lead-owned direct search tasks only; do not allocate researcher subagents.

After writing the plan, stop and ask for explicit confirmation before gathering evidence. Summarize the plan briefly and ask:

> Proceed with this deep research plan? Reply "yes" to continue, or tell me what to change.

Do not run searches, fetch sources, spawn subagents, draft, cite, review, or deliver final artifacts until the user confirms. If the user requests changes, update the plan file first, then ask again.

---

## Step 2: Scale Decision

### Use direct search for:
- Single fact or narrow question, including "what is X" explainers
- Work you can answer with 3–10 tool calls
- No subagents unless the user explicitly asks for comprehensive coverage

### Use researcher subagents (from `.claude/agents/researcher.md`) when:
- Direct comparison of 2–3 items: 2 researcher subagents
- Broad survey or multi-faceted topic: 3–4 researcher subagents
- Complex multi-domain research: 4–6 researcher subagents

---

## Step 3: Gather Evidence

### Integrity Commandments (from researcher agent)
1. **Never fabricate a source.** Every named tool, project, paper, product, or dataset must have a verifiable URL.
2. **Never claim a project exists without checking.** Before citing a GitHub repo, search for it. Before citing a paper, find it.
3. **Never extrapolate details you haven't read.** If you haven't fetched and inspected a source, you may note its existence but must not describe its contents.
4. **URL or it didn't happen.** Every entry in your evidence table must include a direct, checkable URL.
5. **Read before you summarize.** Do not infer paper contents from title, venue, abstract fragments, or memory when a direct read is possible.
6. **Mark status honestly.** Distinguish between claims read directly, claims inferred from multiple sources, and unresolved questions.

### Avoid crash-prone PDF parsing
Do not fetch `.pdf` URLs unless the user explicitly asks for PDF extraction. Prefer paper metadata, abstracts, HTML pages, official docs, and web snippets. If only a PDF exists, cite the PDF URL from search metadata and mark full-text PDF parsing as blocked instead of fetching it.

### If direct search was chosen:
- Search and fetch sources yourself.
- Use multiple search terms/angles before drafting. Minimum: 3 distinct queries for direct-mode research.
- Record the exact search terms used in the session notebook.
- Write notes to `.research-loop/sessions/<slug>/research-notes.md`.
- Continue to synthesis.

### If subagents were chosen:
- Write a per-researcher brief: `.research-loop/sessions/<slug>/brief-T1.md`, `-T2.md`, etc.
- Keep subagent tool-call JSON small and valid.
- Always set `failFast: false`.
- Prefer broad guidance: "use paper search and web search"; if PDF parsing fails, continue from metadata.
- Example subagent dispatch:
```
Agent: researcher
Task: Read .research-loop/sessions/<slug>/brief-T1.md and write <slug>-findings-web.md.
Output: .research-loop/sessions/<slug>/findings-web.md
```

After evidence gathering, update the plan task ledger and verification log.

---

## Step 4: Draft

Write the report yourself. Do not delegate synthesis.

Save to `.research-loop/sessions/<slug>/draft.md`.

Include:
- Executive summary
- Findings organized by question/theme
- Evidence-backed caveats and disagreements
- Open questions
- No invented sources, results, figures, benchmarks, images, charts, or tables

Before citation, sweep the draft:
- Every critical claim, number, figure, table, or benchmark must map to a source URL, research note, raw artifact path, or command/script output.
- Remove or downgrade unsupported claims.
- Mark inferences as inferences.

---

## Step 5: Cite

### If direct search was chosen:
- Do citation yourself.
- Verify reachable URLs with available fetch/search tools.
- Copy or rewrite draft to `.research-loop/sessions/<slug>/draft-cited.md` with inline citations and a Sources section.

### If researcher subagents were used:
- Dispatch the verifier agent (`.claude/agents/verifier.md`) after the draft exists.
- This step is mandatory and must complete before any reviewer runs.
- Do not run verifier and reviewer in the same parallel subagent call.
- After verifier returns, verify on disk that `draft-cited.md` exists.

---

## Step 6: Review

### If direct search was chosen:
- Review the cited draft yourself.
- Write `.research-loop/sessions/<slug>/verification.md` with FATAL / MAJOR / MINOR findings.
- Fix FATAL issues before delivery.

### If researcher subagents were used:
- Dispatch the reviewer agent (`.claude/agents/reviewer.md`) against the cited draft.
- If the reviewer flags FATAL issues, fix them before delivery and run one more review pass.
- Note MAJOR issues in Open Questions. Accept MINOR issues.

When applying reviewer fixes:
- Use small localized edits for 1–3 simple corrections.
- For section rewrites or more than 3 substantive fixes, write a corrected full file to `.research-loop/sessions/<slug>/draft-revised.md`.
- After fixes, run explicit on-disk verification (grep/diff/stat) before saying fixes landed.

The final candidate is `draft-revised.md` if it exists; otherwise `draft-cited.md`.

---

## Step 7: Deliver

Copy the final candidate to `.research-loop/sessions/<slug>/report.md`.

Write provenance as `.research-loop/sessions/<slug>/provenance.md`:

```markdown
# Provenance: [topic]

- **Date:** [date]
- **Rounds:** [number of research rounds]
- **Sources consulted:** [count and/or list]
- **Sources accepted:** [count and/or list]
- **Sources rejected:** [dead, unverifiable, or removed]
- **Verification:** [PASS / PASS WITH NOTES / BLOCKED]
- **Plan:** .research-loop/sessions/<slug>/plan.md
- **Research files:** [files used]
```

Before responding, verify on disk that all required artifacts exist. If verification could not be completed, set `Verification: BLOCKED` or `PASS WITH NOTES` and list the missing checks.

Verify that any fixes claimed in the provenance are reflected in the final candidate. Do not claim "all patches applied" unless these checks succeed.

---

## Subagent Reference

This skill dispatches the following subagents (defined in `.claude/agents/`):

| Agent | File | When dispatched |
|-------|------|-----------------|
| researcher | `.claude/agents/researcher.md` | Evidence gathering (Step 3) for multi-faceted topics |
| verifier | `.claude/agents/verifier.md` | Citation verification (Step 5) after researcher subagents |
| reviewer | `.claude/agents/reviewer.md` | Review pass (Step 6) on cited draft |

All subagents load their definitions from `.claude/agents/`. These definitions encode integrity commandments, output formats, and operating rules specific to each role.

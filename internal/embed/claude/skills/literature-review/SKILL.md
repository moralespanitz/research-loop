---
name: literature-review
description: Run a structured literature review on a topic using parallel search, evidence tables with quality scoring, and primary-source synthesis.
---

<SUBAGENT-STOP>
If you were dispatched as a subagent to execute a specific review task, skip this skill. Do the task and return structured results immediately.
</SUBAGENT-STOP>

<HARD-GATE>
Do NOT launch searches or dispatch subagents until you have written the plan. Write the plan and continue immediately. Do not ask for confirmation unless the user explicitly requested plan review.
</HARD-GATE>

# Literature Review Skill — Parallel Search + Evidence Synthesis

## Workflow Overview

```
PLAN → GATHER (parallel researcher subagents) → SYNTHESIZE → CITE (verifier) → VERIFY (reviewer) → DELIVER
```

---

## Step 1: Plan

Derive a short slug from the topic (lowercase, hyphens, no filler words, ≤5 words). Use this slug for all files.

Create `.research-loop/sessions/<slug>/plan.md` with:
- Key questions the review aims to answer
- Search angles (3–6 distinct angles for parallel dispatch)
- Source types to search: papers, web docs, repos, benchmarks, conference proceedings
- Time period / recency window
- Expected sections for the final review
- Task ledger tracking each angle
- Verification log

Write the plan and continue immediately. Summarize briefly. Do not ask for confirmation.

---

## Step 2: Gather Evidence

### Parallel search across multiple angles

Dispatch researcher subagents (`.claude/agents/researcher.md`) for each search angle. For a typical literature review, use 3–4 subagents in parallel:

| Agent | Task |
|-------|------|
| researcher | Search academic papers: foundational works, recent advances, surveys |
| researcher | Search official documentation, benchmarks, and implementation repos |
| researcher | Search for field disagreements, unresolved debates, open problems |
| researcher (if broad scope) | Search adjacent fields, cross-domain applications, alternative approaches |

Each researcher writes to `.research-loop/sessions/<slug>/findings-<angle>.md`.

### Evidence table with quality scoring

Each researcher must produce an evidence table with:

| # | Source | URL | Key claim | Type | Quality |
|---|--------|-----|-----------|------|---------|
| 1 | ... | ... | ... | primary / secondary / self-reported | A / B / C |

### Source quality tiers

| Tier | Description | Examples |
|------|-------------|---------|
| **A (Highest)** | Peer-reviewed papers, official documentation, verified benchmarks, primary datasets, government filings | arXiv proceedings, IEEE/ACM, official docs |
| **B (Good)** | Reputable secondary sources, expert technical blogs, well-cited surveys, established trade publications | Distill.pub, high-quality blog posts |
| **C (Accept with caveats)** | Undated posts, content aggregators, social media with primary links, vendor claims without independent verification | Listicles, Medium posts without citations |
| **Reject** | No author + no date, AI-generated content without primary backing, anonymous claims | — |

When initial results skew toward low-quality sources, re-search with domain filters targeting authoritative domains (`.edu`, `.gov`, arXiv, ACL, NeurIPS, etc.).

### Context hygiene for researchers

- Write findings progressively to files. Do not accumulate full page contents in working memory.
- Use `includeContent: true` for top candidates only; triage by title/snippet first.
- Triage 10+ results by title before fetching full content.
- Track assigned questions explicitly: mark each as `done`, `blocked`, or `needs follow-up`.

---

## Step 3: Synthesize

Write the synthesis yourself. Do not delegate to subagents.

Save to `.research-loop/sessions/<slug>/draft.md`.

Structure:
- **Executive Summary** — 2–3 paragraph overview of findings
- **Topic Clusters** — Group findings by thematic clusters, not by source
- **Consensus Points** — What the literature agrees on
- **Disagreements** — Where experts diverge, with steel-manned arguments for each side
- **Research Gaps** — Explicit gaps identified in the literature; unanswered questions
- **Open Questions** — What remains unresolved and why it matters

Synthesis rules:
- Separate consensus, disagreements, and open questions clearly
- When useful, propose concrete next experiments or follow-up reading
- Before finishing the draft, sweep every strong claim against the verification log
- Downgrade anything that is inferred or single-source critical

---

## Step 4: Cite

Dispatch the verifier agent (`.claude/agents/verifier.md`) to add inline citations and verify every source URL.

The verifier will:
- Anchor every factual claim to a numbered source citation
- Verify every URL resolves and contains the claimed content
- Remove unsourced factual claims
- Build a unified Sources section
- Deduplicate sources across multiple research files
- Enforce: every `[N]` in body maps to Sources, every Sources entry is cited at least once

Save the cited draft to `.research-loop/sessions/<slug>/draft-cited.md`.

---

## Step 5: Verify

Dispatch the reviewer agent (`.claude/agents/reviewer.md`) to check the cited draft for:
- Unsupported claims
- Logical gaps
- Zombie sections (content surviving from earlier drafts without support)
- Single-source critical findings
- Overstated confidence

The reviewer classifies issues as:
- **FATAL** — must fix before delivery
- **MAJOR** — note in Open Questions
- **MINOR** — accept as-is

Fix FATAL issues before delivery. If FATAL issues were found, run one more verification pass after the fixes.

---

## Step 6: Deliver

Save the final literature review to `.research-loop/sessions/<slug>/report.md`.

Write a provenance sidecar to `.research-loop/sessions/<slug>/provenance.md`:

```markdown
# Provenance: Literature Review — [topic]

- **Date:** [date]
- **Search angles:** [list of angles]
- **Sources consulted:** [count and/or list]
- **Sources accepted:** [count]
- **Sources rejected:** [count] — [reasons]
- **Verification:** [PASS / PASS WITH NOTES / BLOCKED]
- **Plan:** .research-loop/sessions/<slug>/plan.md
- **Research files:** [files used]
```

Before responding, verify on disk that both `report.md` and `provenance.md` exist.

---

## Subagent Reference

| Agent | When dispatched |
|-------|-----------------|
| researcher | Step 2 — evidence gathering across parallel angles |
| verifier | Step 4 — citation verification and URL checking |
| reviewer | Step 5 — review pass on cited draft |

All subagent definitions live in `.claude/agents/`.

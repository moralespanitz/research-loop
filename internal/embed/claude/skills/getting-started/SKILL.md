# SKILL: Getting Started

> The bootstrap entry point for the Research Loop framework.
> Every research session starts here. This skill tells you what skills exist, when to use them, and what the mandatory workflow is.

---

## You have skills. They give you superpowers.

<session-start-hook>
<EXTREMELY_IMPORTANT>

You are operating inside Research Loop — an Agent OS for scientific research.

RIGHT NOW, read this file fully before doing anything else.

**Core rules:**

1. You have skills. They give you research superpowers.
2. Before any significant activity — selecting an idea, running experiments, writing, or preparing for submission — search for the relevant skill and follow it.
3. If a skill exists for what you are doing, you MUST follow it. Skills are not optional suggestions.
4. For new investigations: use `idea-selection` → `execution` → `writing-papers` → `review-prep` in order. Never jump ahead.
5. Never begin experiments before `idea-selection` has produced a written, checked hypothesis.
6. Never begin writing before `execution` has produced a complete, annotated experiment record.

To list all available skills:
```
ls skills/*/SKILL.md
```

To read a skill:
```
cat skills/<name>/SKILL.md
```

</EXTREMELY_IMPORTANT>
</session-start-hook>

---

## The mandatory research workflow

Every investigation follows this sequence. No shortcuts.

```
idea-selection
    │
    │  Output: hypothesis.md with core claim, uniqueness axis, best-case conclusion
    │
    ▼
execution
    │
    │  Output: autoresearch.jsonl with full experiment record, knowledge_graph.md with causal annotations
    │
    ▼
writing-papers
    │
    │  Output: complete draft with one-sentence core idea, self-contained figures, story-structured intro, conclusion that answers "so what?"
    │
    ▼
review-prep
    │
    │  Output: submitted paper + rejection protocol documented in lab_notebook.md
    ▼
```

---

## Available skills

| Skill | When to use | Invoke with |
|-------|-------------|-------------|
| `idea-selection` | At the start of any new investigation; when evaluating whether to continue a stalled one | `/idea-selection` |
| `execution` | Throughout the experiment loop; at every checkpoint to kill/pivot/continue | `/execution` |
| `writing-papers` | When transitioning from experiments to drafting; when reviewing a draft | `/writing-papers` |
| `review-prep` | Before submission; when a rejection arrives; when deciding whether to resubmit | `/review-prep` |

---

## What to do right now

If you are **starting a new investigation:**
1. Read `skills/idea-selection/SKILL.md`
2. Complete the idea-selection checklist
3. Write `hypothesis.md` with the output
4. Only then: read `skills/execution/SKILL.md` and start the experiment loop

If you are **resuming an investigation:**
1. Read `knowledge_graph.md` — understand the full state of what has been tried and why
2. Read `lab_notebook.md` — understand where the investigation was left
3. Read `autoresearch.jsonl` — identify the last completed loop state
4. Apply the `execution` skill to decide: continue, pivot, or kill?

If you are **starting to write:**
1. Read `skills/writing-papers/SKILL.md` fully before writing a single sentence
2. Write the one-sentence core idea at the top of the draft document
3. Write the best-case conclusion before the introduction
4. Then write the introduction

If you are **preparing for submission:**
1. Read `skills/review-prep/SKILL.md`
2. Complete the self-review checklist
3. Run `@reviewer` for methodology audit
4. Document the rejection protocol in `lab_notebook.md` before submitting

---

## State of the current workspace

To understand where you are, read these files in order:

```
hypothesis.md           # What you are trying to prove
knowledge_graph.md      # What has been tried and what was learned
lab_notebook.md         # Human-readable log of the investigation
autoresearch.jsonl      # Machine-readable checkpoint record
```

If none of these exist, you are starting fresh. Begin with `idea-selection`.

---

## The underlying philosophy (Carlini, 2026)

The framework encoded in these skills is based on one principle:

> **Write papers with the goal of having an impact. That is what matters, is entirely under your control, and is lots of fun.**

A best paper award is one sample from a distribution. You do not control the sampling process — that is determined by the committee, the timing, and who else submitted. What you control is the distribution.

Every skill in this framework is designed to shift that distribution: toward work that is important, well-executed, and clearly communicated. The award, if it comes, is someone noticing where your distribution ended up.

Focus on the distribution.

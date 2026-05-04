---
name: explore
description: Use when user wants to explore a topic, find papers, map a field, or understand the research landscape. Not for explaining concepts — use learn for that.
---

<SUBAGENT-STOP>
If you were dispatched as a subagent to execute a specific search task, skip this skill. Do the task and return structured results immediately.
</SUBAGENT-STOP>

<HARD-GATE>
Do NOT launch any searches until the user has confirmed the topic framing. One sentence of confirmation is enough. No searches before confirmation.
</HARD-GATE>

# Explore Skill — Parallel + Persistent

## Step 0 — Create session

Before anything else, create the session directory and open the lab notebook:

```bash
mkdir -p .research-loop/sessions/<slug>/
```

Where `<slug>` = topic lowercased, spaces to dashes, e.g. `policy-compression-ai-agents`.

Create `.research-loop/sessions/<slug>/lab_notebook.md` with:
```markdown
# Lab Notebook — <topic>
Started: <date>
Status: exploring

---
## Session log
```

This file is the **single source of truth** for the entire research session. Every phase appends to it.

## Step 1 — Confirm the topic

Say:
> "I'll explore **[topic]**. Is this the right framing, or do you want to adjust it first?"

Wait for confirmation. Then append to lab_notebook.md:
```markdown
## Framing
Topic: <confirmed topic>
Date: <date>
Researcher notes: <anything they said about prior knowledge>
```

## Step 2 — Launch parallel searches

Dispatch ALL four researcher subagents simultaneously. Each task references the researcher agent definition (`.claude/agents/researcher.md`) and must follow its integrity commandments (no fabricated sources, URL-or-it-didn't-happen, read-before-summarize), use the numbered evidence table format (ID, Source, URL, Key claim, Type, Confidence), and apply the source quality tiers (A = peer-reviewed papers/official docs, B = reputable secondary, C = accept with caveats).

```
Agent 1 (researcher):
  Task: Find the top 10 most important papers on [topic].
  Use web search. For each paper return: title, year, authors, 1-sentence contribution.
  Prioritize foundational and recent (2020–2025) papers. Return a clean numbered list.
  Use the numbered evidence table format from the researcher agent definition.
  Save findings to .research-loop/sessions/<slug>/findings-papers.md.

Agent 2 (researcher):
  Task: Find active GitHub repos and benchmarks for [topic].
  Use web search. For each repo return: name, URL, stars (if available), 1-sentence description.
  Return 5–10 actively maintained repos. Use cited evidence table format.
  Save findings to .research-loop/sessions/<slug>/findings-repos.md.

Agent 3 (researcher):
  Task: Find 3 places where experts in [topic] fundamentally disagree.
  Use web search. For each debate: name it, state Side A's strongest argument, state Side B's strongest argument, explain why it matters.
  Use evidence table format with numbered sources.
  Save findings to .research-loop/sessions/<slug>/findings-debates.md.

Agent 4 (researcher):
  Task: Find explicit open problems stated in the [topic] literature.
  Use web search. Look for "future work" sections, unsolved challenges, gaps mentioned by field leaders.
  For each gap: state the problem, cite where it appears, explain why it's hard.
  Return 5–8 specific open problems. Use numbered evidence table format.
  Save findings to .research-loop/sessions/<slug>/findings-gaps.md.
```

Tell the user: `Searching papers, repos, debates, and open problems in parallel using researcher agents...`

IMPORTANT: Each dispatched subagent must follow the researcher agent's 6 integrity commandments and output format defined in `.claude/agents/researcher.md`. Key rules: every entry needs a URL, no fabricated sources, mark status honestly, triage before fetching full content, write findings to file progressively.

## Step 3 — Save everything, show synthesis

When all 4 return, **append the full results** to lab_notebook.md:

```markdown
## Literature (Phase 1)
<full list of papers with title, year, contribution>

## Active Repos
<list of repos>

## Field Debates
<3 debates with both sides and strongest arguments>

## Stated Open Problems
<list of gaps found in literature>
```

Then **show the user a synthesis** (not the raw dump):

> "Here's the landscape in 3 sentences: [synthesis]
>
> The 3 most interesting angles I see:
> **A)** [angle] — [why interesting]
> **B)** [angle] — [why interesting]
> **C)** [angle] — [why interesting]
>
> Which direction do you want to go deeper on?"

Wait for response. Append their choice to lab_notebook.md:
```markdown
## Researcher direction choice
Chose: <their answer>
Reasoning they gave: <anything they said>
```

## Step 4 — Extract mental models for chosen direction

Now spawn 1 focused search:
```
Agent: What are the 5 core mental models every expert in [chosen direction] carries?
       Not facts — the intuitions and ways of thinking that take years to develop.
```

Show the models conversationally, one at a time:
> "Here's the first mental model experts in this space share: [model 1 + explanation]. Does this match your intuition?"

Append all 5 to lab_notebook.md:
```markdown
## Mental Models
1. [name]: [description]
2. ...
```

## Step 5 — Generate diagnostic questions

Say:
> "Let me give you 3 questions that would expose whether someone truly understands this vs. memorized it. Try answering them — every wrong answer tells us something."

Show questions one at a time. Wait for answers. For each wrong answer:
> "Here's what you're missing: [explanation]"

Append to lab_notebook.md:
```markdown
## Diagnostic Q&A
Q1: [question]
Researcher answer: [their answer]
Expert answer: [correct answer]
Gap identified: [what they didn't know]
...
```

## Step 6 — Transition

Say:
> "Exploration complete. Lab notebook saved to `.research-loop/sessions/<slug>/lab_notebook.md`
>
> Ready to find the gaps and run the Carlini gate? → `/gap` or just tell me which open problem interests you most."

## Subagent Reference

This skill dispatches the `researcher` subagent (`.claude/agents/researcher.md`) for parallel evidence gathering. The researcher agent definition encodes:
- 6 integrity commandments (no fabrication, URL required, read-before-summarize)
- Numbered evidence table format
- Source quality tiers (A/B/C/Reject)
- Context hygiene rules (progressively write to files, triage before fetching)
- Coverage status tracking per question

## What NOT to do
- Do NOT dump all papers at once
- Do NOT skip saving to lab_notebook.md after each phase
- Do NOT advance phases without user input
- Do NOT dispatch subagents without adhering to researcher integrity commandments

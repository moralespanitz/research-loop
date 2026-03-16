# SKILL: Writing Papers

> How to write a paper that people actually read — with singular focus, a story that lands, self-contained figures, and a conclusion that does not waste its moment.
> Based on Carlini (2026) — "How to win a best paper award."

---

## When to use this skill

Use this skill when transitioning from the execution phase to drafting. Also use it when reviewing a draft in progress.

Invoke with: `/writing-papers` or `@scribe: apply writing skill`

---

## The first rule

**A research paper communicates exactly one idea.**

Write this idea in one sentence before you write anything else. Put it at the top of your draft document. Every section, paragraph, figure, and sentence you write must connect to it. If a sentence does not serve this idea, cut it.

If you cannot express your idea in one sentence, your paper is trying to do too many things. Fix the paper, not the sentence.

---

## Step 1: Know your reader

Pick a specific person as your target reader. Write for them, not for a generic "research community."

A useful default: the six-months-younger version of yourself. What would that person need to hear to be motivated to read this paper? What background do they have? What misconceptions do they bring?

The background section exists to expand who "the reader" can be. After the background, you can assume the reader knows everything you wrote there. Write the background accordingly — not as a block of citations to appease reviewers, but as the minimal set of concepts your ideal reader needs to follow your argument.

Before writing, state explicitly:
- Who is the target reader?
- What do they already believe?
- What do you want them to believe after reading?
- What is the one thing they must take away?

---

## Step 2: Write the abstract last, but get it right

The abstract has two jobs: (1) convey the entirety of the paper in a few sentences, and (2) convince the right reader to keep reading.

A reliable structure:

| Sentence | Content |
|----------|---------|
| 1 | The topic/field you are working in |
| 2 | The specific problem you are solving |
| 3 | Your method or key insight |
| 4 | Your results, or what sentence 3 did not cover |
| 5 | Why this matters |

Alternative structure for broad topics or narrow audiences:
1. Your main claim or result
2. Evidence: method or data that supports it
3. Impact: what this changes

Rules for the abstract:
- Include at least one specific number
- Do not hedge: state the clean true version of your result
- Do not end with "we discuss implications" — state the implication
- Rewrite it four or five times; this is normal

---

## Step 3: Write the introduction as a story

The introduction is the beginning of a story. Its job is to move the reader from where they currently are — their existing beliefs — to the frame of mind where your contribution makes sense.

**Structure:**

1. **Meet the reader where they are.** Open with what they already believe to be true, or with a problem they already accept as real.
2. **Introduce tension.** What is unresolved, broken, or unknown? Why does it matter?
3. **Enter your contribution.** Now — and only now — state what you did and why it resolves the tension.
4. **State your results concretely.** Not "we show X is possible" but "we achieve X with Y on benchmark Z."
5. **Summarize the paper structure** only if the paper is long enough that the reader genuinely needs a roadmap.

**How much story you need depends on how accepted the premise is:**

| Situation | Approach |
|-----------|----------|
| Well-studied problem, clear contribution | One paragraph is enough. "We solve unsolved problem X." |
| Moderately known problem | Brief reminder, then contribution |
| Underexplored area | Extended setup — you must convince the reader the problem matters before they will care about your solution |
| Heretical claim | Do not state the conclusion upfront. Lay evidence first. Let the reader arrive at the conclusion themselves. |

The introduction has at most two pages. Readers have short attention spans. Do not write a novel before getting to the point.

---

## Step 4: Make every figure self-contained

Most readers skim before they commit to reading. The skimming path is: title → abstract → figures. Your figures must work on this path.

Rules:
- Every figure must be interpretable without reading the surrounding text
- Every figure caption must include: (1) what the figure shows, and (2) the one-sentence takeaway
- "Figure 7. Our method performs 3% better than all prior methods on benchmark X" is a good caption
- "Figure 7. Results." is not a caption
- If you cannot explain a figure in a caption, it is too complicated — split it into multiple figures
- If a figure does not have a single-sentence takeaway, it does not belong in the paper

A typical figure arc for an experimental paper:
- Figure 1: The problem
- Figure 2: Your approach or algorithm
- Figures 3–4: Method details
- Figures 5–8: Results and analysis

---

## Step 5: Write a conclusion that earns its moment

The conclusion is not the abstract in the past tense. It is not a summary.

**The conclusion is a moment of reflection.** The reader has just spent an hour in the technical details of your paper. They have surfaced. Your job is to tell them clearly what they should take away from the experience.

Structure:
1. **Brief grounding.** One or two sentences reminding the reader of your core method and setup — they need the anchor
2. **The "so what?"** Answer this question explicitly and directly. Do not imply it. Say it.
3. **The broader implication.** What does this mean for the field? What should people do differently now?
4. **What you deliberately left open.** Invite others to engage. Do not leave gaps accidentally — leave them intentionally as directions for future work.

A test: if your conclusion could be replaced by "our method achieves X% on Y" and nothing would be lost, your conclusion is not doing its job. The "so what" must be more than the result — it must be the meaning of the result.

Write your best-case conclusion before running any experiments (per the `idea-selection` skill). When writing the actual paper, this becomes your target. Hold yourself to it.

---

## Step 6: Make the writing itself not embarrassing

You do not need to be a great writer. You need to be clear enough that the reader can follow your argument without getting distracted by the prose.

Practical rules:
- Read your paper aloud, or use text-to-speech, and listen for sentences that do not land
- Kill sentences with dual meanings, especially when one meaning is the wrong interpretation
- Vary sentence length: long sentences build complexity; short sentences provide relief. Use both.
- Use the active voice by default; use passive voice when appropriate, not by habit
- Use jargon only when the alternative is imprecision — not to sound authoritative
- Cut words aggressively; every sentence should be as short as it can be without losing meaning
- Contractions are fine; do not overdo them
- Proofread, but do not let proofreading displace time better spent on experiments or structure

The reader will forgive occasional typos. They will not forgive an argument they cannot follow.

---

## Step 7: Don't obsess over the title; do care about it

The title must be accurate. It need not be clever. It should let the right readers identify that this paper is for them.

If you are struggling to write a title, it is often a sign the paper is trying to do more than one thing. Fix the paper, then title it.

---

## Writing checklist

- [ ] The core idea is expressed in one sentence at the top of the draft
- [ ] The target reader is explicitly identified
- [ ] The abstract contains at least one specific number and states the result without hedging
- [ ] The introduction guides the reader to the right frame of mind before stating the contribution
- [ ] Every figure has a caption with a one-sentence takeaway and is interpretable without reading the text
- [ ] The conclusion answers "so what?" explicitly and does not repeat the abstract
- [ ] The paper has been read aloud or run through text-to-speech
- [ ] Every section, figure, and paragraph traces back to the one-sentence core idea

---

## Anti-patterns to avoid

- **The everything paper.** Two good ideas in one paper means readers remember neither. Pick one.
- **The introduction-as-literature-review.** Background is for teaching the reader; citations are not teaching.
- **The passive abstract.** "We investigate X" tells no one anything. "We show X achieves Y on Z" does.
- **The summary conclusion.** If your conclusion is just the abstract with different verb tenses, rewrite it.
- **Figures without takeaways.** A figure that makes the reader think "so what?" is a figure that is failing.
- **Writing for all readers.** You cannot. Pick one reader and write for them. Others will adapt.

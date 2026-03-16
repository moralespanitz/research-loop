# SKILL: Review Preparation

> What to do after the paper is written: surviving rejection, positioning for awards, and persisting without losing the thread.
> Based on Carlini (2026) — "How to win a best paper award."

---

## When to use this skill

Use this skill when a paper is complete and being prepared for submission, when a rejection arrives, and when evaluating whether a rejected paper is worth resubmitting or should be killed.

Invoke with: `/review-prep` or `@reviewer: prepare for submission`

---

## The honest truth about awards and acceptance

A best paper award is one sample from a distribution. You do not control the sampling process. What you control is the distribution.

The factors outside your control:
- Whether the specific committee members like your topic area
- Whether your paper arrived too early (premise not yet accepted) or too late (field already moved on)
- Whether someone else published the same paper first
- Whether this year's competition is unusually strong
- Whether the committee is using the award to signal a new direction for the field

Focus on what you can control: the quality, clarity, and timing of the work. Everything else is luck, and luck is not a lever.

---

## Step 1: Timing audit before submission

Before submitting, assess timing on two axes:

**Too early:**
- Does your paper assume premises the community has not yet accepted?
- Are reviewers likely to say "this isn't a real problem yet"?
- If yes: write a stronger introduction that argues *why this will matter*, not just that it matters now. Connect the problem to something real today, even tangentially.

**Too late:**
- Has the field moved on? Is your result now an incremental contribution to an area that was exciting two years ago?
- If yes: either pivot the framing toward what the field cares about today, or accept that this is craft practice rather than high-impact work.

**The sweet spot:** You are 6–18 months ahead of where the community will arrive. The problem is real enough that reviewers can see it, but not so worked-over that your contribution is marginal.

---

## Step 2: Prepare for the skeptical reviewer

Before submitting, list every objection a skeptical reviewer might raise. For each objection, answer:

| Objection | Have you addressed it in the paper? | Should you? |
|-----------|-------------------------------------|-------------|
| "This problem isn't important" | The introduction must answer this | Yes, always |
| "This setup is unrealistic" | The experimental design section | Yes, if the objection is valid |
| "Why not just do X instead?" | Related work or discussion | Yes, if X is obvious |
| "The results don't generalize" | Ablations and scope section | Yes, if true |
| "The evaluation metric is wrong" | Metric justification | Yes, this kills papers |

If there is a valid objection you have not addressed, either run the experiment that answers it, or address it explicitly in the paper with a honest discussion of the limitation.

Do not paper over valid objections. Reviewers notice.

---

## Step 3: Handle rejection without sunk cost

Most papers that eventually receive recognition were rejected at least once first.

The pattern:
1. Paper makes a claim that is slightly outside what reviewers currently believe
2. Reviewers leave confused remarks, not outright rejection of the claim
3. Authors revise to make the argument clearer and stronger — even a skeptical reader should now understand it
4. Resubmit at the next cycle when the idea is slightly less heretical
5. Paper is accepted and sometimes receives an award

The revision between rejection and acceptance is often the most important work on the paper. Use rejection feedback as a diagnostic: if reviewers are confused about X, your explanation of X is not good enough. Fix it.

**When to keep revising and resubmitting:**
- The core idea is sound and important
- Rejection feedback reveals fixable communication problems, not fundamental flaws
- The timing is improving (the community is moving toward accepting the premise)

**When to kill the paper after rejection:**
- The core assumption turned out to be wrong (not just that reviewers don't believe it yet — that it actually is wrong)
- A stronger paper has appeared that makes yours redundant
- Your honest assessment is that the contribution is not compelling enough to be worth more reviewer time

Log the decision in `knowledge_graph.md` with status `rejected-killed` or `rejected-revising` and the reasoning.

---

## Step 4: Self-review against the Reviewer agent criteria

Before submitting, run the Reviewer agent (`@reviewer`) or manually check:

**Methodology:**
- Is the evaluation metric measuring what you claim it measures?
- Are the baselines the right baselines?
- Are the results cherry-picked, or do they hold across seeds, datasets, and settings?
- Is there a confounder that could explain the results without your hypothesis being true?

**Claims:**
- Does every claim in the abstract and introduction have a corresponding experiment?
- Are all numbers stated in the paper accurate?
- Is every "we show" statement actually shown, not implied?

**Reproducibility:**
- Can someone reproduce your main result from what you have written?
- Are hyperparameters, datasets, and compute requirements specified?

**The causal annotation test:** For every experiment that did not work in the way expected, is there a causal explanation in the paper? "We tried X and it did not work" is not a causal annotation. "We tried X; it did not work because Y, which implies Z" is.

---

## Step 5: Position the contribution correctly

Awards are sometimes given for reasons beyond pure technical quality:
- To encourage more research in an underexplored direction
- To correct a misunderstanding in the community
- To recognize work that changed how the field frames a problem

If your paper has done any of these things, make it obvious. Do not hide this in the conclusion. The introduction should signal clearly what the paper is changing, not just what it is adding.

The papers most likely to receive recognition are those that:
1. Do something important
2. Are well enough written that the committee can see it is important
3. Arrive at a moment when the community is ready to receive the idea

You control (1) and (2). Maximize them.

---

## Step 6: Thick skin protocol

Accept that:
- Marginal and boring papers are statistically more likely to be accepted than truly interesting ones
- Truly important papers are more likely to be rejected on the first submission
- This is not a signal to write worse papers

When a rejection arrives:
1. Read the reviews fully once
2. Wait 24 hours before responding or deciding
3. Identify which objections are communication problems vs. fundamental problems
4. For communication problems: fix the paper
5. For fundamental problems: either run the experiment or kill the investigation
6. Decide: revise-and-resubmit, or kill

Do not write angry rebuttals. Do not submit a paper you know is not ready. Do not internalize rejection as a judgment of the idea — it is often a judgment of the timing or the framing.

---

## Review-prep checklist

- [ ] Timing audit complete: the paper is not too early or too late
- [ ] Skeptical reviewer objections listed and addressed in the paper
- [ ] Reviewer agent (`@reviewer`) has audited methodology and claims
- [ ] Every claim in the abstract has a corresponding experiment
- [ ] The introduction signals what the paper is *changing*, not just what it is *adding*
- [ ] A rejection protocol is decided in advance: what conditions would cause you to kill vs. revise

---

## Anti-patterns to avoid

- **Optimizing for acceptance over impact.** These are not the same objective. Papers written for acceptance tend to be mediocre; papers written for impact tend to eventually get accepted.
- **Treating the first rejection as fatal.** Most important papers are rejected at least once. This is normal.
- **Submitting before you're ready.** An early rejection that reveals fundamental problems is better than a late acceptance that reveals fundamental problems after people read the paper. But submitting before the paper is ready wastes reviewer time and your credibility.
- **Ignoring the committee's direction.** Awards are partly political. If the committee is trying to push the field in a direction, and your paper genuinely contributes to that direction, make sure the paper says so clearly.
- **Conflating "I was rejected" with "I was wrong."** Rejection means reviewers at this conference at this time did not accept it. It does not mean the idea is wrong.

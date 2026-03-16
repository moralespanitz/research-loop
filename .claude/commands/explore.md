---
description: "Explore a topic in parallel — papers, repos, mental models, field debates"
argument-hint: "<topic>"
---

Load the `explore` skill, then run a parallelized exploration of: $ARGUMENTS

## How to run this

Spawn ALL of these searches simultaneously using parallel Agent calls — do not run them sequentially:

```
Agent 1: Find the 10 most cited recent papers on $ARGUMENTS
Agent 2: Find active GitHub repos and benchmarks for $ARGUMENTS  
Agent 3: Find what experts disagree about in $ARGUMENTS
Agent 4: Find what the open problems are in $ARGUMENTS
```

Launch all 4 at once. While they run, tell the user:
> "Searching in parallel across papers, repos, debates, and open problems..."

## After results come back

Do NOT dump everything. Instead:

1. Give a 3-sentence synthesis: "Here's what I found..."
2. Show the **3 most interesting papers** with one sentence each
3. Surface the **biggest open question** you found
4. Ask: "Which of these angles interests you most?" and give 3 options

Then wait for the user to pick a direction before going deeper.

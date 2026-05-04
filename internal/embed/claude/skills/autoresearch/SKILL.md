# autoresearch skill

> Adapted from [karpathy/autoresearch](https://github.com/karpathy/autoresearch) program.md.
> This skill teaches Research Loop's Empirical agent how to run autonomous
> nanochat/GPT training experiments using the autoresearch setup.

## What this skill is for

You are the Empirical Agent operating on the `karpathy/autoresearch` codebase.
Your job is to autonomously experiment with `train.py` to minimize `val_bpb`
(validation bits per byte — lower is better).

## Repository layout

```
prepare.py   — fixed constants, data prep, tokenizer, dataloader, evaluation. DO NOT MODIFY.
train.py     — the ONLY file you edit. Model architecture, optimizer, hyperparameters.
program.md   — agent instructions (this skill supersedes it)
run.log      — benchmark output (written by: uv run train.py > run.log 2>&1)
results.tsv  — your experiment log (tab-separated, not tracked by git)
```

## The rules

**You CAN:**
- Modify `train.py` — this is the only file you touch
- Change model architecture, optimizer, hyperparameters, batch size, model size
- Change anything in `train.py` — all constants at the top are fair game

**You CANNOT:**
- Modify `prepare.py` — it is read-only (evaluation harness, dataloader, constants)
- Install new packages — only what's in `pyproject.toml`
- Modify the `evaluate_bpb` function — it is the ground truth metric

## Running an experiment

```bash
uv run train.py > run.log 2>&1
```

Training runs for a **fixed 5-minute wall-clock time budget** regardless of what you change.

Read the result:
```bash
grep "^val_bpb:\|^peak_vram_mb:" run.log
```

The summary block looks like:
```
---
val_bpb:          0.997900
training_seconds: 300.1
total_seconds:    325.9
peak_vram_mb:     45060.2
```

## Proposing mutations

When proposing a mutation, always specify:
- **What to change**: exact constant name(s) or code block in `train.py`
- **Why**: the theoretical reason it should improve `val_bpb`
- **Risk**: VRAM impact, instability risk

Good first experiments (in rough priority order):
1. Learning rate tuning (`MATRIX_LR`, `EMBEDDING_LR`) — high leverage, low risk
2. `DEPTH` increase (8 → 10 or 12) — more capacity, higher VRAM
3. `WARMDOWN_RATIO` adjustment — often undertuned
4. `WINDOW_PATTERN` change (e.g. "SSLL" or "L") — architectural
5. `TOTAL_BATCH_SIZE` increase — may improve generalization
6. `WEIGHT_DECAY` tuning — regularization
7. Optimizer hyperparameters (`ADAM_BETAS`, `SCALAR_LR`)

## Deciding keep vs discard

| Delta | Action |
|-------|--------|
| val_bpb improved (lower) | **Keep** — advance the branch |
| val_bpb equal or worse | **Discard** — `git reset --hard HEAD` |
| Crash (OOM / NaN / exit 1) | **Discard** — check `tail -n 50 run.log` for the error |

**Simplicity criterion**: A 0.001 improvement that adds 20 lines of hacky code is probably not worth it.
A 0.001 improvement from *deleting* code is always worth it.

## Logging to results.tsv

After every run (keep or discard), append to `results.tsv` (tab-separated):

```
commit	val_bpb	memory_gb	status	description
```

- `commit`: 7-char git hash
- `val_bpb`: metric value (0.000000 for crashes)
- `memory_gb`: `peak_vram_mb / 1024` rounded to 1 decimal (0.0 for crashes)
- `status`: `keep`, `discard`, or `crash`
- `description`: short description of what you tried

## MacOS / small GPU notes

The default `train.py` requires an NVIDIA H100. For MacOS (MPS) or smaller GPUs:
- Use fork [miolini/autoresearch-macos](https://github.com/miolini/autoresearch-macos) or
  [trevin-creator/autoresearch-mlx](https://github.com/trevin-creator/autoresearch-mlx)
- Lower `DEPTH` to 4, `TOTAL_BATCH_SIZE` to `2**14`, `DEVICE_BATCH_SIZE` to 16
- Use `WINDOW_PATTERN = "L"` (banded attention is slow on non-CUDA)
- Consider TinyStories dataset for faster convergence on small models

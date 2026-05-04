---
name: figure-agent
description: >
  Publication-quality figure generation for research papers. Decision agent
  selects figure type (code plot vs architecture diagram). Generates
  Matplotlib/Seaborn code for quantitative figures with iterative improvement
  loop. Style-matches conference templates (NeurIPS, ICML, ICLR). Use when
  the paper-pipeline reaches the figure generation phase, or when a user
  requests figures for an existing draft.
---

<SUBAGENT-STOP>
If you were dispatched as a subagent to generate a specific figure from data
or create a specific diagram, skip this skill. Do the task and return
structured results immediately.
</SUBAGENT-STOP>

<HARD-GATE>
NEVER fabricate figure data. Every figure must be backed by real experimental
results or a traceable data source. If no data is provided, ask for it before
generating anything.
</HARD-GATE>

# Figure Agent Skill

> Decision-driven figure generation for conference papers. Analyzes the paper
> content to determine figure types, generates Matplotlib/Seaborn code for
> quantitative figures, and iterates on quality via critic feedback.

## Figure Type Decision Agent

When generating figures for a paper, first analyze the content to determine
the appropriate figure type:

### Quantitative Figures (Code-generated)
Use when the figure shows actual experimental results, metrics, or data:

| Figure Type | Best For | Matplotlib Function |
|-------------|----------|-------------------|
| Bar chart | Comparing discrete conditions, ablations | `plt.bar()` |
| Line plot | Convergence curves, training trajectories | `plt.plot()` |
| Scatter plot | Correlation analysis, parameter sensitivity | `plt.scatter()` |
| Heatmap | Ablation matrices, hyperparameter grids | `plt.imshow()` or `sns.heatmap()` |
| Box plot / Violin | Distribution of results across seeds | `plt.boxplot()` or `sns.violinplot()` |
| Histogram | Result distributions | `plt.hist()` |
| Stacked bar | Comparison across multiple metrics | `plt.bar(stacked=True)` |
| Grouped bar | Multiple methods across conditions | `plt.bar()` with offset positions |

### Architecture/Conceptual Figures (Diagram)
Use when the figure shows the proposed method, system architecture, or
conceptual framework. These should be drawn with Excalidraw or a similar
diagramming tool. For code-generated alternatives, use Matplotlib patches.

### Selection Criteria

```
Is the figure based on experimental data?
├── YES → Is it a comparison (method A vs B) or relationship (parameter vs metric)?
│   ├── Bar/Line/Scatter/Box → Generate Matplotlib code
│   └── Architecture/Flowchart → Generate diagram code
└── NO → It must be conceptual or illustrative
    └── Architecture/Flowchart → Generate diagram code
```

## Matplotlib/Seaborn Code Generation

### Template

```python
import matplotlib.pyplot as plt
import matplotlib as mpl
import numpy as np
import json

# === LOAD DATA ===
with open("results.json") as f:
    data = json.load(f)

# === STYLE SETUP ===
plt.rcParams.update({
    "figure.figsize": (5.0, 3.5),  # NeurIPS column width
    "font.family": "serif",
    "font.size": 10,
    "axes.titlesize": 11,
    "axes.labelsize": 10,
    "xtick.labelsize": 9,
    "ytick.labelsize": 9,
    "legend.fontsize": 9,
    "lines.linewidth": 1.5,
    "axes.linewidth": 0.8,
    "grid.alpha": 0.3,
    "savefig.dpi": 300,
    "savefig.bbox": "tight",
})

# === PLOT COLOURS (Color-blind friendly) ===
CB_COLORS = ["#0072B2", "#E69F00", "#009E73", "#CC79A7", "#56B4E9", "#F0E442"]

# === FIGURE ===
fig, ax = plt.subplots(1, 1)

# ... plotting code ...

# === STYLE ===
ax.set_xlabel("X Label")
ax.set_ylabel("Y Label")
ax.set_title("Title")
ax.legend()
ax.grid(True, alpha=0.3)

# === SAVE ===
plt.tight_layout()
plt.savefig("figure_1.pdf", dpi=300)
plt.savefig("figure_1.png", dpi=300)
plt.close()
print("FIGURE_SAVED: figure_1.pdf")
```

### Conference Style Presets

#### NeurIPS (default)
```python
plt.rcParams.update({
    "figure.figsize": (5.0, 3.5),   # Single column
    "font.family": "serif",
    "font.size": 10,
})
```

#### ICML
```python
plt.rcParams.update({
    "figure.figsize": (5.5, 3.5),   # Slightly wider
    "font.family": "serif",
    "font.size": 9,
})
```

#### ICLR
```python
plt.rcParams.update({
    "figure.figsize": (5.0, 3.5),
    "font.family": "serif",
    "font.size": 10,
})
```

### Color Schemes

Use color-blind friendly palettes by default:
```python
# Qualitative (for categories)
CB_COLORS = ["#0072B2", "#E69F00", "#009E73", "#CC79A7",
             "#56B4E9", "#F0E442", "#D55E00", "#000000"]

# Sequential (for heatmaps)
CB_SEQUENTIAL = ["#FFFFFF", "#F0E442", "#D55E00", "#000000"]

# Diverging (for difference plots)
CB_DIVERGING = ["#0072B2", "#FFFFFF", "#D55E00"]
```

## Iterative Improvement Loop

For each figure, run a 3-step improvement loop:

### Step 1 — Generate
Generate the initial figure code from the data and figure specification.

### Step 2 — Critic
Analyze the generated figure:
- Is the data visualization accurate? No misleading axis scales?
- Is the figure readable at conference page width (~5 inches)?
- Are labels clear and font sizes appropriate?
- Is the color scheme accessible (color-blind friendly)?
- Is the caption informative (takeaway in one sentence)?
- Does the figure serve the paper's core claim?

### Step 3 — Improve
Apply critic feedback. Iterate up to 3 times or until the critic judges
the figure as "publishable."

## Figure Specification Format

Each figure must be specified with:

```yaml
figure:
  id: 1
  type: "bar"           # bar | line | scatter | heatmap | box | histogram
  caption: "Comparison of convergence rates across optimizers. SGD-Adam achieves 15% faster convergence than baselines."
  data_source: "results.json"  # or inline data
  x: "optimizer"        # column or variable name
  y: "convergence_time" # column or variable name
  hue: "dataset"        # optional grouping variable
  style: "NeurIPS"      # NeurIPS | ICML | ICLR
  width: 5.0            # inches
  height: 3.5
  color_scheme: "colorblind"
```

## Output Structure

Figures are saved to `sessions/<slug>/figures/`:

```
figures/
├── figure_1.pdf      # Vector format for paper
├── figure_1.png      # Raster format for preview
├── figure_1_caption.md  # Caption text
├── figure_1_code.py  # Regeneratable code
├── figure_2.pdf
├── figure_2.png
├── figure_2_caption.md
└── figure_2_code.py
```

## Anti-patterns

- **Jupyter notebook output as a figure.** Conference papers need vector
  graphics (.pdf). Never submit screenshots or Jupyter inline plots.
- **Rainbow color schemes.** Use color-blind friendly palettes. Never use
  jet/rainbow colormaps for quantitative data.
- **3D plots for 2D data.** 3D plots rarely add value and often obscure the
  data. Use 2D alternatives unless 3D is essential.
- **Figures without captions.** Every figure must have a caption with a
  takeaway. If there is no takeaway, reconsider whether the figure belongs.
- **Axis manipulation.** Never truncate the y-axis to exaggerate differences.
  Always start at 0 for bar charts unless there is a principled reason.

## Verification Checklist

- [ ] Figure type matches the data type (quantitative vs conceptual)
- [ ] All data values are real (not fabricated)
- [ ] Color-blind friendly palette is used
- [ ] Font sizes are readable at conference page width
- [ ] Caption includes a one-sentence takeaway
- [ ] Figure is saved as PDF (vector) and PNG (raster)
- [ ] Code is saved for regeneration
- [ ] Critic judged the figure as "publishable"
- [ ] Axis scales are not misleading

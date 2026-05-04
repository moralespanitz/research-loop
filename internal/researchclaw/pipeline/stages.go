// Package pipeline defines the stages and lifecycle of the Research Loop
// paper generation pipeline. Each stage represents a discrete phase of
// the end-to-end research workflow — from topic initiation through to
// formatted publication export.
//
// The pipeline is a 14-stage state machine ported from AutoResearchClaw
// (Aiming Lab). Stages have defined inputs, outputs, validation gates,
// and rollback rules. Gate stages (Literature Collection, Hypothesis
// Generation, Experiment Design, Paper Draft, Revision) require human
// approval before proceeding.
//
// File layout (future):
//   stages.go      — PipelineStage enum, lifecycle helpers
//   contracts.go   — StageContract definitions (inputs, outputs, gate flags)
//   executor.go    — One executor function per stage
//   runner.go      — Pipeline runner (sequencing, state, error handling)
package pipeline

// PipelineStage enumerates the 14 phases of the research paper pipeline.
// Values map to the phases defined in skills/paper-pipeline/SKILL.md.
// The lifecycle order matches the numeric sequence:
//
//	Phase A (1-2):   Topic scoping and literature discovery
//	Phase B (3-6):   Knowledge synthesis and hypothesis formation
//	Phase C (7-10):  Experiment design, execution, and analysis
//	Phase D (11-14): Paper drafting, review, revision, and export
type PipelineStage int

const (
	// StageTopicInit defines the research question, scope, SMART goal,
	// success criteria, and constraints. Produces sessions/<slug>/01-topic-init.md.
	StageTopicInit PipelineStage = iota + 1

	// StageLiteratureCollect gathers candidate papers from arxiv, Semantic
	// Scholar, and web sources. Requires >= 20 candidates from >= 3 sources.
	// GATE stage — requires human approval.
	StageLiteratureCollect

	// StageLiteratureScreen filters candidates by relevance + quality scoring.
	// Produces a shortlist of 8-15 papers with scores and keep reasons.
	StageLiteratureScreen

	// StageKnowledgeExtract produces structured evidence cards for each
	// shortlisted paper — problem, method, metrics, findings, limitations.
	StageKnowledgeExtract

	// StageHypothesisGen generates >= 2 falsifiable predictions from the
	// evidence base. Each hypothesis includes rationale, measurable prediction,
	// and failure condition. GATE stage — requires human approval.
	StageHypothesisGen

	// StageSynthesis organizes evidence into topic clusters and identifies
	// research gaps. Produces cluster overview + prioritized gap list.
	StageSynthesis

	// StageExperimentDesign defines experimental conditions, baselines,
	// ablations, metrics, and compute budget. GATE stage — requires approval.
	StageExperimentDesign

	// StageCodeGen generates executable Python experiment code implementing
	// real algorithms with validation, security scanning, and auto-repair.
	StageCodeGen

	// StageExecution runs experiments in a sandboxed environment (local venv,
	// Docker, SSH remote, or Colab). Collects metrics and results.json.
	StageExecution

	// StageResultAnalysis produces statistical interpretation of experimental
	// results — summary stats, comparisons, ablation analysis.
	StageResultAnalysis

	// StagePaperDraft writes a full-length conference-quality paper draft
	// (NeurIPS/ICML/ICLR) with data-driven metrics. GATE stage.
	StagePaperDraft

	// StagePeerReview simulates peer review from >= 2 reviewer perspectives,
	// scoring 1-10 with actionable revision requests.
	StagePeerReview

	// StageRevision addresses all reviewer feedback while maintaining or
	// increasing section word counts. GATE stage — requires human approval.
	StageRevision

	// StageExport formats the final paper for the target venue template
	// and generates publication-quality figures via the figure-agent.
	StageExport
)

// String returns the human-readable name of the stage.
func (s PipelineStage) String() string {
	switch s {
	case StageTopicInit:
		return "TopicInit"
	case StageLiteratureCollect:
		return "LiteratureCollect"
	case StageLiteratureScreen:
		return "LiteratureScreen"
	case StageKnowledgeExtract:
		return "KnowledgeExtract"
	case StageHypothesisGen:
		return "HypothesisGen"
	case StageSynthesis:
		return "Synthesis"
	case StageExperimentDesign:
		return "ExperimentDesign"
	case StageCodeGen:
		return "CodeGen"
	case StageExecution:
		return "Execution"
	case StageResultAnalysis:
		return "ResultAnalysis"
	case StagePaperDraft:
		return "PaperDraft"
	case StagePeerReview:
		return "PeerReview"
	case StageRevision:
		return "Revision"
	case StageExport:
		return "Export"
	default:
		return "Unknown"
	}
}

// IsGate returns true if the stage requires human-in-the-loop approval.
// Gate stages block pipeline progression until explicitly approved.
func (s PipelineStage) IsGate() bool {
	switch s {
	case StageLiteratureCollect, StageHypothesisGen,
		StageExperimentDesign, StagePaperDraft, StageRevision:
		return true
	default:
		return false
	}
}

// Phase returns the high-level phase group this stage belongs to.
func (s PipelineStage) Phase() string {
	switch {
	case s >= StageTopicInit && s <= StageLiteratureCollect:
		return "A: Topic & Scope"
	case s >= StageLiteratureScreen && s <= StageSynthesis:
		return "B: Knowledge & Synthesis"
	case s >= StageExperimentDesign && s <= StageResultAnalysis:
		return "C: Experiment"
	case s >= StagePaperDraft && s <= StageExport:
		return "D: Paper & Review"
	default:
		return "Unknown"
	}
}

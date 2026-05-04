// Package paper provides publication venue definitions, template resolution,
// formatting rules, and export utilities for Research Loop's paper pipeline.
//
// Supported venues include the three major machine learning conferences:
//   - NeurIPS (Neural Information Processing Systems)
//   - ICML   (International Conference on Machine Learning)
//   - ICLR   (International Conference on Learning Representations)
//
// Each venue has page limits, formatting requirements, section structure,
// and submission checklist items. The package resolves template paths and
// validates papers against venue-specific requirements.
//
// File layout (future):
//   templates.go  — Venue enum, template path resolution, formatting rules
//   export.go     — Paper export to LaTeX/PDF/markdown with venue template
//   validate.go   — Venue-specific validation (page count, sections, format)
package paper

// Venue enumerates supported publication venues for paper submission.
type Venue int

const (
	// VenueNeurIPS is the Conference on Neural Information Processing Systems.
	// Page limit: 9 pages + references. Double-blind. Requires checklist.
	VenueNeurIPS Venue = iota + 1

	// VenueICML is the International Conference on Machine Learning.
	// Page limit: 8 pages + references. Double-blind. Requires broader impact.
	VenueICML

	// VenueICLR is the International Conference on Learning Representations.
	// Page limit: 9 pages + references. Double-blind. Requires checklist.
	VenueICLR
)

// String returns the three-letter venue abbreviation.
func (v Venue) String() string {
	switch v {
	case VenueNeurIPS:
		return "NeurIPS"
	case VenueICML:
		return "ICML"
	case VenueICLR:
		return "ICLR"
	default:
		return "Unknown"
	}
}

// FullName returns the full conference name for the venue.
func (v Venue) FullName() string {
	switch v {
	case VenueNeurIPS:
		return "Conference on Neural Information Processing Systems"
	case VenueICML:
		return "International Conference on Machine Learning"
	case VenueICLR:
		return "International Conference on Learning Representations"
	default:
		return ""
	}
}

// VenueFromString parses a venue from its string representation.
// Returns false if the string does not match a known venue.
func VenueFromString(s string) (Venue, bool) {
	switch s {
	case "NeurIPS", "neurips", "nips":
		return VenueNeurIPS, true
	case "ICML", "icml":
		return VenueICML, true
	case "ICLR", "iclr":
		return VenueICLR, true
	default:
		return 0, false
	}
}

// TemplateInfo holds formatting guidelines and requirements for a venue.
type TemplateInfo struct {
	// Venue is the target publication venue.
	Venue Venue `json:"venue" yaml:"venue"`

	// PageLimit is the maximum number of pages for the main content
	// (excluding references and appendix).
	PageLimit int `json:"page_limit" yaml:"page_limit"`

	// FontSize is the recommended font size in points.
	FontSize int `json:"font_size" yaml:"font_size"`

	// FontFamily is the recommended font (e.g., "Times", "serif").
	FontFamily string `json:"font_family" yaml:"font_family"`

	// DoubleBlind indicates whether the venue requires anonymous submission.
	DoubleBlind bool `json:"double_blind" yaml:"double_blind"`

	// RequiredSections lists sections that must be present in the paper.
	RequiredSections []string `json:"required_sections" yaml:"required_sections"`

	// RequiresChecklist indicates whether a submission checklist is required.
	RequiresChecklist bool `json:"requires_checklist" yaml:"requires_checklist"`

	// TemplateURL is the official template/style file URL for this venue.
	TemplateURL string `json:"template_url" yaml:"template_url"`
}

// Template returns the formatting guidelines for this venue.
func (v Venue) Template() TemplateInfo {
	switch v {
	case VenueNeurIPS:
		return TemplateInfo{
			Venue:             v,
			PageLimit:         9,
			FontSize:          10,
			FontFamily:        "serif",
			DoubleBlind:       true,
			RequiredSections:  []string{"Abstract", "Introduction", "Related Work", "Method", "Experiments", "Discussion", "Conclusion"},
			RequiresChecklist: true,
			TemplateURL:       "https://nips.cc/Conferences/2025/PaperInformation/StyleFiles",
		}
	case VenueICML:
		return TemplateInfo{
			Venue:             v,
			PageLimit:         8,
			FontSize:          10,
			FontFamily:        "serif",
			DoubleBlind:       true,
			RequiredSections:  []string{"Abstract", "Introduction", "Background", "Method", "Experiments", "Related Work", "Conclusion", "Broader Impact"},
			RequiresChecklist: false,
			TemplateURL:       "https://icml.cc/Conferences/2025/StyleFiles",
		}
	case VenueICLR:
		return TemplateInfo{
			Venue:             v,
			PageLimit:         9,
			FontSize:          10,
			FontFamily:        "serif",
			DoubleBlind:       true,
			RequiredSections:  []string{"Abstract", "Introduction", "Related Work", "Preliminaries", "Method", "Experiments", "Conclusion"},
			RequiresChecklist: true,
			TemplateURL:       "https://iclr.cc/Conferences/2025/StyleFiles",
		}
	default:
		return TemplateInfo{}
	}
}

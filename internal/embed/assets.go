// Package embed holds the distributable Claude Code integration assets.
// These files are installed into a user's project by `research-loop skills`.
package embed

import "embed"

// Assets contains the .claude/ directory template.
// The session-start.sh hook references getting-started (not research-loop) so it
// works in fresh user projects that have never run `research-loop start`.
//
//go:embed all:claude
var Assets embed.FS

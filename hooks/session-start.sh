#!/usr/bin/env bash
# Research Loop — SessionStart hook
# Injects research context into every Claude Code session opened in this workspace.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]:-$0}")" && pwd)"
PLUGIN_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Read the entry-point skill
SKILL_CONTENT=$(cat "${PLUGIN_ROOT}/skills/research-loop/SKILL.md" 2>/dev/null || echo "")

# Detect workspace state
WORKSPACE="${CLAUDE_PROJECT_DIR:-$(pwd)}"
STATE_DIR="${WORKSPACE}/.research-loop"

state_summary=""
if [ -d "${STATE_DIR}/sessions" ]; then
  SESSION_COUNT=$(ls -1 "${STATE_DIR}/sessions" 2>/dev/null | wc -l | tr -d ' ')
  state_summary="Active sessions: ${SESSION_COUNT}"
fi
if [ -d "${STATE_DIR}/explorations" ]; then
  EXPLORE_COUNT=$(ls -1 "${STATE_DIR}/explorations" 2>/dev/null | wc -l | tr -d ' ')
  state_summary="${state_summary} | Explorations: ${EXPLORE_COUNT}"
fi
if [ -d "${STATE_DIR}/discoveries" ]; then
  DISC_COUNT=$(ls -1 "${STATE_DIR}/discoveries" 2>/dev/null | wc -l | tr -d ' ')
  state_summary="${state_summary} | Discoveries: ${DISC_COUNT}"
fi

escape_for_json() {
  local s="$1"
  s="${s//\\/\\\\}"
  s="${s//\"/\\\"}"
  s="${s//$'\n'/\\n}"
  s="${s//$'\r'/\\r}"
  s="${s//$'\t'/\\t}"
  printf '%s' "$s"
}

SKILL_ESCAPED=$(escape_for_json "$SKILL_CONTENT")
STATE_ESCAPED=$(escape_for_json "$state_summary")

CONTEXT="<EXTREMELY_IMPORTANT>\nYou have Research Loop superpowers.\n\n${SKILL_ESCAPED}\n\nWorkspace state: ${STATE_ESCAPED}\n\nSlash commands available: /explore, /discover, /loop, /research\n</EXTREMELY_IMPORTANT>"

cat <<EOF
{
  "hookSpecificOutput": {
    "hookEventName": "SessionStart",
    "additionalContext": "${CONTEXT}"
  }
}
EOF

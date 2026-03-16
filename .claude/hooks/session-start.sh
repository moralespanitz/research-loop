#!/usr/bin/env bash
# Research Loop — SessionStart hook
# Injects skill system + active session briefing into every Claude Code session.

set -euo pipefail

WORKSPACE_DIR="${CLAUDE_PROJECT_DIR:-$(pwd)}"
STATE_DIR="${WORKSPACE_DIR}/.research-loop"

# Read entry-point skill
SKILL_CONTENT=$(cat "${WORKSPACE_DIR}/.claude/skills/research-loop/SKILL.md" 2>/dev/null || echo "")

# Build session context — prefer SESSION.md (full briefing) over summary
session_context=""
if [ -d "${STATE_DIR}/sessions" ]; then
  LATEST=$(ls -t "${STATE_DIR}/sessions" 2>/dev/null | head -1)
  if [ -n "$LATEST" ]; then
    SESSION_FILE="${STATE_DIR}/sessions/${LATEST}/SESSION.md"
    NOTEBOOK="${STATE_DIR}/sessions/${LATEST}/lab_notebook.md"
    if [ -f "$SESSION_FILE" ]; then
      # Full briefing available — inject it directly
      SESSION_CONTENT=$(cat "$SESSION_FILE" 2>/dev/null || echo "")
      session_context="RESUMING SESSION\n\n${SESSION_CONTENT}\n\nSay 'show me the status' for the full decision tree, or continue from where you left off."
    elif [ -f "$NOTEBOOK" ]; then
      # No SESSION.md yet — fall back to summary
      TOPIC=$(grep "^# Lab Notebook" "$NOTEBOOK" 2>/dev/null | sed 's/# Lab Notebook — //' || echo "$LATEST")
      GAP_COUNT=$(grep -c "^GAP " "$NOTEBOOK" 2>/dev/null || echo "0")
      HYPOTHESIS_COUNT=$(grep -c "^### IDEA " "$NOTEBOOK" 2>/dev/null || echo "0")
      LANE_SELECTED=$(grep "^Lane: " "$NOTEBOOK" 2>/dev/null | tail -1 | sed 's/Lane: //' || echo "none")
      session_context="ACTIVE SESSION: ${TOPIC}\nGaps: ${GAP_COUNT} | Hypotheses: ${HYPOTHESIS_COUNT} | Lane: ${LANE_SELECTED}\nSay 'show me the status' to generate your session briefing (SESSION.md)."
    else
      session_context="ACTIVE SESSION: ${LATEST} — no notebook yet. Say what you want to research."
    fi
  fi
fi

if [ -z "$session_context" ]; then
  session_context="No active sessions. Just tell me what you want to research or understand."
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
SESSION_ESCAPED=$(escape_for_json "$session_context")

CONTEXT="<EXTREMELY_IMPORTANT>\nYou have Research Loop superpowers.\n\n${SKILL_ESCAPED}\n\n--- SESSION BRIEFING ---\n${SESSION_ESCAPED}\n\nSkills: research-loop, status, learn, explore, idea-selection, discover, plan, loop, execution\n</EXTREMELY_IMPORTANT>"

cat <<EOF
{
  "hookSpecificOutput": {
    "hookEventName": "SessionStart",
    "additionalContext": "${CONTEXT}"
  }
}
EOF

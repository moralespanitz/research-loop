#!/usr/bin/env bash
# Research Loop — SessionStart hook
# Injects research superpowers + active session state into every Claude Code session.

set -euo pipefail

WORKSPACE_DIR="${CLAUDE_PROJECT_DIR:-$(pwd)}"
STATE_DIR="${WORKSPACE_DIR}/.research-loop"

# Read entry-point skill
SKILL_CONTENT=$(cat "${WORKSPACE_DIR}/.claude/skills/research-loop/SKILL.md" 2>/dev/null || echo "")

# Build session context
session_context=""
if [ -d "${STATE_DIR}/sessions" ]; then
  LATEST=$(ls -t "${STATE_DIR}/sessions" 2>/dev/null | head -1)
  if [ -n "$LATEST" ]; then
    NOTEBOOK="${STATE_DIR}/sessions/${LATEST}/lab_notebook.md"
    if [ -f "$NOTEBOOK" ]; then
      # Get last status line and run count
      LAST_STATUS=$(grep "^## Status" -A2 "$NOTEBOOK" 2>/dev/null | tail -1 | sed 's/^[[:space:]]*//' || echo "")
      RUN_COUNT=$(grep -c "^## Run #" "$NOTEBOOK" 2>/dev/null || echo "0")
      TOPIC=$(grep "^# Lab Notebook" "$NOTEBOOK" 2>/dev/null | sed 's/# Lab Notebook — //' || echo "$LATEST")
      GAP_COUNT=$(grep -c "^GAP " "$NOTEBOOK" 2>/dev/null || echo "0")
      HYPOTHESIS_COUNT=$(grep -c "^### IDEA " "$NOTEBOOK" 2>/dev/null || echo "0")
      LANE_SELECTED=$(grep "^Lane: " "$NOTEBOOK" 2>/dev/null | tail -1 | sed 's/Lane: //' || echo "none")
      INSIGHTS_FILE="${STATE_DIR}/sessions/${LATEST}/insights.md"
      INSIGHT_COUNT=$(grep -c "^## Insight" "$INSIGHTS_FILE" 2>/dev/null || echo "0")
      session_context="ACTIVE SESSION: ${TOPIC}\nGaps: ${GAP_COUNT} | Hypotheses: ${HYPOTHESIS_COUNT} | Lane: ${LANE_SELECTED} | Experiments run: ${RUN_COUNT} | Insights logged: ${INSIGHT_COUNT}\nSay 'show me the status' to see the full decision tree."
    else
      session_context="ACTIVE SESSION: ${LATEST} (no lab notebook yet)"
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

CONTEXT="<EXTREMELY_IMPORTANT>\nYou have Research Loop superpowers.\n\n${SKILL_ESCAPED}\n\n--- WORKSPACE STATE ---\n${SESSION_ESCAPED}\n\nSkills available: research-loop, status, learn, explore, idea-selection, discover, plan, loop, execution\nJust tell me what you want to research or understand. Say 'show me the status' to see the decision tree.\n</EXTREMELY_IMPORTANT>"

cat <<EOF
{
  "hookSpecificOutput": {
    "hookEventName": "SessionStart",
    "additionalContext": "${CONTEXT}"
  }
}
EOF

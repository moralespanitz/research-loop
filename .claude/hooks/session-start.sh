#!/bin/bash
# Research Loop — SessionStart hook
# Injects active session context into every Claude Code session in this workspace.
# Claude Code reads stdout from this hook and adds it to the conversation context.

WORKSPACE_DIR="${CLAUDE_PROJECT_DIR:-$(pwd)}"
SESSIONS_DIR="$WORKSPACE_DIR/.research-loop/sessions"

if [ ! -d "$SESSIONS_DIR" ]; then
  exit 0
fi

# Find the most recent session
LATEST_SESSION=$(ls -t "$SESSIONS_DIR" 2>/dev/null | head -1)
if [ -z "$LATEST_SESSION" ]; then
  exit 0
fi

SESSION_PATH="$SESSIONS_DIR/$LATEST_SESSION"
HYPOTHESIS_FILE="$SESSION_PATH/hypothesis.md"
KG_FILE="$SESSION_PATH/knowledge_graph.md"

if [ ! -f "$HYPOTHESIS_FILE" ]; then
  exit 0
fi

# Read the core claim from hypothesis.md
CORE_CLAIM=$(grep -A2 "## Core Claim" "$HYPOTHESIS_FILE" 2>/dev/null | tail -1 | xargs)

# Count experiments run
RUNS=0
if [ -f "$SESSION_PATH/autoresearch.jsonl" ]; then
  RUNS=$(wc -l < "$SESSION_PATH/autoresearch.jsonl" 2>/dev/null || echo 0)
fi

# Emit context for Claude — this is added to Claude's context window
cat <<EOF
[Research Loop context]
Active session: $LATEST_SESSION
Core claim: $CORE_CLAIM
Experiments run: $RUNS
Session files:
  hypothesis.md      → $HYPOTHESIS_FILE
  knowledge_graph.md → $KG_FILE
  lab_notebook.md    → $SESSION_PATH/lab_notebook.md
  autoresearch.jsonl → $SESSION_PATH/autoresearch.jsonl

Research Loop MCP tools available: research_ingest_paper, research_kg_query, research_update_kg, research_session_status
Dashboard: http://localhost:4321
EOF

exit 0

#!/bin/bash
# Research Loop — PostToolUse hook (async, Write|Edit)
# Pings the dashboard API after any file edit so the UI refreshes.
# Runs asynchronously — does not block Claude Code.

WORKSPACE_DIR="${CLAUDE_PROJECT_DIR:-$(pwd)}"
DASHBOARD_PORT="${RESEARCH_LOOP_PORT:-4321}"

# Only ping if dashboard is running
curl -sf --max-time 1 "http://localhost:$DASHBOARD_PORT/api/health" > /dev/null 2>&1 || exit 0

# Extract edited file from hook input (stdin is JSON)
EDITED_FILE=$(cat /dev/stdin 2>/dev/null | python3 -c "
import json, sys
try:
  d = json.load(sys.stdin)
  p = d.get('tool_input', {})
  print(p.get('file_path') or p.get('path') or '')
except:
  pass
" 2>/dev/null)

# Log to audit trail if we got a file path
if [ -n "$EDITED_FILE" ]; then
  SESSION_DIR=$(ls -td "$WORKSPACE_DIR/.research-loop/sessions/"*/ 2>/dev/null | head -1)
  if [ -n "$SESSION_DIR" ]; then
    JSONL_FILE="$SESSION_DIR/autoresearch.jsonl"
    echo "{\"event\":\"file_edited\",\"file\":\"$EDITED_FILE\",\"timestamp\":\"$(date -u +%Y-%m-%dT%H:%M:%SZ)\"}" >> "$JSONL_FILE" 2>/dev/null
  fi
fi

exit 0

#!/usr/bin/env bash
# セッション起動ルーチン
set -euo pipefail

START_DEV=false
SKIP_CHECKS=false

for arg in "$@"; do
  case "$arg" in
    --dev) START_DEV=true ;;
    --skip-checks) SKIP_CHECKS=true ;;
  esac
done

echo "=== Session Startup ==="

[ -d ".git" ] || { echo "ERROR: Not in git repository"; exit 1; }

echo "=== Recent commits ==="
git log --oneline -10

echo "=== Current context ==="
if [ -f "CONTEXT.json" ]; then
  echo "Stage: $(jq -r '.current_stage // "unknown"' CONTEXT.json)"
  echo "Goal: $(jq -r '.goal // "not set"' CONTEXT.json)"
  echo "Next action: $(jq -r '.next_best_action // "not set"' CONTEXT.json)"
else
  echo "WARN: CONTEXT.json not found. Creating from template..."
fi

echo "=== Session history ==="
if [ -f "progress.json" ]; then
  TOTAL_SESSIONS=$(jq '.sessions | length' progress.json)
  echo "Total sessions: $TOTAL_SESSIONS"
  if [ "$TOTAL_SESSIONS" -gt 0 ]; then
    echo "Last session:"
    jq '.sessions[-1] | {ended_at, stage, summary, checks, next_action}' progress.json
  fi
else
  echo "INFO: progress.json not found. This appears to be the first session."
fi

if [ "$SKIP_CHECKS" = true ]; then
  echo "=== Health check SKIPPED (--skip-checks) ==="
else
  echo "=== Health check ==="
  if make check 2>&1 | tail -10; then
    echo "All checks passed. Ready to work."
  else
    echo "WARN: Checks failed. Review issues before proceeding."
  fi
fi

export SESSION_START_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
echo ""
echo "=== Session started at $SESSION_START_TIME ==="
echo "Run 'bash .claude/session-end.sh \"summary\" \"next action\"' when done."

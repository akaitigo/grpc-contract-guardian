#!/usr/bin/env bash
# セッション終了ルーチン
set -euo pipefail

SUMMARY="${1:?Usage: session-end.sh <summary> <next_action>}"
NEXT_ACTION="${2:?Usage: session-end.sh <summary> <next_action>}"

[ -d ".git" ] || { echo "ERROR: Not in git repository"; exit 1; }

STAGE="build"
if [ -f "CONTEXT.json" ]; then
  STAGE=$(jq -r '.current_stage // "build"' CONTEXT.json)
fi

echo "=== State checkpoint ==="
BACKUP_DIR=".session-backups/$(date -u +%Y%m%d-%H%M%S)"
mkdir -p "$BACKUP_DIR"
cp CONTEXT.json "$BACKUP_DIR/CONTEXT.json" 2>/dev/null || true
cp progress.json "$BACKUP_DIR/progress.json" 2>/dev/null || true
echo "Backup saved to $BACKUP_DIR"

echo "=== Running quality checks ==="
LINT_RESULT="skip"
TEST_RESULT="skip"
TYPECHECK_RESULT="skip"

if make lint >/dev/null 2>&1; then LINT_RESULT="pass"; else LINT_RESULT="fail"; fi
if make test >/dev/null 2>&1; then TEST_RESULT="pass"; else TEST_RESULT="fail"; fi

SESSION_ID=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
STARTED_AT="${SESSION_START_TIME:-$SESSION_ID}"
ENDED_AT="$SESSION_ID"

COMMITS_JSON=$(git log --oneline -10 --format='{"hash":"%h","message":"%s"}' | jq -s '.')
FILES_CHANGED=$(git diff --name-only HEAD~5 HEAD 2>/dev/null | jq -R -s 'split("\n") | map(select(. != ""))')

if [ ! -f "progress.json" ]; then
  echo '{"project":"grpc-contract-guardian","sessions":[]}' > progress.json
fi

NEW_SESSION=$(jq -n \
  --arg session_id "$SESSION_ID" \
  --arg started_at "$STARTED_AT" \
  --arg ended_at "$ENDED_AT" \
  --arg stage "$STAGE" \
  --arg summary "$SUMMARY" \
  --argjson commits "$COMMITS_JSON" \
  --argjson files_changed "$FILES_CHANGED" \
  --arg lint "$LINT_RESULT" \
  --arg test "$TEST_RESULT" \
  --arg typecheck "$TYPECHECK_RESULT" \
  --arg next_action "$NEXT_ACTION" \
  '{
    session_id: $session_id,
    started_at: $started_at,
    ended_at: $ended_at,
    stage: $stage,
    summary: $summary,
    commits: $commits,
    files_changed: $files_changed,
    checks: {
      lint: $lint,
      test: $test,
      typecheck: $typecheck
    },
    blockers: [],
    next_action: $next_action
  }')

jq --argjson session "$NEW_SESSION" '.sessions += [$session]' progress.json > progress.json.tmp && mv progress.json.tmp progress.json

if [ -f "CONTEXT.json" ]; then
  jq --arg action "$NEXT_ACTION" --arg ts "$ENDED_AT" \
    '.next_best_action = $action | .last_checks.timestamp = $ts' \
    CONTEXT.json > CONTEXT.json.tmp && mv CONTEXT.json.tmp CONTEXT.json
fi

echo "=== Creating session-end commit ==="
git add progress.json CONTEXT.json 2>/dev/null || git add progress.json
git commit -m "session-end: ${SUMMARY}

Stage: ${STAGE}
Checks: lint=${LINT_RESULT}, test=${TEST_RESULT}
Next: ${NEXT_ACTION}"

echo "=== Session ended successfully ==="
echo "Summary: ${SUMMARY}"
echo "Next action: ${NEXT_ACTION}"

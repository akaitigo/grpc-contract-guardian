#!/usr/bin/env bash
set -euo pipefail

input="$(cat)"
file="$(jq -r '.tool_input.file_path // .tool_input.path // empty' <<< "$input")"

case "$file" in *.go) ;; *) exit 0 ;; esac

cd "$(git rev-parse --show-toplevel 2>/dev/null || dirname "$file")"

# 1. 自動修正を先に（サイレント）
golangci-lint run --fix "$file" >/dev/null 2>&1 || true

# 2. 残った違反だけをJSON返却
diag="$(golangci-lint run "$file" 2>&1 | head -20)"
if [ -n "$diag" ]; then
  jq -Rn --arg msg "$diag" \
    '{ hookSpecificOutput: { hookEventName: "PostToolUse", additionalContext: $msg } }'
fi

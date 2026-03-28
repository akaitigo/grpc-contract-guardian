#!/usr/bin/env bash
# セッション起動ルーチン
set -euo pipefail

SKIP_CHECKS=false

for arg in "$@"; do
  case "$arg" in
    --skip-checks) SKIP_CHECKS=true ;;
  esac
done

echo "=== Session Startup ==="

[ -d ".git" ] || { echo "ERROR: Not in git repository"; exit 1; }

echo "=== Recent commits ==="
git log --oneline -10

echo "=== Open issues ==="
gh issue list --repo akaitigo/grpc-contract-guardian --state open --limit 5 2>/dev/null || echo "INFO: Could not fetch issues (gh CLI not configured or offline)"

if [ "$SKIP_CHECKS" = true ]; then
  echo "=== Health check SKIPPED (--skip-checks) ==="
else
  echo "=== Health check ==="

  # Auto-install tools
  command -v golangci-lint >/dev/null || {
    echo "Installing golangci-lint..."
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
  }
  command -v gofumpt >/dev/null || {
    echo "Installing gofumpt..."
    go install mvdan.cc/gofumpt@latest
  }

  if make check 2>&1 | tail -10; then
    echo "All checks passed. Ready to work."
  else
    echo "WARN: Checks failed. Review issues before proceeding."
  fi
fi

echo ""
echo "=== Session started at $(date -u +"%Y-%m-%dT%H:%M:%SZ") ==="

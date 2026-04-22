#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

TMP="$(mktemp)"
trap 'rm -f "$TMP"' EXIT

PATTERN='ms-[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}|sk-[A-Za-z0-9_-]{16,}|Bearer[[:space:]]+[A-Za-z0-9._-]{12,}'

grep -R -n -E "$PATTERN" . \
  --exclude-dir=.git \
  --exclude-dir=.omx \
  --exclude-dir=.codex \
  --exclude-dir=artifacts \
  --exclude=*.log \
  > "$TMP" || true

# 允许的测试夹具（纯假数据）
grep -v 'internal/runner/redaction_test.go' "$TMP" > "${TMP}.filtered" || true
mv "${TMP}.filtered" "$TMP"

if [[ -s "$TMP" ]]; then
  echo '[audit_secrets] potential secret-like strings found:' >&2
  cat "$TMP" >&2
  exit 1
fi

echo '[audit_secrets] PASS: no non-whitelisted secret-like strings found'

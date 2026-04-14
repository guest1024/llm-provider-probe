#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

echo '[smoke] go test ./...'
go test ./...

echo '[smoke] list built-in cases'
go run ./cmd/provider-probe -list-cases

echo '[smoke] run compare self-check using existing reports when present'
LATEST_JSON="$(find artifacts -maxdepth 1 -name '*.json' | sort | tail -n 1 || true)"
if [[ -n "$LATEST_JSON" ]]; then
  go run ./cmd/provider-probe compare -left "$LATEST_JSON" -right "$LATEST_JSON"
  echo '[smoke] run history summary'
  ./scripts/history_summary.sh artifacts || true
  echo '[smoke] render history summary files'
  ./scripts/render_history.sh artifacts artifacts/history-smoke || true
fi

if [[ -n "${MODELSCOPE_API_KEY:-}" ]]; then
  echo '[smoke] live ModelScope probe'
  ./scripts/run_modelscope.sh --label smoke-modelscope --repeat 1
else
  echo '[smoke] skip live ModelScope probe because MODELSCOPE_API_KEY is unset'
fi

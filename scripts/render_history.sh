#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

DIR="${1:-artifacts}"
PREFIX="${2:-artifacts/history-summary}"
PROVIDER="${3:-}"
GROUP_BY="${4:-provider}"
BENCHMARK="${5:-}"

if [[ "$PROVIDER" == "provider" || "$PROVIDER" == "benchmark" || "$PROVIDER" == "provider-benchmark" ]]; then
  BENCHMARK="${4:-}"
  GROUP_BY="$PROVIDER"
  PROVIDER=""
fi

CMD=(go run ./cmd/provider-probe history -dir "$DIR" -group-by "$GROUP_BY" -md-out "${PREFIX}.md" -html-out "${PREFIX}.html")
if [[ -n "$PROVIDER" ]]; then
  CMD+=(-provider "$PROVIDER")
fi
if [[ -n "$BENCHMARK" ]]; then
  CMD+=(-benchmark "$BENCHMARK")
fi
"${CMD[@]}"

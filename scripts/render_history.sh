#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

DIR="${1:-artifacts}"
PREFIX="${2:-artifacts/history-summary}"
PROVIDER="${3:-}"

CMD=(go run ./cmd/provider-probe history -dir "$DIR" -md-out "${PREFIX}.md" -html-out "${PREFIX}.html")
if [[ -n "$PROVIDER" ]]; then
  CMD+=(-provider "$PROVIDER")
fi
"${CMD[@]}"

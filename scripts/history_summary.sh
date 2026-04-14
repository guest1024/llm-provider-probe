#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

DIR="${1:-artifacts}"
PROVIDER="${2:-}"

if [[ -n "$PROVIDER" ]]; then
  go run ./cmd/provider-probe history -dir "$DIR" -provider "$PROVIDER"
else
  go run ./cmd/provider-probe history -dir "$DIR"
fi

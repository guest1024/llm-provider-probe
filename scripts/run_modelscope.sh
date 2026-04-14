#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

if [[ -z "${MODELSCOPE_API_KEY:-}" ]]; then
  echo "MODELSCOPE_API_KEY is required" >&2
  exit 2
fi

LABEL="modelscope-qwen"
if [[ $# -ge 2 && "$1" == "--label" ]]; then
  LABEL="$2"
  shift 2
fi

exec ./scripts/run_probe.sh --config examples/modelscope-qwen.json --label "$LABEL" "$@"

#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

CONFIG="${1:-examples/modelscope-vs-openai-template.json}"
LABEL="${2:-ab}"

if [[ -z "${MODELSCOPE_API_KEY:-}" ]]; then
  echo "MODELSCOPE_API_KEY is required" >&2
  exit 2
fi
if [[ -z "${OPENAI_API_KEY:-}" ]]; then
  echo "OPENAI_API_KEY is required" >&2
  exit 2
fi

exec ./scripts/run_probe.sh --config "$CONFIG" --label "$LABEL"

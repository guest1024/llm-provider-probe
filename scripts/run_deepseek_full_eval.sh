#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

if [[ -z "${OPENAI_API_KEY:-}" ]]; then
  echo "OPENAI_API_KEY is required" >&2
  exit 2
fi
if [[ -z "${BASE_URL:-}" ]]; then
  echo "BASE_URL is required (e.g. https://api.deepseek.com/v1)" >&2
  exit 2
fi
if [[ -z "${MODEL:-}" ]]; then
  echo "MODEL is required (e.g. deepseek-v4-flash)" >&2
  exit 2
fi

exec ./scripts/run_probe.sh --config examples/deepseek-full-eval.json --label deepseek-full-eval --fail-on high "$@"

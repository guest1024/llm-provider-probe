#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

CONFIG=""
LABEL=""
FAIL_ON=""
REPEAT=""
TIMEOUT=""
EXTRA_ARGS=()

while [[ $# -gt 0 ]]; do
  case "$1" in
    --config)
      CONFIG="$2"; shift 2 ;;
    --label)
      LABEL="$2"; shift 2 ;;
    --fail-on)
      FAIL_ON="$2"; shift 2 ;;
    --repeat)
      REPEAT="$2"; shift 2 ;;
    --timeout)
      TIMEOUT="$2"; shift 2 ;;
    --)
      shift
      EXTRA_ARGS+=("$@")
      break ;;
    *)
      EXTRA_ARGS+=("$1")
      shift ;;
  esac
done

if [[ -z "$CONFIG" ]]; then
  echo "usage: $0 --config <path> [--label <name>] [--fail-on medium|high] [--repeat N] [--timeout SEC] [-- extra go args]" >&2
  exit 2
fi

STAMP="$(date +%Y%m%d-%H%M%S)"
if [[ -n "$LABEL" ]]; then
  SAFE_LABEL="$(echo "$LABEL" | tr ' /:' '---')"
else
  SAFE_LABEL="run"
fi

OUT_JSON="artifacts/${SAFE_LABEL}-${STAMP}.json"
OUT_MD="artifacts/${SAFE_LABEL}-${STAMP}.md"
CMD=(go run ./cmd/provider-probe -config "$CONFIG" -out "$OUT_JSON" -md-out "$OUT_MD")

if [[ -n "$FAIL_ON" ]]; then
  CMD+=(-fail-on "$FAIL_ON")
fi
if [[ -n "$REPEAT" ]]; then
  CMD+=(-repeat "$REPEAT")
fi
if [[ -n "$TIMEOUT" ]]; then
  CMD+=(-timeout "$TIMEOUT")
fi
if [[ ${#EXTRA_ARGS[@]} -gt 0 ]]; then
  CMD+=("${EXTRA_ARGS[@]}")
fi

echo "[run_probe] config=$CONFIG"
echo "[run_probe] out_json=$OUT_JSON"
echo "[run_probe] out_md=$OUT_MD"
"${CMD[@]}"

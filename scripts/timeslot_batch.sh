#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

CONFIG=""
SLOTS="凌晨,上午,晚高峰,周末高峰"
REPEAT="1"
FAIL_ON=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --config)
      CONFIG="$2"; shift 2 ;;
    --slots)
      SLOTS="$2"; shift 2 ;;
    --repeat)
      REPEAT="$2"; shift 2 ;;
    --fail-on)
      FAIL_ON="$2"; shift 2 ;;
    *)
      echo "unknown arg: $1" >&2
      exit 2 ;;
  esac
done

if [[ -z "$CONFIG" ]]; then
  echo "usage: $0 --config <path> [--slots a,b,c] [--repeat N] [--fail-on medium|high]" >&2
  exit 2
fi

IFS=',' read -r -a SLOT_ITEMS <<< "$SLOTS"
for slot in "${SLOT_ITEMS[@]}"; do
  echo "=== running slot: $slot ==="
  if [[ -n "$FAIL_ON" ]]; then
    ./scripts/run_probe.sh --config "$CONFIG" --label "$slot" --repeat "$REPEAT" --fail-on "$FAIL_ON"
  else
    ./scripts/run_probe.sh --config "$CONFIG" --label "$slot" --repeat "$REPEAT"
  fi
done

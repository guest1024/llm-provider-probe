#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

if [[ $# -ne 2 ]]; then
  echo "usage: $0 <left.json> <right.json>" >&2
  exit 2
fi

go run ./cmd/provider-probe compare -left "$1" -right "$2"

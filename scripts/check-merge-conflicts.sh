#!/usr/bin/env bash
set -Eeuo pipefail
ROOT="${1:-$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)}"
cd "$ROOT"
mapfile -t bad < <(grep -RIl --exclude-dir=.git --exclude='*.zip' --exclude='*.pdf' -E '^(<<<<<<< |>>>>>>> )' . || true)
if ((${#bad[@]})); then
  printf 'Unresolved merge conflicts:\n'; printf '  %s\n' "${bad[@]}"; exit 1
fi
echo "Merge conflict markers: none"

#!/usr/bin/env bash
set -Eeuo pipefail
ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
if ! command -v python3 >/dev/null 2>&1; then
  if (( EUID != 0 )); then echo "Run sudo ./install.sh (Python3 is required)" >&2; exit 1; fi
  command -v apt-get >/dev/null 2>&1 || { echo "Install Python3 first" >&2; exit 1; }
  for arg in "$@"; do [[ "$arg" != --offline && "$arg" != --dry-run ]] || { echo "Python3 is required for this mode; no packages installed" >&2; exit 1; }; done
  apt-get update
  apt-get install -y --no-install-recommends python3 ca-certificates
fi
exec python3 "$ROOT/scripts/installer/auto_install.py" "$@"

#!/usr/bin/env bash
set -Eeuo pipefail
ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT"
bash -n install.sh
bash -n trestctl.sh
python3 -m unittest discover -s tests/installer -v
echo 'Installer unit/structure tests PASS; no Docker runtime claim.'

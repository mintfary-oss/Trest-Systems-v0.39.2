#!/usr/bin/env bash
set -Eeuo pipefail
ROOT="${1:-$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd)}"
exec python3 "$ROOT/scripts/installer/auto_install.py" --install-dir "$ROOT" --doctor

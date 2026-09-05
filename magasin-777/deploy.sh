#!/usr/bin/env bash
set -Eeuo pipefail
ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
echo 'This component is installed by the unified Trest Systems 0.39.2 installer.'
exec bash "$ROOT/install.sh" "$@"

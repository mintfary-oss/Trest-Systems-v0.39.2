#!/usr/bin/env bash
set -Eeuo pipefail
ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
bash "$ROOT/install.sh" --dry-run --install-dir /tmp/trest-clean-host-check
command -v docker >/dev/null || { echo 'BLOCKED: Docker absent; package validation is not installation acceptance' >&2; exit 2; }
docker info >/dev/null
docker compose version
echo 'Prerequisites only; clean-host installation/E2E has NOT been performed by this check.'

#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
status=0
for f in scripts/*.sh; do bash -n "$f" || status=1; done
./scripts/security-baseline.sh || status=1
./scripts/clean-host-check.sh || status=1
./scripts/validate-compose.sh || status=1
if command -v go >/dev/null 2>&1; then
  go test ./... || status=1
  go vet ./... || status=1
else
  echo 'FAIL: Go toolchain missing' >&2; status=1
fi
if command -v govulncheck >/dev/null 2>&1; then govulncheck ./... || status=1; else echo 'FAIL: govulncheck missing' >&2; status=1; fi
if [[ -f magasin-777/frontend/package-lock.json ]] && command -v npm >/dev/null 2>&1; then (cd magasin-777/frontend && npm audit --omit=dev) || status=1; else echo 'FAIL: npm audit prerequisites missing' >&2; status=1; fi
if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then docker compose -f deployments/docker-compose.yml config >/dev/null || status=1; else echo 'FAIL: Docker/Compose missing' >&2; status=1; fi
exit "$status"

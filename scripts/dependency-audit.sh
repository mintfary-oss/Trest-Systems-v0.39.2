#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
status=0
if command -v govulncheck >/dev/null 2>&1; then
  govulncheck ./... || status=$?
else
  echo "WARN: govulncheck is not installed; install it in CI/release environment."
fi
if [ -f magasin-777/frontend/package-lock.json ] && command -v npm >/dev/null 2>&1; then
  (cd magasin-777/frontend && npm audit --omit=dev) || status=$?
else
  echo "WARN: npm lockfile or npm unavailable; frontend audit deferred."
fi
exit "$status"

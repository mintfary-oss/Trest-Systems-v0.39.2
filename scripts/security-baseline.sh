#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
fail=0
check(){ if "$@" >/dev/null 2>&1; then echo "[PASS] $*"; else echo "[FAIL] $*"; fail=1; fi; }
check test -f .env.example
check test -f deployments/docker-compose.yml
check test -f Dockerfile.api
check test -f Dockerfile.worker
if grep -RInE --exclude-dir=.git --exclude='*.lock' '(BEGIN (RSA |EC )?PRIVATE KEY|AKIA[0-9A-Z]{16})' . >/tmp/trest-secret-scan.$$ 2>/dev/null; then
  echo "[FAIL] possible hard-coded key material:"; sed -n '1,20p' /tmp/trest-secret-scan.$$; fail=1
else echo "[PASS] no obvious private-key/AWS-key patterns"; fi
rm -f /tmp/trest-secret-scan.$$
if find . -type f -perm -o+w -not -path './.git/*' | grep -q .; then echo "[FAIL] world-writable files found"; find . -type f -perm -o+w -not -path './.git/*' | head -20; fail=1; else echo "[PASS] no world-writable files"; fi
if grep -RInE --exclude-dir=.git '(0\.0\.0\.0:[0-9]+|:[0-9]+:.*0\.0\.0\.0)' deployments super-sistema 2>/dev/null | grep -vE '127\.0\.0\.1|localhost' >/tmp/trest-bind.$$; then echo "[WARN] externally bound ports require deployment review"; sed -n '1,20p' /tmp/trest-bind.$$; else echo "[PASS] no obvious unrestricted host bindings detected"; fi
rm -f /tmp/trest-bind.$$
if ((fail)); then exit 1; fi
echo "security-baseline: PASS"

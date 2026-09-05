#!/usr/bin/env bash
set -Eeuo pipefail
ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
python3 - "$ROOT/verification/v0.39.2/VERIFICATION.json" <<'PYCODE'
import json,sys
from pathlib import Path
p=Path(sys.argv[1])
if not p.is_file():raise SystemExit('NOT APPROVED: verification evidence absent')
d=json.loads(p.read_text())
if not d.get('production_ready_claim') or not d.get('full_product_e2e_verified'):
 raise SystemExit('NOT APPROVED FOR PRODUCTION: runtime/E2E/security gates not verified. This package is a release candidate.')
if d.get('counts',{}).get('FAIL',0) or d.get('counts',{}).get('BLOCKED',0):raise SystemExit('NOT APPROVED: open gates')
print('Recorded production gates: PASS; validate report provenance separately')
PYCODE

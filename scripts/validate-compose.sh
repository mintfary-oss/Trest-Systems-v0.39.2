#!/usr/bin/env bash
set -Eeuo pipefail
ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
[[ -f "$ROOT/.env" ]] || { echo 'Missing installation .env; run install.sh or use a disposable test config' >&2; exit 2; }
ROOT="$ROOT" python3 - <<'PYCODE'
import sys,os,subprocess
from pathlib import Path
root=Path(os.environ['ROOT']);sys.path.insert(0,str(root/'scripts/installer'))
from auto_install import read_env,compose_args
e=read_env(root/'.env');clean=os.environ.copy()
for k in e:clean.pop(k,None)
subprocess.run(compose_args(root,e)+['config','--quiet'],env=clean,check=True)
print('Compose syntax/interpolation: PASS (not runtime)')
PYCODE

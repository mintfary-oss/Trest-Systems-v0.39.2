#!/usr/bin/env bash
set -Eeuo pipefail
ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
cmd="${1:-status}"; shift || true
case "$cmd" in
 install) exec bash "$ROOT/install.sh" "$@";;
 doctor|check) exec python3 "$ROOT/scripts/installer/auto_install.py" --install-dir "$ROOT" --doctor "$@";;
 backup) exec python3 "$ROOT/scripts/installer/auto_install.py" --install-dir "$ROOT" --backup-only "$@";;
 restore-drill) exec python3 "$ROOT/scripts/installer/auto_install.py" --install-dir "$ROOT" --restore-drill-only "$@";;
 migrate) exec python3 "$ROOT/scripts/installer/auto_install.py" --install-dir "$ROOT" --migrate-only "$@";;
 version) echo "Trest Systems 0.39.2";;
 status|logs|start|stop|restart) exec python3 "$ROOT/scripts/installer/lifecycle.py" "$ROOT" "$cmd" "$@";;
 *) echo "Commands: install, status, logs, start, stop, restart, doctor, backup, restore-drill, migrate, version";exit 2;;
esac

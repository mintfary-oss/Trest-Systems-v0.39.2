#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
FILE="${1:-}"
if [[ -z "$FILE" ]]; then echo "usage: $0 <backup.sql>" >&2; exit 2; fi
[[ -f "$FILE" ]] || { echo "backup not found: $FILE" >&2; exit 2; }
[[ "${FILE##*.}" == "sql" ]] || { echo "backup must be .sql" >&2; exit 2; }
SIZE=$(wc -c <"$FILE")
(( SIZE > 32 )) || { echo "backup is suspiciously small: $SIZE bytes" >&2; exit 1; }
if ! grep -qE '^(--|SET |CREATE |INSERT |COPY |ALTER |SELECT |BEGIN|COMMIT|GRANT |REVOKE |COMMENT )' "$FILE"; then echo "backup does not look like pg_dump SQL" >&2; exit 1; fi
sha256sum "$FILE"
echo "backup-verify: PASS"

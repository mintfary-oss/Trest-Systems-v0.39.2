# Stage 11 — Security Baseline + Backup/Restore Hardening

## Сделано
- добавлен `scripts/security-baseline.sh`;
- добавлен `scripts/backup-verify.sh`;
- `trestctl backup` дополнительно проверяет, что dump не пустой;
- добавлена документация `docs/install/SECURITY_BASELINE_RU.md`;
- добавлена документация `docs/install/BACKUP_RESTORE_RU.md`;
- release-check включает security baseline;
- `PATENT_RU/` и `READABLE_TXT_RU/` синхронизированы.

## Проверки
- security baseline: PASS;
- shell syntax: PASS;
- целевые Go-тесты: попытка запуска заблокирована отсутствием `go.sum` для части зависимостей и недоступностью `proxy.golang.org` в текущей среде;
- Docker runtime/E2E: не выполнен, Docker отсутствует в текущей среде.

## Важно
Это baseline, а не полноценный penetration test/SAST/DAST. Restore остаётся разрушительной операцией и требует явного `TREST_CONFIRM_RESTORE=YES`.

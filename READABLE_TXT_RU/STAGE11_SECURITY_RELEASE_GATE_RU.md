# Stage 11 — Security Release Gate

Версия: `v0.35-stage11-security-release-gate`

## Назначение
Финальный автоматизированный gate перед release candidate. Он не заменяет внешний penetration test.

## Проверки
- syntax всех shell scripts;
- security baseline;
- clean-host prerequisites;
- Docker Compose config;
- `go test ./...`;
- `go vet ./...`;
- `govulncheck ./...`;
- `npm audit --omit=dev`;
- oversized/malicious IFC fixture checks выполняются disposable E2E;
- backup verification;
- restore drill только на disposable Compose project и только при `TREST_CONFIRM_RESTORE=YES`.

## Важно
В среде без Docker/Compose или без доступа к Go module registry gate закономерно завершается FAIL. Это означает отсутствие доказательства, а не успешный runtime.

## Restore drill
```bash
TREST_CONFIRM_RESTORE=YES ./scripts/backup-restore-drill.sh
```
Команда уничтожает только собственный disposable Compose project, указанный через `COMPOSE_PROJECT_NAME`; не запускайте её против production.

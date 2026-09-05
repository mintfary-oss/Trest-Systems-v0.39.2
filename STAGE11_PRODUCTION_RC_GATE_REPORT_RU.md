# Stage 11 — Production RC Gate Report

## v0.36

### Цель
Перевести Security Release Gate v0.35 в воспроизводимый Production RC gate на CI/Docker host, не называя систему Production Ready до фактического runtime-прохождения.

### Добавлено
- `scripts/production-rc-gate.sh` — единая проверка shell/security/compose/domain/Go/dependency gates.
- CI job `security` с `govulncheck`.
- CI job `restore-drill` для disposable PostgreSQL backup/restore.
- CI job `clean-host` для проверки release prerequisites без отдельной установки Go.
- Существующие E2E/security jobs сохранены.

### Статус
Локально Docker runtime и внешний Go module registry недоступны, поэтому CI-only gates здесь не объявляются пройденными.

### Production blockers
1. Фактический GitHub Actions run.
2. Полный Docker E2E.
3. Реальный backup/restore drill.
4. Успешный `go test ./...`, `go vet ./...`, release build на CI.
5. Dependency vulnerability audit без критических блокеров.
6. Clean-host install/rollback verification.

**Классификация:** Engineering MVP / Pre-Beta. Production Ready не подтверждён.

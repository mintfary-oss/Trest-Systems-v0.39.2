# Stage 11 — Full E2E preparation report

## Версия
v0.32-stage11-full-e2e

## Сделано
- Добавлен `scripts/e2e-full.sh`.
- Покрыт базовый пользовательский поток: register → login → me → project → estimate → approve → order → idempotency → BIM model/version → PostgreSQL backup.
- Добавлена CI job `e2e`.
- Обнаружены и исправлены два реальных integration gap: `createProject` существовал, но `POST /api/v1/projects` не был зарегистрирован; `/health` и `/ready` не были явно проксированы Nginx в API. Оба исправлены.

## Верификация в текущем окружении
- `gofmt`: выполнен.
- Docker runtime: недоступен, поэтому фактический Compose E2E не запускался.
- Production readiness: не подтверждается.

## Следующий P0
Запустить `./scripts/e2e-full.sh` на Linux runner с Docker/Compose и зафиксировать фактический PASS/FAIL. Затем выполнить отдельный disposable restore drill и IFC multipart import/tessellation E2E.

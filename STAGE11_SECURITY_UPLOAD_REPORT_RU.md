# Stage 11 — Upload / Rate-Limit Hardening Report

**Версия:** `v0.34-stage11-security-upload-rate-limit`
**Дата:** 2026-09-03
**Статус:** подготовлено; runtime Docker ещё не подтверждён в текущем окружении.

## Выполнено

1. IFC multipart upload ограничен `http.MaxBytesReader` на 64 MiB.
2. Multipart parsing ограничен тем же лимитом.
3. Имя IFC-файла валидируется до формирования object-storage key.
4. Символы `/` и `\\` в исходном имени файла отклоняются, исключая path traversal через имя объекта.
5. Object-storage key формируется только из серверных `projectID`, `versionID` и проверенного имени файла.
6. Добавлен in-process rate limiter: 120 запросов на IP за 60 секунд.
7. `X-Forwarded-For` не доверяется по умолчанию; для доверенного proxy требуется явный `X-Trust-Proxy: 1`.
8. Добавлен `scripts/dependency-audit.sh` для `govulncheck` и frontend `npm audit` в release/CI environment.
9. Добавлены unit tests для rate limiter и определения client IP.

## Ограничения

- Rate limiter in-process не заменяет edge/gateway rate limiting при нескольких API replicas.
- Полный dependency audit требует среды с доступом к vulnerability databases/registry.
- Полный Docker E2E и malicious IFC runtime test в текущем окружении не запускались, потому что Docker/Compose недоступны.

## Release gate

Нельзя считать production security gate полностью пройденным до выполнения на Linux/Docker runner:

- `go test ./...`
- `go vet ./...`
- `go build ./...`
- `docker compose config`
- full E2E
- oversized IFC rejection
- traversal filename rejection
- backup/restore drill
- dependency vulnerability audit
- clean-host installation.

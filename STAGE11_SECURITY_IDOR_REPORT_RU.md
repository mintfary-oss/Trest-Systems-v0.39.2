# Stage 11 — Security / IDOR Hardening Report

Дата: 2026-09-03
Версия: v0.33-stage11-security-idor

## Изменения
1. `GET /api/v1/users` закрыт авторизацией и ролью `admin`.
2. `GET /api/v1/projects` закрыт авторизацией; обычный пользователь видит только собственные проекты.
3. `GET /api/v1/orders` закрыт авторизацией; обычный пользователь видит заказы только собственных проектов.
4. Добавлен `scripts/clean-host-check.sh` для проверки release prerequisites на хосте без Go.
5. Добавлен `docs/install/SECURITY_E2E_RU.md`.

## Проверки
- `bash -n` для E2E/security/clean-host scripts: PASS.
- Dependency-free/domain/BIM tests: PASS.
- Полный API test suite: не подтверждён из-за ограничений окружения с Go module registry.
- Docker/Compose E2E: не запускался, поскольку Docker отсутствует в текущей среде.
- Clean-host runtime: не подтверждён; скрипт подготовлен для Linux/Docker runner.

## Важное
Это hardening-контроль доступа, а не заявление о полном security audit. Перед Production остаются auth/IDOR fuzzing, upload/path traversal, oversized IFC, rate limiting, dependency audit и внешний security review.

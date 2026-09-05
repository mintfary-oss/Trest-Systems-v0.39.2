# Security / E2E Release Gate — Trest Systems

## Цель
Проверить, что публичный API не раскрывает списки пользователей, проектов и заказов без авторизации, а обычный пользователь получает только собственные проекты/заказы.

## Выполненные hardening-изменения
- `GET /api/v1/users` теперь требует Bearer + роль `admin`.
- `GET /api/v1/projects` теперь требует Bearer; customer видит только свои проекты.
- `GET /api/v1/orders` теперь требует Bearer; customer видит заказы только своих проектов.
- Admin сохраняет обзорный доступ.

## Runtime gate
Запускать на Linux/Docker host:

```bash
./scripts/e2e-full.sh
./scripts/clean-host-check.sh
```

`e2e-full.sh` является disposable-тестом и использует отдельный Compose project. Production environment запускать этим скриптом нельзя.

## Ограничения
Локальная среда разработки без Docker не подтверждает runtime PASS. Полный `go test ./...`, `go vet ./...`, Compose E2E и clean-host gate должны быть выполнены в CI/Linux runner.

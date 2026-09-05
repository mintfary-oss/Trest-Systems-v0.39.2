# Полный E2E Trest Systems

## Назначение
`scripts/e2e-full.sh` — disposable smoke/E2E-проверка полного базового потока на Docker Compose.

Проверяет:
1. Compose config и запуск стека;
2. PostgreSQL/Redis/Marketplace API;
3. `/health` и `/ready` через Nginx → unified API;
4. регистрацию и login;
5. `/api/v1/me`;
6. создание проекта;
7. создание и approve estimate;
8. создание order;
9. идемпотентность order по `idempotency_key`;
10. создание BIM model/version;
11. PostgreSQL backup non-empty check.

## Запуск

```bash
./scripts/e2e-full.sh
```

Тест использует отдельный Compose project `trest-e2e`, генерирует временные секреты и по завершении удаляет созданные E2E containers/volumes.

Для отладки можно оставить окружение:

```bash
TREST_E2E_KEEP=YES ./scripts/e2e-full.sh
```

Удаление E2E окружения выполняется только самим тестовым скриптом для его собственного Compose project.

## Ограничения

Это не production acceptance test. В текущей разработческой среде Docker runtime недоступен, поэтому фактический PASS должен быть получен на Linux host/CI runner с Docker.

IFC upload/tessellation и restore drill остаются отдельными P0-проверками; данный smoke не подменяет полноценный security/load test.

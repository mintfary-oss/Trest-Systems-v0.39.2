# Contributing

1. Прочитать `AGENTS.md`, `CURRENT_STATE.md`, `TASKS.md`.
2. Проверить `git status`.
3. Делать небольшие проверяемые изменения.
4. Для API обновлять документацию.
5. Для БД добавлять миграции.
6. Для архитектуры добавлять ADR.
7. Не добавлять секреты.
8. Перед commit запускать:

```bash
git diff --check
go build ./...
go vet ./...
go test ./...
cd deployments && docker compose config
```

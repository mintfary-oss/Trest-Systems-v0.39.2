# CI/CD релиз Trest Systems

## Назначение

GitHub Actions собирает проверенный релиз без необходимости компилировать проект на production-сервере.

Pipeline:

1. `go test ./...`
2. `go vet ./...`
3. сборка статических release-бинарников;
4. `docker compose config`;
5. упаковка release-архива;
6. для Git tag `v*` — сборка и публикация OCI-образов API и Worker в GHCR.

## Production-принцип

Production-хост не обязан иметь Go toolchain. Он получает заранее собранные бинарники/образы и запускает их через Docker Compose.

## Важное ограничение

CI-конфигурация подготовлена, но её фактическое выполнение в текущей среде не заявляется. Реальный green build появляется после запуска workflow в GitHub Actions.

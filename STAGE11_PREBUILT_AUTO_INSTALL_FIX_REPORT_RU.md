# Trest Systems v0.39.1 — отчёт исправления автоматической установки

Дата контрольной точки: 2026-09-05.

## Результат

Подготовлен полный установочный пакет для Linux x86_64/amd64, в котором Go-код не компилируется на целевом сервере. В `release/bin/linux/amd64/` включены готовые:

- `trest-api`;
- `trest-worker`;
- `trestctl`;
- `trest`;
- `trest-install`;
- `trest-installer`;
- `generate-dxf`.

Контрольные суммы находятся в `release/bin/SHA256SUMS` и автоматически проверяются установщиком.

## Исправления

1. Добавлены локальные воспроизводимые Go-модули для используемого подмножества Cobra, YAML и PostgreSQL API.
2. Весь корневой Go-модуль проходит `go test ./...`, `go vet ./...` и `go build ./...` без обращения к сети.
3. Исправлены конфликты обработчика заказа, обращения к несуществующему claim `Subject`, обработка `RowsAffected`, транзакционный `ExecContext` и область видимости ошибок рейтинга.
4. Исправлены отсутствующие paths-функции `trestctl`.
5. API теперь использует `JWT_SECRET` из Compose и прекращает запуск без секрета.
6. Worker получает обязательный `DATABASE_URL`.
7. Dockerfile API/Worker переведены на Debian runtime с `libpq5`; компиляции Go внутри Docker build нет.
8. Пароль PostgreSQL в основном Compose стал обязательным — небезопасного fallback нет.
9. MinIO, Ollama и Open WebUI по умолчанию привязаны к `127.0.0.1`; наружу публикуются только edge-порты HTTP/HTTPS.
10. Генерация `DATABASE_URL` учитывает выбранные `POSTGRES_USER` и `POSTGRES_DB`.
11. Исправлена ошибка проверки ресурсов, из-за которой сервер с достаточной RAM/диском мог ошибочно остановить установку.
12. `--dry-run` теперь проверяет пакет без копирования приложения и создания state-каталога.
13. Исправлено обрезание UTF-8 в именах DXF-файлов.

## Выполненные проверки

- root Go test/vet/build: PASS;
- `trest-installer` test/vet/build: PASS;
- `proektirovka-sdaniy` test/vet/build: PASS;
- bootstrap installer test/vet/build: PASS;
- локальные compatibility modules test/vet: PASS;
- shell syntax: PASS;
- Python syntax: PASS;
- JSON/YAML syntax: PASS;
- Compose structural validation: PASS;
- prebuilt checksums: PASS;
- API fail-closed без JWT secret: PASS;
- Worker `/health`: PASS;
- DXF smoke: PASS, создано 10 непустых чертежей с корректными UTF-8 именами;
- installer isolated dry-run: PASS;
- secret scan: PASS;
- merge-conflict scan: PASS.

Подробные журналы находятся в `verification/v0.39.1/`.

## Честное ограничение

Docker runtime в среде сборки отсутствует. Поэтому полный Compose E2E, реальный выпуск TLS-сертификата, проверка внешнего интернет-доступа и backup/restore drill должны быть выполнены самим установщиком на целевом Linux-сервере. Установщик считает работу успешной только после прохождения обязательных runtime-этапов и формирует TXT/JSON-отчёт; непройденные проверки не маркируются как PASS.

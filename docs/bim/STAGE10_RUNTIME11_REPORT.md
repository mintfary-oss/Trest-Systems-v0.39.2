# Stage 10 Runtime 11 — MinIO/S3 + IFC Import API

## Что добавлено

- Dependency-free AWS Signature V4 клиент `internal/bim/storage/s3.go`, совместимый с MinIO/S3.
- Multipart API `POST /api/v1/bim/ifc/import`.
- Проверка project membership перед импортом.
- Сохранение исходного IFC в object storage с SHA-256/размером.
- Парсинг IFC STEP и атомарный semantic import в `bim_elements`.
- Аудит каждой попытки импорта.
- Migration `0019_bim_storage_import_runtime11.sql`.
- LocalStore остаётся fallback для development.

## Проверка

Прошли тесты:

- `internal/bim/storage`
- `internal/bim/ifc`
- `internal/bim/ifcsemantic`
- `internal/bim/diff`
- `internal/bim/converter`
- `internal/bim`

Docker/реальный MinIO в этой среде не запускался, поэтому подтверждена совместимость на уровне протокола и unit-тестов, но не end-to-end подключение к живому MinIO.

## Ограничения

- Для production необходимо задать S3_ENDPOINT/S3_BUCKET/S3_ACCESS_KEY/S3_SECRET_KEY.
- API пока принимает IFC как multipart upload и импортирует semantic слой; генерация полноценной IFC geometry mesh manifest остаётся следующим этапом.

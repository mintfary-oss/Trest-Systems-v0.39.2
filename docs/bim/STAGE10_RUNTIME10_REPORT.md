# Stage 10 Runtime 10 — Mesh Picking / Storage / Authorization

## Сделано

1. `internal/bim/meshmanifest.go`
   - immutable mapping диапазонов triangle indices к `external_id` и `IFC GlobalId`;
   - validation triangle alignment and bounds;
   - lookup по renderer index.
2. `internal/api/bim_runtime10.go`
   - `GET /api/v1/bim/elements/pick`;
   - получает mesh manifest из версии модели;
   - проверяет владельца проекта;
   - возвращает BIM element и mesh range.
3. `migrations/0018_bim_runtime10.sql`
   - `mesh_manifest`;
   - `object_storage_key`;
   - storage index.
4. `internal/bim/storage/object.go`
   - `Store` interface;
   - self-hosted `LocalStore`;
   - SHA-256 и размер объекта.
5. Unit tests для manifest/storage.

## Проверка

Доступные чистые Go пакеты Stage 10: PASS.
Полный API test в текущей среде блокируется отсутствующим `go.sum` для pgx, как и ранее.
Docker/Compose в среде отсутствует.

## Ограничения

- S3/MinIO transport пока представлен абстракцией Store; конкретный SigV4/MinIO adapter переносится в production runtime после подтверждения зависимости/CI.
- API picking использует manifest из БД, но WebGL ray/triangle intersection в viewer ещё не подключён.
- Текущая схема `projects` содержит `owner_id`; полноценное project membership через организации потребует отдельной миграции модели проектов.

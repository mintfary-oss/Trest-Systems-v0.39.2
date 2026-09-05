# Stage 10 Runtime 9 — IFC Properties, Quantities, Atomic Import and Semantic Diff

## Реализовано
- Декодер `IfcPropertySet` / `IfcPropertySingleValue`.
- Декодер распространённых `IfcQuantity*` с единицей и исходным STEP token.
- Атомарный importer `internal/bim/importer`: все upsert-операции выполняются в одной SQL-транзакции; ошибка откатывает импорт целиком.
- Аудит импорта `bim_semantic_imports` (migration 0017).
- Semantic diff по стабильному IFC `GlobalId`.
- Self-hosted local object-storage adapter с SHA-256 manifest.
- Viewer progress overlay для `planned_percent` / `actual_percent` из semantic JSON.

## Ограничения
- PostgreSQL runtime не запускался в этой среде: Docker/Compose и внешний module registry недоступны.
- Object storage adapter сейчас локальный filesystem abstraction; MinIO/S3 connector подключается следующим runtime-этапом.
- 3D picking пока не считается production-ready: для точной связи треугольника с IFC GlobalId нужен mesh-element manifest.

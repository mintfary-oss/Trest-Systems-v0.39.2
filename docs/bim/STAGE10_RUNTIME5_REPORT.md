# Stage 10 Runtime 5 — IFC semantic foundation

## Реализовано

- Добавлен `internal/bim/ifc` — self-hosted парсер физического IFC STEP-слоя без внешнего IFC SDK.
- Парсер извлекает `#ID`, тип сущности IFC и список атрибутов.
- Сохраняется вложенность tuple/list-атрибутов и запятые внутри строк.
- Размер строки ограничен безопасным буфером scanner до 8 MiB.
- Добавлены unit-тесты для IFC entity parsing.

## Границы текущего этапа

Это фундамент semantic layer, а не полный IFC implementation. Пока не реализованы IFC schema-driven типизация всех классов, inverse attributes, property sets, geometry extraction, placements, materials и relationship graph.

## Проверка

Прошли:

- `go test ./internal/bim/ifc`
- `go test ./internal/bim/converter ./internal/bim/diff ./internal/bim`

Полный `go test ./internal/bim/...` по-прежнему зависит от доступности внешнего pgx-модуля для queue в текущей среде.

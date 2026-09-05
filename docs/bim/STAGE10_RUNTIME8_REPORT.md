# Stage 10 Runtime 8 — IFC schema-aware core + Viewer semantic tree

Дата: 2026-09-03

## Сделано
- Исправлено отображение `IfcRelAggregates`: первый reference — parent, остальные — children.
- Добавлен whitelist основных IFC-классов: Project, Site, Building, BuildingStorey, Space, Wall, Slab, Roof, Door, Window, Column, Beam.
- Добавлен `MapToBIMElements()` для безопасного преобразования IFC semantic данных в нейтральную модель `bim_elements` без выдумывания геометрии.
- Добавлен `SpatialPath()` для построения цепочки пространственных контейнеров.
- Добавлена migration `0016_bim_ifc_import_metadata.sql` для traceability source format/entity type.
- BIM Viewer очищен от дублирующей `loadGLTF` и получил semantic JSON tree + properties panel + selection.

## Ограничения
- Это schema-aware core subset, а не полный IFC EXPRESS schema decoder.
- PropertySet decoding и geometry-to-element mapping остаются следующим этапом.
- Viewer selection сейчас выбирает semantic element в дереве; пиксельный picking 3D geometry ещё не реализован.
- Транзакционный DB importer должен использовать этот mapper вместе с существующим queue processor после подключения object storage.

## Проверки
`go test ./internal/bim/ifc ./internal/bim/ifcsemantic ./internal/bim/converter ./internal/bim/diff ./internal/bim` — PASS.

## Следующий этап
IFC PropertySets/Quantities → transactional importer → GlobalId semantic diff → 3D picking → progress overlays → object storage.

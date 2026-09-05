# Stage 10 Runtime 7 — IFC semantic mapping

Дата: 2026-09-03

## Реализовано
- Добавлен консервативный semantic relation mapper для `IfcRelAggregates`.
- Добавлен mapper для `IfcRelContainedInSpatialStructure`.
- В `bim_elements` добавлены IFC semantic identity/containment поля:
  - `ifc_entity_id`
  - `ifc_global_id`
  - `parent_external_id`
  - `relation_type`
- Добавлены индексы для GlobalId и parent relation.
- Добавлены unit-тесты relation mapping.

## Ограничения
Это не полный IFC schema engine. Mapping намеренно работает с распространёнными relation entities и не делает предположений о неизвестных schema variants.

## Следующий шаг
- schema-aware decoders для Project/Site/Building/Storey/Space и основных строительных элементов;
- импорт IFC в `bim_elements` транзакционно;
- viewer tree/properties/selection;
- semantic diff по IFC GlobalId;
- richer DXF/OBJ exchange.

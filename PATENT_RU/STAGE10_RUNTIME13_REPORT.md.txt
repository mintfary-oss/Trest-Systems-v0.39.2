# Stage 10 Runtime 13 — IFC Geometry Engine Foundation

**Дата:** 2026-09-03  
**Версия:** v0.27

## Выполнено
- Добавлен `internal/bim/ifcgeometry`.
- Реализовано извлечение геометрии для ограниченного IFC-подмножества.
- Поддержан `IfcFacetedBrep` через `IfcClosedShell`/`IfcFace`/`IfcFaceOuterBound`/`IfcPolyLoop`/`IfcCartesianPoint`.
- Поддержан `IfcTriangulatedFaceSet` с `IfcCartesianPointList3D` и `CoordIndex`.
- Полигональные контуры триангулируются детерминированным fan-подходом для поддерживаемого простого случая.
- Выход — triangle mesh с vertices/indices, совместимый с существующим mesh/GLB pipeline.
- Добавлены unit tests.

## Что НЕ заявляется
Это не полная реализация IFC geometry engine.

Пока отсутствуют или требуют отдельной реализации:
- полноценная `IfcProductDefinitionShape` → `IfcShapeRepresentation` привязка;
- `IfcObjectPlacement` и полная цепочка local→world coordinates;
- swept solids;
- CSG/boolean geometry;
- mapped representations;
- curved geometry;
- advanced BRep;
- полная schema coverage;
- production tessellation quality для произвольных IFC-файлов;
- BVH/accelerated ray picking;
- streaming/LOD для очень больших моделей.

## Проверки
Runtime 13 domain/geometry tests проходят в доступной среде. Полный production runtime не подтверждён, поскольку Docker/Compose недоступен, а полный Go dependency build ограничен сетевым окружением.

## Следующий шаг
Stage 11 Production Hardening. В BIM-подсистеме сначала реализовать ProductDefinitionShape/ShapeRepresentation/ObjectPlacement → world coordinates → per-element mesh batches → GlobalId/triangle ranges → точный ray picking.

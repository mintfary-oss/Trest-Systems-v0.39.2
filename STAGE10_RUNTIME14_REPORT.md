# Stage 10 — Runtime 14: IFC Placement + Production Picking Foundation

## Цель
Устранить один из ключевых ограничителей Runtime 13: корректно связать типовой IFC Product → ProductDefinitionShape → ShapeRepresentation → geometry item и перевести локальную геометрию в мировую систему координат через цепочку IfcLocalPlacement.

## Реализовано
- `internal/bim/ifcgeometry/placement.go`
  - 4x4 affine matrices;
  - `IfcAxis2Placement3D`;
  - `IfcLocalPlacement`;
  - рекурсивный `PlacementRelTo`;
  - защита от циклов;
  - нормализация осей и построение правой системы координат.
- `internal/bim/ifcgeometry/product.go`
  - ProductDefinitionShape → ShapeRepresentation → Items;
  - выбор поддерживаемой геометрии;
  - применение полного world transform к вершинам.
- `internal/bim/ifcgeometry/scene.go`
  - детерминированная сборка нескольких продуктов;
  - `GlobalId → triangle range`;
  - сохранение product entity ID.
- `internal/bim/ifcgeometry/bvh.go`
  - AABB BVH;
  - nearest-hit traversal;
  - Möller–Trumbore ray/triangle intersection;
  - реальный triangle picking на backend geometry layer.

## Тесты
PASS:
- placement chain;
- nested placement;
- ProductDefinitionShape/ShapeRepresentation linkage;
- world-coordinate mesh transformation;
- GlobalId triangle ranges;
- BVH nearest triangle;
- IFC parser;
- IFC semantic;
- converter;
- geometry diff.

Полный `go test ./internal/bim/...` в текущем окружении не проходит только из-за отсутствующего `go.sum`/недоступного скачивания pgx; это инфраструктурное ограничение среды, а не объявленный PASS полного набора.

## Ограничения
Runtime 14 всё ещё не является полным IFC geometry engine. Не реализованы в полном объёме swept solids, CSG/boolean, mapped representations, curved geometry, advanced BRep, schema-wide representation selection и streaming/LOD.

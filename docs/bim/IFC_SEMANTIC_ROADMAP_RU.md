# IFC Semantic Layer — roadmap

1. STEP entity parser — выполнено.
2. IFC entity index по ID/type — следующий шаг.
3. Reference graph (`#123` -> `#456`) — следующий шаг.
4. Schema-aware decoding для основных строительных сущностей.
5. `IfcProject`, `IfcSite`, `IfcBuilding`, `IfcBuildingStorey`, `IfcSpace`.
6. Elements: wall/slab/roof/door/window/column/beam.
7. Property sets и quantities.
8. Object placement и локальные координаты.
9. Geometry extraction в внутренний BIM mesh.
10. Связка semantic element ↔ `bim_elements`.
11. Viewer tree/properties/selection.
12. Semantic geometry diff по IFC GlobalId.

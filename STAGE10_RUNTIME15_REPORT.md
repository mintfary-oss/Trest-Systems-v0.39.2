# Trest Systems — Stage 10 Runtime 15

## Geometry Engine

Runtime 15 расширяет self-hosted IFC tessellation без внешнего IFC SDK:

- `IfcPolygonalFaceSet` + `IfcIndexedPolygon`;
- `IfcExtrudedAreaSolid`;
- `IfcRectangleProfileDef`;
- `IfcArbitraryClosedProfileDef` с `IfcPolyline`;
- extrusion caps + side walls;
- сохранён полный `IfcLocalPlacement` world-transform pipeline;
- сохранён ProductDefinitionShape → ShapeRepresentation → geometry linkage;
- deterministic mesh output.

## Validation

PASS:
- `go test ./internal/bim/ifcgeometry`
- `go test ./internal/bim/ifc`
- `go test ./internal/bim/ifcsemantic`
- `go test ./internal/bim/converter`
- `go test ./internal/bim/diff`
- `go test ./internal/bim/storage`
- `gofmt` for modified Go sources.

Not claimed:
- complete IFC schema coverage;
- arbitrary swept/BRep/CSG/curved geometry;
- production validation against all commercial IFC exporters;
- Docker/production runtime E2E in this environment.

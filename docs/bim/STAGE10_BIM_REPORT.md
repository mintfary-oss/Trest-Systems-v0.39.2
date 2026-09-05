# Stage 10 — 3D/BIM Report

**Status:** implemented at source/domain/API/migration level.

## Implemented

- Unified BIM model registry linked to `projects` and optionally `project_versions`.
- BIM model lifecycle: `draft`, `review`, `approved`, `archived`.
- Versioned BIM artifacts with source/geometry URIs, manifests and checksums.
- Element registry with external IDs, types, properties and geometry metadata.
- Import/export job registry for IFC, glTF/GLB, OBJ and DXF.
- Planned-vs-actual progress snapshots linked to projects and BIM versions.
- API endpoints for BIM model listing/creation and progress listing/creation.
- Domain validation tests for formats, status, versions, progress and import/export operations.

## API

- `GET /api/v1/bim/models`
- `POST /api/v1/bim/models`
- `GET /api/v1/bim/progress`
- `POST /api/v1/bim/progress`

## Safety / integrity

- BIM status is explicit and auditable.
- Progress is constrained to 0–100.
- Version numbers are positive and unique per model.
- Element external IDs are unique inside a BIM model version.
- Storage paths and geometry are metadata references; no destructive filesystem operation is performed by this stage.

## Validation

`go test ./internal/bim` passes in the available environment.

Full production runtime, database migration execution, and complete 3D viewer/import/export integration remain environment-dependent and require the target Docker/PostgreSQL host.

## Продолжение Stage 10 — BIM API expansion

Добавлены API для:
- списка и создания версий BIM-модели;
- списка и создания BIM-элементов;
- постановки import/export задач в очередь;
- просмотра истории import/export задач по проекту.

Новые маршруты:
- `GET /api/v1/bim/model-versions`
- `POST /api/v1/bim/model-versions`
- `GET /api/v1/bim/elements`
- `POST /api/v1/bim/elements`
- `GET /api/v1/bim/exchanges`
- `POST /api/v1/bim/exchanges`

Операции обмена пока регистрируются как задачи `queued`; фактический IFC/glTF/OBJ/DXF converter worker и браузерный 3D viewer остаются следующим шагом. Автоматическое удаление или перезапись исходных BIM-файлов не выполняется.


## Stage 10 continuation — queue and converter foundation

Added a deterministic OBJ parser and glTF 2.0 JSON exporter under `internal/bim/converter`. Added a PostgreSQL-backed queue processor under `internal/bim/queue` and connected the worker to process queued BIM exchange jobs when `DATABASE_URL` is configured. The current worker adapter intentionally supports only OBJ import to glTF; IFC/DXF/GLB and export adapters remain explicit pending work. A minimal self-hosted viewer shell is provided at `web/bim-viewer/index.html`.

Environment note: full repository builds may still be blocked where external Go modules cannot be downloaded; this does not imply production build success.

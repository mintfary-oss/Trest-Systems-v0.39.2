# Stage 4 — Constructor

## Implemented

- Object type catalog with parameter schemas.
- Material catalog with categories, units and parameters.
- Engineering system catalog.
- Finish catalog with optional material references.
- Project constructor fields: object type, location, parameters, architecture version and BIM URL.
- Immutable project version records with JSON snapshots.
- Project version workflow: `draft -> review -> approved`, plus `archived`.
- Authorization: project owner or admin can read/create version records; only admin can approve an architecture version.
- Catalog and version API endpoints.
- Seed data for common residential/commercial construction concepts.
- Route-level authorization tests.

## API

- `GET /api/v1/catalog/object-types`
- `GET /api/v1/catalog/materials?category=...`
- `GET /api/v1/catalog/engineering-systems`
- `GET /api/v1/catalog/finishes`
- `GET /api/v1/projects/{projectID}/versions`
- `POST /api/v1/projects/{projectID}/versions`
- `POST /api/v1/projects/{projectID}/versions/{version}/status`

## Safety/domain constraints

Architecture versions are stored as explicit snapshots. Approval is a separate transition and is not performed by AI. Legal/financial data remains outside autonomous AI control.

## Validation

- `gofmt` applied to new/changed Go files.
- Constructor route authorization tests pass when the package can be compiled.
- Full repository Go tests require external Go modules that are unavailable in the current offline environment.
- Docker Compose runtime validation requires Docker, which is not installed in the current environment.

# Stage 6 — Contractors

## Implemented
- Contractor application workflow with pending/approved/rejected/withdrawn states.
- Contractor profile with experience and service-area JSON.
- Competencies with level 1–5 and evidence snapshot.
- Admin verification/suspension workflow.
- Contractor discovery endpoint with status and competency filters.
- Deterministic eligibility domain helper for verified+active contractors, competencies and geography.

## API
- `POST /api/v1/contractors/applications`
- `POST /api/v1/contractors/profile`
- `GET /api/v1/contractors?status=verified&competency=...`
- `POST /api/v1/contractors/{contractorID}/competencies`
- `POST /api/v1/contractors/verify` (admin)

## Validation
- Pure contractor eligibility tests pass.
- Full API/database integration and Docker runtime validation require external dependencies and a Docker-capable Linux host.

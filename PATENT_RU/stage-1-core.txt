# Stage 1 — Core

Stage 1 adds the first executable backend foundation:

- PostgreSQL schema for users, projects, objects and orders.
- Go database connector using pgx.
- Idempotent SQL migration runner.
- REST read endpoints under `/api/v1`.
- `/health` and `/ready` endpoints.
- `trestctl api` and `trestctl migrate` commands.

This stage does not yet implement authentication, write endpoints, payments, contracts, matching, quality decisions or autonomous AI actions.

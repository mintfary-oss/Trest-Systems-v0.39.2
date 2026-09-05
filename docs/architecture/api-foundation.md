# API foundation

Base endpoints:

- `GET /health` — process health.
- `GET /ready` — database readiness.
- `GET /api/v1/users` — users (foundation read model).
- `GET /api/v1/projects` — projects.
- `GET /api/v1/orders` — orders; optional `?status=` filter.

Authentication and authorization are intentionally deferred to the next security stage. Do not expose the foundation API directly to an untrusted network in production.

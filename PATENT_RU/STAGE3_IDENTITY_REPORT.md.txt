# Stage 3 — Identity

Implemented:
- organizations and organization memberships;
- role/permission catalog and role mappings;
- organization verification workflow for admins;
- authenticated organization endpoints;
- permissions endpoint;
- role guard middleware;
- migration `0004_identity.sql`;
- API wiring with `TREST_AUTH_SECRET`.

Runtime Docker verification remains pending until a Docker-capable Linux host is available.

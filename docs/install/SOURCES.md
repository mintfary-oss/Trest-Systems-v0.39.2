# Primary references consulted for this patch

- Docker Compose up: https://docs.docker.com/reference/cli/docker/compose/up/
- PostgreSQL 16 psql (SQL input vs -c, ON_ERROR_STOP): https://www.postgresql.org/docs/16/app-psql.html
- Open WebUI environment configuration and persistence: https://docs.openwebui.com/reference/env-configuration/
- Open WebUI headless admin creation: https://docs.openwebui.com/features/authentication-access/rbac/roles/
- Open WebUI offline limitations: https://docs.openwebui.com/tutorials/maintenance/offline-mode/
- Next.js published patch baseline: https://nextjs.org/blog/security-update-2025-12-11

Server diagnosis basis: user-provided installation logs in this conversation. Their PASS/HTTP 200 results describe the manually repaired previous installation, not this new package. No credentials from those logs were copied into this release.

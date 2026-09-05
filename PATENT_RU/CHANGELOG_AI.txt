# Changelog AI

## 2026-09-02

### Добавлено
- Документация проекта.
- Go-скелет `trestctl`.
- Docker Compose для PostgreSQL, Redis, MinIO и Nginx.
- `.env.example` и `.gitignore`.

### Архитектура
Self-hosted Go + Docker Compose вместо обязательной облачной инфраструктуры.

### Известные ограничения
`install`, `repair`, `backup`, `update`, полноценные API/web/DB/AI/3D ещё находятся в разработке.

## 2026-09-02 — Stage 2 Docker unified stack
- Added real Go API entrypoint with migrations and readiness endpoint.
- Added worker health service.
- Unified Compose stack now includes PostgreSQL, Redis, MinIO, Go API, worker, real magasin-777 marketplace API, real Next.js web, Ollama, Super Sistema WebUI and Nginx.
- Added container healthchecks and reverse-proxy routing.

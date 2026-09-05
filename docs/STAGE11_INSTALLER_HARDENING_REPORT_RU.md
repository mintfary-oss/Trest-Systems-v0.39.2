# Stage 11 — Installer/Lifecycle Hardening

## Выполнено
- `trestctl start` — `docker compose up -d`.
- `trestctl stop` — остановка compose-сервисов.
- `trestctl restart` — перезапуск compose-сервисов.
- `trestctl logs [service...]` — просмотр последних логов.
- `trestctl backup [name]` — PostgreSQL dump в `.trest/backups/`.
- `trestctl restore <file.sql>` — восстановление через `psql` только при явном `TREST_CONFIRM_RESTORE=YES`.
- `trestctl update` — pull образов + запуск сервисов.
- `trestctl repair` — безопасная диагностика `compose ps --all` без разрушительных действий.
- версия `trestctl` обновлена до `v0.28-stage11-installer-hardening`.

## Ограничение проверки
В текущем окружении Docker/Compose и доступ к Go module registry отсутствуют, поэтому реальный Linux/Docker E2E не подтверждён. Код подготовлен, но production readiness по P0 не объявляется.

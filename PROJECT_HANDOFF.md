# PROJECT_HANDOFF.md — точка передачи проекта между сессиями AI

## Порядок старта новой сессии
1. Прочитать `PROJECT_HANDOFF.md`.
2. Прочитать `CURRENT_STATE.md`.
3. Прочитать `NEXT_STEPS.md`.
4. Прочитать конец `AI_CONVERSATION_LOG.md`.
5. Проверить фактические migrations/internal/cmd/deployments/web.
6. Не повторять завершённые Runtime 1–13 без причины.
7. Следовать Stage 11 P0/P1 из `NEXT_STEPS.md`.
8. Перед каждым релизом обновлять continuity и `PATENT_RU`.

## Проект
**Trest Systems** — self-hosted строительная платформа с единым Go core/API/CLI, web, marketplace, constructor, estimates/orders, contractors, suppliers, ratings, AI/agents и BIM/IFC.

## Сохранённые исходные подсистемы
- `magasin-777/`
- `proektirovka-sdaniy/`
- `super-sistema/`
- `trest-installer/`

## Контрольная точка
**Последний релиз:** `v0.32-stage11-full-e2e`.

## Текущее состояние
Engineering MVP / Pre-Beta, ориентировочно 60–65% общей готовности. Backend/domain значительно опережает productization. Production runtime не подтверждён.

## Runtime 13
IFC geometry extraction foundation реализована для FacetedBrep и TriangulatedFaceSet. Результат — deterministic triangle mesh. Это НЕ полный IFC geometry engine.

## Ключевые ограничения
- Docker/Compose E2E не выполнен в текущей среде.
- Полный Go build/test/vet не подтверждён из-за окружения.
- Реальное GPU training не выполнено здесь.
- Installer ещё не полный production installer.
- Unified production frontend ещё не завершён.
- IFC placement/world coordinates/full tessellation ещё не завершены.

## Следующий этап
**Stage 11 — Production Hardening / Release Candidate.**
Приоритет: runtime proof → installer → prebuilt release → clean-host → security → E2E → frontend integration → BIM hardening.

## Архитектурные правила
- Self-hosted first; no mandatory cloud.
- Frontend → API, no direct DB.
- Approved estimates versioned/immutable.
- Contractual events auditable.
- Financial operations idempotent.
- AI recommendations stored with input/result/model version/confidence.
- AI cannot autonomously modify legal/financial records.
- Schema changes require migrations.
- New API endpoints require docs/tests.
- No secrets in source/docs.
- Preserve real functionality.

## Опасные операции
Удаление/recreate DB или volumes, public exposure, firewall/SSH/domain/TLS changes и иные необратимые операции требуют подтверждения пользователя.

## Release preservation rule
`PATENT_RU/` — обязательная отдельная папка в каждом FULL ZIP. `презентация/` также сохраняется. Никакие существующие пути не удаляются молча.

## Stage 11 — Security + Backup hardening
- Latest stage: security baseline + backup/restore hardening. Read `STAGE11_SECURITY_BACKUP_REPORT_RU.md`.


## v0.31 — E2E/Release hardening
- `trestctl migrate` now exposed.
- Root CI uses Go 1.23.x; release workflow uses root `go.mod`.
- `scripts/e2e-smoke.sh` added for Docker Compose disposable smoke tests.
- `scripts/e2e-full.sh` added for auth/project/estimate/order/BIM/backup flow.
- `POST /api/v1/projects` is now explicitly registered in the unified router.
- `docs/install/E2E_SMOKE_RU.md` added.
- Docker E2E is not claimed as passed until executed on a host with Docker/Compose.

## Latest checkpoint — v0.33
Security/IDOR hardening applied to users/projects/orders list endpoints; clean-host checker added. Runtime remains unverified until Docker/Linux CI.

## Latest checkpoint — v0.34

Stage 11 security upload/rate-limit hardening добавил IFC upload size/path guards, API rate limiting и dependency audit helper. Следующая обязательная проверка — Linux/Docker runtime, oversized/malicious IFC E2E, backup/restore drill, dependency audit и clean-host gate.


## Latest checkpoint — v0.35
Stage 11 security release gate prepared: bounded rate limiter, IFC malicious/traversal fixtures, oversized/traversal E2E regressions, disposable backup/restore drill, and unified security-release gate. Docker/Compose and full Go/dependency runtime remain unverified in the current environment.


## Stage 10 — Runtime 14 (v0.37)
- Added IFC ProductDefinitionShape → ShapeRepresentation → geometry linkage.
- Added IfcAxis2Placement3D + recursive IfcLocalPlacement world transforms.
- Added deterministic GlobalId → triangle ranges for concatenated product scenes.
- Added backend BVH and real ray/triangle intersection for nearest-hit picking.
- Tests for placement, product linkage, scene ranges and BVH PASS.
- Full repository runtime remains pending until CI/Docker host is available.


## Runtime 15 update
Stage 10 Runtime 15 completed: PolygonalFaceSet and ExtrudedAreaSolid geometry support added with regression tests. See STAGE10_RUNTIME15_REPORT.md.


## Последний артефакт

`v0.38.1-verified-release`; подробный машинный отчёт: `verification/VERIFICATION_SUMMARY.json`. Production Ready: НЕТ — ожидаются runtime gates.

## Checkpoint v0.39-stage11-automatic-installer

Главная команда установки: `sudo ./install.sh ...`. Установщик обязан сам проверить и подготовить хост, использовать prebuilt Go binaries, проверить Compose/migrations/health/ports/web/TLS/logs и создать TXT/JSON report. Не объявлять production PASS без реального отчёта целевого Docker-хоста.

# HANDOFF v0.39.1 — PREBUILT AUTO INSTALLER

Последняя полная сборка: `v0.39.1-stage11-prebuilt-auto-installer`. Перед изменениями читать `STAGE11_PREBUILT_AUTO_INSTALL_FIX_REPORT_RU.md` и `verification/v0.39.1/STATIC_VALIDATION.json`. Не удалять `release/bin/linux/amd64`, `PATENT_RU`, `READABLE_TXT_RU` и `презентация`.

Корневые зависимости заменены локальными compatibility modules в `third_party/`, что позволяет воспроизводимо тестировать и собирать проект без сети. API/Worker/trestctl используют system libpq; Docker runtime основан на Debian с `libpq5`. Полный Docker E2E в среде сборки не выполнялся и должен быть проведён на сервере.


# Trest Systems — Спецификация патентных фигур

## FIG.1 — Общая архитектура
Показать Unified Go Core/API, PostgreSQL, Redis, MinIO, web, worker, marketplace, CAD/GOST, local AI и training center.

## FIG.2 — Version lineage
Показать ProjectVersion → EstimateVersion → approved estimate → Order.

## FIG.3 — Eligibility
Показать входные требования проекта → verification → competency/region checks → eligible/review/reject.

## FIG.4 — Agent gate
Показать agent request → allowlist → approval flag → deny / human approval / execute.

## FIG.5 — AI audit lineage
Показать input snapshot → model version → inference → result → confidence → human decision → audit event.

## FIG.6 — Local AI topology
Показать application server → controlled Ollama client → local model runtime → response.

## FIG.7 — Training
Показать source dataset → preparation → JSONL/manifest → QLoRA → evaluation → registry/promotion.

## FIG.8 — CAD/GOST
Показать project parameters → Go generator → DXF/GOST document → project artifact.

## FIG.9 — BIM
Добавить только после фактической реализации Stage 10.

## FIG.10 — Self-hosted deployment
Показать control CLI → services → databases/storage/AI → health/diagnostics.

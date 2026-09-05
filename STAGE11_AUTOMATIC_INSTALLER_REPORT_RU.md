# Stage 11 — Automatic Installer Report

Версия: `v0.39-stage11-automatic-installer`

## Выполнено

- Добавлен Linux bootstrap installer `install.sh`.
- Добавлена Windows PC установка `install.ps1`.
- Автоматизированы preflight, package manager, Docker/Compose, firewall, secrets, Compose, migrations, health/readiness, ports, web, TLS, logs и reports.
- Caddy edge добавлен для автоматического публичного HTTPS и renewal.
- Go-контейнеры переведены на prebuilt binaries; Go не компилируется на сервере клиента.
- Добавлен независимый stdlib `trest-bootstrap` и multi-platform builds.
- Удалён небезопасный fallback `[REMOVED_INSECURE_DEFAULT]`; `ADMIN_PASSWORD` и `SECRET_KEY` обязательны.
- Worker получает обязательный `DATABASE_URL`.
- Go directives нормализованы до 1.23.0.
- Добавлен точный merge-conflict scanner.
- Добавлены static/mocked installer verification и production-host post-install gate.

## Честная граница проверки

Статические тесты, компиляция и архивная целостность выполняются при сборке релиза. Реальные Docker E2E, выпуск публичного сертификата и backup/restore должны выполняться установщиком на целевом Linux-хосте и отражаются в его итоговом отчёте.

## Дополнение

Установщик теперь вызывает обязательный `runtime-acceptance.sh`, который автоматически запускает smoke/full E2E и backup/restore drill после старта стека.

## Объём кода v0.39

- Программный код по основной методике: **28 586 строк**.
- Код вместе с PowerShell/Batch/NSIS/Dockerfile: **29 521 строк**.
- Подробности: `CODE_LINE_COUNT_V0.39.json`.

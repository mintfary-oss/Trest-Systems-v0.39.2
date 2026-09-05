# E2E smoke test

`./scripts/e2e-smoke.sh` проверяет Compose-конфигурацию и запускает полный стек на тестовом Docker-хосте.

Проверяются:
- Docker Compose config;
- сборка API/Worker/Web/marketplace;
- PostgreSQL readiness;
- Redis;
- marketplace API health;
- публичная точка Nginx `/` и API `/ready`;
- автоматическое удаление тестовых контейнеров после завершения.

Скрипт не должен запускаться на production-хосте с существующими данными без отдельного compose project/окружения.

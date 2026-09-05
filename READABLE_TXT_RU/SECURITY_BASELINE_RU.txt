# Security baseline Trest Systems

## Цель
Минимальный gate перед публикацией release: отсутствие очевидных секретов в исходниках, наличие контейнерных артефактов, отсутствие world-writable файлов и отдельная проверка host bindings.

## Проверка

```bash
./scripts/security-baseline.sh
```

Проверка не заменяет полноценный security audit, SAST/DAST, dependency scanning, penetration test или review production firewall.

## Секреты
Секреты должны передаваться через environment/secret manager. `.env.example` содержит только шаблонные значения.

## Docker
Публикация портов наружу требует отдельного решения администратора. Production deployment должен ограничивать доступ firewall/reverse proxy и не публиковать PostgreSQL/Redis без необходимости.

## Backup
После создания дампа:

```bash
trestctl backup
./scripts/backup-verify.sh .trest/backups/<file>.sql
```

`restore` является разрушительной операцией и требует `TREST_CONFIRM_RESTORE=YES`.

## Ограничение
В текущей среде Docker недоступен, поэтому runtime security/E2E здесь не подтверждается.

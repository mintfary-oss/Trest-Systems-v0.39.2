# Backup / Restore

## Backup

```bash
trestctl backup
```

Дамп PostgreSQL сохраняется в `.trest/backups/` с правами `0600`.

Проверка:

```bash
./scripts/backup-verify.sh .trest/backups/20260903-120000.sql
```

## Restore

Восстановление специально защищено явным подтверждением:

```bash
TREST_CONFIRM_RESTORE=YES trestctl restore .trest/backups/20260903-120000.sql
```

Перед restore рекомендуется создать свежий backup.

## Production policy
- хранить backup вне основного host;
- шифровать backup at rest;
- регулярно выполнять test restore;
- хранить несколько поколений;
- проверять SHA-256 и размер;
- документировать RPO/RTO.

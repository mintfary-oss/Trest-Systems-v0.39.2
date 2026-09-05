# Эксплуатация

## Установка

```bash
trestctl install
```

## Проверка

```bash
trestctl doctor
trestctl status
```

## Логи

```bash
trestctl logs
trestctl logs --service api
trestctl logs --follow
```

## Управление

```bash
trestctl start
trestctl stop
trestctl restart
```

## Backup/update

```bash
trestctl backup create
trestctl update
```

Backup должен включать PostgreSQL, конфигурацию, пользователей, проекты, сметы, документы, BIM-модели, фотографии и аудит.

Отчёты: `.trest/reports/` и `.trest/diagnostics/`.

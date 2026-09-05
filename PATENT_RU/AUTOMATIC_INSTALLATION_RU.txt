# Полностью автоматическая установка Trest Systems

## Одна команда

После распаковки релиза на Linux-сервере:

```bash
sudo ./install.sh --domain app.example.com --email admin@example.com --non-interactive
```

Для локального сервера без домена:

```bash
sudo ./install.sh --tls off --non-interactive
```

## Что выполняется автоматически

1. Проверка root/sudo, ОС, архитектуры, CPU, RAM, диска, DNS и сети.
2. Установка или обновление системных утилит, Docker Engine и Docker Compose v2.
3. Проверка занятости портов 80/443 и безопасная настройка активного firewall.
4. Резервирование предыдущего `.env` при обновлении.
5. Генерация криптографически случайных паролей и секретов; небезопасных defaults нет.
6. Проверка SHA-256 и наличия prebuilt Go binaries для архитектуры сервера.
7. Docker Compose validation, загрузка образов и быстрая сборка контейнеров без компиляции Go.
8. Запуск PostgreSQL, Redis, MinIO, API, Worker, Web/NGINX и Caddy edge.
9. Идемпотентное применение SQL migrations с таблицей `schema_migrations` и контролем checksum.
10. Ожидание readiness, проверка контейнеров, портов, HTTP/HTTPS и веб-интерфейса.
11. При домене Caddy автоматически получает и продлевает доверенный ACME-сертификат.
12. Проверка TLS chain без `-k`, security headers, DNS и доступности URL.
13. Анализ журналов контейнеров по критическим сигнатурам.
14. Итоговый TXT/JSON-отчёт и отдельная папка diagnostics.

## Отчёты

- `/var/log/trest-systems/install-*.txt`
- `/var/log/trest-systems/install-*.json`
- `/var/log/trest-systems/install-*.log`
- `/var/log/trest-systems/diagnostics-*/`
- `/var/lib/trest-systems/last-install-report.*`

Если ошибок нет, отчёт содержит: `ОШИБКИ: нет`.

## TLS и браузеры

Доверенный сертификат без предупреждений браузера автоматически выпускается только при наличии публичного домена, корректного DNS, доступных портов 80/443 и email для ACME. Без домена установщик запускает HTTP по IP и явно пишет это в отчёте. Режим `--tls internal` создаёт внутреннюю CA, но клиентские устройства должны доверять её корневому сертификату.

## Обновление, ремонт и диагностика

```bash
sudo ./install.sh --update --domain app.example.com --email admin@example.com --non-interactive
sudo ./install.sh --repair --domain app.example.com --email admin@example.com --non-interactive
sudo ./install.sh --doctor --install-dir /opt/trest-systems
```

## После первого входа

Начальные реквизиты создаются в `/var/lib/trest-systems/credentials/admin.txt` с правами `0600`. Пароль необходимо изменить после первого входа.

## Обязательный runtime acceptance

После запуска установщик автоматически выполняет `scripts/installer/runtime-acceptance.sh`: smoke E2E, full E2E, backup→restore drill, проверку TLS, состояний контейнеров, прав `.env` и критических сигнатур журналов. Нулевой код всех проверок обязателен; иначе установка помечается как неуспешная и сохраняет логи. На сервере Go не компилируется.

## Порядок безопасного запуска

Образы готовятся заранее, затем PostgreSQL/Redis/MinIO запускаются до API и Worker. Миграции применяются перед запуском прикладных сервисов. Redis получает обязательный случайный пароль, MinIO и клиентское хранилище используют согласованные credentials. Для существующей схемы migration runner формирует checksum baseline и не повторяет уже выполненные `CREATE`.

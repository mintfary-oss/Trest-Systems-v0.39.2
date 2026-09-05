# Быстрая установка v0.39.1 с готовыми Go-бинарниками

Поддерживаемая цель этого пакета: Linux x86_64/amd64. Проверка:

```bash
uname -m
```

Ожидается `x86_64`.

## Проверка без изменений

```bash
sudo ./install.sh --dry-run --tls off --non-interactive --no-firewall
```

## Установка с доменом и доверенным HTTPS

```bash
sudo ./install.sh \
  --domain app.example.com \
  --email admin@example.com \
  --non-interactive
```

DNS A-запись домена должна указывать на IP сервера, а внешние TCP-порты 80 и 443 должны быть разрешены в firewall хостинг-провайдера.

## Установка по IP без TLS

```bash
sudo ./install.sh --tls off --non-interactive
```

Go toolchain на сервере не нужен. API и Worker копируются из `release/bin/linux/amd64/` в runtime-образы. Установщик проверяет SHA-256 до запуска.

Отчёты:

```text
/var/log/trest-systems/install-*.txt
/var/log/trest-systems/install-*.json
/var/log/trest-systems/install-*.log
/var/lib/trest-systems/last-install-report.txt
/var/lib/trest-systems/last-install-report.json
```

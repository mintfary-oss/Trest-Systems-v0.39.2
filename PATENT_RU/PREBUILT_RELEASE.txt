# Prebuilt Release Policy

## Цель

Production-сервер **не должен компилировать Go-проект при обычной установке**. Это устраняет задержку в десятки минут/час и уменьшает зависимость от сети, Go toolchain и внешнего module proxy.

## Release Factory

CI/release machine выполняет один раз:

```text
checkout
  ↓
resolve dependencies
  ↓
run tests/vet
  ↓
compile Go binaries
  ↓
build web/FastAPI artifacts
  ↓
build OCI images
  ↓
create checksums + manifest
  ↓
package release
```

Клиентский сервер выполняет:

```text
verify checksum
  ↓
install binaries/images
  ↓
load/start
  ↓
migrations
  ↓
health/readiness
```

## Артефакты

Рекомендуемая структура production bundle:

```text
release/
  manifest.json
  checksums.sha256
  bin/
    linux-amd64/trest
    linux-amd64/trestctl
    linux-amd64/trest-api
    linux-amd64/trest-worker
  images/
    trest-images.tar.zst
  migrations/
  deployments/
  docs/
```

## Docker

Если используются Docker/OCI images, production compose должен использовать уже собранные images и `docker load`, а не `docker compose build`.

## Offline installation

Release bundle должен быть самодостаточным настолько, насколько позволяют лицензии и размер артефактов. Установка не должна требовать скачивания Go modules.

## Fallback

Исходный `go build` остаётся development/CI fallback. Он не является штатным production install path.

## Проверка

Перед выпуском:

- SHA-256 manifest;
- version consistency;
- migration sequence;
- binary `--version`;
- smoke tests;
- image inspection;
- clean-host installation test.

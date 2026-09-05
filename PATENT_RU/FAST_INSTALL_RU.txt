# Быстрая установка Trest Systems

## Главное

Не запускать на production-сервере:

```bash
go mod download
go build ./...
docker compose build
```

Эти операции предназначены для CI/release factory.

## Production

Используется заранее собранный release bundle:

```text
Trest release
├── готовые Go-бинарники
├── готовые web artifacts
├── готовые Docker/OCI images
├── migrations
├── config templates
├── checksums
└── installer
```

После распаковки сервер только устанавливает/запускает готовые артефакты.

## Почему это быстрее

Компиляция и загрузка зависимостей происходят один раз на release machine. Сервер клиента не зависит от доступности `proxy.golang.org` и не тратит время на компиляцию.

## Важное

Текущий архив содержит исходный код и release tooling, но **не содержит подтверждённых production Go-бинарников**, поскольку в среде подготовки релиза отсутствует доступ к Go module registry. Поэтому нельзя честно утверждать, что сервер уже сейчас может установить этот архив без сборки. Этот документ фиксирует целевую production-модель и release gate.

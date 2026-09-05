# Trest Systems 0.39.2 — полный исправленный пакет

**Статус: исправленный исходный релиз с проверенными Go-бинарниками; release candidate. Не сертификат «100% готов без ошибок».** Авторитетный отчёт новой сборки: `verification/v0.39.2/VERIFICATION.json`. Старые отчёты и юридические/патентные материалы сохранены как исторические документы.

## Что внутри

Полный исходный проект v0.39.1 с исправлениями, 7 новых Go-программ для Linux/amd64, единый установщик, Windows/WSL launcher, миграции, Docker Compose, frontend/API/worker, исходные тесты, новые regression-тесты, `PATENT_RU`, `READABLE_TXT_RU`, PDF-презентация в `презентация/`, документация, проверки SHA-256 и экспортёр офлайн-пакета. Копии изменённых старых файлов находятся в `archive/v0.39.1-changed/` с расширением `.original` у исполняемых исходников.

## Новая установка на Linux x86_64

Распакуйте ZIP, перейдите в каталог `Trest-Systems-v0.39.2` и запустите:

```bash
sudo bash install.sh --tls off --non-interactive
```

Никаких GitHub-токенов не нужно. На новой Ubuntu/Debian установщик проверяет ресурсы и инструменты, при необходимости устанавливает Docker; сохраняет существующие пароли; выбирает свободные порты, не останавливая чужие сервисы; создаёт конфигурацию; готовит образы; запускает инфраструктуру; выполняет единственный мигратор; готовит модели; ждёт healthchecks; проверяет HTTP и генерацию/embeddings; делает SQL backup/restore в отдельной временной БД и пишет отчёт.

Проверка пакета без установки:

```bash
bash install.sh --dry-run --install-dir /tmp/trest-check
```

Свой домен + доверенный HTTPS:

```bash
sudo bash install.sh --domain app.example.com --email admin@example.com --tls auto --non-interactive
```

Начальные credentials: `/var/lib/trest-systems/credentials/admin.txt` (0600). Отчёты: `/var/log/trest-systems/install-*.json`, `.txt`, `.log`. Пароли не выводятся в отчёт. `.env`: `/opt/trest-systems/.env`; менять через секретный файл, не коммитить.

## Уже работающий сервер

**Не распаковывать поверх `/opt/trest-systems` и не запускать обновление вслепую.** Прочитайте `docs/install/UPGRADE_RU.md`. Старый общий `public.users/orders` несовместим с полноценными моделями магазина; новый пакет изолирует магазин в `marketplace.*`. Автоматический перенос существующих коммерческих данных намеренно блокируется до отдельной проверки. База и volumes не удаляются.

## Управление

```bash
./trestctl.sh status
sudo ./trestctl.sh doctor
sudo ./trestctl.sh backup
sudo ./trestctl.sh restore-drill
```

`doctor` — HTTP/container/AI smoke, НЕ вся продуктовая E2E. `scripts/e2e-full.sh --confirm-disposable` — только выбранные бизнес-сценарии с тестовыми записями на отдельной тестовой установке.

Windows: `install.ps1 -Tls off -NonInteractive` делегирует в установленную и инициализированную Ubuntu под WSL2. Это НЕ нативная Windows-server установка. Может потребоваться включение WSL2 и перезагрузка. Нативный Windows runtime в текущей среде не тестировался.

## Скорость и офлайн

Go не компилируется на целевом сервере. При первом обычном запуске Docker/Node/Python образы и Ollama-модели всё ещё требуют интернета. После подготовки пакета `scripts/release/export-offline.py` экспортирует images/models для `--offline`; подробности в `docs/install/OFFLINE_RU.md`. **Многогигабайтные образы и веса в текущий ZIP не включены**, это явно отмечено в машинном отчёте.

## Ограничения доказательной базы

В среде сборки нет Docker daemon/PostgreSQL server и недоступна загрузка зависимостей через DNS. Новый полный Compose/E2E/clean-host/TLS/browser/restore runtime здесь не пройден. Go unit/vet/build, offline installer regressions, Python syntax/metadata, shell/JSON/YAML и целостность архива проверяются отдельно и отражаются в отчёте. `SKIP`/`BLOCKED` не равны `PASS`.

См. `docs/install/SECURITY_NOTES_RU.md` перед размещением пользовательских данных. В частности, старые опубликованные секреты нельзя использовать повторно, а HTTP — не защищённый production-доступ.

# Быстрый и полностью офлайн-пакет

Основной ZIP v0.39.2 содержит полный проект и Go-бинарники, **но не многогигабайтные Docker images и модели**. Для первой обычной установки нужен доступ к Docker Registry, Debian, npm, PyPI и Ollama. Повторный запуск не делает принудительный pull/build уже имеющихся образов. API/Worker используют готовые Go-бинарники; Node/Python при первой сборке устанавливают зависимости.

После подготовки именно v0.39.2 на отдельной Linux/amd64 машине:

```bash
sudo python3 scripts/release/export-offline.py --root /opt/trest-systems --output /root/trest-offline-0.39.2
```

Экспорт собирает `images.tar`, `ollama-models.tar`, `bundle.json`, `SHA256SUMS`; `.env`, пароли, SSH-ключи, PostgreSQL/MinIO и чаты не экспортируются. Копируются только Ollama `models`, а не весь домашний каталог Ollama.

На машине без интернета уже должны быть Docker Engine, Compose >=2.24, Python3 и базовые системные инструменты. Установка:

```bash
sudo ./install.sh --offline --image-bundle /root/trest-offline-0.39.2 --tls off --non-interactive
```

Нет образа/модели или не совпала SHA-256 — установка прекращается с FAIL; скрытой онлайн-сборки в --offline нет. Сам bundle не создан и не приложен к текущему ZIP: в среде подготовки нет Docker daemon и доступа к образам вашего сервера. Экспортёр включён как исполняемый инструмент, а не как фиктивный готовый bundle.

В Open WebUI используются локальные Ollama embeddings (`nomic-embed-text`) и `OFFLINE_MODE=true`, `HF_HUB_OFFLINE=1`; плагины, которым нужны неустановленные внешние Python-зависимости, могут не работать в таком режиме. Это не сетевой firewall и не обещание полной работоспособности всех внешних интеграций.

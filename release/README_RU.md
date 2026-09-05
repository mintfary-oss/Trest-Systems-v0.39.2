0.39.2: release/bin/linux/amd64 содержит вновь собранные программы. API/Worker используют CGO/libpq; контейнеры включают libpq5. Контрольные суммы — release/bin/SHA256SUMS.

Этот ZIP содержит исходники и Go-бинарники, но не Docker-образы и веса AI-моделей. Первая обычная установка скачивает их и собирает веб/Python-образы. Повторный запуск использует имеющиеся образы. Для air-gap пакета предусмотрен scripts/release/export-offline.py; успешный экспорт создаёт images.tar, ollama-models.tar, bundle.json и SHA256SUMS. Не называйте обычный ZIP полностью офлайн-пакетом.

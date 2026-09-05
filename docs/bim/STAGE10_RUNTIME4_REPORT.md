# Stage 10 Runtime 4 — BIM exchange lifecycle

## Сделано
- Добавлен GLB 2.0 decoder для первого triangle mesh primitive: POSITION + индексы 8/16/32-bit.
- Viewer теперь принимает `.gltf`, `.json` и `.glb`, без CDN/облака.
- Queue processor получил attempts/max_attempts, started_at, output SHA-256 и JSON manifest.
- Успешный exchange с привязанной BIM-моделью автоматически создаёт следующую `bim_model_versions` с checksum/manifest.
- Добавлены API статуса и безопасной отмены queued job.
- Migration `0014_bim_exchange_runtime.sql` добавляет статус `cancelled` и queue index.

## Проверки
- `go test ./internal/bim/converter ./internal/bim/diff ./internal/bim` — PASS.
- Полный runtime с PostgreSQL/Docker не запускался: в текущей среде нет Docker и доступного registry для внешних Go modules.

## Ограничения
- IFC semantic parser ещё не реализован.
- DXF остаётся консервативным 3DFACE subset.
- Queue пока работает с локальными файловыми URI; object-storage adapter и потоковая обработка больших моделей — следующий блок.
- Project membership/ownership hardening для BIM endpoints остаётся отдельной production-задачей.

# Stage 10 Runtime 3 — BIM Viewer + Geometry Diff API

Добавлено:
- `GET/POST /api/v1/bim/geometry-diffs`;
- сохранение результата geometry diff в `bim_geometry_diffs`;
- полнофункциональный self-hosted WebGL viewer без обязательных внешних CDN;
- загрузка self-contained glTF 2.0;
- orbit camera, zoom, pan и reset;
- отображение базовой статистики модели;
- migration `0013_bim_diff.sql`.

Ограничения:
- viewer сейчас рассчитан на первый mesh primitive и POSITION/indices;
- GLB binary loading и сложные glTF accessors остаются следующим расширением;
- IFC semantic layer пока не реализован;
- API diff принимает геометрию запроса и сохраняет агрегированный результат; загрузка больших mesh payload через API требует дальнейшего object-storage workflow.

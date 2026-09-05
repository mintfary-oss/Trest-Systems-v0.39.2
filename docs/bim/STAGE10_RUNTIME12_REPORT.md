# Stage 10 Runtime 12

Дата: 2026-09-03

## Реализовано
- Детерминированная генерация `MeshElementManifest` из результатов конвертера, связывающая диапазоны индексов треугольников с `external_id` и IFC `GlobalId`.
- API `GET /api/v1/bim/model-versions/manifest` с обязательной проверкой доступа к проекту.
- Manifest API возвращает geometry/source URI и неизменяемый semantic manifest.
- Добавлены unit tests для генерации и triangle-index picking.

## Ограничение
Полноценная IFC BRep/mesh геометрия не реализована этим этапом: manifest строится из geometry batches, предоставленных конвертером. Поэтому автоматическое получение triangle ranges непосредственно из произвольного IFC файла требует следующего geometry runtime.

## Следующее
- IFC geometry extraction / tessellation.
- WebGL CPU ray picking и вызов manifest/pick API.
- Progress lookup для выбранного элемента.
- Docker/MinIO E2E profile.

# STAGE8_RATINGS_REPORT.md

## Status
Stage 8 — Ratings is implemented in the current working tree.

## Database
Migration `0009_ratings.sql` adds:
- ratings for contractors and suppliers;
- rating dimensions JSON;
- aggregate reputation snapshots and history versions;
- reputation bonuses and sanctions with immutable audit records;
- rating disputes and dispute event history;
- anti-abuse uniqueness for one authoritative rating per order/reviewer/target and one open dispute per rating.

## Rules
- Only completed orders can create ratings.
- The project owner (or admin) can submit the rating.
- The rated contractor/supplier must be the verified active party attached to the completed order.
- Ratings can be edited only within 30 days; edits remain auditable through updated timestamps and aggregate history.
- Removing a rating through a resolved dispute hides it instead of deleting its historical record.
- Bonuses/sanctions are explicit admin actions and do not erase rating history.
- Rating/dispute operations do not mutate legal or financial order records.

## API
- `POST /api/v1/ratings`
- `GET /api/v1/ratings?target_type=&target_id=`
- `GET /api/v1/ratings/aggregate?target_type=&target_id=`
- `GET /api/v1/ratings/history?target_type=&target_id=`
- `POST /api/v1/ratings/disputes`
- `POST /api/v1/ratings/disputes/resolve` (admin)
- `POST /api/v1/reputation/actions` (admin)

## Validation
- dependency-free ratings domain tests;
- Go formatting;
- migration sequence/integrity checks;
- Python/YAML/ZIP checks where applicable.
- Full Docker/runtime and unrestricted dependency builds remain pending because the current environment does not provide Docker/Compose and has restricted module downloads.

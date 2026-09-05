# Stage 3 — Marketplace core

Stage 3 introduces the first transactional marketplace primitives:

- project objects;
- estimate items with quantity and unit price;
- order items;
- contractor/supplier bids;
- authenticated creation of customer-owned objects, estimates and orders;
- bids only against published/matching orders.

The database migration is `migrations/0003_marketplace.sql`.

## Endpoints

- `POST /api/v1/objects`
- `GET /api/v1/objects?project_id=<id>`
- `POST /api/v1/estimate/items`
- `GET /api/v1/estimate?project_id=<id>`
- `POST /api/v1/orders`
- `GET /api/v1/order-items?order_id=<id>`
- `POST /api/v1/order-items?order_id=<id>`
- `GET /api/v1/bids?order_id=<id>`
- `POST /api/v1/bids?order_id=<id>`

## Scope boundary

Stage 3 does not implement payment execution, contract signing, quality acceptance, dispute resolution, or autonomous AI decisions.

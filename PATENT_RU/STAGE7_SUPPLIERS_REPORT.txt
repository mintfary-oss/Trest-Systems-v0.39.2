# STAGE7_SUPPLIERS_REPORT.md

## Status
Stage 7 — Suppliers is implemented in the current working tree.

## Database
Migration `0008_suppliers.sql` adds:
- supplier profiles and applications;
- verification/active state;
- categories and delivery regions/terms;
- supplier offers with SKU, unit, price, currency, MOQ, stock and lead time;
- certificate/document metadata;
- supplier offer references from estimate items and orders;
- indexes and uniqueness constraints.

## API
Added:
- `POST /api/v1/suppliers/applications`
- `POST /api/v1/suppliers/profile`
- `GET /api/v1/suppliers`
- `POST /api/v1/suppliers/verify` (admin)
- `POST /api/v1/suppliers/offers`
- `GET /api/v1/suppliers/offers`
- `POST /api/v1/suppliers/offers/{offerID}/publish`
- `POST /api/v1/orders/{id}/supplier-offer`

The order attachment checks owner/admin authorization, published + verified + active supplier state, MOQ and stock before changing the financial order amount.

## Domain
`internal/suppliers/eligibility.go` implements deterministic supplier matching baseline: verified, active, published, category/region compatibility and sufficient stock.

## Validation
- supplier domain tests: available dependency-free tests;
- migration sequence checked;
- Go formatting applied;
- full runtime stack remains pending because Docker/Compose and unrestricted module downloads are unavailable in the current environment.

## Next
Stage 8 — Ratings: contractor/supplier ratings, aggregation/history, bonuses, sanctions and disputes.

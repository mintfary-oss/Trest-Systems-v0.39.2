# Stage 5 — Estimates & Orders

Implemented the TЗ layer for versioned estimates and controlled orders.

## Delivered
- Immutable estimate versions and versioned line items.
- Estimate lifecycle and approval timestamp.
- Order references to estimate versions.
- Planned start/end dates.
- Contractor/supplier assignment fields.
- Idempotent order creation via unique idempotency key.
- Explicit order transition state machine with domain tests.
- Protected estimate/order APIs.

## API
- `POST /api/v1/estimates`
- `POST /api/v1/estimates/approve`
- `POST /api/v1/orders`
- `POST /api/v1/orders/{id}/transition`

## Validation
- Order transitions are validated in `internal/orders`.
- SQL migration is `migrations/0006_estimates_orders.sql`.
- Full Docker/integration validation remains blocked on this environment because Docker is unavailable.

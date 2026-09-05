-- Stage 5: immutable estimate versions, specifications, order lifecycle and idempotency.
CREATE TABLE IF NOT EXISTS estimates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    object_id UUID REFERENCES objects(id) ON DELETE SET NULL,
    title TEXT NOT NULL,
    currency TEXT NOT NULL DEFAULT 'RUB',
    status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','review','approved','archived')),
    current_version INTEGER NOT NULL DEFAULT 1 CHECK (current_version > 0),
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS estimate_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    estimate_id UUID NOT NULL REFERENCES estimates(id) ON DELETE CASCADE,
    version INTEGER NOT NULL CHECK (version > 0),
    status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','review','approved','archived')),
    subtotal NUMERIC(16,2) NOT NULL DEFAULT 0 CHECK (subtotal >= 0),
    total NUMERIC(16,2) NOT NULL DEFAULT 0 CHECK (total >= 0),
    snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    approved_at TIMESTAMPTZ,
    UNIQUE(estimate_id, version)
);
CREATE TABLE IF NOT EXISTS estimate_version_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    estimate_version_id UUID NOT NULL REFERENCES estimate_versions(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    category TEXT NOT NULL DEFAULT 'general',
    unit TEXT NOT NULL DEFAULT 'pcs',
    quantity NUMERIC(16,3) NOT NULL CHECK (quantity >= 0),
    unit_price NUMERIC(16,2) NOT NULL CHECK (unit_price >= 0),
    total NUMERIC(16,2) NOT NULL CHECK (total >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
ALTER TABLE orders ADD COLUMN IF NOT EXISTS estimate_version_id UUID REFERENCES estimate_versions(id) ON DELETE SET NULL;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS planned_start_at TIMESTAMPTZ;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS planned_end_at TIMESTAMPTZ;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS contractor_id UUID REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS supplier_id UUID REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS approved_at TIMESTAMPTZ;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS idempotency_key TEXT;
CREATE UNIQUE INDEX IF NOT EXISTS uq_orders_idempotency_key ON orders(idempotency_key) WHERE idempotency_key IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_estimates_project ON estimates(project_id);
CREATE INDEX IF NOT EXISTS idx_estimate_versions_estimate ON estimate_versions(estimate_id, version DESC);
CREATE INDEX IF NOT EXISTS idx_estimate_version_items_version ON estimate_version_items(estimate_version_id);
CREATE INDEX IF NOT EXISTS idx_orders_estimate_version ON orders(estimate_version_id);
CREATE INDEX IF NOT EXISTS idx_orders_contractor ON orders(contractor_id);

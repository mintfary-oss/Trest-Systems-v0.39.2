-- Stage 3: objects, estimates, order positions and marketplace bids.
CREATE TABLE IF NOT EXISTS estimate_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    object_id UUID REFERENCES objects(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    category TEXT NOT NULL DEFAULT 'general',
    unit TEXT NOT NULL DEFAULT 'pcs',
    quantity NUMERIC(16,3) NOT NULL DEFAULT 0 CHECK (quantity >= 0),
    unit_price NUMERIC(16,2) NOT NULL DEFAULT 0 CHECK (unit_price >= 0),
    approved BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    estimate_item_id UUID REFERENCES estimate_items(id) ON DELETE SET NULL,
    name TEXT NOT NULL,
    unit TEXT NOT NULL DEFAULT 'pcs',
    quantity NUMERIC(16,3) NOT NULL DEFAULT 0 CHECK (quantity >= 0),
    unit_price NUMERIC(16,2) NOT NULL DEFAULT 0 CHECK (unit_price >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS marketplace_bids (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    bidder_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    bid_type TEXT NOT NULL CHECK (bid_type IN ('contractor','supplier')),
    amount NUMERIC(16,2) NOT NULL CHECK (amount >= 0),
    comment TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'submitted' CHECK (status IN ('submitted','shortlisted','accepted','rejected','withdrawn')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(order_id, bidder_id)
);

CREATE INDEX IF NOT EXISTS idx_estimate_items_project ON estimate_items(project_id);
CREATE INDEX IF NOT EXISTS idx_estimate_items_object ON estimate_items(object_id);
CREATE INDEX IF NOT EXISTS idx_order_items_order ON order_items(order_id);
CREATE INDEX IF NOT EXISTS idx_bids_order ON marketplace_bids(order_id);
CREATE INDEX IF NOT EXISTS idx_bids_bidder ON marketplace_bids(bidder_id);

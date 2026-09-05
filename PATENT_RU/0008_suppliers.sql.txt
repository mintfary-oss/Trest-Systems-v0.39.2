-- Stage 7: supplier profiles, offers, stock, documents and delivery terms.
CREATE TABLE IF NOT EXISTS supplier_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL UNIQUE REFERENCES organizations(id) ON DELETE CASCADE,
    summary TEXT NOT NULL DEFAULT '',
    categories JSONB NOT NULL DEFAULT '[]'::jsonb,
    delivery_regions JSONB NOT NULL DEFAULT '[]'::jsonb,
    delivery_terms JSONB NOT NULL DEFAULT '{}'::jsonb,
    verification_status TEXT NOT NULL DEFAULT 'pending' CHECK (verification_status IN ('pending','verified','rejected','suspended')),
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS supplier_applications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    applicant_user_id UUID NOT NULL REFERENCES users(id),
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','approved','rejected','withdrawn')),
    note TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    reviewed_at TIMESTAMPTZ,
    UNIQUE(organization_id, applicant_user_id)
);
CREATE TABLE IF NOT EXISTS supplier_offers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    supplier_id UUID NOT NULL REFERENCES supplier_profiles(id) ON DELETE CASCADE,
    sku TEXT NOT NULL,
    name TEXT NOT NULL,
    category TEXT NOT NULL DEFAULT 'general',
    unit TEXT NOT NULL DEFAULT 'pcs',
    price NUMERIC(16,2) NOT NULL CHECK (price >= 0),
    currency TEXT NOT NULL DEFAULT 'RUB',
    min_order_quantity NUMERIC(16,3) NOT NULL DEFAULT 1 CHECK (min_order_quantity > 0),
    stock_quantity NUMERIC(16,3) NOT NULL DEFAULT 0 CHECK (stock_quantity >= 0),
    lead_time_days INTEGER NOT NULL DEFAULT 0 CHECK (lead_time_days >= 0),
    status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','published','archived')),
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(supplier_id, sku)
);
CREATE TABLE IF NOT EXISTS supplier_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    supplier_id UUID NOT NULL REFERENCES supplier_profiles(id) ON DELETE CASCADE,
    document_type TEXT NOT NULL,
    title TEXT NOT NULL,
    document_ref TEXT NOT NULL,
    expires_at TIMESTAMPTZ,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
ALTER TABLE estimate_version_items ADD COLUMN IF NOT EXISTS supplier_offer_id UUID REFERENCES supplier_offers(id) ON DELETE SET NULL;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS supplier_offer_id UUID REFERENCES supplier_offers(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_supplier_profiles_status ON supplier_profiles(verification_status, active);
CREATE INDEX IF NOT EXISTS idx_supplier_applications_status ON supplier_applications(status);
CREATE INDEX IF NOT EXISTS idx_supplier_offers_lookup ON supplier_offers(category, status, currency);
CREATE INDEX IF NOT EXISTS idx_supplier_offers_supplier ON supplier_offers(supplier_id);
CREATE INDEX IF NOT EXISTS idx_supplier_documents_supplier ON supplier_documents(supplier_id);
CREATE INDEX IF NOT EXISTS idx_estimate_items_supplier_offer ON estimate_version_items(supplier_offer_id);
CREATE INDEX IF NOT EXISTS idx_orders_supplier_offer ON orders(supplier_offer_id);

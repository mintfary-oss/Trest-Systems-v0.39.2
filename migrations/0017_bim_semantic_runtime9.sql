-- Stage 10 Runtime 9: semantic import audit and GlobalId diff support.
CREATE TABLE IF NOT EXISTS bim_semantic_imports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bim_model_version_id UUID NOT NULL REFERENCES bim_model_versions(id) ON DELETE CASCADE,
    source_checksum TEXT NOT NULL DEFAULT '',
    element_count INTEGER NOT NULL DEFAULT 0,
    property_count INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'completed' CHECK (status IN ('running','completed','failed')),
    error TEXT NOT NULL DEFAULT '',
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_bim_semantic_imports_version ON bim_semantic_imports(bim_model_version_id, created_at DESC);

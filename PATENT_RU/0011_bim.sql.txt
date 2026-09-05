-- Stage 10: unified 3D/BIM model registry, versioning, exchange and progress snapshots.
CREATE TABLE IF NOT EXISTS bim_models (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    project_version_id UUID REFERENCES project_versions(id) ON DELETE SET NULL,
    name TEXT NOT NULL,
    format TEXT NOT NULL CHECK (format IN ('native','ifc','gltf','glb','obj','dxf')),
    storage_url TEXT NOT NULL DEFAULT '',
    schema_version TEXT NOT NULL DEFAULT '1.0',
    status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','review','approved','archived')),
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS bim_model_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bim_model_id UUID NOT NULL REFERENCES bim_models(id) ON DELETE CASCADE,
    version INTEGER NOT NULL CHECK (version > 0),
    source_format TEXT NOT NULL,
    source_uri TEXT NOT NULL DEFAULT '',
    geometry_uri TEXT NOT NULL DEFAULT '',
    manifest JSONB NOT NULL DEFAULT '{}'::jsonb,
    checksum TEXT NOT NULL DEFAULT '',
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(bim_model_id, version)
);

CREATE TABLE IF NOT EXISTS bim_elements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bim_model_version_id UUID NOT NULL REFERENCES bim_model_versions(id) ON DELETE CASCADE,
    external_id TEXT NOT NULL,
    element_type TEXT NOT NULL,
    name TEXT NOT NULL DEFAULT '',
    properties JSONB NOT NULL DEFAULT '{}'::jsonb,
    geometry JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(bim_model_version_id, external_id)
);

CREATE TABLE IF NOT EXISTS bim_progress_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    bim_model_version_id UUID REFERENCES bim_model_versions(id) ON DELETE SET NULL,
    snapshot_date DATE NOT NULL,
    planned_percent NUMERIC(6,3) NOT NULL DEFAULT 0 CHECK (planned_percent BETWEEN 0 AND 100),
    actual_percent NUMERIC(6,3) NOT NULL DEFAULT 0 CHECK (actual_percent BETWEEN 0 AND 100),
    source TEXT NOT NULL DEFAULT 'manual',
    details JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(project_id, snapshot_date)
);

CREATE TABLE IF NOT EXISTS bim_import_exports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    bim_model_id UUID REFERENCES bim_models(id) ON DELETE SET NULL,
    operation TEXT NOT NULL CHECK (operation IN ('import','export')),
    format TEXT NOT NULL CHECK (format IN ('ifc','gltf','glb','obj','dxf')),
    status TEXT NOT NULL DEFAULT 'queued' CHECK (status IN ('queued','running','completed','failed')),
    input_uri TEXT NOT NULL DEFAULT '',
    output_uri TEXT NOT NULL DEFAULT '',
    error TEXT NOT NULL DEFAULT '',
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_bim_models_project ON bim_models(project_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_bim_model_versions_model ON bim_model_versions(bim_model_id, version DESC);
CREATE INDEX IF NOT EXISTS idx_bim_elements_version ON bim_elements(bim_model_version_id);
CREATE INDEX IF NOT EXISTS idx_bim_progress_project_date ON bim_progress_snapshots(project_id, snapshot_date DESC);
CREATE INDEX IF NOT EXISTS idx_bim_io_project ON bim_import_exports(project_id, created_at DESC);

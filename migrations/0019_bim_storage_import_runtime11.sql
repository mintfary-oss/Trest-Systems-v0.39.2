-- Stage 10 Runtime 11: stored IFC source metadata and import audit.
ALTER TABLE bim_model_versions ADD COLUMN IF NOT EXISTS source_sha256 TEXT NOT NULL DEFAULT '';
ALTER TABLE bim_model_versions ADD COLUMN IF NOT EXISTS source_size BIGINT NOT NULL DEFAULT 0;
CREATE TABLE IF NOT EXISTS bim_import_audits (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
 bim_model_version_id UUID NOT NULL REFERENCES bim_model_versions(id) ON DELETE CASCADE,
 source_key TEXT NOT NULL,
 source_sha256 TEXT NOT NULL,
 inserted_count INT NOT NULL DEFAULT 0,
 updated_count INT NOT NULL DEFAULT 0,
 status TEXT NOT NULL CHECK(status IN ('completed','failed')),
 error TEXT NOT NULL DEFAULT '',
 created_by UUID NOT NULL REFERENCES users(id),
 created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_bim_import_audits_version ON bim_import_audits(bim_model_version_id,created_at DESC);

-- Stage 10 Runtime 10: mesh-element picking and BIM authorization metadata.
ALTER TABLE bim_model_versions ADD COLUMN IF NOT EXISTS mesh_manifest JSONB NOT NULL DEFAULT '{}'::jsonb;
ALTER TABLE bim_model_versions ADD COLUMN IF NOT EXISTS object_storage_key TEXT NOT NULL DEFAULT '';
CREATE INDEX IF NOT EXISTS idx_bim_model_versions_storage_key ON bim_model_versions(object_storage_key) WHERE object_storage_key <> '';

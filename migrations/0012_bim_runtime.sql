-- Stage 10 runtime: exchange outputs, geometry diff and retry metadata.
ALTER TABLE bim_import_exports ADD COLUMN IF NOT EXISTS attempts integer NOT NULL DEFAULT 0;
ALTER TABLE bim_import_exports ADD COLUMN IF NOT EXISTS max_attempts integer NOT NULL DEFAULT 3;
ALTER TABLE bim_import_exports ADD COLUMN IF NOT EXISTS started_at timestamptz;
ALTER TABLE bim_import_exports ADD COLUMN IF NOT EXISTS output_checksum text NOT NULL DEFAULT '';
ALTER TABLE bim_import_exports ADD COLUMN IF NOT EXISTS output_manifest jsonb NOT NULL DEFAULT '{}'::jsonb;
CREATE TABLE IF NOT EXISTS bim_geometry_diffs (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
 from_version_id uuid NOT NULL REFERENCES bim_model_versions(id) ON DELETE CASCADE,
 to_version_id uuid NOT NULL REFERENCES bim_model_versions(id) ON DELETE CASCADE,
 tolerance numeric NOT NULL DEFAULT 0.001,
 result jsonb NOT NULL,
 created_by uuid NOT NULL REFERENCES users(id),
 created_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(from_version_id,to_version_id)
);
CREATE INDEX IF NOT EXISTS idx_bim_geometry_diffs_project ON bim_geometry_diffs(project_id,created_at DESC);

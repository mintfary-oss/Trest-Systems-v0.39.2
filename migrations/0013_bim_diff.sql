-- Stage 10 Runtime 3: geometry diff API/runtime indexes.
CREATE INDEX IF NOT EXISTS idx_bim_geometry_diffs_versions ON bim_geometry_diffs(from_version_id,to_version_id);

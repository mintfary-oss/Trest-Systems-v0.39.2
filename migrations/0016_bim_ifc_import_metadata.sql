-- Stage 10 Runtime 8: IFC semantic import metadata and traceability.
ALTER TABLE bim_elements ADD COLUMN IF NOT EXISTS source_format TEXT NOT NULL DEFAULT '';
ALTER TABLE bim_elements ADD COLUMN IF NOT EXISTS source_entity_type TEXT NOT NULL DEFAULT '';
CREATE INDEX IF NOT EXISTS idx_bim_elements_source_entity_type ON bim_elements(bim_model_version_id, source_entity_type);

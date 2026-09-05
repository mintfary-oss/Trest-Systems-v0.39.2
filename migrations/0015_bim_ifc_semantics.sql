-- Stage 10 Runtime 7: IFC semantic identity and containment metadata.
ALTER TABLE bim_elements ADD COLUMN IF NOT EXISTS ifc_entity_id INTEGER;
ALTER TABLE bim_elements ADD COLUMN IF NOT EXISTS ifc_global_id TEXT NOT NULL DEFAULT '';
ALTER TABLE bim_elements ADD COLUMN IF NOT EXISTS parent_external_id TEXT NOT NULL DEFAULT '';
ALTER TABLE bim_elements ADD COLUMN IF NOT EXISTS relation_type TEXT NOT NULL DEFAULT '';
CREATE INDEX IF NOT EXISTS idx_bim_elements_ifc_global_id ON bim_elements(bim_model_version_id, ifc_global_id) WHERE ifc_global_id <> '';
CREATE INDEX IF NOT EXISTS idx_bim_elements_parent ON bim_elements(bim_model_version_id, parent_external_id) WHERE parent_external_id <> '';

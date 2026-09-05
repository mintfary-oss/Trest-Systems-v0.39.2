package importer

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/mintfary-oss/trest-sistems/internal/bim/ifc"
	"github.com/mintfary-oss/trest-sistems/internal/bim/ifcsemantic"
)

type Result struct {
	Inserted int `json:"inserted"`
	Updated  int `json:"updated"`
}

// ImportIFC atomically upserts semantic IFC elements into one existing model version.
// A failure rolls back the whole import, preventing partially imported BIM versions.
func ImportIFC(ctx context.Context, db *sql.DB, versionID string, entities []ifc.Entity) (Result, error) {
	if db == nil {
		return Result{}, fmt.Errorf("database is nil")
	}
	if versionID == "" {
		return Result{}, fmt.Errorf("version id is empty")
	}
	mappings := ifcsemantic.MapToBIMElements(entities)
	props := ifcsemantic.DecodeProperties(entities)
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return Result{}, err
	}
	defer tx.Rollback()
	var out Result
	for _, m := range mappings {
		p := map[string]any{}
		for k, v := range m.Properties {
			p[k] = v
		}
		if x, ok := props[m.IFCEntityID]; ok {
			p["property_sets"] = x
		}
		raw, err := json.Marshal(p)
		if err != nil {
			return Result{}, err
		}
		res, err := tx.ExecContext(ctx, `INSERT INTO bim_elements (bim_model_version_id,external_id,element_type,name,properties,geometry,ifc_entity_id,ifc_global_id,parent_external_id,relation_type,source_format,source_entity_type) VALUES ($1,$2,$3,$4,$5::jsonb,'{}'::jsonb,$6,$7,$8,$9,'ifc',$10) ON CONFLICT (bim_model_version_id,external_id) DO UPDATE SET element_type=EXCLUDED.element_type,name=EXCLUDED.name,properties=EXCLUDED.properties,ifc_entity_id=EXCLUDED.ifc_entity_id,ifc_global_id=EXCLUDED.ifc_global_id,parent_external_id=EXCLUDED.parent_external_id,relation_type=EXCLUDED.relation_type,source_format='ifc',source_entity_type=EXCLUDED.source_entity_type`, versionID, m.ExternalID, m.ElementType, m.Name, string(raw), m.IFCEntityID, m.IFCGlobalID, m.ParentExternalID, m.RelationType, m.ElementType)
		if err != nil {
			return Result{}, err
		}
		n, _ := res.RowsAffected()
		if n == 1 {
			out.Inserted++
		} else {
			out.Updated++
		}
	}
	if err := tx.Commit(); err != nil {
		return Result{}, err
	}
	return out, nil
}

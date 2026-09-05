package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/mintfary-oss/trest-sistems/internal/bim"
)

// bimElementPick resolves a renderer index to the semantic BIM element using the
// immutable mesh manifest stored with the model version.
func (s *Server) bimElementPick(w http.ResponseWriter, r *http.Request) {
	versionID := strings.TrimSpace(r.URL.Query().Get("bim_model_version_id"))
	if versionID == "" {
		writeJSON(w, 400, map[string]string{"error": "bim_model_version_id is required"})
		return
	}
	userID := claimsUserID(r.Context())
	var projectID, raw string
	if err := s.DB.QueryRowContext(r.Context(), `SELECT bm.project_id::text,bmv.mesh_manifest::text FROM bim_model_versions bmv JOIN bim_models bm ON bm.id=bmv.bim_model_id WHERE bmv.id=$1`, versionID).Scan(&projectID, &raw); err != nil {
		writeJSON(w, 404, map[string]string{"error": "model version not found"})
		return
	}
	if err := s.requireProjectAccess(r.Context(), projectID, userID); err != nil {
		writeJSON(w, 403, map[string]string{"error": "forbidden"})
		return
	}
	var index int
	if _, err := fmt.Sscanf(strings.TrimSpace(r.URL.Query().Get("index")), "%d", &index); err != nil || index < 0 {
		writeJSON(w, 400, map[string]string{"error": "valid index is required"})
		return
	}
	manifest, err := bim.ParseMeshElementManifest([]byte(raw))
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": "invalid mesh manifest"})
		return
	}
	rng, ok := manifest.FindIndex(index)
	if !ok {
		writeJSON(w, 404, map[string]string{"error": "index is not mapped to an element"})
		return
	}
	var element map[string]any
	var props, geom string
	var id, typ, name, global, parent string
	err = s.DB.QueryRowContext(r.Context(), `SELECT id::text,element_type,name,properties::text,geometry::text,ifc_global_id,parent_external_id FROM bim_elements WHERE bim_model_version_id=$1 AND external_id=$2`, versionID, rng.ElementExternalID).Scan(&id, &typ, &name, &props, &geom, &global, &parent)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON(w, 404, map[string]string{"error": "element not found"})
			return
		}
		writeError(w, err)
		return
	}
	_ = json.Unmarshal([]byte(props), &element)
	if element == nil {
		element = map[string]any{}
	}
	element["id"] = id
	element["element_type"] = typ
	element["name"] = name
	element["ifc_global_id"] = global
	element["parent_external_id"] = parent
	element["mesh_range"] = rng
	writeJSON(w, 200, element)
}

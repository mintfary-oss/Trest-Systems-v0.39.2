package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

// GET /api/v1/bim/model-versions/manifest returns the immutable renderer-to-element manifest.
func (s *Server) bimModelVersionManifest(w http.ResponseWriter, r *http.Request) {
	versionID := strings.TrimSpace(r.URL.Query().Get("bim_model_version_id"))
	if versionID == "" {
		writeJSON(w, 400, map[string]string{"error": "bim_model_version_id is required"})
		return
	}
	var projectID, manifest, geometry, source string
	err := s.DB.QueryRowContext(r.Context(), `SELECT bm.project_id::text,COALESCE(bmv.mesh_manifest::text,'{}'),COALESCE(bmv.geometry_uri,''),COALESCE(bmv.source_uri,'') FROM bim_model_versions bmv JOIN bim_models bm ON bm.id=bmv.bim_model_id WHERE bmv.id=$1`, versionID).Scan(&projectID, &manifest, &geometry, &source)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON(w, 404, map[string]string{"error": "model version not found"})
			return
		}
		writeError(w, err)
		return
	}
	if err = s.requireProjectAccess(r.Context(), projectID, claimsUserID(r.Context())); err != nil {
		writeJSON(w, 403, map[string]string{"error": "forbidden"})
		return
	}
	var m any
	if err = json.Unmarshal([]byte(manifest), &m); err != nil {
		writeJSON(w, 500, map[string]string{"error": "invalid mesh manifest"})
		return
	}
	writeJSON(w, 200, map[string]any{"version_id": versionID, "project_id": projectID, "manifest": m, "geometry_uri": geometry, "source_uri": source})
}

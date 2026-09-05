package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/mintfary-oss/trest-sistems/internal/bim"
)

type createBIMVersionRequest struct {
	BIMModelID   string         `json:"bim_model_id"`
	Version      int            `json:"version"`
	SourceFormat string         `json:"source_format"`
	SourceURI    string         `json:"source_uri"`
	GeometryURI  string         `json:"geometry_uri"`
	Manifest     map[string]any `json:"manifest"`
	Checksum     string         `json:"checksum"`
}

type upsertBIMElementRequest struct {
	BIMModelVersionID string         `json:"bim_model_version_id"`
	ExternalID        string         `json:"external_id"`
	ElementType       string         `json:"element_type"`
	Name              string         `json:"name"`
	Properties        map[string]any `json:"properties"`
	Geometry          map[string]any `json:"geometry"`
}

type createBIMExchangeRequest struct {
	ProjectID  string `json:"project_id"`
	BIMModelID string `json:"bim_model_id"`
	Operation  string `json:"operation"`
	Format     string `json:"format"`
	InputURI   string `json:"input_uri"`
	OutputURI  string `json:"output_uri"`
}

func (s *Server) createBIMModelVersion(w http.ResponseWriter, r *http.Request) {
	var req createBIMVersionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid JSON"})
		return
	}
	if strings.TrimSpace(req.BIMModelID) == "" {
		writeJSON(w, 400, map[string]string{"error": "bim_model_id is required"})
		return
	}
	if err := bim.ValidateVersion(req.Version); err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}
	if err := bim.ValidateFormat(req.SourceFormat); err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}
	if req.Manifest == nil {
		req.Manifest = map[string]any{}
	}
	raw, _ := json.Marshal(req.Manifest)
	var id string
	err := s.DB.QueryRowContext(r.Context(), `INSERT INTO bim_model_versions(bim_model_id,version,source_format,source_uri,geometry_uri,manifest,checksum,created_by) VALUES($1,$2,$3,$4,$5,$6::jsonb,$7,$8) RETURNING id`, req.BIMModelID, req.Version, strings.ToLower(req.SourceFormat), req.SourceURI, req.GeometryURI, string(raw), req.Checksum, claimsUserID(r.Context())).Scan(&id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, 201, map[string]any{"id": id, "version": req.Version})
}

func (s *Server) bimModelVersions(w http.ResponseWriter, r *http.Request) {
	modelID := strings.TrimSpace(r.URL.Query().Get("bim_model_id"))
	if modelID == "" {
		writeJSON(w, 400, map[string]string{"error": "bim_model_id is required"})
		return
	}
	rows, err := s.DB.QueryContext(r.Context(), `SELECT id, version, source_format, source_uri, geometry_uri, manifest, checksum, created_at FROM bim_model_versions WHERE bim_model_id=$1 ORDER BY version DESC`, modelID)
	if err != nil {
		writeError(w, err)
		return
	}
	defer rows.Close()
	out := make([]map[string]any, 0)
	for rows.Next() {
		var id, sf, su, gu, checksum string
		var version int
		var manifest, created any
		if err := rows.Scan(&id, &version, &sf, &su, &gu, &manifest, &checksum, &created); err != nil {
			writeError(w, err)
			return
		}
		out = append(out, map[string]any{"id": id, "version": version, "source_format": sf, "source_uri": su, "geometry_uri": gu, "manifest": manifest, "checksum": checksum, "created_at": created})
	}
	if err := rows.Err(); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, 200, out)
}

func (s *Server) createBIMElement(w http.ResponseWriter, r *http.Request) {
	var req upsertBIMElementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid JSON"})
		return
	}
	if strings.TrimSpace(req.BIMModelVersionID) == "" || strings.TrimSpace(req.ExternalID) == "" || strings.TrimSpace(req.ElementType) == "" {
		writeJSON(w, 400, map[string]string{"error": "bim_model_version_id, external_id and element_type are required"})
		return
	}
	if req.Properties == nil {
		req.Properties = map[string]any{}
	}
	if req.Geometry == nil {
		req.Geometry = map[string]any{}
	}
	props, _ := json.Marshal(req.Properties)
	geom, _ := json.Marshal(req.Geometry)
	var id string
	err := s.DB.QueryRowContext(r.Context(), `INSERT INTO bim_elements(bim_model_version_id,external_id,element_type,name,properties,geometry) VALUES($1,$2,$3,$4,$5::jsonb,$6::jsonb) RETURNING id`, req.BIMModelVersionID, req.ExternalID, req.ElementType, req.Name, string(props), string(geom)).Scan(&id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, 201, map[string]any{"id": id, "external_id": req.ExternalID})
}

func (s *Server) bimElements(w http.ResponseWriter, r *http.Request) {
	versionID := strings.TrimSpace(r.URL.Query().Get("bim_model_version_id"))
	if versionID == "" {
		writeJSON(w, 400, map[string]string{"error": "bim_model_version_id is required"})
		return
	}
	limit := 100
	if v, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && v > 0 && v <= 500 {
		limit = v
	}
	rows, err := s.DB.QueryContext(r.Context(), `SELECT id,external_id,element_type,name,properties,geometry,created_at FROM bim_elements WHERE bim_model_version_id=$1 ORDER BY created_at, external_id LIMIT $2`, versionID, limit)
	if err != nil {
		writeError(w, err)
		return
	}
	defer rows.Close()
	out := make([]map[string]any, 0)
	for rows.Next() {
		var id, eid, typ, name string
		var props, geom, created any
		if err := rows.Scan(&id, &eid, &typ, &name, &props, &geom, &created); err != nil {
			writeError(w, err)
			return
		}
		out = append(out, map[string]any{"id": id, "external_id": eid, "element_type": typ, "name": name, "properties": props, "geometry": geom, "created_at": created})
	}
	if err := rows.Err(); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, 200, out)
}

func (s *Server) createBIMExchange(w http.ResponseWriter, r *http.Request) {
	var req createBIMExchangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid JSON"})
		return
	}
	req.Operation = strings.ToLower(strings.TrimSpace(req.Operation))
	req.Format = strings.ToLower(strings.TrimSpace(req.Format))
	if strings.TrimSpace(req.ProjectID) == "" {
		writeJSON(w, 400, map[string]string{"error": "project_id is required"})
		return
	}
	if err := bim.ValidateImportExport(req.Operation, req.Format); err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}
	var id string
	err := s.DB.QueryRowContext(r.Context(), `INSERT INTO bim_import_exports(project_id,bim_model_id,operation,format,input_uri,output_uri,created_by) VALUES($1,NULLIF($2,'')::uuid,$3,$4,$5,$6,$7) RETURNING id`, req.ProjectID, req.BIMModelID, req.Operation, req.Format, req.InputURI, req.OutputURI, claimsUserID(r.Context())).Scan(&id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, 201, map[string]any{"id": id, "status": "queued"})
}

func (s *Server) bimExchanges(w http.ResponseWriter, r *http.Request) {
	projectID := strings.TrimSpace(r.URL.Query().Get("project_id"))
	if projectID == "" {
		writeJSON(w, 400, map[string]string{"error": "project_id is required"})
		return
	}
	rows, err := s.DB.QueryContext(r.Context(), `SELECT id,bim_model_id,operation,format,status,input_uri,output_uri,error,created_at,completed_at FROM bim_import_exports WHERE project_id=$1 ORDER BY created_at DESC LIMIT 100`, projectID)
	if err != nil {
		writeError(w, err)
		return
	}
	defer rows.Close()
	out := make([]map[string]any, 0)
	for rows.Next() {
		var id, op, format, status, in, outURI, errText string
		var model any
		var created, completed any
		if err := rows.Scan(&id, &model, &op, &format, &status, &in, &outURI, &errText, &created, &completed); err != nil {
			writeError(w, err)
			return
		}
		out = append(out, map[string]any{"id": id, "bim_model_id": model, "operation": op, "format": format, "status": status, "input_uri": in, "output_uri": outURI, "error": errText, "created_at": created, "completed_at": completed})
	}
	if err := rows.Err(); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, 200, out)
}

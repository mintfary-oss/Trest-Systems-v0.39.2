package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/mintfary-oss/trest-sistems/internal/bim"
)

type createBIMModelRequest struct {
	ProjectID        string         `json:"project_id"`
	ProjectVersionID string         `json:"project_version_id"`
	Name             string         `json:"name"`
	Format           string         `json:"format"`
	StorageURL       string         `json:"storage_url"`
	SchemaVersion    string         `json:"schema_version"`
	Metadata         map[string]any `json:"metadata"`
}

type createBIMProgressRequest struct {
	ProjectID         string         `json:"project_id"`
	BIMModelVersionID string         `json:"bim_model_version_id"`
	SnapshotDate      string         `json:"snapshot_date"`
	PlannedPercent    float64        `json:"planned_percent"`
	ActualPercent     float64        `json:"actual_percent"`
	Source            string         `json:"source"`
	Details           map[string]any `json:"details"`
}

func claimsUserID(ctx context.Context) string {
	c, _ := claimsFromContext(ctx)
	return c.UserID
}

func (s *Server) bimModels(w http.ResponseWriter, r *http.Request) {
	projectID := strings.TrimSpace(r.URL.Query().Get("project_id"))
	q := `SELECT id, project_id, COALESCE(project_version_id::text,''), name, format, storage_url, schema_version, status, created_at, updated_at FROM bim_models`
	args := []any{}
	if projectID != "" {
		q += ` WHERE project_id=$1`
		args = append(args, projectID)
	}
	q += ` ORDER BY created_at DESC LIMIT 100`
	rows, err := s.DB.QueryContext(r.Context(), q, args...)
	if err != nil {
		writeError(w, err)
		return
	}
	defer rows.Close()
	out := make([]map[string]any, 0)
	for rows.Next() {
		var id, pid, pvid, name, format, url, schema, status string
		var created, updated any
		if err := rows.Scan(&id, &pid, &pvid, &name, &format, &url, &schema, &status, &created, &updated); err != nil {
			writeError(w, err)
			return
		}
		out = append(out, map[string]any{"id": id, "project_id": pid, "project_version_id": pvid, "name": name, "format": format, "storage_url": url, "schema_version": schema, "status": status, "created_at": created, "updated_at": updated})
	}
	if err := rows.Err(); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) createBIMModel(w http.ResponseWriter, r *http.Request) {
	var req createBIMModelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid JSON"})
		return
	}
	if strings.TrimSpace(req.ProjectID) == "" || strings.TrimSpace(req.Name) == "" {
		writeJSON(w, 400, map[string]string{"error": "project_id and name are required"})
		return
	}
	if err := bim.ValidateFormat(req.Format); err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}
	schema := req.SchemaVersion
	if schema == "" {
		schema = "1.0"
	}
	meta := req.Metadata
	if meta == nil {
		meta = map[string]any{}
	}
	raw, _ := json.Marshal(meta)
	var id string
	err := s.DB.QueryRowContext(r.Context(), `INSERT INTO bim_models(project_id, project_version_id, name, format, storage_url, schema_version, metadata, created_by) VALUES($1,NULLIF($2,'')::uuid,$3,$4,$5,$6,$7::jsonb,$8) RETURNING id`, req.ProjectID, req.ProjectVersionID, req.Name, strings.ToLower(req.Format), req.StorageURL, schema, string(raw), claimsUserID(r.Context())).Scan(&id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"id": id, "status": "draft"})
}

func (s *Server) bimProgress(w http.ResponseWriter, r *http.Request) {
	projectID := strings.TrimSpace(r.URL.Query().Get("project_id"))
	if projectID == "" {
		writeJSON(w, 400, map[string]string{"error": "project_id is required"})
		return
	}
	rows, err := s.DB.QueryContext(r.Context(), `SELECT id, snapshot_date, planned_percent, actual_percent, source, details, created_at FROM bim_progress_snapshots WHERE project_id=$1 ORDER BY snapshot_date DESC LIMIT 100`, projectID)
	if err != nil {
		writeError(w, err)
		return
	}
	defer rows.Close()
	out := make([]map[string]any, 0)
	for rows.Next() {
		var id, source string
		var date, planned, actual, details, created any
		if err := rows.Scan(&id, &date, &planned, &actual, &source, &details, &created); err != nil {
			writeError(w, err)
			return
		}
		out = append(out, map[string]any{"id": id, "snapshot_date": date, "planned_percent": planned, "actual_percent": actual, "source": source, "details": details, "created_at": created})
	}
	if err := rows.Err(); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, 200, out)
}

func (s *Server) createBIMProgress(w http.ResponseWriter, r *http.Request) {
	var req createBIMProgressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid JSON"})
		return
	}
	if req.ProjectID == "" || req.SnapshotDate == "" {
		writeJSON(w, 400, map[string]string{"error": "project_id and snapshot_date are required"})
		return
	}
	if err := bim.ValidateProgress(req.PlannedPercent, req.ActualPercent); err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}
	source := req.Source
	if source == "" {
		source = "manual"
	}
	details := req.Details
	if details == nil {
		details = map[string]any{}
	}
	raw, _ := json.Marshal(details)
	var id string
	err := s.DB.QueryRowContext(r.Context(), `INSERT INTO bim_progress_snapshots(project_id,bim_model_version_id,snapshot_date,planned_percent,actual_percent,source,details,created_by) VALUES($1,NULLIF($2,'')::uuid,$3,$4,$5,$6,$7::jsonb,$8) RETURNING id`, req.ProjectID, req.BIMModelVersionID, req.SnapshotDate, req.PlannedPercent, req.ActualPercent, source, string(raw), claimsUserID(r.Context())).Scan(&id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, 201, map[string]any{"id": id})
}

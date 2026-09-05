package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/mintfary-oss/trest-sistems/internal/bim/diff"
)

type geometryDiffRequest struct {
	ProjectID     string    `json:"project_id"`
	FromVersionID string    `json:"from_version_id"`
	ToVersionID   string    `json:"to_version_id"`
	Tolerance     float64   `json:"tolerance"`
	OldPositions  []float64 `json:"old_positions"`
	OldIndices    []uint32  `json:"old_indices"`
	NewPositions  []float64 `json:"new_positions"`
	NewIndices    []uint32  `json:"new_indices"`
}

func (s *Server) createBIMGeometryDiff(w http.ResponseWriter, r *http.Request) {
	var req geometryDiffRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid JSON"})
		return
	}
	if strings.TrimSpace(req.ProjectID) == "" || strings.TrimSpace(req.FromVersionID) == "" || strings.TrimSpace(req.ToVersionID) == "" {
		writeJSON(w, 400, map[string]string{"error": "project_id, from_version_id and to_version_id are required"})
		return
	}
	if req.Tolerance < 0 {
		req.Tolerance = 0
	}
	result := diff.Compare(req.OldPositions, req.OldIndices, req.NewPositions, req.NewIndices, req.Tolerance)
	raw, _ := json.Marshal(result)
	var id string
	err := s.DB.QueryRowContext(r.Context(), `INSERT INTO bim_geometry_diffs(project_id,from_version_id,to_version_id,tolerance,result,created_by) VALUES($1,$2,$3,$4,$5::jsonb,$6) ON CONFLICT(from_version_id,to_version_id) DO UPDATE SET tolerance=EXCLUDED.tolerance,result=EXCLUDED.result,created_by=EXCLUDED.created_by,created_at=now() RETURNING id`, req.ProjectID, req.FromVersionID, req.ToVersionID, req.Tolerance, string(raw), claimsUserID(r.Context())).Scan(&id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, 201, map[string]any{"id": id, "project_id": req.ProjectID, "from_version_id": req.FromVersionID, "to_version_id": req.ToVersionID, "tolerance": req.Tolerance, "result": result})
}

func (s *Server) bimGeometryDiffs(w http.ResponseWriter, r *http.Request) {
	projectID := strings.TrimSpace(r.URL.Query().Get("project_id"))
	if projectID == "" {
		writeJSON(w, 400, map[string]string{"error": "project_id is required"})
		return
	}
	limit := 100
	if v, e := strconv.Atoi(r.URL.Query().Get("limit")); e == nil && v > 0 && v <= 500 {
		limit = v
	}
	rows, err := s.DB.QueryContext(r.Context(), `SELECT id,from_version_id,to_version_id,tolerance,result,created_by,created_at FROM bim_geometry_diffs WHERE project_id=$1 ORDER BY created_at DESC LIMIT $2`, projectID, limit)
	if err != nil {
		writeError(w, err)
		return
	}
	defer rows.Close()
	out := make([]map[string]any, 0)
	for rows.Next() {
		var id, from, to, createdBy string
		var tol float64
		var result any
		var created any
		if err := rows.Scan(&id, &from, &to, &tol, &result, &createdBy, &created); err != nil {
			writeError(w, err)
			return
		}
		out = append(out, map[string]any{"id": id, "from_version_id": from, "to_version_id": to, "tolerance": tol, "result": result, "created_by": createdBy, "created_at": created})
	}
	if err := rows.Err(); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, 200, out)
}

package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

func (s *Server) objectTypes(w http.ResponseWriter, r *http.Request) {
	rows, err := s.DB.QueryContext(r.Context(), `SELECT id,code,name,description,parameters_schema,active,created_at FROM object_types WHERE active=true ORDER BY name`)
	if err != nil {
		writeError(w, err)
		return
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id, code, name, desc string
		var schema any
		var active bool
		var created any
		if err := rows.Scan(&id, &code, &name, &desc, &schema, &active, &created); err != nil {
			writeError(w, err)
			return
		}
		out = append(out, map[string]any{"id": id, "code": code, "name": name, "description": desc, "parameters_schema": schema, "active": active, "created_at": created})
	}
	writeJSON(w, 200, out)
}

func (s *Server) materials(w http.ResponseWriter, r *http.Request) {
	category := strings.TrimSpace(r.URL.Query().Get("category"))
	q := `SELECT id,code,name,category,unit,parameters,active,created_at FROM materials WHERE active=true`
	args := []any{}
	if category != "" {
		q += ` AND category=$1`
		args = append(args, category)
	}
	q += ` ORDER BY name`
	rows, err := s.DB.QueryContext(r.Context(), q, args...)
	if err != nil {
		writeError(w, err)
		return
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id, code, name, cat, unit string
		var params any
		var active bool
		var created any
		if err := rows.Scan(&id, &code, &name, &cat, &unit, &params, &active, &created); err != nil {
			writeError(w, err)
			return
		}
		out = append(out, map[string]any{"id": id, "code": code, "name": name, "category": cat, "unit": unit, "parameters": params, "active": active, "created_at": created})
	}
	writeJSON(w, 200, out)
}

func (s *Server) engineeringSystems(w http.ResponseWriter, r *http.Request) {
	rows, err := s.DB.QueryContext(r.Context(), `SELECT id,code,name,category,parameters,active,created_at FROM engineering_systems WHERE active=true ORDER BY category,name`)
	if err != nil {
		writeError(w, err)
		return
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id, code, name, cat string
		var params any
		var active bool
		var created any
		if err := rows.Scan(&id, &code, &name, &cat, &params, &active, &created); err != nil {
			writeError(w, err)
			return
		}
		out = append(out, map[string]any{"id": id, "code": code, "name": name, "category": cat, "parameters": params, "active": active, "created_at": created})
	}
	writeJSON(w, 200, out)
}

func (s *Server) finishes(w http.ResponseWriter, r *http.Request) {
	rows, err := s.DB.QueryContext(r.Context(), `SELECT id,code,name,category,material_id,parameters,active,created_at FROM finishes WHERE active=true ORDER BY category,name`)
	if err != nil {
		writeError(w, err)
		return
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id, code, name, cat string
		var materialID any
		var params any
		var active bool
		var created any
		if err := rows.Scan(&id, &code, &name, &cat, &materialID, &params, &active, &created); err != nil {
			writeError(w, err)
			return
		}
		out = append(out, map[string]any{"id": id, "code": code, "name": name, "category": cat, "material_id": materialID, "parameters": params, "active": active, "created_at": created})
	}
	writeJSON(w, 200, out)
}

func (s *Server) projectVersions(w http.ResponseWriter, r *http.Request) {
	c, ok := claimsFromContext(r.Context())
	if !ok {
		writeJSON(w, 401, map[string]string{"error": "unauthorized"})
		return
	}
	projectID := strings.TrimSpace(r.PathValue("projectID"))
	if projectID == "" {
		writeJSON(w, 400, map[string]string{"error": "project id is required"})
		return
	}
	var owner string
	if err := s.DB.QueryRowContext(r.Context(), `SELECT owner_id FROM projects WHERE id=$1`, projectID).Scan(&owner); err != nil {
		writeJSON(w, 404, map[string]string{"error": "project not found"})
		return
	}
	if owner != c.UserID && c.Role != "admin" {
		writeJSON(w, 403, map[string]string{"error": "forbidden"})
		return
	}
	rows, err := s.DB.QueryContext(r.Context(), `SELECT id,version,status,snapshot,created_by,created_at,approved_at FROM project_versions WHERE project_id=$1 ORDER BY version DESC`, projectID)
	if err != nil {
		writeError(w, err)
		return
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id, status, createdBy string
		var version int
		var snapshot any
		var created, approved any
		if err := rows.Scan(&id, &version, &status, &snapshot, &createdBy, &created, &approved); err != nil {
			writeError(w, err)
			return
		}
		out = append(out, map[string]any{"id": id, "version": version, "status": status, "snapshot": snapshot, "created_by": createdBy, "created_at": created, "approved_at": approved})
	}
	writeJSON(w, 200, out)
}

func (s *Server) createProjectVersion(w http.ResponseWriter, r *http.Request) {
	c, ok := claimsFromContext(r.Context())
	if !ok {
		writeJSON(w, 401, map[string]string{"error": "unauthorized"})
		return
	}
	projectID := strings.TrimSpace(r.PathValue("projectID"))
	if projectID == "" {
		writeJSON(w, 400, map[string]string{"error": "project id is required"})
		return
	}
	var owner string
	if err := s.DB.QueryRowContext(r.Context(), `SELECT owner_id FROM projects WHERE id=$1`, projectID).Scan(&owner); err != nil {
		writeJSON(w, 404, map[string]string{"error": "project not found"})
		return
	}
	if owner != c.UserID && c.Role != "admin" {
		writeJSON(w, 403, map[string]string{"error": "forbidden"})
		return
	}
	var in struct {
		Snapshot map[string]any `json:"snapshot"`
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid json"})
		return
	}
	snap, _ := json.Marshal(in.Snapshot)
	tx, err := s.DB.BeginTx(r.Context(), nil)
	if err != nil {
		writeError(w, err)
		return
	}
	defer tx.Rollback()
	var version int
	if err = tx.QueryRowContext(r.Context(), `SELECT COALESCE(MAX(version),0)+1 FROM project_versions WHERE project_id=$1`, projectID).Scan(&version); err != nil {
		writeError(w, err)
		return
	}
	var id string
	if err = tx.QueryRowContext(r.Context(), `INSERT INTO project_versions(project_id,version,status,snapshot,created_by) VALUES($1,$2,'draft',$3,$4) RETURNING id`, projectID, version, snap, c.UserID).Scan(&id); err != nil {
		writeError(w, err)
		return
	}
	if _, err = tx.ExecContext(r.Context(), `UPDATE projects SET architecture_version=$1,parameters=$2 WHERE id=$3`, version, snap, projectID); err != nil {
		writeError(w, err)
		return
	}
	if err = tx.Commit(); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, 201, map[string]any{"id": id, "project_id": projectID, "version": version, "status": "draft"})
}

func (s *Server) projectVersionAction(w http.ResponseWriter, r *http.Request) {
	c, ok := claimsFromContext(r.Context())
	if !ok {
		writeJSON(w, 401, map[string]string{"error": "unauthorized"})
		return
	}
	projectID := strings.TrimSpace(r.PathValue("projectID"))
	versionText := strings.TrimSpace(r.PathValue("version"))
	version, err := strconv.Atoi(versionText)
	if err != nil || version < 1 {
		writeJSON(w, 400, map[string]string{"error": "invalid version"})
		return
	}
	var owner string
	if err = s.DB.QueryRowContext(r.Context(), `SELECT owner_id FROM projects WHERE id=$1`, projectID).Scan(&owner); err != nil {
		writeJSON(w, 404, map[string]string{"error": "project not found"})
		return
	}
	if owner != c.UserID && c.Role != "admin" {
		writeJSON(w, 403, map[string]string{"error": "forbidden"})
		return
	}
	var in struct {
		Status string `json:"status"`
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid json"})
		return
	}
	if in.Status != "review" && in.Status != "approved" && in.Status != "archived" {
		writeJSON(w, 400, map[string]string{"error": "invalid status"})
		return
	}
	if in.Status == "approved" && c.Role != "admin" {
		writeJSON(w, 403, map[string]string{"error": "only admin can approve architecture versions"})
		return
	}
	res, err := s.DB.ExecContext(r.Context(), `UPDATE project_versions SET status=$1,approved_at=CASE WHEN $1='approved' THEN now() ELSE approved_at END WHERE project_id=$2 AND version=$3`, in.Status, projectID, version)
	if err != nil {
		writeError(w, err)
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		writeJSON(w, 404, map[string]string{"error": "version not found"})
		return
	}
	writeJSON(w, 200, map[string]any{"project_id": projectID, "version": version, "status": in.Status})
}

func (s *Server) createProject(w http.ResponseWriter, r *http.Request) {
	c, ok := claimsFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	var in struct {
		Name         string         `json:"name"`
		ObjectTypeID string         `json:"object_type_id"`
		Location     map[string]any `json:"location"`
		Parameters   map[string]any `json:"parameters"`
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid json"})
		return
	}
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		writeJSON(w, 400, map[string]string{"error": "name is required"})
		return
	}
	if in.Location == nil {
		in.Location = map[string]any{}
	}
	if in.Parameters == nil {
		in.Parameters = map[string]any{}
	}
	location, _ := json.Marshal(in.Location)
	parameters, _ := json.Marshal(in.Parameters)
	if in.ObjectTypeID != "" {
		var active bool
		if err := s.DB.QueryRowContext(r.Context(), `SELECT active FROM object_types WHERE id=$1`, in.ObjectTypeID).Scan(&active); err != nil || !active {
			writeJSON(w, 400, map[string]string{"error": "invalid object_type_id"})
			return
		}
	}
	var id string
	err := s.DB.QueryRowContext(r.Context(), `INSERT INTO projects(owner_id,name,object_type_id,location,parameters) VALUES($1,$2,NULLIF($3,'')::uuid,$4,$5) RETURNING id`, c.UserID, in.Name, in.ObjectTypeID, location, parameters).Scan(&id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"id": id, "owner_id": c.UserID, "name": in.Name, "architecture_version": 1, "status": "draft"})
}

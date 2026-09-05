package api

import (
	"encoding/json"
	"net/http"
	"strings"
)

func (s *Server) organizations(w http.ResponseWriter, r *http.Request) {
	c, _ := claimsFromContext(r.Context())
	rows, err := s.DB.QueryContext(r.Context(), `SELECT o.id,o.type,o.legal_name,o.verification_status,o.rating,o.created_at FROM organizations o JOIN organization_members m ON m.organization_id=o.id WHERE m.user_id=$1 ORDER BY o.created_at DESC`, c.UserID)
	if err != nil {
		writeError(w, err)
		return
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id, typ, name, status string
		var rating, created any
		if err := rows.Scan(&id, &typ, &name, &status, &rating, &created); err != nil {
			writeError(w, err)
			return
		}
		out = append(out, map[string]any{"id": id, "type": typ, "legal_name": name, "verification_status": status, "rating": rating, "created_at": created})
	}
	writeJSON(w, 200, out)
}
func (s *Server) createOrganization(w http.ResponseWriter, r *http.Request) {
	c, _ := claimsFromContext(r.Context())
	var in struct {
		Type, LegalName             string
		RegistrationData, Geography map[string]any
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid json"})
		return
	}
	in.Type = strings.TrimSpace(strings.ToLower(in.Type))
	in.LegalName = strings.TrimSpace(in.LegalName)
	if in.Type == "" {
		in.Type = c.Role
	}
	if in.Type != c.Role && c.Role != "admin" {
		writeJSON(w, 403, map[string]string{"error": "organization type must match user role"})
		return
	}
	if in.LegalName == "" {
		writeJSON(w, 400, map[string]string{"error": "legal_name is required"})
		return
	}
	reg, _ := json.Marshal(in.RegistrationData)
	geo, _ := json.Marshal(in.Geography)
	var id string
	err := s.DB.QueryRowContext(r.Context(), `INSERT INTO organizations(type,legal_name,registration_data,geography) VALUES($1,$2,$3,$4) RETURNING id`, in.Type, in.LegalName, reg, geo).Scan(&id)
	if err != nil {
		writeError(w, err)
		return
	}
	if _, err = s.DB.ExecContext(r.Context(), `INSERT INTO organization_members(organization_id,user_id,membership_role) VALUES($1,$2,'owner')`, id, c.UserID); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, 201, map[string]any{"id": id, "verification_status": "pending"})
}
func (s *Server) verifyOrganization(w http.ResponseWriter, r *http.Request) {
	var in struct{ OrganizationID, Status string }
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid json"})
		return
	}
	if in.Status != "verified" && in.Status != "rejected" {
		writeJSON(w, 400, map[string]string{"error": "status must be verified or rejected"})
		return
	}
	res, err := s.DB.ExecContext(r.Context(), `UPDATE organizations SET verification_status=$1 WHERE id=$2`, in.Status, in.OrganizationID)
	if err != nil {
		writeError(w, err)
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		writeJSON(w, 404, map[string]string{"error": "organization not found"})
		return
	}
	writeJSON(w, 200, map[string]string{"status": in.Status})
}
func (s *Server) myPermissions(w http.ResponseWriter, r *http.Request) {
	c, _ := claimsFromContext(r.Context())
	rows, err := s.DB.QueryContext(r.Context(), `SELECT p.code FROM permissions p JOIN role_permissions rp ON rp.permission_id=p.id WHERE rp.role=$1 ORDER BY p.code`, c.Role)
	if err != nil {
		writeError(w, err)
		return
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var v string
		if rows.Scan(&v) == nil {
			out = append(out, v)
		}
	}
	writeJSON(w, 200, map[string]any{"role": c.Role, "permissions": out})
}

package api

import (
	"encoding/json"
	"net/http"
	"strings"
)

func (s *Server) createContractorApplication(w http.ResponseWriter, r *http.Request) {
	c, ok := claimsFromContext(r.Context())
	if !ok {
		writeJSON(w, 401, map[string]string{"error": "unauthorized"})
		return
	}
	var in struct {
		OrganizationID string `json:"organization_id"`
		Note           string `json:"note"`
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil || strings.TrimSpace(in.OrganizationID) == "" {
		writeJSON(w, 400, map[string]string{"error": "organization_id is required"})
		return
	}
	var typ string
	if err := s.DB.QueryRowContext(r.Context(), `SELECT type FROM organizations WHERE id=$1`, in.OrganizationID).Scan(&typ); err != nil {
		writeJSON(w, 404, map[string]string{"error": "organization not found"})
		return
	}
	if typ != "contractor" && c.Role != "admin" {
		writeJSON(w, 403, map[string]string{"error": "organization must be contractor type"})
		return
	}
	var id string
	err := s.DB.QueryRowContext(r.Context(), `INSERT INTO contractor_applications(organization_id,applicant_user_id,note) VALUES($1,$2,$3) ON CONFLICT(organization_id,applicant_user_id) DO UPDATE SET status='pending',note=EXCLUDED.note,reviewed_at=NULL RETURNING id`, in.OrganizationID, c.UserID, in.Note).Scan(&id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, 201, map[string]any{"id": id, "status": "pending"})
}

func (s *Server) upsertContractorProfile(w http.ResponseWriter, r *http.Request) {
	c, ok := claimsFromContext(r.Context())
	if !ok {
		writeJSON(w, 401, map[string]string{"error": "unauthorized"})
		return
	}
	var in struct {
		OrganizationID  string         `json:"organization_id"`
		Summary         string         `json:"summary"`
		ExperienceYears int            `json:"experience_years"`
		ServiceArea     map[string]any `json:"service_area"`
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil || strings.TrimSpace(in.OrganizationID) == "" {
		writeJSON(w, 400, map[string]string{"error": "organization_id is required"})
		return
	}
	var allowed int
	err := s.DB.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM organization_members WHERE organization_id=$1 AND user_id=$2 AND membership_role IN ('owner','admin')`, in.OrganizationID, c.UserID).Scan(&allowed)
	if err != nil || (allowed == 0 && c.Role != "admin") {
		writeJSON(w, 403, map[string]string{"error": "forbidden"})
		return
	}
	if in.ExperienceYears < 0 {
		writeJSON(w, 400, map[string]string{"error": "experience_years must be >= 0"})
		return
	}
	if in.ServiceArea == nil {
		in.ServiceArea = map[string]any{}
	}
	area, _ := json.Marshal(in.ServiceArea)
	var id string
	err = s.DB.QueryRowContext(r.Context(), `INSERT INTO contractor_profiles(organization_id,summary,experience_years,service_area) VALUES($1,$2,$3,$4) ON CONFLICT(organization_id) DO UPDATE SET summary=EXCLUDED.summary,experience_years=EXCLUDED.experience_years,service_area=EXCLUDED.service_area,updated_at=now() RETURNING id`, in.OrganizationID, in.Summary, in.ExperienceYears, area).Scan(&id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, 201, map[string]any{"id": id, "organization_id": in.OrganizationID, "verification_status": "pending"})
}

func (s *Server) addContractorCompetency(w http.ResponseWriter, r *http.Request) {
	c, ok := claimsFromContext(r.Context())
	if !ok {
		writeJSON(w, 401, map[string]string{"error": "unauthorized"})
		return
	}
	id := strings.TrimSpace(r.PathValue("contractorID"))
	if id == "" {
		writeJSON(w, 400, map[string]string{"error": "contractor id is required"})
		return
	}
	var in struct {
		Code     string         `json:"code"`
		Name     string         `json:"name"`
		Level    int            `json:"level"`
		Evidence map[string]any `json:"evidence"`
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid json"})
		return
	}
	in.Code = strings.TrimSpace(strings.ToLower(in.Code))
	in.Name = strings.TrimSpace(in.Name)
	if in.Level == 0 {
		in.Level = 1
	}
	if in.Code == "" || in.Name == "" || in.Level < 1 || in.Level > 5 {
		writeJSON(w, 400, map[string]string{"error": "code, name and level 1..5 are required"})
		return
	}
	var allowed int
	if err := s.DB.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM contractor_profiles cp JOIN organization_members om ON om.organization_id=cp.organization_id WHERE cp.id=$1 AND om.user_id=$2 AND om.membership_role IN ('owner','admin')`, id, c.UserID).Scan(&allowed); err != nil || (allowed == 0 && c.Role != "admin") {
		writeJSON(w, 403, map[string]string{"error": "forbidden"})
		return
	}
	if in.Evidence == nil {
		in.Evidence = map[string]any{}
	}
	evidence, _ := json.Marshal(in.Evidence)
	var cid string
	err := s.DB.QueryRowContext(r.Context(), `INSERT INTO contractor_competencies(contractor_id,code,name,level,evidence) VALUES($1,$2,$3,$4,$5) ON CONFLICT(contractor_id,code) DO UPDATE SET name=EXCLUDED.name,level=EXCLUDED.level,evidence=EXCLUDED.evidence RETURNING id`, id, in.Code, in.Name, in.Level, evidence).Scan(&cid)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, 201, map[string]any{"id": cid, "contractor_id": id, "code": in.Code, "level": in.Level})
}

func (s *Server) verifyContractor(w http.ResponseWriter, r *http.Request) {
	var in struct {
		ContractorID string `json:"contractor_id"`
		Status       string `json:"status"`
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid json"})
		return
	}
	in.Status = strings.TrimSpace(strings.ToLower(in.Status))
	if in.Status != "verified" && in.Status != "rejected" && in.Status != "suspended" {
		writeJSON(w, 400, map[string]string{"error": "invalid status"})
		return
	}
	res, err := s.DB.ExecContext(r.Context(), `UPDATE contractor_profiles SET verification_status=$1,active=CASE WHEN $1='suspended' THEN false WHEN $1='verified' THEN true ELSE active END,updated_at=now() WHERE id=$2`, in.Status, in.ContractorID)
	if err != nil {
		writeError(w, err)
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		writeJSON(w, 404, map[string]string{"error": "contractor not found"})
		return
	}
	writeJSON(w, 200, map[string]string{"status": in.Status})
}

func (s *Server) contractors(w http.ResponseWriter, r *http.Request) {
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	skill := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("competency")))
	q := `SELECT cp.id,cp.organization_id,o.legal_name,cp.summary,cp.experience_years,cp.service_area,cp.verification_status,cp.active,cp.created_at FROM contractor_profiles cp JOIN organizations o ON o.id=cp.organization_id WHERE 1=1`
	args := []any{}
	n := 1
	if status != "" {
		q += ` AND cp.verification_status=$` + itoa(n)
		args = append(args, status)
		n++
	}
	if skill != "" {
		q += ` AND EXISTS (SELECT 1 FROM contractor_competencies cc WHERE cc.contractor_id=cp.id AND cc.code=$` + itoa(n) + `)`
		args = append(args, skill)
		n++
	}
	q += ` ORDER BY cp.created_at DESC LIMIT 100`
	rows, err := s.DB.QueryContext(r.Context(), q, args...)
	if err != nil {
		writeError(w, err)
		return
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id, org, name, summary, status string
		var years int
		var area any
		var active bool
		var created any
		if err := rows.Scan(&id, &org, &name, &summary, &years, &area, &status, &active, &created); err != nil {
			writeError(w, err)
			return
		}
		out = append(out, map[string]any{"id": id, "organization_id": org, "legal_name": name, "summary": summary, "experience_years": years, "service_area": area, "verification_status": status, "active": active, "created_at": created})
	}
	writeJSON(w, 200, out)
}

func itoa(v int) string {
	if v == 1 {
		return "1"
	}
	if v == 2 {
		return "2"
	}
	return "3"
}

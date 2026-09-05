package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/mintfary-oss/trest-sistems/internal/orders"
)

func (s *Server) createEstimate(w http.ResponseWriter, r *http.Request) {
	c, ok := claimsFromContext(r.Context())
	if !ok {
		writeJSON(w, 401, map[string]string{"error": "unauthorized"})
		return
	}
	var in struct {
		ProjectID string `json:"project_id"`
		ObjectID  string `json:"object_id"`
		Title     string `json:"title"`
		Currency  string `json:"currency"`
		Items     []struct {
			Name      string  `json:"name"`
			Category  string  `json:"category"`
			Unit      string  `json:"unit"`
			Quantity  float64 `json:"quantity"`
			UnitPrice float64 `json:"unit_price"`
		} `json:"items"`
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil || strings.TrimSpace(in.ProjectID) == "" || strings.TrimSpace(in.Title) == "" {
		writeJSON(w, 400, map[string]string{"error": "project_id and title are required"})
		return
	}
	if in.Currency == "" {
		in.Currency = "RUB"
	}
	tx, err := s.DB.BeginTx(r.Context(), nil)
	if err != nil {
		writeError(w, err)
		return
	}
	defer tx.Rollback()
	var id string
	if err = tx.QueryRowContext(r.Context(), `INSERT INTO estimates(project_id,object_id,title,currency,created_by) VALUES($1,NULLIF($2,'')::uuid,$3,$4,$5) RETURNING id`, in.ProjectID, in.ObjectID, in.Title, in.Currency, c.UserID).Scan(&id); err != nil {
		writeError(w, err)
		return
	}
	var vid string
	if err = tx.QueryRowContext(r.Context(), `INSERT INTO estimate_versions(estimate_id,version,status,created_by) VALUES($1,1,'draft',$2) RETURNING id`, id, c.UserID).Scan(&vid); err != nil {
		writeError(w, err)
		return
	}
	var total float64
	for _, item := range in.Items {
		line := item.Quantity * item.UnitPrice
		total += line
		if _, err = tx.ExecContext(r.Context(), `INSERT INTO estimate_version_items(estimate_version_id,name,category,unit,quantity,unit_price,total) VALUES($1,$2,$3,$4,$5,$6,$7)`, vid, item.Name, item.Category, item.Unit, item.Quantity, item.UnitPrice, line); err != nil {
			writeError(w, err)
			return
		}
	}
	if _, err = tx.ExecContext(r.Context(), `UPDATE estimate_versions SET subtotal=$1,total=$1 WHERE id=$2`, total, vid); err != nil {
		writeError(w, err)
		return
	}
	if err = tx.Commit(); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, 201, map[string]any{"id": id, "version_id": vid, "version": 1, "total": total, "status": "draft"})
}

func (s *Server) approveEstimate(w http.ResponseWriter, r *http.Request) {
	c, ok := claimsFromContext(r.Context())
	if !ok {
		writeJSON(w, 401, map[string]string{"error": "unauthorized"})
		return
	}
	var in struct {
		EstimateID string `json:"estimate_id"`
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil || in.EstimateID == "" {
		writeJSON(w, 400, map[string]string{"error": "estimate_id is required"})
		return
	}
	var owner string
	err := s.DB.QueryRowContext(r.Context(), `SELECT created_by FROM estimates WHERE id=$1`, in.EstimateID).Scan(&owner)
	if err != nil {
		writeJSON(w, 404, map[string]string{"error": "estimate not found"})
		return
	}
	if owner != c.UserID && c.Role != "admin" {
		writeJSON(w, 403, map[string]string{"error": "forbidden"})
		return
	}
	_, err = s.DB.ExecContext(r.Context(), `UPDATE estimates SET status='approved' WHERE id=$1`, in.EstimateID)
	if err != nil {
		writeError(w, err)
		return
	}
	_, err = s.DB.ExecContext(r.Context(), `UPDATE estimate_versions SET status='approved',approved_at=now() WHERE estimate_id=$1 AND version=(SELECT current_version FROM estimates WHERE id=$1)`, in.EstimateID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, 200, map[string]string{"status": "approved"})
}

func (s *Server) createOrder(w http.ResponseWriter, r *http.Request) {
	c, ok := claimsFromContext(r.Context())
	if !ok {
		writeJSON(w, 401, map[string]string{"error": "unauthorized"})
		return
	}
	var in struct {
		ProjectID         string  `json:"project_id"`
		ObjectID          string  `json:"object_id"`
		Title             string  `json:"title"`
		EstimateVersionID string  `json:"estimate_version_id"`
		IdempotencyKey    string  `json:"idempotency_key"`
		Amount            float64 `json:"amount"`
		PlannedStart      string  `json:"planned_start"`
		PlannedEnd        string  `json:"planned_end"`
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil || strings.TrimSpace(in.ProjectID) == "" || strings.TrimSpace(in.Title) == "" {
		writeJSON(w, 400, map[string]string{"error": "project_id and title are required"})
		return
	}
	if in.IdempotencyKey != "" {
		var id string
		err := s.DB.QueryRowContext(r.Context(), `SELECT id FROM orders WHERE idempotency_key=$1`, in.IdempotencyKey).Scan(&id)
		if err == nil {
			writeJSON(w, 200, map[string]any{"id": id, "status": orders.Draft, "idempotent": true})
			return
		}
	}
	var id string
	var start, end any
	if in.PlannedStart != "" {
		t, err := time.Parse(time.RFC3339, in.PlannedStart)
		if err != nil {
			writeJSON(w, 400, map[string]string{"error": "invalid planned_start"})
			return
		}
		start = t
	}
	if in.PlannedEnd != "" {
		t, err := time.Parse(time.RFC3339, in.PlannedEnd)
		if err != nil {
			writeJSON(w, 400, map[string]string{"error": "invalid planned_end"})
			return
		}
		end = t
	}
	err := s.DB.QueryRowContext(r.Context(), `INSERT INTO orders(project_id,object_id,title,status,amount,estimate_version_id,planned_start_at,planned_end_at,idempotency_key) VALUES($1,NULLIF($2,'')::uuid,$3,'draft',$4,NULLIF($5,'')::uuid,$6,$7,NULLIF($8,'')) RETURNING id`, in.ProjectID, in.ObjectID, in.Title, in.Amount, in.EstimateVersionID, start, end, in.IdempotencyKey).Scan(&id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, 201, map[string]any{"id": id, "status": orders.Draft, "created_by": c.UserID})
}

func (s *Server) transitionOrder(w http.ResponseWriter, r *http.Request) {
	c, ok := claimsFromContext(r.Context())
	if !ok {
		writeJSON(w, 401, map[string]string{"error": "unauthorized"})
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/orders/")
	id = strings.TrimSuffix(id, "/transition")
	var in struct {
		Status string `json:"status"`
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid json"})
		return
	}
	var current, owner string
	err := s.DB.QueryRowContext(r.Context(), `SELECT o.status,COALESCE(e.created_by,p.owner_id) FROM orders o JOIN projects p ON p.id=o.project_id LEFT JOIN estimate_versions ev ON ev.id=o.estimate_version_id LEFT JOIN estimates e ON e.id=ev.estimate_id WHERE o.id=$1`, id).Scan(&current, &owner)
	if err != nil {
		writeJSON(w, 404, map[string]string{"error": "order not found"})
		return
	}
	if owner != "" && owner != c.UserID && c.Role != "admin" {
		writeJSON(w, 403, map[string]string{"error": "forbidden"})
		return
	}
	if err = orders.ValidateTransition(current, in.Status); err != nil {
		writeJSON(w, 409, map[string]string{"error": err.Error()})
		return
	}
	_, err = s.DB.ExecContext(r.Context(), `UPDATE orders SET status=$1, approved_at=CASE WHEN $1='contracted' THEN COALESCE(approved_at,now()) ELSE approved_at END WHERE id=$2`, in.Status, id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, 200, map[string]any{"id": id, "from": current, "status": in.Status})
}

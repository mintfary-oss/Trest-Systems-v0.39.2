package api

import (
	"encoding/json"
	"net/http"
	"strings"
)

func (s *Server) bimExchangeStatus(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.URL.Query().Get("id"))
	if id == "" {
		writeJSON(w, 400, map[string]string{"error": "id is required"})
		return
	}
	var status, errText, input, output string
	var attempts, maxAttempts int
	var started, completed any
	err := s.DB.QueryRowContext(r.Context(), `SELECT status,error,input_uri,output_uri,attempts,max_attempts,started_at,completed_at FROM bim_import_exports WHERE id=$1`, id).Scan(&status, &errText, &input, &output, &attempts, &maxAttempts, &started, &completed)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, 200, map[string]any{"id": id, "status": status, "error": errText, "input_uri": input, "output_uri": output, "attempts": attempts, "max_attempts": maxAttempts, "started_at": started, "completed_at": completed})
}

func (s *Server) cancelBIMExchange(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID string `json:"id"`
	}
	if json.NewDecoder(r.Body).Decode(&req) != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid JSON"})
		return
	}
	req.ID = strings.TrimSpace(req.ID)
	if req.ID == "" {
		writeJSON(w, 400, map[string]string{"error": "id is required"})
		return
	}
	res, err := s.DB.ExecContext(r.Context(), `UPDATE bim_import_exports SET status='cancelled',error='cancelled by user',completed_at=now() WHERE id=$1 AND status='queued'`, req.ID)
	if err != nil {
		writeError(w, err)
		return
	}
	affected, err := res.RowsAffected()
	if err != nil {
		writeError(w, err)
		return
	}
	if affected == 0 {
		writeJSON(w, 409, map[string]string{"error": "job is not queued or does not exist"})
		return
	}
	writeJSON(w, 200, map[string]any{"id": req.ID, "status": "cancelled"})
}

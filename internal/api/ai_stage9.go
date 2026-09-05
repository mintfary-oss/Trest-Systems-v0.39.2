package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/mintfary-oss/trest-sistems/internal/ai"
)

func (s *Server) aiModels(w http.ResponseWriter, r *http.Request) {
	rows, err := s.DB.QueryContext(r.Context(), `SELECT m.id,p.name,m.name,m.family,m.capabilities,m.active,m.is_default FROM ai_models m JOIN ai_providers p ON p.id=m.provider_id ORDER BY p.name,m.name`)
	if err != nil {
		writeError(w, err)
		return
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id, p, n, f string
		var caps any
		var active, def bool
		if err := rows.Scan(&id, &p, &n, &f, &caps, &active, &def); err != nil {
			writeError(w, err)
			return
		}
		out = append(out, map[string]any{"id": id, "provider": p, "name": n, "family": f, "capabilities": caps, "active": active, "is_default": def})
	}
	writeJSON(w, 200, out)
}

func (s *Server) createAgent(w http.ResponseWriter, r *http.Request) {
	c, _ := claimsFromContext(r.Context())
	var in struct {
		Name, Description, Instructions                 string
		ModelVersionID                                  string
		Tools, Permissions, MemoryPolicy, SandboxPolicy json.RawMessage
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid json"})
		return
	}
	if strings.TrimSpace(in.Name) == "" || !ai.ValidateInstructions(in.Instructions) {
		writeJSON(w, 400, map[string]string{"error": "name and instructions are required"})
		return
	}
	defaults := func(v json.RawMessage, d string) json.RawMessage {
		if len(v) == 0 {
			return json.RawMessage(d)
		}
		return v
	}
	var id string
	err := s.DB.QueryRowContext(r.Context(), `INSERT INTO agents(owner_user_id,name,description,instructions,model_version_id,tools,permissions,memory_policy,sandbox_policy) VALUES($1,$2,$3,$4,NULLIF($5,'')::uuid,$6,$7,$8,$9) RETURNING id`, c.UserID, in.Name, in.Description, in.Instructions, in.ModelVersionID, defaults(in.Tools, "[]"), defaults(in.Permissions, "[]"), defaults(in.MemoryPolicy, "{}"), defaults(in.SandboxPolicy, "{}")).Scan(&id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, 201, map[string]any{"id": id, "name": in.Name})
}

func (s *Server) agents(w http.ResponseWriter, r *http.Request) {
	c, _ := claimsFromContext(r.Context())
	rows, err := s.DB.QueryContext(r.Context(), `SELECT id,name,description,instructions,model_version_id,tools,permissions,memory_policy,sandbox_policy,enabled,created_at,updated_at FROM agents WHERE owner_user_id=$1 OR $2='admin' ORDER BY created_at DESC`, c.UserID, c.Role)
	if err != nil {
		writeError(w, err)
		return
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id, n, d, i string
		var mv any
		var tools, perm, mem, sand any
		var en bool
		var ca, ua any
		if err := rows.Scan(&id, &n, &d, &i, &mv, &tools, &perm, &mem, &sand, &en, &ca, &ua); err != nil {
			writeError(w, err)
			return
		}
		out = append(out, map[string]any{"id": id, "name": n, "description": d, "instructions": i, "model_version_id": mv, "tools": tools, "permissions": perm, "memory_policy": mem, "sandbox_policy": sand, "enabled": en, "created_at": ca, "updated_at": ua})
	}
	writeJSON(w, 200, out)
}

func (s *Server) createTrainingDataset(w http.ResponseWriter, r *http.Request) {
	c, _ := claimsFromContext(r.Context())
	var in struct {
		Name, Description, Format, SourceURI, Version, Checksum string
		RowCount                                                int
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid json"})
		return
	}
	if strings.TrimSpace(in.Name) == "" || strings.TrimSpace(in.Version) == "" {
		writeJSON(w, 400, map[string]string{"error": "name and version are required"})
		return
	}
	if in.Format == "" {
		in.Format = "jsonl"
	}
	var id string
	err := s.DB.QueryRowContext(r.Context(), `INSERT INTO ai_datasets(owner_user_id,name,description,format,source_uri,version,row_count,checksum,status) VALUES($1,$2,$3,$4,$5,$6,$7,$8,'ready') RETURNING id`, c.UserID, in.Name, in.Description, in.Format, in.SourceURI, in.Version, in.RowCount, in.Checksum).Scan(&id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, 201, map[string]any{"id": id, "status": "ready"})
}

func (s *Server) createTrainingJob(w http.ResponseWriter, r *http.Request) {
	c, _ := claimsFromContext(r.Context())
	var in struct {
		DatasetID, BaseModelVersionID, Method string
		Config                                json.RawMessage
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid json"})
		return
	}
	if in.Method == "" {
		in.Method = "qlora"
	}
	if in.Method != "lora" && in.Method != "qlora" && in.Method != "full" {
		writeJSON(w, 400, map[string]string{"error": "unsupported training method"})
		return
	}
	if len(in.Config) == 0 {
		in.Config = json.RawMessage(`{}`)
	}
	var id string
	err := s.DB.QueryRowContext(r.Context(), `INSERT INTO ai_training_jobs(dataset_id,base_model_version_id,method,config,created_by) VALUES($1,$2,$3,$4,$5) RETURNING id`, in.DatasetID, in.BaseModelVersionID, in.Method, in.Config, c.UserID).Scan(&id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, 202, map[string]any{"id": id, "status": "queued", "method": in.Method})
}

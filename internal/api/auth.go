package api

import (
	"context"
	"encoding/json"
	"github.com/mintfary-oss/trest-sistems/internal/auth"
	"net/http"
	"strings"
	"time"
)

type ctxKey string

const claimsKey ctxKey = "auth_claims"

func withAuth(svc *auth.Service, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			writeJSON(w, 401, map[string]string{"error": "missing bearer token"})
			return
		}
		c, err := svc.Parse(strings.TrimSpace(strings.TrimPrefix(h, "Bearer ")))
		if err != nil {
			writeJSON(w, 401, map[string]string{"error": "unauthorized"})
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), claimsKey, c)))
	})
}
func claimsFromContext(ctx context.Context) (auth.Claims, bool) {
	c, ok := ctx.Value(claimsKey).(auth.Claims)
	return c, ok
}

func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	var in struct{ Email, Name, Password, Role string }
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid json"})
		return
	}
	in.Email = strings.TrimSpace(strings.ToLower(in.Email))
	in.Name = strings.TrimSpace(in.Name)
	if in.Role == "" {
		in.Role = "customer"
	}
	if in.Email == "" || in.Name == "" {
		writeJSON(w, 400, map[string]string{"error": "email and name are required"})
		return
	}
	if in.Role != "customer" && in.Role != "contractor" && in.Role != "supplier" {
		writeJSON(w, 400, map[string]string{"error": "invalid role"})
		return
	}
	ph, err := s.Auth.HashPassword(in.Password)
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}
	var id string
	err = s.DB.QueryRowContext(r.Context(), `INSERT INTO users(email,name,role,password_hash) VALUES($1,$2,$3,$4) RETURNING id`, in.Email, in.Name, in.Role, ph).Scan(&id)
	if err != nil {
		writeJSON(w, 409, map[string]string{"error": "email already exists or database error"})
		return
	}
	writeJSON(w, 201, map[string]any{"id": id, "email": in.Email, "name": in.Name, "role": in.Role})
}
func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var in struct{ Email, Password string }
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid json"})
		return
	}
	var id, name, role, ph string
	err := s.DB.QueryRowContext(r.Context(), `SELECT id,name,role,password_hash FROM users WHERE lower(email)=lower($1)`, strings.TrimSpace(in.Email)).Scan(&id, &name, &role, &ph)
	if err != nil || !s.Auth.VerifyPassword(in.Password, ph) {
		writeJSON(w, 401, map[string]string{"error": "invalid credentials"})
		return
	}
	tok, _ := s.Auth.Issue(auth.Claims{UserID: id, Role: role, ExpiresAt: time.Now().Add(24 * time.Hour)})
	writeJSON(w, 200, map[string]any{"access_token": tok, "token_type": "Bearer", "expires_in": 86400, "user": map[string]string{"id": id, "name": name, "role": role}})
}
func (s *Server) me(w http.ResponseWriter, r *http.Request) {
	c, _ := claimsFromContext(r.Context())
	var email, name, role string
	err := s.DB.QueryRowContext(r.Context(), `SELECT email,name,role FROM users WHERE id=$1`, c.UserID).Scan(&email, &name, &role)
	if err != nil {
		writeJSON(w, 404, map[string]string{"error": "user not found"})
		return
	}
	writeJSON(w, 200, map[string]string{"id": c.UserID, "email": email, "name": name, "role": role})
}

func requireRole(role string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, ok := claimsFromContext(r.Context())
		if !ok || c.Role != role {
			writeJSON(w, 403, map[string]string{"error": "forbidden"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

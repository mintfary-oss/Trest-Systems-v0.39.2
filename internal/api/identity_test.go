package api

import (
	"context"
	"github.com/mintfary-oss/trest-sistems/internal/auth"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequireRole(t *testing.T) {
	h := requireRole("admin", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) }))
	req := httptest.NewRequest("GET", "/", nil).WithContext(context.WithValue(context.Background(), claimsKey, auth.Claims{UserID: "1", Role: "customer"}))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != 403 {
		t.Fatalf("got %d", rr.Code)
	}
	req = httptest.NewRequest("GET", "/", nil).WithContext(context.WithValue(context.Background(), claimsKey, auth.Claims{UserID: "1", Role: "admin"}))
	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != 204 {
		t.Fatalf("got %d", rr.Code)
	}
}

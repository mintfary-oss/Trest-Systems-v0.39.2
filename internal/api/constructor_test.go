package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mintfary-oss/trest-sistems/internal/auth"
)

func TestProjectVersionRouteRequiresAuth(t *testing.T) {
	s := &Server{Auth: auth.New("test-secret")}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/1/versions", nil)
	s.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("got %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestConstructorCatalogRoutesRequireAuth(t *testing.T) {
	s := &Server{Auth: auth.New("test-secret")}
	for _, path := range []string{"/api/v1/catalog/object-types", "/api/v1/catalog/materials", "/api/v1/catalog/engineering-systems", "/api/v1/catalog/finishes"} {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		s.Handler().ServeHTTP(rr, req)
		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("%s: got %d, want %d", path, rr.Code, http.StatusUnauthorized)
		}
	}
}

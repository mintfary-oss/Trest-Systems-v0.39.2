package api

import (
	"fmt"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiter(t *testing.T) {
	l := newRateLimiter(2, time.Minute)
	if !l.Allow("a") || !l.Allow("a") {
		t.Fatal("first two requests must pass")
	}
	if l.Allow("a") {
		t.Fatal("third request must be rejected")
	}
	if !l.Allow("b") {
		t.Fatal("different client must have its own bucket")
	}
}

func TestClientIPDoesNotTrustForwardedByDefault(t *testing.T) {
	r := httptest.NewRequest("GET", "/health", nil)
	r.RemoteAddr = "10.0.0.8:1234"
	r.Header.Set("X-Forwarded-For", "192.0.2.1")
	if got := clientIP(r); got != "10.0.0.8" {
		t.Fatalf("clientIP=%q", got)
	}
}

func TestRateLimiterBoundsDistinctClientKeys(t *testing.T) {
	l := newRateLimiter(1, time.Minute)
	for i := 0; i < l.maxBuckets; i++ {
		if !l.Allow(fmt.Sprintf("client-%d", i)) {
			t.Fatalf("client %d unexpectedly rejected", i)
		}
	}
	if l.Allow("overflow-client") {
		t.Fatal("expected bounded limiter to reject overflow key")
	}
}

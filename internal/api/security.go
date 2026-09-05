package api

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// simpleRateLimiter is an in-process safety guard. Production deployments behind
// multiple API replicas should additionally enforce limits at the edge/gateway.
type simpleRateLimiter struct {
	mu         sync.Mutex
	buckets    map[string]*rateBucket
	limit      int
	window     time.Duration
	maxBuckets int
}

type rateBucket struct {
	start time.Time
	count int
}

func newRateLimiter(limit int, window time.Duration) *simpleRateLimiter {
	return &simpleRateLimiter{buckets: make(map[string]*rateBucket), limit: limit, window: window, maxBuckets: 10000}
}

func (l *simpleRateLimiter) Allow(key string) bool {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()
	// Bound memory usage when many distinct client keys are presented.
	if len(l.buckets) >= l.maxBuckets {
		for k, b := range l.buckets {
			if now.Sub(b.start) >= l.window {
				delete(l.buckets, k)
			}
		}
		if len(l.buckets) >= l.maxBuckets {
			return false
		}
	}
	b := l.buckets[key]
	if b == nil || now.Sub(b.start) >= l.window {
		l.buckets[key] = &rateBucket{start: now, count: 1}
		return true
	}
	if b.count >= l.limit {
		return false
	}
	b.count++
	return true
}

func clientIP(r *http.Request) string {
	// Trust X-Forwarded-For only when the deployment explicitly places the API
	// behind a trusted reverse proxy. Otherwise RemoteAddr is authoritative.
	if r.Header.Get("X-Trust-Proxy") == "1" {
		if x := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0]); x != "" {
			return x
		}
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil && host != "" {
		return host
	}
	return strings.TrimSpace(r.RemoteAddr)
}

func rateLimit(next http.Handler) http.Handler {
	limiter := newRateLimiter(120, time.Minute)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow(clientIP(r)) {
			w.Header().Set("Retry-After", "60")
			writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "rate limit exceeded"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

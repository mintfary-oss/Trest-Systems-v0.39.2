// Package web serves the orchestrator's latest report as a live HTTP
// dashboard, so users get a single URL to check the status of all three
// sub-projects instead of tailing individual logs and docker compose ps
// output.
package web

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/mintfary-oss/trest-sistems/trest-installer/internal/report"
)

// Server serves the most recently set report over HTTP. It is safe for
// concurrent use: SetReport may be called from an orchestration loop while
// handlers concurrently serve requests.
type Server struct {
	mu             sync.RWMutex
	report         *report.Report
	webhookSecret  string
	webhookHandler func()
}

// NewServer returns a Server with no report set yet. Handlers respond with
// 503 until SetReport is called at least once.
func NewServer() *Server {
	return &Server{}
}

// SetReport replaces the report served by subsequent requests.
func (s *Server) SetReport(r *report.Report) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.report = r
}

// current returns the most recently set report, or nil if none has been
// set yet.
func (s *Server) current() *report.Report {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.report
}

// SetWebhookHandler enables POST /webhook: requests bearing a valid
// HMAC-SHA256 signature (GitHub's "X-Hub-Signature-256" scheme) over
// secret trigger fn in a new goroutine. secret must be non-empty and fn
// non-nil, or the endpoint stays disabled (501) — there is no unsigned
// fallback, since an unauthenticated trigger would let anyone on the
// network force a rebuild/redeploy.
func (s *Server) SetWebhookHandler(secret string, fn func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.webhookSecret = secret
	s.webhookHandler = fn
}

// webhookConfigured reports whether SetWebhookHandler was called with a
// non-empty secret and handler.
func (s *Server) webhookConfigured() (secret string, fn func(), ok bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.webhookSecret, s.webhookHandler, s.webhookSecret != "" && s.webhookHandler != nil
}

// Handler returns the HTTP handler for the dashboard: "/" redirects to
// "/report", "/report" serves the HTML dashboard, "/report.json" serves the
// same data as JSON for programmatic consumers, and "/webhook" triggers an
// immediate orchestration run when configured via SetWebhookHandler.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/report", s.handleReportHTML)
	mux.HandleFunc("/report.json", s.handleReportJSON)
	mux.HandleFunc("/webhook", s.handleWebhook)
	return mux
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, "/report", http.StatusFound)
}

func (s *Server) handleReportHTML(w http.ResponseWriter, r *http.Request) {
	rep := s.current()
	if rep == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintln(w, "no report available yet; run the orchestrator first")
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := rep.Render(w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleReportJSON(w http.ResponseWriter, r *http.Request) {
	rep := s.current()
	if rep == nil {
		http.Error(w, "no report available yet", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(rep); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleWebhook(w http.ResponseWriter, r *http.Request) {
	secret, fn, ok := s.webhookConfigured()
	if !ok {
		http.Error(w, "webhook not configured", http.StatusNotImplemented)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	if !validSignature(secret, body, r.Header.Get("X-Hub-Signature-256")) {
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	// Run asynchronously and acknowledge immediately: the orchestration
	// pass (git fetch, docker compose up, go build) can easily exceed the
	// short timeout webhook senders (e.g. GitHub) enforce on the response.
	go fn()

	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintln(w, "accepted")
}

// validSignature reports whether header matches GitHub's
// "X-Hub-Signature-256: sha256=<hex-hmac>" scheme for body, computed with
// secret.
func validSignature(secret string, body []byte, header string) bool {
	const prefix = "sha256="
	if secret == "" || !strings.HasPrefix(header, prefix) {
		return false
	}

	sig, err := hex.DecodeString(strings.TrimPrefix(header, prefix))
	if err != nil {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hmac.Equal(sig, mac.Sum(nil))
}

// ListenAndServe starts the dashboard on addr (e.g. ":9091"), blocking
// until the server stops or an error occurs.
func (s *Server) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, s.Handler())
}

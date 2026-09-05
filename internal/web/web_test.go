package web

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mintfary-oss/trest-sistems/internal/orchestrator"
	"github.com/mintfary-oss/trest-sistems/internal/report"
)

// sign computes the "sha256=<hex>" GitHub-style signature for body under
// secret.
func sign(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func TestReportRoutesReturn503BeforeAnyReport(t *testing.T) {
	s := NewServer()
	srv := httptest.NewServer(s.Handler())
	defer srv.Close()

	for _, path := range []string{"/report", "/report.json"} {
		resp, err := http.Get(srv.URL + path)
		if err != nil {
			t.Fatalf("GET %s: %v", path, err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusServiceUnavailable {
			t.Errorf("GET %s: status = %d, want %d", path, resp.StatusCode, http.StatusServiceUnavailable)
		}
	}
}

func TestIndexRedirectsToReport(t *testing.T) {
	s := NewServer()
	srv := httptest.NewServer(s.Handler())
	defer srv.Close()

	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	resp, err := client.Get(srv.URL + "/")
	if err != nil {
		t.Fatalf("GET /: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusFound)
	}
	if loc := resp.Header.Get("Location"); loc != "/report" {
		t.Errorf("Location = %q, want %q", loc, "/report")
	}
}

func TestReportHTMLServesLatestReport(t *testing.T) {
	s := NewServer()
	s.SetReport(report.Build([]orchestrator.Result{{Project: "magasin-777", Action: "compose-up"}}, nil, nil))

	srv := httptest.NewServer(s.Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/report")
	if err != nil {
		t.Fatalf("GET /report: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if !strings.Contains(string(body), "magasin-777") {
		t.Errorf("body missing project name, got: %s", body)
	}
}

func TestReportJSONServesLatestReport(t *testing.T) {
	s := NewServer()
	s.SetReport(report.Build([]orchestrator.Result{{Project: "proektirovka-sdaniy", Action: "go-build"}}, nil, nil))

	srv := httptest.NewServer(s.Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/report.json")
	if err != nil {
		t.Fatalf("GET /report.json: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if !strings.Contains(string(body), "proektirovka-sdaniy") {
		t.Errorf("body missing project name, got: %s", body)
	}
}

func TestWebhookDisabledByDefault(t *testing.T) {
	s := NewServer()
	srv := httptest.NewServer(s.Handler())
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/webhook", "application/json", bytes.NewReader([]byte("{}")))
	if err != nil {
		t.Fatalf("POST /webhook: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotImplemented {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusNotImplemented)
	}
}

func TestWebhookRejectsMissingOrInvalidSignature(t *testing.T) {
	s := NewServer()
	var calls int32
	s.SetWebhookHandler("shh", func() { atomic.AddInt32(&calls, 1) })

	srv := httptest.NewServer(s.Handler())
	defer srv.Close()

	body := []byte(`{"ref":"refs/heads/main"}`)

	cases := []struct {
		name   string
		header string
	}{
		{"missing", ""},
		{"wrong secret", sign("wrong-secret", body)},
		{"malformed", "sha256=not-hex"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, srv.URL+"/webhook", bytes.NewReader(body))
			if err != nil {
				t.Fatalf("NewRequest: %v", err)
			}
			if tc.header != "" {
				req.Header.Set("X-Hub-Signature-256", tc.header)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("POST /webhook: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusUnauthorized {
				t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
			}
		})
	}

	if got := atomic.LoadInt32(&calls); got != 0 {
		t.Errorf("handler called %d times, want 0", got)
	}
}

func TestWebhookRejectsNonPostMethod(t *testing.T) {
	s := NewServer()
	s.SetWebhookHandler("shh", func() {})

	srv := httptest.NewServer(s.Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/webhook")
	if err != nil {
		t.Fatalf("GET /webhook: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusMethodNotAllowed)
	}
}

func TestWebhookTriggersHandlerOnValidSignature(t *testing.T) {
	s := NewServer()
	done := make(chan struct{})
	s.SetWebhookHandler("shh", func() { close(done) })

	srv := httptest.NewServer(s.Handler())
	defer srv.Close()

	body := []byte(`{"ref":"refs/heads/main"}`)
	req, err := http.NewRequest(http.MethodPost, srv.URL+"/webhook", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	req.Header.Set("X-Hub-Signature-256", sign("shh", body))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST /webhook: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusAccepted)
	}

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("webhook handler was not invoked within timeout")
	}
}

package report

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mintfary-oss/trest-sistems/trest-installer/internal/diagnostics"
	"github.com/mintfary-oss/trest-sistems/trest-installer/internal/orchestrator"
)

func TestBuildAndHealthy(t *testing.T) {
	results := []orchestrator.Result{
		{Project: "a", Action: "compose-up", Output: "started"},
		{Project: "b", Action: "go-build", Output: "built"},
	}

	r := Build(results, nil, nil)

	if !r.Healthy() {
		t.Error("Healthy() = false, want true for all-successful results with no alerts")
	}
	if len(r.Projects) != 2 {
		t.Fatalf("Projects = %d, want 2", len(r.Projects))
	}
	if !r.Projects[0].Success || !r.Projects[1].Success {
		t.Errorf("Projects = %+v, want all successful", r.Projects)
	}
}

func TestBuildUnhealthyOnProjectFailure(t *testing.T) {
	results := []orchestrator.Result{
		{Project: "a", Action: "compose-up", Err: errors.New("port in use")},
	}

	r := Build(results, nil, nil)

	if r.Healthy() {
		t.Error("Healthy() = true, want false when a project failed")
	}
	if r.Projects[0].Success {
		t.Error("Projects[0].Success = true, want false")
	}
	if r.Projects[0].Error != "port in use" {
		t.Errorf("Projects[0].Error = %q, want %q", r.Projects[0].Error, "port in use")
	}
}

func TestBuildUnhealthyOnLogAlerts(t *testing.T) {
	results := []orchestrator.Result{{Project: "a", Action: "compose-up"}}
	alerts := []diagnostics.LogAlert{{Source: "a.log", Line: 1, Level: "ERROR", Text: "boom"}}

	r := Build(results, nil, alerts)

	if r.Healthy() {
		t.Error("Healthy() = true, want false when log alerts are present")
	}
}

func TestWriteJSONRoundTrip(t *testing.T) {
	r := Build([]orchestrator.Result{{Project: "a", Action: "compose-up"}}, nil, nil)
	path := filepath.Join(t.TempDir(), "report.json")

	if err := r.WriteJSON(path); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read report: %v", err)
	}

	var decoded Report
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal report: %v", err)
	}
	if len(decoded.Projects) != 1 || decoded.Projects[0].Name != "a" {
		t.Errorf("decoded.Projects = %+v, want one project named %q", decoded.Projects, "a")
	}
}

func TestLoadReadsWrittenReport(t *testing.T) {
	r := Build([]orchestrator.Result{{Project: "a", Action: "compose-up"}}, nil, nil)
	path := filepath.Join(t.TempDir(), "report.json")
	if err := r.WriteJSON(path); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(loaded.Projects) != 1 || loaded.Projects[0].Name != "a" {
		t.Errorf("loaded.Projects = %+v, want one project named %q", loaded.Projects, "a")
	}
}

func TestLoadMissingFile(t *testing.T) {
	if _, err := Load(filepath.Join(t.TempDir(), "missing.json")); err == nil {
		t.Fatal("Load: want error for missing file, got nil")
	}
}

func TestWriteHTMLContainsProjectAndAlertData(t *testing.T) {
	results := []orchestrator.Result{{Project: "magasin-777", Action: "compose-up", Err: errors.New("port 8080 busy")}}
	ports := []diagnostics.PortStatus{{Port: 8080, Free: false}}
	alerts := []diagnostics.LogAlert{{Source: "app.log", Line: 3, Level: "ERROR", Text: "db down"}}

	r := Build(results, ports, alerts)
	path := filepath.Join(t.TempDir(), "report.html")

	if err := r.WriteHTML(path); err != nil {
		t.Fatalf("WriteHTML: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read report: %v", err)
	}
	html := string(data)

	for _, want := range []string{"magasin-777", "port 8080 busy", "8080", "db down"} {
		if !strings.Contains(html, want) {
			t.Errorf("report.html missing expected content %q", want)
		}
	}
}

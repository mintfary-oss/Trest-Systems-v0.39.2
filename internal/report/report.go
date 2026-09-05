// Package report assembles the orchestrator's findings (project start
// results, port availability, log alerts) into a machine-readable
// report.json and a human-readable report.html.
package report

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"os"
	"time"

	"github.com/mintfary-oss/trest-sistems/internal/diagnostics"
	"github.com/mintfary-oss/trest-sistems/internal/orchestrator"
)

// ProjectReport is the reported outcome for a single project.
type ProjectReport struct {
	Name    string `json:"name"`
	Action  string `json:"action"`
	Success bool   `json:"success"`
	Output  string `json:"output,omitempty"`
	Error   string `json:"error,omitempty"`
}

// Report is the full set of findings from one orchestration run.
type Report struct {
	GeneratedAt time.Time                `json:"generated_at"`
	Projects    []ProjectReport          `json:"projects"`
	Ports       []diagnostics.PortStatus `json:"ports,omitempty"`
	LogAlerts   []diagnostics.LogAlert   `json:"log_alerts,omitempty"`
}

// Healthy reports whether every project succeeded, every checked port was
// free (or otherwise not flagged), and no log alerts were found.
func (r *Report) Healthy() bool {
	for _, p := range r.Projects {
		if !p.Success {
			return false
		}
	}
	return len(r.LogAlerts) == 0
}

// Build assembles a Report from orchestration and diagnostics results.
func Build(results []orchestrator.Result, ports []diagnostics.PortStatus, alerts []diagnostics.LogAlert) *Report {
	projects := make([]ProjectReport, 0, len(results))
	for _, res := range results {
		pr := ProjectReport{Name: res.Project, Action: res.Action, Success: res.Err == nil, Output: res.Output}
		if res.Err != nil {
			pr.Error = res.Err.Error()
		}
		projects = append(projects, pr)
	}

	return &Report{
		GeneratedAt: time.Now().UTC(),
		Projects:    projects,
		Ports:       ports,
		LogAlerts:   alerts,
	}
}

// Load reads a previously written report.json from path.
func Load(path string) (*Report, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read report %q: %w", path, err)
	}
	var r Report
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("parse report %q: %w", path, err)
	}
	return &r, nil
}

// WriteJSON writes r to path as indented JSON.
func (r *Report) WriteJSON(path string) error {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal report: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write report %q: %w", path, err)
	}
	return nil
}

// WriteHTML renders r as a static HTML page at path.
func (r *Report) WriteHTML(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create report %q: %w", path, err)
	}
	defer f.Close()

	if err := r.Render(f); err != nil {
		return fmt.Errorf("render report %q: %w", path, err)
	}
	return nil
}

// Render writes r as an HTML page to w. It is exported so the live web
// dashboard (see internal/web) can serve the same markup as the static
// report.html without duplicating the template.
func (r *Report) Render(w io.Writer) error {
	return htmlTemplate.Execute(w, r)
}

// htmlTemplate renders a Report as a self-contained HTML page. It is also
// reused by the live web dashboard (see internal/web) to keep the served
// status page and the static report.html visually consistent.
var htmlTemplate = template.Must(template.New("report").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>Trest Installer Report</title>
  <style>
    body { font-family: system-ui, sans-serif; margin: 2rem; color: #1a1a1a; }
    h1 { margin-bottom: 0.25rem; }
    .meta { color: #666; margin-bottom: 1.5rem; }
    table { border-collapse: collapse; width: 100%; margin-bottom: 2rem; }
    th, td { border: 1px solid #ddd; padding: 0.5rem 0.75rem; text-align: left; vertical-align: top; }
    th { background: #f5f5f5; }
    .ok { color: #1a7f37; font-weight: 600; }
    .fail { color: #c0392b; font-weight: 600; }
    .free { color: #1a7f37; }
    .busy { color: #c0392b; }
    pre { white-space: pre-wrap; margin: 0; font-size: 0.85em; }
  </style>
</head>
<body>
  <h1>Trest Installer Report</h1>
  <p class="meta">Generated at {{ .GeneratedAt }}</p>

  <h2>Projects</h2>
  <table>
    <tr><th>Name</th><th>Action</th><th>Status</th><th>Output / Error</th></tr>
    {{ range .Projects }}
    <tr>
      <td>{{ .Name }}</td>
      <td>{{ .Action }}</td>
      <td>{{ if .Success }}<span class="ok">OK</span>{{ else }}<span class="fail">FAILED</span>{{ end }}</td>
      <td><pre>{{ if .Error }}{{ .Error }}{{ else }}{{ .Output }}{{ end }}</pre></td>
    </tr>
    {{ end }}
  </table>

  {{ if .Ports }}
  <h2>Ports</h2>
  <table>
    <tr><th>Port</th><th>Status</th></tr>
    {{ range .Ports }}
    <tr>
      <td>{{ .Port }}</td>
      <td>{{ if .Free }}<span class="free">free</span>{{ else }}<span class="busy">busy</span>{{ end }}</td>
    </tr>
    {{ end }}
  </table>
  {{ end }}

  {{ if .LogAlerts }}
  <h2>Log Alerts</h2>
  <table>
    <tr><th>Source</th><th>Line</th><th>Level</th><th>Text</th></tr>
    {{ range .LogAlerts }}
    <tr>
      <td>{{ .Source }}</td>
      <td>{{ .Line }}</td>
      <td>{{ .Level }}</td>
      <td><pre>{{ .Text }}</pre></td>
    </tr>
    {{ end }}
  </table>
  {{ end }}
</body>
</html>
`))

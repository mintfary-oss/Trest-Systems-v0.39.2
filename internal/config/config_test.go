package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempConfig(t *testing.T, contents string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write temp config: %v", err)
	}
	return path
}

func TestLoadAppliesDefaults(t *testing.T) {
	path := writeTempConfig(t, `
repo_url: https://example.com/org/repo.git
audit_port: 9091
projects:
  - name: svc
    type: compose
    path: svc
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.Branch != defaultBranch {
		t.Errorf("Branch = %q, want %q", cfg.Branch, defaultBranch)
	}
	if got := cfg.Projects[0].ComposeFile; got != defaultComposeFile {
		t.Errorf("ComposeFile = %q, want %q", got, defaultComposeFile)
	}
}

func TestLoadRejectsMissingRepoURL(t *testing.T) {
	path := writeTempConfig(t, `
audit_port: 9091
projects:
  - name: svc
    type: compose
    path: svc
`)

	if _, err := Load(path); err == nil {
		t.Fatal("Load: want error for missing repo_url, got nil")
	}
}

func TestLoadRejectsDuplicateProjectNames(t *testing.T) {
	path := writeTempConfig(t, `
repo_url: https://example.com/org/repo.git
audit_port: 9091
projects:
  - name: svc
    type: compose
    path: a
  - name: svc
    type: compose
    path: b
`)

	if _, err := Load(path); err == nil {
		t.Fatal("Load: want error for duplicate project name, got nil")
	}
}

func TestLoadRejectsGoBuildWithoutBuildCmd(t *testing.T) {
	path := writeTempConfig(t, `
repo_url: https://example.com/org/repo.git
audit_port: 9091
projects:
  - name: cli
    type: go-build
    path: cli
`)

	if _, err := Load(path); err == nil {
		t.Fatal("Load: want error for go-build without build_cmd, got nil")
	}
}

func TestLoadRejectsInvalidAuditPort(t *testing.T) {
	path := writeTempConfig(t, `
repo_url: https://example.com/org/repo.git
audit_port: 0
projects:
  - name: svc
    type: compose
    path: svc
`)

	if _, err := Load(path); err == nil {
		t.Fatal("Load: want error for invalid audit_port, got nil")
	}
}

func TestLoadMissingFile(t *testing.T) {
	if _, err := Load(filepath.Join(t.TempDir(), "missing.yaml")); err == nil {
		t.Fatal("Load: want error for missing file, got nil")
	}
}

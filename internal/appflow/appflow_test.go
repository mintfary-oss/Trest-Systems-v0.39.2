package appflow

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mintfary-oss/trest-sistems/internal/config"
)

// mustRun runs git in dir and fails the test on error.
func mustRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// newRemote creates a local repository with an initial commit on branch and
// returns its filesystem path, usable as a clone source.
func newRemote(t *testing.T, branch string) string {
	t.Helper()
	dir := t.TempDir()

	mustRun(t, dir, "init", "--initial-branch="+branch)
	mustRun(t, dir, "config", "user.email", "test@example.com")
	mustRun(t, dir, "config", "user.name", "Test")

	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("write README: %v", err)
	}
	mustRun(t, dir, "add", "README.md")
	mustRun(t, dir, "commit", "-m", "initial")

	return dir
}

func TestRunClonesStartsAndReportsSuccess(t *testing.T) {
	remote := newRemote(t, "main")
	repoDir := filepath.Join(t.TempDir(), "checkout")
	statePath := filepath.Join(t.TempDir(), "state.json")
	errorLogPath := filepath.Join(t.TempDir(), "error_log.txt")

	cfg := &config.Config{
		RepoURL:   remote,
		Branch:    "main",
		AuditPort: 9091,
		Projects: []config.Project{
			{Name: "ok-project", Type: config.ProjectTypeGoBuild, Path: ".", BuildCmd: []string{"true"}},
		},
	}

	rep, err := Run(context.Background(), cfg, repoDir, statePath, errorLogPath)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !rep.Healthy() {
		t.Errorf("report = %+v, want healthy", rep)
	}

	if _, err := os.Stat(filepath.Join(repoDir, "README.md")); err != nil {
		t.Errorf("expected repoDir to contain cloned files: %v", err)
	}
	if _, err := os.Stat(statePath); err != nil {
		t.Errorf("expected state file to be written: %v", err)
	}
	if _, err := os.Stat(errorLogPath); !os.IsNotExist(err) {
		t.Errorf("expected no error_log.txt on success, stat err = %v", err)
	}
}

func TestRunReportsProjectFailureWithoutAbortingOtherProjects(t *testing.T) {
	remote := newRemote(t, "main")
	repoDir := filepath.Join(t.TempDir(), "checkout")
	statePath := filepath.Join(t.TempDir(), "state.json")
	errorLogPath := filepath.Join(t.TempDir(), "error_log.txt")

	cfg := &config.Config{
		RepoURL:   remote,
		Branch:    "main",
		AuditPort: 9091,
		Projects: []config.Project{
			{Name: "failing-project", Type: config.ProjectTypeGoBuild, Path: ".", BuildCmd: []string{"false"}},
			{Name: "ok-project", Type: config.ProjectTypeGoBuild, Path: ".", BuildCmd: []string{"true"}},
		},
	}

	rep, err := Run(context.Background(), cfg, repoDir, statePath, errorLogPath)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if rep.Healthy() {
		t.Error("report.Healthy() = true, want false due to failing-project")
	}
	if len(rep.Projects) != 2 {
		t.Fatalf("Projects = %d, want 2", len(rep.Projects))
	}
	if rep.Projects[0].Success {
		t.Error("Projects[0] (failing-project) reported success, want failure")
	}
	if !rep.Projects[1].Success {
		t.Error("Projects[1] (ok-project) reported failure, want success")
	}
}

func TestRunFailsFastWhenRepoUnreachable(t *testing.T) {
	repoDir := filepath.Join(t.TempDir(), "checkout")
	statePath := filepath.Join(t.TempDir(), "state.json")
	errorLogPath := filepath.Join(t.TempDir(), "error_log.txt")

	cfg := &config.Config{
		RepoURL:   filepath.Join(t.TempDir(), "does-not-exist"),
		Branch:    "main",
		AuditPort: 9091,
		Projects:  []config.Project{{Name: "p", Type: config.ProjectTypeGoBuild, Path: ".", BuildCmd: []string{"true"}}},
	}

	if _, err := Run(context.Background(), cfg, repoDir, statePath, errorLogPath); err == nil {
		t.Fatal("Run: want error when repo is unreachable, got nil")
	}

	data, err := os.ReadFile(errorLogPath)
	if err != nil {
		t.Fatalf("read error_log.txt: %v", err)
	}
	if !strings.Contains(string(data), "clone/update repo") {
		t.Errorf("error_log.txt = %q, want it to mention the clone failure", data)
	}
}

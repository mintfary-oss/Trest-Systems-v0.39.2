package orchestrator

import (
	"context"
	"errors"
	"testing"

	"github.com/mintfary-oss/trest-sistems/trest-installer/internal/config"
)

// call records a single invocation made through a fake CommandRunner.
type call struct {
	dir  string
	name string
	args []string
}

func fakeRunner(calls *[]call, err error) CommandRunner {
	return func(_ context.Context, dir, name string, args ...string) (string, error) {
		*calls = append(*calls, call{dir: dir, name: name, args: args})
		if err != nil {
			return "boom", err
		}
		return "ok", nil
	}
}

func TestStartProjectCompose(t *testing.T) {
	var calls []call
	p := config.Project{Name: "magasin-777", Type: config.ProjectTypeCompose, Path: "magasin-777", ComposeFile: "docker-compose.yml"}

	res := StartProject(context.Background(), fakeRunner(&calls, nil), "/repo", p)

	if res.Err != nil {
		t.Fatalf("StartProject: %v", res.Err)
	}
	if res.Action != "compose-up" {
		t.Errorf("Action = %q, want %q", res.Action, "compose-up")
	}
	if len(calls) != 1 {
		t.Fatalf("calls = %d, want 1", len(calls))
	}
	if calls[0].dir != "/repo/magasin-777" {
		t.Errorf("dir = %q, want %q", calls[0].dir, "/repo/magasin-777")
	}
	if calls[0].name != "docker" {
		t.Errorf("name = %q, want %q", calls[0].name, "docker")
	}
}

func TestStartProjectGoBuild(t *testing.T) {
	var calls []call
	p := config.Project{
		Name:     "proektirovka-sdaniy",
		Type:     config.ProjectTypeGoBuild,
		Path:     "proektirovka-sdaniy",
		BuildCmd: []string{"go", "build", "-o", "gencad", "./cmd/generate"},
	}

	res := StartProject(context.Background(), fakeRunner(&calls, nil), "/repo", p)

	if res.Err != nil {
		t.Fatalf("StartProject: %v", res.Err)
	}
	if res.Action != "go-build" {
		t.Errorf("Action = %q, want %q", res.Action, "go-build")
	}
	if calls[0].name != "go" {
		t.Errorf("name = %q, want %q", calls[0].name, "go")
	}
	if len(calls[0].args) != 4 {
		t.Errorf("args = %v, want 4 elements", calls[0].args)
	}
}

func TestStartProjectUnknownType(t *testing.T) {
	var calls []call
	p := config.Project{Name: "mystery", Type: "unknown", Path: "mystery"}

	res := StartProject(context.Background(), fakeRunner(&calls, nil), "/repo", p)

	if res.Err == nil {
		t.Fatal("StartProject: want error for unknown type, got nil")
	}
	if len(calls) != 0 {
		t.Errorf("calls = %d, want 0", len(calls))
	}
}

func TestStartAllContinuesPastFailure(t *testing.T) {
	var calls []call
	failingRunner := fakeRunner(&calls, errors.New("port in use"))

	projects := []config.Project{
		{Name: "a", Type: config.ProjectTypeCompose, Path: "a", ComposeFile: "docker-compose.yml"},
		{Name: "b", Type: config.ProjectTypeCompose, Path: "b", ComposeFile: "docker-compose.yml"},
	}

	results := StartAll(context.Background(), failingRunner, "/repo", projects)

	if len(results) != 2 {
		t.Fatalf("results = %d, want 2", len(results))
	}
	for _, r := range results {
		if r.Err == nil {
			t.Errorf("project %q: want error, got nil", r.Project)
		}
	}
	if len(calls) != 2 {
		t.Errorf("calls = %d, want 2 (should not stop after first failure)", len(calls))
	}
}

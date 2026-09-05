// Package orchestrator starts each configured sub-project: bringing up
// docker-compose projects, or building on-demand Go tools, and reporting
// the outcome of each attempt.
package orchestrator

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mintfary-oss/trest-sistems/internal/config"
)

// CommandRunner executes name with args in dir and returns trimmed combined
// output. It is a seam for tests to substitute a fake process runner.
type CommandRunner func(ctx context.Context, dir, name string, args ...string) (string, error)

// RunCommand is the default CommandRunner, backed by os/exec.
func RunCommand(ctx context.Context, dir, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		return strings.TrimSpace(out.String()), fmt.Errorf("%s %s: %w", name, strings.Join(args, " "), err)
	}
	return strings.TrimSpace(out.String()), nil
}

// Result records the outcome of starting or building a single project.
type Result struct {
	// Project is the config.Project.Name this result belongs to.
	Project string
	// Action names the operation performed, e.g. "compose-up" or
	// "go-build".
	Action string
	// Output is the captured combined stdout/stderr of the operation.
	Output string
	// Err is non-nil if the operation failed.
	Err error
}

// StartProject starts a single project according to its type, using run to
// execute the underlying command in repoRoot/p.Path.
func StartProject(ctx context.Context, run CommandRunner, repoRoot string, p config.Project) Result {
	dir := filepath.Join(repoRoot, p.Path)

	switch p.Type {
	case config.ProjectTypeCompose:
		out, err := run(ctx, dir, "docker", "compose", "-f", p.ComposeFile, "up", "-d", "--build")
		return Result{Project: p.Name, Action: "compose-up", Output: out, Err: err}
	case config.ProjectTypeGoBuild:
		if len(p.BuildCmd) == 0 {
			return Result{Project: p.Name, Action: "go-build", Err: fmt.Errorf("project %q: build_cmd is empty", p.Name)}
		}
		out, err := run(ctx, dir, p.BuildCmd[0], p.BuildCmd[1:]...)
		return Result{Project: p.Name, Action: "go-build", Output: out, Err: err}
	default:
		return Result{Project: p.Name, Err: fmt.Errorf("project %q: unknown type %q", p.Name, p.Type)}
	}
}

// StartAll starts every project in order, continuing past individual
// failures so a single broken project does not block the others. Callers
// should inspect each Result's Err field.
func StartAll(ctx context.Context, run CommandRunner, repoRoot string, projects []config.Project) []Result {
	results := make([]Result, 0, len(projects))
	for _, p := range projects {
		results = append(results, StartProject(ctx, run, repoRoot, p))
	}
	return results
}

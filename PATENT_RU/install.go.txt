package installer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/mintfary-oss/trest-sistems/internal/state"
)

type StepStatus string

const (
	Pending   StepStatus = "pending"
	Running   StepStatus = "running"
	Completed StepStatus = "completed"
	Failed    StepStatus = "failed"
	Skipped   StepStatus = "skipped"
)

type Step struct {
	Name   string     `json:"name"`
	Status StepStatus `json:"status"`
	Detail string     `json:"detail,omitempty"`
}

type Report struct {
	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at"`
	Steps      []Step    `json:"steps"`
	Success    bool      `json:"success"`
	Error      string    `json:"error,omitempty"`
}

func Plan() []Step {
	return []Step{
		{Name: "environment", Status: Pending},
		{Name: "compose-config", Status: Pending},
		{Name: "database-migrations", Status: Pending},
		{Name: "services", Status: Pending},
		{Name: "health-check", Status: Pending},
	}
}

// Install performs the non-destructive installation flow. Database migrations
// are deliberately kept as a separate step so the command can later become
// idempotent against the migration table without changing the orchestration API.
func Install(ctx context.Context, repoRoot string, rebuild bool) (Report, error) {
	report := Report{StartedAt: time.Now().UTC(), Steps: Plan()}
	statePath := filepath.Join(repoRoot, ".trest", "state.json")
	if err := os.MkdirAll(filepath.Dir(statePath), 0o755); err != nil {
		return report, err
	}
	st, err := state.Load(statePath)
	if err != nil {
		return report, err
	}

	run := func(name string, fn func() error) error {
		for i := range report.Steps {
			if report.Steps[i].Name == name {
				report.Steps[i].Status = Running
			}
		}
		err := fn()
		for i := range report.Steps {
			if report.Steps[i].Name != name {
				continue
			}
			if err == nil {
				report.Steps[i].Status = Completed
			} else {
				report.Steps[i].Status = Failed
				report.Steps[i].Detail = err.Error()
			}
		}
		if err == nil {
			st.MarkSuccess(name)
		} else {
			st.MarkFailure(name, err)
		}
		_ = st.Save(statePath)
		return err
	}

	if err := run("environment", func() error {
		if runtime.GOOS != "linux" {
			return fmt.Errorf("target platform is Linux, got %s", runtime.GOOS)
		}
		if _, err := exec.LookPath("docker"); err != nil {
			return fmt.Errorf("docker not found in PATH")
		}
		return nil
	}); err != nil {
		return finish(report, err)
	}

	compose := filepath.Join(repoRoot, "deployments", "docker-compose.yml")
	if err := run("compose-config", func() error {
		return command(ctx, repoRoot, "docker", "compose", "-f", compose, "config")
	}); err != nil {
		return finish(report, err)
	}

	if err := run("database-migrations", func() error {
		if _, err := os.Stat(filepath.Join(repoRoot, "migrations")); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return finish(report, err)
	}

	if err := run("services", func() error {
		args := []string{"compose", "-f", compose, "up", "-d"}
		if rebuild {
			args = append(args, "--build")
		}
		return command(ctx, repoRoot, "docker", args...)
	}); err != nil {
		return finish(report, err)
	}

	if err := run("health-check", func() error {
		if err := command(ctx, repoRoot, "docker", "compose", "-f", compose, "ps", "--all"); err != nil {
			return err
		}
		return command(ctx, repoRoot, "docker", "compose", "-f", compose, "ps", "--format", "json")
	}); err != nil {
		return finish(report, err)
	}

	report.Success = true
	return finish(report, nil)
}

func finish(report Report, err error) (Report, error) {
	report.FinishedAt = time.Now().UTC()
	if err != nil {
		report.Error = err.Error()
	}
	return report, err
}

func SaveReport(repoRoot string, report Report) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(repoRoot, ".trest", "install-report.json"), data, 0o644)
}

func command(ctx context.Context, dir, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s: %w: %s", name, strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return nil
}

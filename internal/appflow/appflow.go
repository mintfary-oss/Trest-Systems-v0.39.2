// Package appflow wires config, git, orchestration, diagnostics, state, and
// report generation into the single orchestration pass used by the "run"
// and "audit" CLI commands.
//
// Every underlying step (git fetch, docker compose up, go build) is
// idempotent, so state.json is kept purely as an audit trail of the most
// recent outcome per step rather than as skip-logic: it is always safe,
// and often necessary (to pick up new commits or restart a crashed
// container), to redo a step that previously succeeded.
package appflow

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/mintfary-oss/trest-sistems/internal/config"
	"github.com/mintfary-oss/trest-sistems/internal/diagnostics"
	"github.com/mintfary-oss/trest-sistems/internal/errorlog"
	"github.com/mintfary-oss/trest-sistems/internal/gitrepo"
	"github.com/mintfary-oss/trest-sistems/internal/orchestrator"
	"github.com/mintfary-oss/trest-sistems/internal/report"
	"github.com/mintfary-oss/trest-sistems/internal/state"
)

// logsSubdir is the conventional per-project directory the orchestrator
// scans for ERROR/FATAL/WARN markers. A project without this directory is
// skipped without error.
const logsSubdir = "logs"

// Run performs one full orchestration pass: ensure repoDir holds an
// up-to-date checkout of cfg.RepoURL, start every configured project,
// check their ports, scan their logs, and persist the outcome to
// statePath. It returns the resulting report; callers are responsible for
// writing it to disk or serving it.
//
// A fatal error (repo unreachable, log directory unreadable) is also
// appended to errorLogPath so operators retain a record of failures even
// after state.json and report.json are overwritten by a later run.
func Run(ctx context.Context, cfg *config.Config, repoDir, statePath, errorLogPath string) (*report.Report, error) {
	fail := func(step string, st *state.State, err error) (*report.Report, error) {
		st.MarkFailure(step, err)
		_ = errorlog.Append(errorLogPath, err) // best-effort; never masks err
		return nil, err
	}

	st, err := state.Load(statePath)
	if err != nil {
		return nil, fmt.Errorf("load state: %w", err)
	}
	defer func() {
		// Best-effort: a failed state save should not mask the
		// orchestration result already computed above.
		_ = st.Save(statePath)
	}()

	if _, err := gitrepo.EnsureRepo(ctx, cfg.RepoURL, cfg.Branch, repoDir); err != nil {
		return fail("clone", st, fmt.Errorf("clone/update repo: %w", err))
	}
	st.MarkSuccess("clone")

	var allPorts []int
	for _, p := range cfg.Projects {
		allPorts = append(allPorts, p.Ports...)
	}
	ports := diagnostics.CheckPorts(allPorts)

	results := orchestrator.StartAll(ctx, orchestrator.RunCommand, repoDir, cfg.Projects)
	for _, res := range results {
		step := "start:" + res.Project
		if res.Err != nil {
			st.MarkFailure(step, res.Err)
		} else {
			st.MarkSuccess(step)
		}
	}

	var alerts []diagnostics.LogAlert
	for _, p := range cfg.Projects {
		logDir := filepath.Join(repoDir, p.Path, logsSubdir)
		a, err := diagnostics.ScanLogDir(logDir)
		if err != nil {
			return fail("scan-logs:"+p.Name, st, fmt.Errorf("scan logs for %q: %w", p.Name, err))
		}
		alerts = append(alerts, a...)
	}
	st.MarkSuccess("scan-logs")

	return report.Build(results, ports, alerts), nil
}

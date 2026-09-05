package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mintfary-oss/trest-sistems/trest-installer/internal/appflow"
	"github.com/mintfary-oss/trest-sistems/trest-installer/internal/config"
	"github.com/mintfary-oss/trest-sistems/trest-installer/internal/report"
)

// newRunCmd builds the "trest run" subcommand: a single orchestration
// pass that clones/updates the repo, starts every project, and writes a
// report.
func newRunCmd(flags *globalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "Clone/update the repo, start every project, and write a report",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(flags.configPath)
			if err != nil {
				return err
			}

			rep, err := appflow.Run(cmd.Context(), cfg, flags.repoDir, flags.statePath(), flags.errorLogPath())
			if err != nil {
				return err
			}

			if err := rep.WriteJSON(flags.reportJSONPath()); err != nil {
				return err
			}
			if err := rep.WriteHTML(flags.reportHTMLPath()); err != nil {
				return err
			}

			printSummary(cmd, rep)

			if !rep.Healthy() {
				return fmt.Errorf("one or more projects reported problems; see %s", flags.reportHTMLPath())
			}
			return nil
		},
	}
}

// printSummary writes a short per-project status line to cmd's stdout.
func printSummary(cmd *cobra.Command, rep *report.Report) {
	out := cmd.OutOrStdout()
	for _, p := range rep.Projects {
		status := "OK"
		if !p.Success {
			status = "FAILED: " + p.Error
		}
		fmt.Fprintf(out, "%-24s %-12s %s\n", p.Name, p.Action, status)
	}
	for _, alert := range rep.LogAlerts {
		fmt.Fprintf(out, "log alert: %s:%d [%s] %s\n", alert.Source, alert.Line, alert.Level, alert.Text)
	}
}

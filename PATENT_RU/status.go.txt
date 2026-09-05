package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mintfary-oss/trest-sistems/trest-installer/internal/report"
)

// newStatusCmd builds the "trest status" subcommand: prints the
// most recently written report without performing any orchestration.
func newStatusCmd(flags *globalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Print the most recently generated report",
		RunE: func(cmd *cobra.Command, args []string) error {
			rep, err := report.Load(flags.reportJSONPath())
			if err != nil {
				return fmt.Errorf("no report found at %s (run \"trest run\" first): %w", flags.reportJSONPath(), err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "report generated at %s\n", rep.GeneratedAt)
			printSummary(cmd, rep)

			if !rep.Healthy() {
				return fmt.Errorf("last run reported problems; see %s", flags.reportHTMLPath())
			}
			return nil
		},
	}
}

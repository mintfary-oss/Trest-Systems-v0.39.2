// Command trest is the unified orchestrator/installer for the Trest
// Systems monorepo. It coordinates deployment, health checks, and status
// reporting for the three sub-projects: magasin-777, proektirovka-sdaniy,
// and super-sistema.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// globalFlags holds the flags shared by every subcommand.
type globalFlags struct {
	configPath string
	repoDir    string
}

// statePath returns the path to this run's state.json, derived from
// repoDir so each checkout tracks its own progress independently.
func (g globalFlags) statePath() string {
	return filepath.Join(g.repoDir, ".trest_state.json")
}

// reportJSONPath returns the path to this run's report.json.
func (g globalFlags) reportJSONPath() string {
	return filepath.Join(g.repoDir, "report.json")
}

// reportHTMLPath returns the path to this run's report.html.
func (g globalFlags) reportHTMLPath() string {
	return filepath.Join(g.repoDir, "report.html")
}

// errorLogPath returns the path to this run's append-only error_log.txt.
func (g globalFlags) errorLogPath() string {
	return filepath.Join(g.repoDir, "error_log.txt")
}

func main() {
	if unifiedInstall() {
		return
	}
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	flags := &globalFlags{}

	root := &cobra.Command{
		Use:   "trest",
		Short: "Orchestrator/installer for the Trest Systems monorepo",
		Long: "trest clones or updates the Trest Systems repository, starts each " +
			"configured sub-project (magasin-777, proektirovka-sdaniy, super-sistema), " +
			"checks ports and logs, and reports the result.",
		SilenceUsage: true,
	}

	root.PersistentFlags().StringVar(&flags.configPath, "config", "config.yaml", "path to the orchestrator config file")
	root.PersistentFlags().StringVar(&flags.repoDir, "repo-dir", "./repo", "directory to clone/update the tracked repository into")

	root.AddCommand(newRunCmd(flags))
	root.AddCommand(newStatusCmd(flags))
	root.AddCommand(newAuditCmd(flags))

	return root
}

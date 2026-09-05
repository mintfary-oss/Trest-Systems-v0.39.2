package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/mintfary-oss/trest-sistems/trest-installer/internal/appflow"
	"github.com/mintfary-oss/trest-sistems/trest-installer/internal/config"
	"github.com/mintfary-oss/trest-sistems/trest-installer/internal/report"
	"github.com/mintfary-oss/trest-sistems/trest-installer/internal/web"
)

// webhookSecretEnvVar names the environment variable holding the shared
// secret for POST /webhook. It is read from the environment rather than a
// CLI flag or config.yaml so it never ends up committed to the repository
// or visible in a process listing.
const webhookSecretEnvVar = "TREST_WEBHOOK_SECRET"

// newAuditCmd builds the "trest audit" subcommand: serves the latest
// report over HTTP, optionally re-running the orchestration on a fixed
// interval and/or in response to a signed webhook, to pick up new commits
// (see AGENTS.md section 6, "Мониторинг репозитория").
func newAuditCmd(flags *globalFlags) *cobra.Command {
	var interval time.Duration

	cmd := &cobra.Command{
		Use:   "audit",
		Short: "Serve the latest report over HTTP, optionally polling for updates",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(flags.configPath)
			if err != nil {
				return err
			}

			server := web.NewServer()

			// Serialize orchestration passes: the interval ticker and the
			// webhook can otherwise fire concurrently and race on
			// state.json/report.json/report.html.
			var runMu sync.Mutex
			runOnce := func() {
				runMu.Lock()
				defer runMu.Unlock()

				rep, err := appflow.Run(cmd.Context(), cfg, flags.repoDir, flags.statePath(), flags.errorLogPath())
				if err != nil {
					fmt.Fprintln(cmd.ErrOrStderr(), "audit: orchestration run failed:", err)
					return
				}
				if err := rep.WriteJSON(flags.reportJSONPath()); err != nil {
					fmt.Fprintln(cmd.ErrOrStderr(), "audit: write report.json failed:", err)
				}
				if err := rep.WriteHTML(flags.reportHTMLPath()); err != nil {
					fmt.Fprintln(cmd.ErrOrStderr(), "audit: write report.html failed:", err)
				}
				server.SetReport(rep)
			}

			// Seed the dashboard with whatever report already exists so
			// it responds immediately, then run a fresh pass.
			if existing, err := report.Load(flags.reportJSONPath()); err == nil {
				server.SetReport(existing)
			}
			runOnce()

			if interval > 0 {
				go func() {
					ticker := time.NewTicker(interval)
					defer ticker.Stop()
					for range ticker.C {
						runOnce()
					}
				}()
			}

			if secret := os.Getenv(webhookSecretEnvVar); secret != "" {
				server.SetWebhookHandler(secret, runOnce)
				fmt.Fprintln(cmd.OutOrStdout(), "audit: webhook enabled at /webhook")
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "audit: webhook disabled (set %s to enable)\n", webhookSecretEnvVar)
			}

			addr := fmt.Sprintf(":%d", cfg.AuditPort)
			fmt.Fprintf(cmd.OutOrStdout(), "audit: serving report on http://localhost:%d/report\n", cfg.AuditPort)
			return server.ListenAndServe(addr)
		},
	}

	cmd.Flags().DurationVar(&interval, "interval", 5*time.Minute, "how often to re-run the orchestration; 0 disables polling")

	return cmd
}

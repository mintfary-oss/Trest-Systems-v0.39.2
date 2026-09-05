# Current state: 0.39.2

Release candidate prepared from the complete 0.39.1 archive and user-supplied server logs. New canonical installer: scripts/installer/auto_install.py. New verification is verification/v0.39.2/VERIFICATION.json; all other historical success reports are not current evidence.

Corrections: single checksum migration implementation; simple-query libpq path; migration verification for externally migrated API; .env preservation and no silent credential rotation; isolated marketplace schema; admin password preservation; users.is_active migration; frontend namespace; secure offline WebUI/Ollama; port ownership and coexistence; actual HTTP healthchecks; no production cleanup in smoke/drill; rebuilt amd64 Go binaries; offline exporter.

Pending before production certification: Docker/clean-host runtime for THIS release; legacy marketplace data upgrade review; browser/full functional E2E; complete file/object-storage restore/reboot drill; upstream dependency/toolchain security audit; actual offline image/model bundle. No GitHub push or live server update performed by this packaging task.

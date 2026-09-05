# Offline compatibility modules

This release includes compact, project-specific compatibility implementations for the exact API surface used by Trest Systems:

- `github.com/spf13/cobra` — command/flag subset built on the Go standard library;
- `gopkg.in/yaml.v3` — configuration YAML subset used by `config.yaml` and the DXF building configuration;
- `github.com/jackc/pgx/v5` — the used `stdlib`/`pgxpool` subset backed by the system PostgreSQL `libpq` library.

They were added so release binaries can be built reproducibly without downloading Go modules during server installation. They are not represented as full replacements for the upstream projects. The application imports only the supported subset, and all root/nested Go tests, vet and build checks are run against these modules.

The API, Worker and `trestctl` Linux binaries require glibc and `libpq.so.5`. The automatic installer installs the corresponding OS package, while the API/Worker container images use Debian with `libpq5`.

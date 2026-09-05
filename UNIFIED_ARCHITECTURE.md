# Trest Systems — Unified Go Core

## Current merge

The repository now has a root Go module at `github.com/mintfary-oss/trest-sistems`.
The original Go implementation from `proektirovka-sdaniy` is ported under
`internal/proektirovka/` and is exposed through the unified `trest generate`
command. The original `trest-installer` orchestration packages are ported under
`internal/` and are exposed through the root `trest` CLI.

The original source trees remain intact during migration so functionality and
history are not lost while each component is ported and verified.

## Unified CLI

- `trest run` — orchestration pass for the configured services
- `trest status` — last report
- `trest audit` — HTTP audit dashboard and optional polling/webhook
- `trest generate` — GOST DXF generation from the original architectural module

## Three original systems

1. `magasin-777`: Next.js + FastAPI marketplace. It remains the source of the
   existing marketplace UI/API while its domain logic is progressively ported
   behind the unified Go core.
2. `proektirovka-sdaniy`: Go/GOST DXF generator. Its packages are already
   physically integrated into the root Go module.
3. `super-sistema`: Ollama/Open WebUI local AI stack. Its existing Docker/GPU
   deployment remains available while the unified Go core controls lifecycle,
   diagnostics and reporting.

## Port allocation

The LocalMarket frontend uses host port `3000`. Super Sistema Open WebUI uses
`3001` by default in the unified stack, while Ollama uses `11434`.

## Security

Secrets are never stored in source files. GitHub tokens, webhook secrets and
application credentials must be supplied through environment variables or an
untracked `.env` file.

## Verification

The test host currently has Go 1.23.2, Node.js 22.16.0 and no Docker CLI.
Therefore Go packages that require external modules and Docker integration
cannot be fully executed in this environment until dependencies and Docker
are available. Pure Go packages, Python syntax compilation, JavaScript syntax
checks and Compose YAML parsing have been verified locally.

# Trest Systems — provenance manifest

This manifest is a release checklist, not a declaration that every file is original.

| Area | Path | Treatment | Rights check |
|---|---|---|---|
| Unified Go | `internal/`, `cmd/` | Project implementation | Verify authorship/assignments |
| DB | `migrations/` | Project migrations/schema | Verify authorship |
| AI | `internal/ai/`, `training/` | Project implementation + ML tooling | Verify code + dataset rights |
| Documentation | root/docs | Project documentation | Verify authorship |
| Marketplace | `magasin-777/` | Preserved subsystem | **Separate provenance/license audit required** |
| DXF/GOST | `proektirovka-sdaniy/` | Preserved subsystem | **Separate provenance/license audit required** |
| Local AI infra | `super-sistema/` | Preserved subsystem | Existing `super-sistema/LICENSE` is MIT; the rights-holder declaration states that the project/accounts are owned by the rights holder; preserve the license and verify chain-of-title evidence before filing |
| Installer | `trest-installer/` | Preserved/merged subsystem | **Separate provenance/license audit required** |
| Go dependencies | `go.mod`, `go.sum` | Third-party | SPDX/SBOM scan required |
| Frontend dependencies | `magasin-777/frontend/package.json` | Third-party | lockfile/SBOM scan required |
| Python dependencies | `super-sistema/gpu-panel/requirements.txt` | Third-party | full dependency scan required |

## Release rule

A file can be marked “owned proprietary” only after one of these is documented:
- authored by the rights holder;
- authored by an employee under a valid IP agreement;
- authored by a contractor with a valid assignment/license;
- derived from a permissive third-party component with all conditions satisfied;
- otherwise legally cleared.

The existence of an AI-generated draft does not by itself establish exclusive ownership in every jurisdiction. Human review, contractual chain-of-title and applicable law must be documented.

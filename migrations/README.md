# Database migrations

Migration `0001_foundation.sql` creates the Stage 1 foundation schema: users, projects, objects and orders.

For production, run migrations through the deployment pipeline before starting the API. The SQL is intentionally idempotent for the foundation stage.


Migration `0010_ai.sql` adds the Stage 9 AI/agent/training registry. It is idempotent and records models, versions, prompts, requests, audit events, agents, approvals, datasets, training jobs and evaluations.

- `0011_bim.sql` — Stage 10: 3D/BIM model registry, versions, elements, import/export jobs, progress snapshots.

- `0012_bim_runtime.sql` — runtime exchange attempts/output metadata and BIM geometry diff results.

- `0013_bim_diff.sql` — Stage 10 Runtime 3 geometry diff indexes.

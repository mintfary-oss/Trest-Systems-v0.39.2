-- Stage 9: AI, agents and controlled training registry.
CREATE TABLE IF NOT EXISTS ai_providers (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(), name TEXT NOT NULL UNIQUE, kind TEXT NOT NULL CHECK (kind IN ('ollama','openai_compatible','local')), base_url TEXT NOT NULL, enabled BOOLEAN NOT NULL DEFAULT TRUE, created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS ai_models (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(), provider_id UUID NOT NULL REFERENCES ai_providers(id) ON DELETE RESTRICT, name TEXT NOT NULL, family TEXT NOT NULL DEFAULT '', capabilities JSONB NOT NULL DEFAULT '{}'::jsonb, active BOOLEAN NOT NULL DEFAULT TRUE, is_default BOOLEAN NOT NULL DEFAULT FALSE, created_at TIMESTAMPTZ NOT NULL DEFAULT now(), UNIQUE(provider_id,name)
);
CREATE TABLE IF NOT EXISTS ai_model_versions (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(), model_id UUID NOT NULL REFERENCES ai_models(id) ON DELETE CASCADE, version TEXT NOT NULL, artifact_uri TEXT NOT NULL DEFAULT '', adapter_uri TEXT NOT NULL DEFAULT '', metrics JSONB NOT NULL DEFAULT '{}'::jsonb, status TEXT NOT NULL DEFAULT 'draft' CHECK(status IN ('draft','training','evaluating','published','rolled_back','failed')), created_at TIMESTAMPTZ NOT NULL DEFAULT now(), published_at TIMESTAMPTZ, UNIQUE(model_id,version)
);
CREATE TABLE IF NOT EXISTS ai_prompts (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(), name TEXT NOT NULL UNIQUE, version INTEGER NOT NULL DEFAULT 1, template TEXT NOT NULL, variables JSONB NOT NULL DEFAULT '[]'::jsonb, active BOOLEAN NOT NULL DEFAULT TRUE, created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS ai_requests (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(), requester_user_id UUID REFERENCES users(id), model_version_id UUID REFERENCES ai_model_versions(id), prompt_id UUID REFERENCES ai_prompts(id), input JSONB NOT NULL, output JSONB NOT NULL DEFAULT '{}'::jsonb, confidence NUMERIC(5,4), status TEXT NOT NULL DEFAULT 'queued' CHECK(status IN ('queued','running','completed','failed')), error TEXT NOT NULL DEFAULT '', created_at TIMESTAMPTZ NOT NULL DEFAULT now(), completed_at TIMESTAMPTZ
);
CREATE TABLE IF NOT EXISTS ai_audit_events (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(), ai_request_id UUID REFERENCES ai_requests(id) ON DELETE CASCADE, actor_user_id UUID REFERENCES users(id), event_type TEXT NOT NULL, metadata JSONB NOT NULL DEFAULT '{}'::jsonb, created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS agents (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(), owner_user_id UUID NOT NULL REFERENCES users(id), name TEXT NOT NULL, description TEXT NOT NULL DEFAULT '', instructions TEXT NOT NULL, model_version_id UUID REFERENCES ai_model_versions(id), tools JSONB NOT NULL DEFAULT '[]'::jsonb, permissions JSONB NOT NULL DEFAULT '[]'::jsonb, memory_policy JSONB NOT NULL DEFAULT '{}'::jsonb, sandbox_policy JSONB NOT NULL DEFAULT '{}'::jsonb, enabled BOOLEAN NOT NULL DEFAULT TRUE, created_at TIMESTAMPTZ NOT NULL DEFAULT now(), updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS agent_executions (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(), agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE, requester_user_id UUID NOT NULL REFERENCES users(id), input JSONB NOT NULL, output JSONB NOT NULL DEFAULT '{}'::jsonb, status TEXT NOT NULL DEFAULT 'queued' CHECK(status IN ('queued','running','waiting_approval','completed','failed','cancelled')), tool_calls JSONB NOT NULL DEFAULT '[]'::jsonb, error TEXT NOT NULL DEFAULT '', started_at TIMESTAMPTZ, completed_at TIMESTAMPTZ, created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS agent_approvals (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(), execution_id UUID NOT NULL REFERENCES agent_executions(id) ON DELETE CASCADE, requested_action TEXT NOT NULL, payload JSONB NOT NULL DEFAULT '{}'::jsonb, status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending','approved','rejected','expired')), decided_by UUID REFERENCES users(id), decided_at TIMESTAMPTZ, created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS ai_datasets (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(), owner_user_id UUID REFERENCES users(id), name TEXT NOT NULL, description TEXT NOT NULL DEFAULT '', format TEXT NOT NULL DEFAULT 'jsonl', source_uri TEXT NOT NULL DEFAULT '', version TEXT NOT NULL, row_count INTEGER NOT NULL DEFAULT 0, checksum TEXT NOT NULL DEFAULT '', status TEXT NOT NULL DEFAULT 'draft' CHECK(status IN ('draft','ready','archived')), created_at TIMESTAMPTZ NOT NULL DEFAULT now(), UNIQUE(name,version)
);
CREATE TABLE IF NOT EXISTS ai_training_jobs (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(), dataset_id UUID NOT NULL REFERENCES ai_datasets(id), base_model_version_id UUID NOT NULL REFERENCES ai_model_versions(id), output_model_version_id UUID REFERENCES ai_model_versions(id), method TEXT NOT NULL DEFAULT 'qlora' CHECK(method IN ('lora','qlora','full')), config JSONB NOT NULL DEFAULT '{}'::jsonb, status TEXT NOT NULL DEFAULT 'queued' CHECK(status IN ('queued','preparing','training','evaluating','completed','failed','cancelled')), metrics JSONB NOT NULL DEFAULT '{}'::jsonb, artifact_uri TEXT NOT NULL DEFAULT '', error TEXT NOT NULL DEFAULT '', created_by UUID REFERENCES users(id), started_at TIMESTAMPTZ, completed_at TIMESTAMPTZ, created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS ai_evaluations (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(), training_job_id UUID REFERENCES ai_training_jobs(id) ON DELETE CASCADE, model_version_id UUID REFERENCES ai_model_versions(id), dataset_id UUID REFERENCES ai_datasets(id), metrics JSONB NOT NULL DEFAULT '{}'::jsonb, passed BOOLEAN NOT NULL DEFAULT FALSE, created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_ai_requests_created ON ai_requests(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_agent_executions_created ON agent_executions(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_training_jobs_created ON ai_training_jobs(created_at DESC);
CREATE UNIQUE INDEX IF NOT EXISTS uq_ai_default_model ON ai_models(is_default) WHERE is_default = TRUE;

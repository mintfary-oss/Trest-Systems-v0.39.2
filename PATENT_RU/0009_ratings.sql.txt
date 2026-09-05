-- Stage 8: trusted ratings, reputation history, bonuses, sanctions and disputes.
CREATE TABLE IF NOT EXISTS ratings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE RESTRICT,
    reviewer_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    target_type TEXT NOT NULL CHECK (target_type IN ('contractor','supplier')),
    target_id UUID NOT NULL,
    score NUMERIC(3,2) NOT NULL CHECK (score >= 1 AND score <= 5),
    dimensions JSONB NOT NULL DEFAULT '{}'::jsonb,
    comment TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'published' CHECK (status IN ('published','hidden','disputed')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(order_id, reviewer_user_id, target_type, target_id)
);
CREATE INDEX IF NOT EXISTS idx_ratings_target ON ratings(target_type, target_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_ratings_order ON ratings(order_id);

CREATE TABLE IF NOT EXISTS rating_aggregates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    target_type TEXT NOT NULL CHECK (target_type IN ('contractor','supplier')),
    target_id UUID NOT NULL,
    rating_count INTEGER NOT NULL DEFAULT 0 CHECK (rating_count >= 0),
    average_score NUMERIC(4,3) NOT NULL DEFAULT 0 CHECK (average_score >= 0 AND average_score <= 5),
    dimensions JSONB NOT NULL DEFAULT '{}'::jsonb,
    calculated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    version INTEGER NOT NULL DEFAULT 1 CHECK (version > 0),
    UNIQUE(target_type, target_id)
);

CREATE TABLE IF NOT EXISTS rating_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    target_type TEXT NOT NULL CHECK (target_type IN ('contractor','supplier')),
    target_id UUID NOT NULL,
    rating_count INTEGER NOT NULL,
    average_score NUMERIC(4,3) NOT NULL,
    dimensions JSONB NOT NULL DEFAULT '{}'::jsonb,
    source_rating_id UUID REFERENCES ratings(id) ON DELETE SET NULL,
    version INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_rating_history_target ON rating_history(target_type, target_id, created_at DESC);

CREATE TABLE IF NOT EXISTS reputation_bonuses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    target_type TEXT NOT NULL CHECK (target_type IN ('contractor','supplier')),
    target_id UUID NOT NULL,
    points NUMERIC(8,2) NOT NULL CHECK (points > 0),
    reason TEXT NOT NULL,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_reputation_bonuses_target ON reputation_bonuses(target_type, target_id);

CREATE TABLE IF NOT EXISTS reputation_sanctions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    target_type TEXT NOT NULL CHECK (target_type IN ('contractor','supplier')),
    target_id UUID NOT NULL,
    points NUMERIC(8,2) NOT NULL CHECK (points > 0),
    reason TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','expired','revoked')),
    created_by UUID NOT NULL REFERENCES users(id),
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_reputation_sanctions_target ON reputation_sanctions(target_type, target_id, status);

CREATE TABLE IF NOT EXISTS rating_disputes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rating_id UUID NOT NULL REFERENCES ratings(id) ON DELETE CASCADE,
    opened_by UUID NOT NULL REFERENCES users(id),
    reason TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open','under_review','resolved_upheld','resolved_removed','rejected')),
    resolution_note TEXT NOT NULL DEFAULT '',
    resolved_by UUID REFERENCES users(id),
    resolved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_rating_disputes_rating ON rating_disputes(rating_id);
CREATE INDEX IF NOT EXISTS idx_rating_disputes_status ON rating_disputes(status);

CREATE TABLE IF NOT EXISTS rating_dispute_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dispute_id UUID NOT NULL REFERENCES rating_disputes(id) ON DELETE CASCADE,
    actor_user_id UUID NOT NULL REFERENCES users(id),
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_rating_dispute_events_dispute ON rating_dispute_events(dispute_id, created_at);
CREATE UNIQUE INDEX IF NOT EXISTS uq_rating_open_dispute ON rating_disputes(rating_id) WHERE status IN ('open','under_review');

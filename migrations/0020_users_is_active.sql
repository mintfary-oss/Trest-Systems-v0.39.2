-- Compatibility field; do not change applied 0001-0019 migrations.
ALTER TABLE public.users ADD COLUMN IF NOT EXISTS is_active boolean NOT NULL DEFAULT true;

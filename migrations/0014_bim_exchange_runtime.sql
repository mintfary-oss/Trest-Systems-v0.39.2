-- Stage 10 Runtime 4: auditable cancellation and queue lifecycle.
DO $$
DECLARE c text;
BEGIN
  SELECT conname INTO c FROM pg_constraint WHERE conrelid='bim_import_exports'::regclass AND contype='c' AND pg_get_constraintdef(oid) LIKE '%status%';
  IF c IS NOT NULL THEN EXECUTE format('ALTER TABLE bim_import_exports DROP CONSTRAINT %I', c); END IF;
END $$;
ALTER TABLE bim_import_exports ADD CONSTRAINT bim_import_exports_status_check CHECK (status IN ('queued','running','completed','failed','cancelled'));
CREATE INDEX IF NOT EXISTS idx_bim_io_queue ON bim_import_exports(status,created_at);

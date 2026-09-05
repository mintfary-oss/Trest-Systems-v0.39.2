package db

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type migration struct{ Name, Checksum, Body string }

var migrationName = regexp.MustCompile(`^[0-9]{4,}_[A-Za-z0-9_.-]+\.sql$`)

func loadMigrations(dir string) ([]migration, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var result []migration
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		if !migrationName.MatchString(e.Name()) {
			return nil, fmt.Errorf("invalid migration filename: %s", e.Name())
		}
		if e.Type()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("migration symlink refused: %s", e.Name())
		}
		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		result = append(result, migration{e.Name(), fmt.Sprintf("%x", sha256.Sum256(b)), string(b)})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	if len(result) == 0 {
		return nil, fmt.Errorf("no SQL migrations in %s", dir)
	}
	return result, nil
}

// Migrate is the SINGLE migration implementation used by API, CLI and installer.
// Never fabricate baselines. Never swallow query errors. Every migration and
// tracking insert commit together, serialized by a PostgreSQL advisory lock.
func Migrate(ctx context.Context, d *DB, dir string) error {
	files, err := loadMigrations(dir)
	if err != nil {
		return err
	}
	setup, err := d.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer setup.Rollback()
	if _, err = setup.ExecContext(ctx, `SELECT pg_advisory_xact_lock(7392001)`); err != nil {
		return err
	}
	if _, err = setup.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS public.schema_migrations(version text PRIMARY KEY, checksum text NOT NULL, applied_at timestamptz NOT NULL DEFAULT now())`); err != nil {
		return err
	}
	var typ string
	err = setup.QueryRowContext(ctx, `SELECT data_type FROM information_schema.columns WHERE table_schema='public' AND table_name='schema_migrations' AND column_name='version'`).Scan(&typ)
	if err != nil {
		return err
	}
	if typ != "text" && typ != "character varying" {
		return fmt.Errorf("legacy numeric schema_migrations detected; reviewed migration baseline required; no schema/data changed")
	}
	if err = setup.Commit(); err != nil {
		return err
	}
	for _, m := range files {
		if err = applyMigration(ctx, d, m); err != nil {
			return fmt.Errorf("migration %s: %w", m.Name, err)
		}
	}
	return nil
}
func applyMigration(ctx context.Context, d *DB, m migration) error {
	tx, err := d.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(7392001)`); err != nil {
		return err
	}
	var previous string
	err = tx.QueryRowContext(ctx, `SELECT checksum FROM public.schema_migrations WHERE version=$1`, m.Name).Scan(&previous)
	if err == nil {
		if previous != m.Checksum {
			return fmt.Errorf("checksum mismatch; restore original file, do not edit applied migrations")
		}
		fmt.Println("SKIP", m.Name)
		return tx.Commit()
	}
	if err != sql.ErrNoRows {
		return err
	}
	fmt.Println("APPLY", m.Name)
	if _, err = tx.ExecContext(ctx, m.Body); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO public.schema_migrations(version,checksum) VALUES($1,$2)`, m.Name, m.Checksum); err != nil {
		return err
	}
	return tx.Commit()
}

// CheckMigrations permits externally managed migrations without blindly skipping
// schema validation. The API refuses readiness if even one checksum is absent.
func CheckMigrations(ctx context.Context, d *DB, dir string) error {
	files, err := loadMigrations(dir)
	if err != nil {
		return err
	}
	for _, m := range files {
		var sum string
		if err = d.QueryRowContext(ctx, `SELECT checksum FROM public.schema_migrations WHERE version=$1`, m.Name).Scan(&sum); err != nil {
			return fmt.Errorf("unapplied migration %s: %w", m.Name, err)
		}
		if sum != m.Checksum {
			return fmt.Errorf("checksum mismatch for %s", m.Name)
		}
	}
	return nil
}

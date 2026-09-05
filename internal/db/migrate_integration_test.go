package db

import (
	"context"
	"database/sql"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMigrationsAgainstDisposablePostgres(t *testing.T) {
	dsn := os.Getenv("TREST_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("real PostgreSQL not supplied; set TREST_TEST_DATABASE_URL for a disposable empty database")
	}
	u, err := url.Parse(dsn)
	if err != nil || !strings.HasPrefix(strings.TrimPrefix(u.Path, "/"), "trest_test_") {
		t.Fatal("database name must start with trest_test_")
	}
	d, e := sql.Open("pgx", dsn)
	if e != nil {
		t.Fatal(e)
	}
	defer d.Close()
	db := &DB{d}
	ctx := context.Background()
	var count int
	if e = d.QueryRow(`SELECT count(*) FROM pg_tables WHERE schemaname='public'`).Scan(&count); e != nil {
		t.Fatal(e)
	}
	if count != 0 {
		t.Fatal("test requires an EMPTY disposable database")
	}
	dir, err := filepath.Abs("../../migrations")
	if err != nil {
		t.Fatal(err)
	}
	if err = Migrate(ctx, db, dir); err != nil {
		t.Fatal(err)
	}
	if err = Migrate(ctx, db, dir); err != nil {
		t.Fatal("repeat failed:", err)
	}
	if err = CheckMigrations(ctx, db, dir); err != nil {
		t.Fatal(err)
	}
}

package db

import (
	"context"
	"net/url"
	"testing"
)

func TestDSNRequiresSecret(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("POSTGRES_PASSWORD", "")
	if DSN() != "" {
		t.Fatal("unsafe fallback DSN")
	}
	if _, err := Open(context.Background()); err == nil {
		t.Fatal("missing password accepted")
	}
}
func TestDSNPreservesExplicitURL(t *testing.T) {
	d := "postgres://example:secret@host:5432/db?sslmode=require"
	t.Setenv("DATABASE_URL", d)
	if DSN() != d {
		t.Fatal("explicit URL changed")
	}
}
func TestDSNEscapesUserinfoAndIPv6(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("POSTGRES_USER", "u ser@a")
	t.Setenv("POSTGRES_PASSWORD", "p a+ss:@/\\'?#")
	t.Setenv("POSTGRES_HOST", "::1")
	t.Setenv("POSTGRES_PORT", "5432")
	t.Setenv("POSTGRES_DB", "db name")
	t.Setenv("POSTGRES_SSLMODE", "require")
	u, err := url.Parse(DSN())
	if err != nil {
		t.Fatal(err)
	}
	p, _ := u.User.Password()
	if u.User.Username() != "u ser@a" || p != "p a+ss:@/\\'?#" || u.Hostname() != "::1" || u.Path != "/db name" || u.Query().Get("sslmode") != "require" {
		t.Fatal("DSN round trip failed")
	}
}

package db

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DB struct{ *sql.DB }

func DSN() string {
	if v := os.Getenv("DATABASE_URL"); v != "" {
		return v
	}
	host := getenv("POSTGRES_HOST", "localhost")
	port := getenv("POSTGRES_PORT", "5432")
	user := getenv("POSTGRES_USER", "trest")
	password := os.Getenv("POSTGRES_PASSWORD")
	if password == "" {
		return ""
	}
	name := getenv("POSTGRES_DB", "trest")
	u := url.URL{Scheme: "postgres", User: url.UserPassword(user, password), Host: net.JoinHostPort(host, port), Path: "/" + name}
	q := url.Values{}
	q.Set("sslmode", getenv("POSTGRES_SSLMODE", "disable"))
	u.RawQuery = q.Encode()
	return u.String()
}

func Open(ctx context.Context) (*DB, error) {
	dsn := DSN()
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL or POSTGRES_PASSWORD is required")
	}
	d, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	d.SetMaxOpenConns(20)
	d.SetMaxIdleConns(5)
	d.SetConnMaxLifetime(30 * time.Minute)
	if err := d.PingContext(ctx); err != nil {
		_ = d.Close()
		return nil, err
	}
	return &DB{d}, nil
}

func getenv(k, fallback string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fallback
}

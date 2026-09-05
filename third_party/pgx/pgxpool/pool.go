// Package pgxpool provides the subset of pgxpool used by Trest Systems,
// backed by database/sql and the local libpq driver.
package pgxpool

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Pool struct{ db *sql.DB }

func New(ctx context.Context, connString string) (*Pool, error) {
	db, err := sql.Open("pgx", connString)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &Pool{db: db}, nil
}

func (p *Pool) Close() { _ = p.db.Close() }

func (p *Pool) Begin(ctx context.Context) (*Tx, error) {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &Tx{tx: tx}, nil
}

func (p *Pool) Exec(ctx context.Context, query string, args ...any) (pgx.CommandTag, error) {
	result, err := p.db.ExecContext(ctx, query, args...)
	if err != nil {
		return pgx.CommandTag{}, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return pgx.CommandTag{}, err
	}
	return pgx.CommandTag{Affected: affected}, nil
}

func (p *Pool) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	return p.db.QueryRowContext(ctx, query, args...)
}

type Tx struct {
	tx   *sql.Tx
	done bool
}

func (t *Tx) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	return t.tx.QueryRowContext(ctx, query, args...)
}

func (t *Tx) Exec(ctx context.Context, query string, args ...any) (pgx.CommandTag, error) {
	result, err := t.tx.ExecContext(ctx, query, args...)
	if err != nil {
		return pgx.CommandTag{}, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return pgx.CommandTag{}, err
	}
	return pgx.CommandTag{Affected: affected}, nil
}

func (t *Tx) Commit(_ context.Context) error {
	if t.done {
		return fmt.Errorf("transaction already closed")
	}
	t.done = true
	return t.tx.Commit()
}

func (t *Tx) Rollback(_ context.Context) error {
	if t.done {
		return nil
	}
	t.done = true
	err := t.tx.Rollback()
	if err == sql.ErrTxDone {
		return nil
	}
	return err
}
